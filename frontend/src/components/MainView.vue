<script setup>
// MainView はカテゴリ一覧と課題一覧を並べて表示する。
// 詳細編集は別コンポーネントに委ねる。
import { computed, onMounted, ref, watch } from 'vue'

import { useAppStore } from '../stores/app'
import { useCategoriesStore } from '../stores/categories'
import { useIssuesStore } from '../stores/issues'
import { formatDate } from '../utils/time'

const emit = defineEmits(['open-issue'])

const appStore = useAppStore()
const categoriesStore = useCategoriesStore()
const issuesStore = useIssuesStore()

const showIssueCreateDialog = ref(false)
const newIssueTitle = ref('')
const newIssueDescription = ref('')
const newIssueDueDate = ref('')
// 優先度は必須入力のため初期値を設定する。
const newIssuePriority = ref('Medium')
const newIssueAssignee = ref('')
const issueCreateErrorMessage = ref('')
const showCreateIssueDatePicker = ref(false)
const createIssuePickerDate = ref(null)

const filterText = ref('')
const filterStatus = ref([])
const filterPriority = ref([])
const filterDueFrom = ref('')
const filterDueTo = ref('')
const filterSchemaInvalid = ref(false)
const showFilterDueFromPicker = ref(false)
const showFilterDueToPicker = ref(false)
const filterDueFromPickerDate = ref(null)
const filterDueToPickerDate = ref(null)

const statusOptions = [
  'Open',
  'Working',
  'Inquiry',
  'Hold',
  'Feedback',
  'Resolved',
  'Closed',
  'Rejected',
]
const priorityOptions = ['High', 'Medium', 'Low']

const selectedCategory = computed(() => categoriesStore.selectedCategory)
const cacheEntry = computed(() => {
  const key = selectedCategory.value
  if (!key) {
    return null
  }
  return issuesStore.issuesByCategory[key] ?? null
})

const currentQuery = computed(() => {
  const key = selectedCategory.value
  if (!key) {
    return issuesStore.defaultQuery
  }
  return issuesStore.getQuery(key)
})

const pageCount = computed(() => {
  if (!cacheEntry.value) {
    return 1
  }
  const total = cacheEntry.value.total ?? 0
  const pageSize = appStore.pageSize || 20
  return Math.max(1, Math.ceil(total / pageSize))
})

const filteredItems = computed(() => {
  const items = cacheEntry.value?.items ?? []
  return items.filter((item) => {
    if (filterSchemaInvalid.value && !item.is_schema_invalid) {
      return false
    }
    if (filterText.value) {
      const text = filterText.value.toLowerCase()
      const target = `${item.title ?? ''} ${item.issue_id ?? ''}`.toLowerCase()
      if (!target.includes(text)) {
        return false
      }
    }
    if (filterStatus.value.length > 0 && !filterStatus.value.includes(item.status)) {
      return false
    }
    if (filterPriority.value.length > 0 && !filterPriority.value.includes(item.priority)) {
      return false
    }
    if (filterDueFrom.value && item.due_date < filterDueFrom.value) {
      return false
    }
    if (filterDueTo.value && item.due_date > filterDueTo.value) {
      return false
    }
    return true
  })
})

watch(selectedCategory, async (value) => {
  if (value && !issuesStore.issuesByCategory[value]) {
    await issuesStore.loadIssues(value)
  }
})



function isEndState(status) {
  return status === 'Closed' || status === 'Rejected'
}

function sortLabel(key) {
  const sort = currentQuery.value.sort
  if (sort.key !== key) {
    return ''
  }
  return sort.dir === 'asc' ? '▲' : '▼'
}

async function toggleSort(key) {
  const sort = currentQuery.value.sort
  const nextDir = sort.key === key && sort.dir === 'asc' ? 'desc' : 'asc'
  if (selectedCategory.value) {
    await issuesStore.setSort(selectedCategory.value, { key, dir: nextDir })
  }
}

