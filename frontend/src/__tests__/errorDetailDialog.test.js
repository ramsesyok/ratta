// errorDetailDialog.test.js は ErrorDetailDialog の表示とフィルタを検証する。
// エラーストアの状態を操作し、UIが期待通り描画されるか確認する。
import { createPinia, setActivePinia } from 'pinia'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { createVuetify } from 'vuetify'

import ErrorDetailDialog from '../components/ErrorDetailDialog.vue'
import { useCategoriesStore } from '../stores/categories'
import { useErrorsStore } from '../stores/errors'

const vuetify = createVuetify()

// setupStores はテスト用のストア状態を準備する。
// 目的: エラー一覧とカテゴリ選択の条件を簡易に作る。
// 入力: selectedCategory は現在選択中カテゴリ名。
// 出力: errors/categories ストア参照。
// エラー: なし。
// 副作用: Pinia の状態を更新する。
// 並行性: テスト単位で独立。
// 不変条件: errors.items は配列として設定する。
// 関連DD: DD-UI-007
function setupStores(selectedCategory) {
  setActivePinia(createPinia())
  const errors = useErrorsStore()
  const categories = useCategoriesStore()
  categories.selectedCategory = selectedCategory
  errors.items = [
    {
      id: 'err-1',
      occurred_at: '2024-01-01T00:00:00Z',
      source: 'issues',
      action: 'loadIssues',
      category: 'Cat',
      severity: 'error',
      api: {
        error_code: 'E_VALIDATION',
        message: 'invalid',
        detail: 'detail text',
        target_path: 'C:/proj/Cat/issue.json',
        hint: ''
      },
      user_message: '読み込みに失敗しました',
      is_read: false
    },
    {
      id: 'err-2',
      occurred_at: '2024-01-02T00:00:00Z',
      source: 'issues',
      action: 'loadIssues',
      category: 'Other',
      severity: 'error',
      api: {
        error_code: 'E_IO',
        message: 'io error',
        detail: '',
        target_path: 'C:/proj/Other/issue.json',
        hint: ''
      },
      user_message: '別カテゴリのエラー',
      is_read: false
    }
  ]
  return { errors, categories }
}

// mountDialog は ErrorDetailDialog をマウントする。
// 目的: Vuetify とテレポートの設定を含めて描画する。
// 入力: なし。
// 出力: Vue Test Utils の wrapper。
// エラー: なし。
// 副作用: DOM へコンポーネントを描画する。
// 並行性: テスト単位で独立。
// 不変条件: VDialog はスタブ化する。
// 関連DD: DD-UI-007
function mountDialog() {
  return mount(ErrorDetailDialog, {
    global: {
      plugins: [vuetify],
      stubs: {
        teleport: true,
        VDialog: { template: '<div><slot /></div>' }
      }
    }
  })
}

describe('ErrorDetailDialog', () => {
  it('filters errors by selected category', async () => {
    // カテゴリフィルタで対象カテゴリのみ表示されることを確認する。
    setupStores('Cat')
    const wrapper = mountDialog()
    await wrapper.vm.$nextTick()

    await wrapper.find('[data-testid="scope-category"]').trigger('click')
    await wrapper.vm.$nextTick()

    const items = wrapper.findAll('[data-testid="error-item"]')
    expect(items.length).toBe(1)
    expect(wrapper.text()).toContain('読み込みに失敗しました')
    expect(wrapper.text()).not.toContain('別カテゴリのエラー')
  })

  it('copies message text via clipboard', async () => {
    // コピー操作でクリップボードAPIが呼ばれることを確認する。
    setupStores('Cat')
    const writeText = vi.fn().mockResolvedValue()
    globalThis.navigator.clipboard = { writeText }

    const wrapper = mountDialog()
    await wrapper.vm.$nextTick()

    await wrapper.find('[data-testid="copy-message"]').trigger('click')

    expect(writeText).toHaveBeenCalledWith('読み込みに失敗しました')
  })
})
