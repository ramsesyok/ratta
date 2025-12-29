<script setup>
// IssueDetailDialog は課題詳細の表示と編集を担う。
// 編集保存やコメント追加の実処理はストアに委ねる。
import MarkdownIt from 'markdown-it'
import { computed, ref, watch } from 'vue'

import { useCategoriesStore } from '../stores/categories'
import { useErrorsStore } from '../stores/errors'
import { useIssueDetailStore } from '../stores/issueDetail'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['update:modelValue', 'open-errors'])

const issueDetailStore = useIssueDetailStore()
const categoriesStore = useCategoriesStore()
const errorsStore = useErrorsStore()

const md = new MarkdownIt({ linkify: true, breaks: true })

const editMode = ref(false)
const errorMessage = ref('')

const editTitle = ref('')
const editDescription = ref('')
const editStatus = ref('')
const editPriority = ref('')
const editDueDate = ref('')
const editAssignee = ref('')

const commentBody = ref('')
const commentAuthor = ref('')
const commentAttachments = ref([])

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const current = computed(() => issueDetailStore.current)
const currentCategory = computed(() => issueDetailStore.currentCategory)

const isReadOnlyCategory = computed(() => {
  const category = categoriesStore.items.find((item) => item.name === currentCategory.value)
  return category?.is_read_only ?? false
})

const isSchemaInvalid = computed(() => current.value?.is_schema_invalid ?? false)
const isBlocked = computed(() => isReadOnlyCategory.value || isSchemaInvalid.value)

// watch(isOpen) はダイアログ表示時に詳細を再読み込みする。
// 目的: 表示の都度ディスク上の最新状態を反映する。
// 入力: value はダイアログ表示の真偽値。
// 出力: なし。
// エラー: 失敗時は issueDetailStore 側のエラー処理に委ねる。
// 副作用: バックエンド呼び出しを行う。
// 並行性: 単一UIイベント前提。
// 不変条件: value が false の場合は再読み込みしない。
// 関連DD: DD-UI-006
watch(
  isOpen,
  async (value) => {
    if (value) {
      await issueDetailStore.reloadCurrent()
    }
  },
  { immediate: true }
)

watch(
  current,
  (value) => {
    if (!value) {
      return
    }
    editTitle.value = value.title ?? ''
    editDescription.value = value.description ?? ''
    editStatus.value = value.status ?? ''
    editPriority.value = value.priority ?? ''
    editDueDate.value = value.due_date ?? ''
    editAssignee.value = value.assignee ?? ''
  },
  { immediate: true }
)

// enterEdit は編集モードへ遷移する。
// 目的: 読み取り専用の抑止を行い、編集モードを開始する。
// 入力: なし。
// 出力: なし。
// エラー: 読み取り専用時は errorsStore に登録する。
// 副作用: エラー登録とエラーダイアログ誘導イベントの発火。
// 並行性: 単一UIイベント前提。
// 不変条件: 読み取り専用時は editMode を true にしない。
// 関連DD: DD-UI-006
function enterEdit() {
  if (isBlocked.value) {
    errorMessage.value = '読み取り専用の課題は編集できません。'
    errorsStore.capture(new Error('read-only issue cannot be edited'), {
      source: 'issueDetail',
      action: 'enterEdit',
      category: currentCategory.value,
      issue_id: current.value?.issue_id
    })
    emit('open-errors')
    return
  }
  editMode.value = true
  errorMessage.value = ''
}

// cancelEdit は編集モードを終了して表示値を復元する。
// 目的: 未保存の変更を破棄して最新の詳細に戻す。
// 入力: なし。
// 出力: なし。
// エラー: なし。
// 副作用: なし。
// 並行性: 単一UIイベント前提。
// 不変条件: current がある場合は編集フィールドを復元する。
// 関連DD: DD-UI-006
function cancelEdit() {
  editMode.value = false
  errorMessage.value = ''
  if (current.value) {
    editTitle.value = current.value.title ?? ''
    editDescription.value = current.value.description ?? ''
    editStatus.value = current.value.status ?? ''
    editPriority.value = current.value.priority ?? ''
    editDueDate.value = current.value.due_date ?? ''
    editAssignee.value = current.value.assignee ?? ''
  }
}