function applyFilter() {
  if (!selectedCategory.value) {
    return
  }
  issuesStore.setFilter(selectedCategory.value, {
    text: filterText.value,
    status: filterStatus.value,
    priority: filterPriority.value,
    dueDateFrom: filterDueFrom.value || null,
    dueDateTo: filterDueTo.value || null,
    schemaInvalidOnly: filterSchemaInvalid.value,
  })
}

async function handlePageChange(page) {
  if (selectedCategory.value) {
    await issuesStore.setPage(selectedCategory.value, page)
  }
}



// handleOpenIssueCreateDialog は新規課題ダイアログを開く。
// 目的: 入力状態を初期化して作成フォームを表示する。
// 入力: なし。
// 出力: なし。
// エラー: なし。
// 副作用: 入力状態とダイアログ表示を更新する。
// 並行性: 単一UIイベント前提。
// 不変条件: ダイアログ表示時はエラーメッセージが空になる。
// 関連DD: DD-UI-006
function handleOpenIssueCreateDialog() {
  resetIssueCreateForm()
  showIssueCreateDialog.value = true
}

// resetIssueCreateForm は新規課題入力を初期化する。
// 目的: ダイアログ再利用時に前回入力を残さない。
// 入力: なし。
// 出力: なし。
// エラー: なし。
// 副作用: 入力用のrefを初期化する。
// 並行性: 単一UIイベント前提。
// 不変条件: 必須項目は空または既定値に戻る。
// 関連DD: DD-UI-006
function resetIssueCreateForm() {
  newIssueTitle.value = ''
  newIssueDescription.value = ''
  newIssueDueDate.value = ''
  newIssuePriority.value = 'Medium'
  newIssueAssignee.value = ''
  issueCreateErrorMessage.value = ''
  createIssuePickerDate.value = null
}

// handleCreateIssue は新規課題の作成を実行する。
// 目的: 入力内容を検証し、課題作成APIを呼び出す。
// 入力: なし。
// 出力: 成功時はダイアログを閉じる。
// エラー: 必須未入力やAPI失敗時はメッセージを表示する。
// 副作用: バックエンド呼び出しと一覧再取得を行う。
// 並行性: 単一UIイベント前提。
// 不変条件: 必須項目が空の場合は作成しない。
// 関連DD: DD-UI-006
async function handleCreateIssue() {
  if (!selectedCategory.value) {
    issueCreateErrorMessage.value = 'カテゴリを選択してください。'
    return
  }
  if (
    !newIssueTitle.value ||
    !newIssueDescription.value ||
    !newIssueDueDate.value ||
    !newIssuePriority.value
  ) {
    issueCreateErrorMessage.value = '必須項目を入力してください。'
    return
  }
  issueCreateErrorMessage.value = ''
  const result = await issuesStore.createIssue(selectedCategory.value, {
    title: newIssueTitle.value,
    description: newIssueDescription.value,
    due_date: newIssueDueDate.value,
    priority: newIssuePriority.value,
    assignee: newIssueAssignee.value,
  })
  if (result) {
    showIssueCreateDialog.value = false
    resetIssueCreateForm()
  } else {
    issueCreateErrorMessage.value = '課題の作成に失敗しました。'
  }
}

function handleCreateIssueDateUpdate(value) {
  newIssueDueDate.value = formatDate(value)
  showCreateIssueDatePicker.value = false
}

function handleFilterDueFromUpdate(value) {
  filterDueFrom.value = formatDate(value)
  showFilterDueFromPicker.value = false
  applyFilter()
}

function handleFilterDueToUpdate(value) {
  filterDueTo.value = formatDate(value)
  showFilterDueToPicker.value = false
  applyFilter()
}


// handleOpenIssue は課題詳細ダイアログの表示を要求する。
// 目的: 選択した課題を上位コンポーネントへ通知する。
// 入力: item は課題行の情報。
// 出力: なし。
// エラー: なし。
// 副作用: emit で open-issue を通知する。
// 並行性: 単一UIイベント前提。
// 不変条件: selectedCategory が無い場合は通知しない。
// 関連DD: DD-UI-006
function handleOpenIssue(item) {
  if (!selectedCategory.value) {
    return
  }
  emit('open-issue', { category: selectedCategory.value, issue_id: item.issue_id })
}

