// issueDetailDialog.test.js は課題詳細ダイアログのUI挙動を検証する。
// API通信は行わず、ストア状態と表示の連携のみを確認する。
import { createPinia, setActivePinia } from 'pinia'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { createVuetify } from 'vuetify'

import IssueDetailDialog from '../components/IssueDetailDialog.vue'
import { useCategoriesStore } from '../stores/categories'
import { useIssueDetailStore } from '../stores/issueDetail'

const vuetify = createVuetify()

// setupStores はテスト用ストア状態を初期化する。
// 目的: 読み取り専用やスキーマ不整合の条件を簡易に切り替える。
// 入力: isReadOnly/isSchemaInvalid は抑止条件の有無。
// 出力: issueDetail/categories ストア参照。
// エラー: なし。
// 副作用: Pinia の状態を上書きする。
// 並行性: テスト単位で独立。
// 不変条件: current と currentCategory を必ず設定する。
// 関連DD: DD-UI-006
function setupStores({ isReadOnly = false, isSchemaInvalid = false } = {}) {
  setActivePinia(createPinia())
  const issueDetail = useIssueDetailStore()
  const categories = useCategoriesStore()

  issueDetail.current = {
    issue_id: 'ISSUE-1',
    title: 'Sample Title',
    description: 'Sample Description',
    status: 'Open',
    priority: 'High',
    due_date: '2024-01-01',
    assignee: '担当者',
    comments: [
      {
        comment_id: 'COMMENT-1',
        author_name: '作成者',
        body: '本文'
      }
    ],
    is_schema_invalid: isSchemaInvalid
  }
  issueDetail.currentCategory = 'Cat'
  issueDetail.saveIssue = vi.fn().mockResolvedValue(issueDetail.current)
  issueDetail.addComment = vi.fn().mockResolvedValue(issueDetail.current)
  issueDetail.reloadCurrent = vi.fn().mockResolvedValue(issueDetail.current)

  categories.items = [{ name: 'Cat', is_read_only: isReadOnly }]

  return { issueDetail, categories }
}

// mountDialog は IssueDetailDialog をマウントする。
// 目的: Vuetify とテレポートの設定を含めて描画する。
// 入力: なし。
// 出力: Vue Test Utils の wrapper。
// エラー: なし。
// 副作用: DOM へコンポーネントを描画する。
// 並行性: テスト単位で独立。
// 不変条件: VDialog はスタブ化する。
// 関連DD: DD-UI-006
function mountDialog() {
  return mount(IssueDetailDialog, {
    global: {
      plugins: [vuetify],
      stubs: {
        teleport: true,
        VDialog: { template: '<div><slot /></div>' }
      }
    }
  })
}

describe('IssueDetailDialog', () => {
  it('toggles edit mode when edit button is clicked', async () => {
    // 編集ボタンで編集モードへ遷移することを確認する。
    setupStores()
    const wrapper = mountDialog()
    await wrapper.vm.$nextTick()

    await wrapper.find('[data-testid="edit"]').trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.find('[data-testid="save"]').exists()).toBe(true)
  })

  it('shows validation message when required fields are empty', async () => {
    // 必須項目が空の場合にバリデーションメッセージが表示されることを確認する。
    setupStores()
    const wrapper = mountDialog()
    await wrapper.vm.$nextTick()

    await wrapper.find('[data-testid="edit"]').trigger('click')
    await wrapper.vm.$nextTick()

    const titleInput = wrapper.find('[data-testid="edit-title"] input')
    await titleInput.setValue('')
    await wrapper.find('[data-testid="save"]').trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('必須項目を入力してください。')
  })

  it('blocks editing and commenting when category is read-only', async () => {
    // 読み取り専用カテゴリでは編集とコメントが抑止されることを確認する。
    setupStores({ isReadOnly: true })
    const wrapper = mountDialog()
    await wrapper.vm.$nextTick()

    expect(wrapper.find('[data-testid="edit"]').attributes('disabled')).toBeDefined()
    expect(wrapper.find('[data-testid="comment-submit"]').attributes('disabled')).toBeDefined()
    expect(wrapper.text()).toContain('スキーマ不整合または読み取り専用のため編集できません。')

    const errorButton = wrapper.findAll('button').find((node) => node.text() === 'エラー詳細')
    await errorButton.trigger('click')

    expect(wrapper.emitted()['open-errors']).toBeTruthy()
  })

  it('reloads detail when dialog opens', async () => {
    // ダイアログ表示時に詳細の再読み込みが行われることを確認する。
    const { issueDetail } = setupStores()
    const wrapper = mountDialog()
    await wrapper.vm.$nextTick()

    expect(issueDetail.reloadCurrent).toHaveBeenCalled()
  })
})
