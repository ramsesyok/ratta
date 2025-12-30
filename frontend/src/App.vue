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

const showProjectSelect = computed(() => !appStore.projectRoot)
const needsContractorAuth = computed(() => appStore.contractorAuthRequired && appStore.mode !== 'Contractor')
const isReady = computed(() => !showProjectSelect.value && !needsContractorAuth.value)

const unreadErrors = computed(() => errorsStore.items.filter((item) => !item.is_read).length)

// onMounted は起動時の初期データを読み込む。
// 目的: 前回のプロジェクトルートと設定を反映する。
// 入力: なし。
// 出力: なし。
// エラー: 取得失敗時は errors ストアへ登録される。
// 副作用: バックエンド呼び出しを行う。
// 並行性: 単一UIイベント前提。
// 不変条件: bootstrapLoaded が true になる。
// 関連DD: DD-UI-004
onMounted(async () => {
  if (!appStore.bootstrapLoaded) {
    await appStore.bootstrap()
  }
})

// watch(showProjectSelect) はプロジェクト選択ダイアログの表示を同期する。
// 目的: projectRoot の有無でダイアログ表示を切り替える。
// 入力: value は表示判定。
// 出力: なし。
// エラー: なし。
// 副作用: showProjectDialog を更新する。
// 並行性: 単一UIイベント前提。
// 不変条件: プロジェクト未選択時は true。
// 関連DD: DD-UI-004
watch(showProjectSelect, (value) => {
  showProjectDialog.value = value
}, { immediate: true })

// watch(needsContractorAuth) は Contractor ダイアログの表示を同期する。
// 目的: モード判定結果に応じて認証を要求する。
// 入力: value は表示判定。
// 出力: なし。
// エラー: なし。
// 副作用: showContractorDialog を更新する。
// 並行性: 単一UIイベント前提。
// 不変条件: contractorAuthRequired が true の間は true。
// 関連DD: DD-UI-008
watch(needsContractorAuth, (value) => {
  showContractorDialog.value = value
}, { immediate: true })

// watch(projectRoot) はプロジェクト選択後にモード判定を行う。
// 目的: Contractor 認証の要否を決定する。
// 入力: value は projectRoot。
// 出力: なし。
// エラー: 失敗時は errors ストアへ登録される。
// 副作用: バックエンド呼び出しを行う。
// 並行性: 単一UIイベント前提。
// 不変条件: projectRoot が空の時は判定しない。
// 関連DD: DD-UI-008
watch(() => appStore.projectRoot, async (value) => {
  if (value) {
    await appStore.detectMode()
  }
})

// handleOpenIssue は課題詳細ダイアログを開く。
// 目的: 選択された課題の詳細を読み込み、ダイアログを表示する。
// 入力: payload は category と issue_id を含む。
// 出力: なし。
// エラー: 取得失敗時は errors ストアへ登録される。
// 副作用: バックエンド呼び出しとダイアログ表示を行う。
// 並行性: 単一UIイベント前提。
// 不変条件: category と issue_id が揃わない場合は何もしない。
// 関連DD: DD-UI-006
async function handleOpenIssue(payload) {
  const category = payload?.category ?? categoriesStore.selectedCategory
  if (!category || !payload?.issue_id) {
    return
  }
  await issueDetailStore.openIssue(category, payload.issue_id)
  showIssueDetailDialog.value = true
}

// handleOpenErrors はエラー詳細ダイアログを開く。
// 目的: エラー一覧の確認を促す。
// 入力: なし。
// 出力: なし。
// エラー: なし。
// 副作用: showErrorDetailDialog を更新する。
// 並行性: 単一UIイベント前提。
// 不変条件: なし。
// 関連DD: DD-UI-007
function handleOpenErrors() {
  showErrorDetailDialog.value = true
}
</script>

<template>
  <v-app>
    <v-app-bar flat>
      <v-toolbar-title>ratta</v-toolbar-title>
      <v-spacer />
      <v-badge
        v-if="unreadErrors > 0"
        :content="unreadErrors"
        color="error"
        class="mr-3"
      >
        <v-btn variant="tonal" @click="handleOpenErrors">
          エラー詳細
        </v-btn>
      </v-badge>
      <v-btn v-else variant="text" @click="handleOpenErrors">
        エラー詳細
      </v-btn>
    </v-app-bar>

    <v-main>
      <MainView v-if="isReady" @open-issue="handleOpenIssue" />
    </v-main>

    <ProjectSelectDialog v-model="showProjectDialog" />
    <ContractorPasswordDialog v-model="showContractorDialog" />
    <IssueDetailDialog v-model="showIssueDetailDialog" @open-errors="handleOpenErrors" />
    <ErrorDetailDialog v-model="showErrorDetailDialog" />
  </v-app>
</template>