// saveEdit は編集内容を保存する。
// 目的: 必須項目の検証を行い、課題更新APIを呼び出す。
// 入力: なし。
// 出力: 成功時は editMode を終了する。
// エラー: 読み取り専用や必須未入力時にメッセージを設定する。
// 副作用: バックエンド呼び出しとエラーストア更新。
// 並行性: 単一UIイベント前提。
// 不変条件: 必須項目が空の場合は更新しない。
// 関連DD: DD-UI-006
async function saveEdit() {
  if (!current.value || !currentCategory.value) {
    return
  }
  if (isBlocked.value) {
    errorMessage.value = '読み取り専用の課題は更新できません。'
    errorsStore.capture(new Error('read-only issue cannot be updated'), {
      source: 'issueDetail',
      action: 'saveIssue',
      category: currentCategory.value,
      issue_id: current.value.issue_id
    })
    emit('open-errors')
    return
  }
  if (!editTitle.value || !editDescription.value || !editStatus.value || !editPriority.value || !editDueDate.value) {
    errorMessage.value = '必須項目を入力してください。'
    return
  }
  errorMessage.value = ''
  const result = await issueDetailStore.saveIssue({
    title: editTitle.value,
    description: editDescription.value,
    status: editStatus.value,
    priority: editPriority.value,
    due_date: editDueDate.value,
    assignee: editAssignee.value
  })
  if (result) {
    editMode.value = false
  }
}

// handleFileChange はコメント添付ファイルを登録する。
// 目的: 追加添付をリストに反映し、最大5件に制限する。
// 入力: event は input[type=file] の change イベント。
// 出力: なし。
// エラー: なし。
// 副作用: commentAttachments を更新し、input 値をクリアする。
// 並行性: 単一UIイベント前提。
// 不変条件: 添付件数は5件を超えない。
// 関連DD: DD-UI-006
function handleFileChange(event) {
  const files = Array.from(event.target.files ?? [])
  const next = commentAttachments.value.concat(
    files.map((file) => ({
      source_path: '',
      original_file_name: file.name,
      mime_type: file.type
    }))
  )
  // 添付上限は UI で超過入力されても切り捨てる。
  commentAttachments.value = next.slice(0, 5)
  event.target.value = ''
}

// addComment はコメントを追加する。
// 目的: 必須項目の検証を行い、コメント追加APIを呼び出す。
// 入力: なし。
// 出力: 成功時は入力欄を初期化する。
// エラー: 読み取り専用や必須未入力時にメッセージを設定する。
// 副作用: バックエンド呼び出しとエラーストア更新。
// 並行性: 単一UIイベント前提。
// 不変条件: 添付件数は5件以内。
// 関連DD: DD-UI-006
async function addComment() {
  if (!current.value || !currentCategory.value) {
    return
  }
  if (isBlocked.value) {
    errorMessage.value = '読み取り専用の課題にはコメントできません。'
    errorsStore.capture(new Error('read-only issue cannot be commented'), {
      source: 'comments',
      action: 'addComment',
      category: currentCategory.value,
      issue_id: current.value.issue_id
    })
    emit('open-errors')
    return
  }
  if (!commentBody.value || !commentAuthor.value) {
    errorMessage.value = 'コメント本文と作成者名を入力してください。'
    return
  }
  if (commentAttachments.value.length > 5) {
    errorMessage.value = '添付は5件までです。'
    return
  }
  errorMessage.value = ''
  const result = await issueDetailStore.addComment({
    body: commentBody.value,
    author_name: commentAuthor.value,
    attachments: commentAttachments.value
  })
  if (result) {
    commentBody.value = ''
    commentAuthor.value = ''
    commentAttachments.value = []
  }
}

