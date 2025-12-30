<script setup>
// App はダイアログ群の表示制御と画面遷移の起点を担う。
// 実際の処理は各ストアとダイアログへ委譲する。
import { computed, onMounted, ref, watch } from 'vue'

import ContractorPasswordDialog from './components/ContractorPasswordDialog.vue'
import ErrorDetailDialog from './components/ErrorDetailDialog.vue'
import IssueDetailDialog from './components/IssueDetailDialog.vue'
import MainView from './components/MainView.vue'
import ProjectSelectDialog from './components/ProjectSelectDialog.vue'
import { useAppStore } from './stores/app'
import { useCategoriesStore } from './stores/categories'
import { useErrorsStore } from './stores/errors'
import { useIssueDetailStore } from './stores/issueDetail'

const appStore = useAppStore()
const categoriesStore = useCategoriesStore()
const errorsStore = useErrorsStore()
const issueDetailStore = useIssueDetailStore()

const showProjectDialog = ref(false)
const showContractorDialog = ref(false)
const showIssueDetailDialog = ref(false)
const showErrorDetailDialog = ref(false)

const drawer = ref(true)
const showCreateDialog = ref(false)
const showRenameDialog = ref(false)
const showDeleteDialog = ref(false)
const newCategoryName = ref('')
const renameCategoryName = ref('')
const targetCategoryName = ref('')

const showProjectSelect = computed(() => !appStore.projectRoot)
const needsContractorAuth = computed(() => appStore.contractorAuthRequired && appStore.mode !== 'Contractor')
const isReady = computed(() => !showProjectSelect.value && !needsContractorAuth.value)

const unreadErrors = computed(() => errorsStore.items.filter((item) => !item.is_read).length)
const selectedCategory = computed(() => categoriesStore.selectedCategory)

// onMounted は起動時の初期データを読み込む。
onMounted(async () => {
  if (!appStore.bootstrapLoaded) {
    await appStore.bootstrap()
  }
})

// プロジェクトロード完了後にカテゴリを読み込む
watch(isReady, async (ready) => {
  if (ready) {
    await categoriesStore.loadCategories()
    if (!categoriesStore.selectedCategory && categoriesStore.items.length > 0) {
      await categoriesStore.selectCategory(categoriesStore.items[0].name)
    }
  }
})

watch(showProjectSelect, (value) => {
  showProjectDialog.value = value
}, { immediate: true })

watch(needsContractorAuth, (value) => {
  showContractorDialog.value = value
}, { immediate: true })

watch(() => appStore.projectRoot, async (value) => {
  if (value) {
    await appStore.detectMode()
  }
})

async function handleOpenIssue(payload) {
  const category = payload?.category ?? categoriesStore.selectedCategory
  if (!category || !payload?.issue_id) {
    return
  }
  await issueDetailStore.openIssue(category, payload.issue_id)
  showIssueDetailDialog.value = true
}

function handleOpenErrors() {
  showErrorDetailDialog.value = true
}

async function handleSelectCategory(name) {
  await categoriesStore.selectCategory(name)
  // モバイルの場合は選択後にドロワーを閉じる (UX向上のため)
  // if (window.innerWidth < 600) drawer.value = false 
}

async function handleCreateCategory() {
  await categoriesStore.createCategory(newCategoryName.value)
  newCategoryName.value = ''
  showCreateDialog.value = false
}

function openRenameDialog(name) {
  targetCategoryName.value = name
  showRenameDialog.value = true
}

async function handleRenameCategory() {
  if (!targetCategoryName.value) return
  await categoriesStore.renameCategory(targetCategoryName.value, renameCategoryName.value)
  renameCategoryName.value = ''
  showRenameDialog.value = false
}

function openDeleteDialog(name) {
  targetCategoryName.value = name
  showDeleteDialog.value = true
}

async function handleDeleteCategory() {
  if (!targetCategoryName.value) return
  await categoriesStore.deleteCategory(targetCategoryName.value)
  showDeleteDialog.value = false
}
</script>