// テスト用にフィルタ操作を公開する。
// コンポーネントマウント時に初期フィルタをセットする
onMounted(() => {
   const query = currentQuery.value
  filterText.value = query.filter.text
  filterStatus.value = query.filter.status
  filterPriority.value = query.filter.priority
  filterDueFrom.value = query.filter.dueDateFrom ?? ''
  filterDueTo.value = query.filter.dueDateTo ?? ''
  filterSchemaInvalid.value = query.filter.schemaInvalidOnly
})

defineExpose({ applyFilter })
</script>

<template>
  <v-container class="py-6" fluid>
    <v-row>

      <v-col cols="12">
        <v-card rounded="lg">
          <v-card-title class="text-subtitle-1 d-flex align-center">
            課題一覧
            <v-spacer />
            <!-- <v-checkbox
              v-model="filterSchemaInvalid"
              label="エラー課題"
              density="compact"
              hide-details
              class="mr-4"
              @update:model-value="applyFilter"
            /> -->
            <v-btn
              size="small"
              variant="tonal"
              color="primary"
              :disabled="!selectedCategory"
              @click="handleOpenIssueCreateDialog"
              prepend-icon="mdi-plus"
            >
              新規課題
            </v-btn>
          </v-card-title>
          <v-card-text>
            <v-row class="mb-4" dense>
              <v-col cols="6">
                <v-text-field
                  v-model="filterText"
                  data-testid="filter-text"
                  label="検索"
                  variant="outlined"
                  density="compact"
                  prepend-inner-icon="mdi-magnify"
                  @update:model-value="applyFilter"
                />
              </v-col>
              <v-col cols="2">
                <v-select
                  v-model="filterStatus"
                  :items="statusOptions"
                  label="ステータス"
                  variant="outlined"
                  density="compact"
                  multiple
                  @update:model-value="applyFilter"
                />
              </v-col>
              <v-col cols="2">
                <v-select
                  v-model="filterPriority"
                  :items="priorityOptions"
                  label="優先度"
                  variant="outlined"
                  density="compact"
                  multiple
                  @update:model-value="applyFilter"
                />
              </v-col>
              <!-- <v-col cols="3">
                <v-menu
                  v-model="showFilterDueFromPicker"
                  :close-on-content-click="false"
                  min-width="auto"
                >
                  <template v-slot:activator="{ props }">
                    <v-text-field
                      v-model="filterDueFrom"
                      label="期限From"
                      readonly
                      v-bind="props"
                      variant="outlined"
                      density="compact"
                      placeholder="YYYY-MM-DD"
                      prepend-inner-icon="mdi-calendar"
                      clearable
                      @click:clear="filterDueFrom = ''; applyFilter()"
                    />
                  </template>
                  <v-date-picker
                    v-model="filterDueFromPickerDate"
                    color="primary"
                    @update:model-value="handleFilterDueFromUpdate"
                  />
                </v-menu>
              </v-col>
              <v-col cols="3">
                <v-menu
                  v-model="showFilterDueToPicker"
                  :close-on-content-click="false"
                  min-width="auto"
                >
                  <template v-slot:activator="{ props }">
                    <v-text-field
                      v-model="filterDueTo"
                      label="期限To"
                      readonly
                      v-bind="props"
                      variant="outlined"
                      density="compact"
                      placeholder="YYYY-MM-DD"
                      prepend-inner-icon="mdi-calendar"
                      clearable
                      @click:clear="filterDueTo = ''; applyFilter()"
                    />
                  </template>
                  <v-date-picker
                    v-model="filterDueToPickerDate"
                    color="primary"
                    @update:model-value="handleFilterDueToUpdate"
                  />
                </v-menu>
              </v-col>  -->
            </v-row>

            <v-table density="compact">
              <thead>
                <tr>
                  <th>
                    <v-btn
                      variant="text"
                      size="small"
                      data-testid="sort-title"
                      @click="toggleSort('title')"
                    >
                      件名 {{ sortLabel('title') }}
                    </v-btn>
                  </th>
                  <th>
                    <v-btn variant="text" size="small" @click="toggleSort('status')">
                      ステータス {{ sortLabel('status') }}
                    </v-btn>
                  </th>
                  <th>
                    <v-btn variant="text" size="small" @click="toggleSort('priority')">
                      優先度 {{ sortLabel('priority') }}
                    </v-btn>
                  </th>
                  <th>
                    <v-btn variant="text" size="small" @click="toggleSort('updated_at')">
                      更新日 {{ sortLabel('updated_at') }}
                    </v-btn>
                  </th>
                  <th>
                    <v-btn variant="text" size="small" @click="toggleSort('due_date')">
                      期限 {{ sortLabel('due_date') }}
                    </v-btn>
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in filteredItems"
                  :key="item.issue_id"
                  :class="{
                    'issue-row--closed': isEndState(item.status),
                    'issue-row--schema': item.is_schema_invalid,
                    'issue-row--clickable': true,
                  }"
                  @click="handleOpenIssue(item)"
                >
                  <td>
                    <v-icon
                      v-if="item.is_schema_invalid"
                      icon="mdi-alert-circle-outline"
                      color="warning"
                      size="x-small"
                      class="mr-1"
                    />
                    {{ item.title }}
                  </td>
                  <td>{{ item.status }}</td>
                  <td>{{ item.priority }}</td>
                  <td>{{ item.updated_at }}</td>
                  <td>{{ item.due_date }}</td>
                </tr>
              </tbody>
            </v-table>
          </v-card-text>
          <v-card-actions class="justify-end">
            <v-pagination
              v-if="selectedCategory"
              :model-value="currentQuery.page"
              :length="pageCount"
              @update:model-value="handlePageChange"
            />
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>



    <v-dialog v-model="showIssueCreateDialog" max-width="640">
      <v-card rounded="lg">
        <v-card-title class="text-subtitle-1">課題の新規作成</v-card-title>
        <v-card-text>
          <v-alert v-if="issueCreateErrorMessage" type="error" variant="tonal" class="mb-4">
            {{ issueCreateErrorMessage }}
          </v-alert>
          <v-text-field
            v-model="newIssueTitle"
            label="件名"
            variant="outlined"
            density="comfortable"
          />
          <v-textarea
            v-model="newIssueDescription"
            label="内容"
            variant="outlined"
            density="comfortable"
            rows="4"
          />
          <v-menu
            v-model="showCreateIssueDatePicker"
            :close-on-content-click="false"
            min-width="auto"
          >
            <template v-slot:activator="{ props }">
              <v-text-field
                v-model="newIssueDueDate"
                label="期限"
                readonly
                v-bind="props"
                variant="outlined"
                density="comfortable"
                placeholder="YYYY-MM-DD"
                prepend-inner-icon="mdi-calendar"
              />
            </template>
            <v-date-picker
              v-model="createIssuePickerDate"
              color="primary"
              @update:model-value="handleCreateIssueDateUpdate"
            />
          </v-menu>
          <v-select
            v-model="newIssuePriority"
            :items="priorityOptions"
            label="優先度"
            variant="outlined"
            density="comfortable"
          />
          <v-text-field
            v-model="newIssueAssignee"
            label="担当者"
            variant="outlined"
            density="comfortable"
          />
        </v-card-text>
        <v-card-actions class="justify-end">
          <v-btn variant="text" @click="showIssueCreateDialog = false">キャンセル</v-btn>
          <v-btn variant="flat" color="primary" @click="handleCreateIssue"> 作成 </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<style scoped>
.issue-row--closed {
  opacity: 0.6;
}

.issue-row--schema {
  background: rgba(255, 160, 0, 0.08);
}

.issue-row--clickable {
  cursor: pointer;
}
</style>