// renderMarkdown はコメント本文を HTML に変換する。
// 目的: Markdown のレンダリング結果を表示する。
// 入力: value は Markdown 文字列。
// 出力: HTML 文字列。
// エラー: なし。
// 副作用: なし。
// 並行性: スレッドセーフ。
// 不変条件: null は空文字として扱う。
// 関連DD: DD-UI-006
function renderMarkdown(value) {
  return md.render(value ?? '')
}
</script>

<template>
  <v-dialog v-model="isOpen" max-width="960">
    <v-card rounded="lg">
      <v-card-title class="text-h6">
        課題詳細
      </v-card-title>
      <v-card-text v-if="current">
        <v-alert
          v-if="isBlocked"
          type="warning"
          variant="tonal"
          class="mb-4"
        >
          スキーマ不整合または読み取り専用のため編集できません。
          <v-btn variant="text" size="small" @click="$emit('open-errors')">
            エラー詳細
          </v-btn>
        </v-alert>
        <v-alert
          v-if="errorMessage"
          type="error"
          variant="tonal"
          class="mb-4"
        >
          {{ errorMessage }}
        </v-alert>

        <div v-if="!editMode">
          <p class="text-h6 mb-2">{{ current.title }}</p>
          <p class="text-body-2 mb-2">{{ current.description }}</p>
          <p class="text-caption mb-1">ステータス: {{ current.status }}</p>
          <p class="text-caption mb-1">優先度: {{ current.priority }}</p>
          <p class="text-caption mb-1">期限: {{ current.due_date }}</p>
          <p class="text-caption mb-1">担当: {{ current.assignee || '未設定' }}</p>
          <v-btn
            data-testid="edit"
            variant="tonal"
            color="primary"
            class="mt-3"
            :disabled="isBlocked"
            @click="enterEdit"
          >
            編集
          </v-btn>
        </div>

        <div v-else>
          <v-text-field v-model="editTitle" label="件名" data-testid="edit-title" />
          <v-textarea v-model="editDescription" label="詳細" rows="4" />
          <v-select
            v-model="editStatus"
            :items="['Open','Working','Inquiry','Hold','Feedback','Resolved','Closed','Rejected']"
            label="ステータス"
          />
          <v-select
            v-model="editPriority"
            :items="['High','Medium','Low']"
            label="優先度"
          />
          <v-text-field v-model="editDueDate" label="期限" placeholder="YYYY-MM-DD" />
          <v-text-field v-model="editAssignee" label="担当者" />
          <v-card-actions class="justify-end">
            <v-btn variant="text" @click="cancelEdit">キャンセル</v-btn>
            <v-btn
              data-testid="save"
              variant="flat"
              color="teal"
              @click="saveEdit"
            >
              保存
            </v-btn>
          </v-card-actions>
        </div>

        <v-divider class="my-4" />

        <div>
          <p class="text-subtitle-2 mb-2">コメント</p>
          <v-text-field v-model="commentAuthor" label="作成者名" data-testid="comment-author" />
          <v-textarea v-model="commentBody" label="コメント本文" rows="3" data-testid="comment-body" />
          <input
            type="file"
            multiple
            :disabled="isBlocked"
            data-testid="comment-files"
            @change="handleFileChange"
          />
          <v-btn
            data-testid="comment-submit"
            variant="flat"
            color="primary"
            class="mt-2"
            :disabled="isBlocked"
            @click="addComment"
          >
            コメント追加
          </v-btn>

          <v-list class="mt-4">
            <v-list-item
              v-for="comment in current.comments"
              :key="comment.comment_id"
            >
              <v-list-item-title>{{ comment.author_name }}</v-list-item-title>
              <v-list-item-subtitle>
                <div v-html="renderMarkdown(comment.body)" />
              </v-list-item-subtitle>
            </v-list-item>
          </v-list>
        </div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>