<template>
  <v-app>
    <v-app-bar density="compact">
      <v-app-bar-nav-icon v-if="isReady" @click="drawer = !drawer" />
      <v-toolbar-title>ratta</v-toolbar-title>
      <v-spacer />
      <v-badge
        v-if="unreadErrors > 0"
        :content="unreadErrors"
        color="error"
        class="mr-3"
      >
        <v-btn variant="tonal" @click="handleOpenErrors" icon="mdi-bell">
          
        </v-btn>
      </v-badge>
      <v-btn v-else variant="text" @click="handleOpenErrors" icon="mdi-bell">
        
      </v-btn>
    </v-app-bar>

    <v-navigation-drawer v-if="isReady" v-model="drawer">
      <v-list density="compact" nav>
        <v-list-item
          v-for="item in categoriesStore.items"
          :key="item.name"
          :active="item.name === selectedCategory"
          @click="handleSelectCategory(item.name)"
        >
          <v-list-item-title>{{ item.name }}</v-list-item-title>
          <template v-slot:append>
             <v-badge v-if="item.issueCount" :content="item.issueCount" inline color="grey-lighten-1" />
             <v-menu v-if="appStore.mode === 'Contractor'">
               <template v-slot:activator="{ props }">
                 <v-btn icon="mdi-dots-vertical" variant="text" size="small" v-bind="props" @click.stop />
               </template>
               <v-list>
                 <v-list-item @click="openRenameDialog(item.name)">
                   <v-list-item-title>変更</v-list-item-title>
                 </v-list-item>
                 <v-list-item @click="openDeleteDialog(item.name)">
                   <v-list-item-title class="text-error">削除</v-list-item-title>
                 </v-list-item>
               </v-list>
             </v-menu>
          </template>
        </v-list-item>
      </v-list>

      <template v-slot:append>
        <div v-if="appStore.mode === 'Contractor'" class="pa-2">
           <v-row dense>
             <v-col cols="12">
               <v-btn block variant="tonal" @click="showCreateDialog = true" class="mb-2" prepend-icon="mdi-plus-circle">カテゴリ追加</v-btn>
             </v-col>
           </v-row>
        </div>
      </template>
    </v-navigation-drawer>

    <v-main>
      <MainView v-if="isReady" @open-issue="handleOpenIssue" />
    </v-main>

    <ProjectSelectDialog v-model="showProjectDialog" />
    <ContractorPasswordDialog v-model="showContractorDialog" />
    <IssueDetailDialog v-model="showIssueDetailDialog" @open-errors="handleOpenErrors" />
    <ErrorDetailDialog v-model="showErrorDetailDialog" />

    <v-dialog v-model="showCreateDialog" max-width="420">
      <v-card rounded="lg">
        <v-card-title class="text-subtitle-1">カテゴリ追加</v-card-title>
        <v-card-text>
          <v-text-field v-model="newCategoryName" label="カテゴリ名" />
        </v-card-text>
        <v-card-actions class="justify-end">
          <v-btn variant="text" @click="showCreateDialog = false">キャンセル</v-btn>
          <v-btn variant="flat" color="primary" @click="handleCreateCategory"> 追加 </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-model="showRenameDialog" max-width="420">
      <v-card rounded="lg">
        <v-card-title class="text-subtitle-1">カテゴリ名変更</v-card-title>
        <v-card-text>
          <v-text-field v-model="renameCategoryName" label="新しいカテゴリ名" />
        </v-card-text>
        <v-card-actions class="justify-end">
          <v-btn variant="text" @click="showRenameDialog = false">キャンセル</v-btn>
          <v-btn variant="flat" color="primary" @click="handleRenameCategory"> 変更 </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-model="showDeleteDialog" max-width="420">
      <v-card rounded="lg">
        <v-card-title class="text-subtitle-1">カテゴリ削除</v-card-title>
        <v-card-text>選択中のカテゴリを削除しますか？</v-card-text>
        <v-card-actions class="justify-end">
          <v-btn variant="text" @click="showDeleteDialog = false">キャンセル</v-btn>
          <v-btn variant="flat" color="error" @click="handleDeleteCategory"> 削除 </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-app>
</template>
