import { createPinia, setActivePinia } from 'pinia'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { createVuetify } from 'vuetify'

import MainView from '../components/MainView.vue'
import { useAppStore } from '../stores/app'
import { useCategoriesStore } from '../stores/categories'
import { useIssuesStore } from '../stores/issues'

vi.mock('../utils/apiClient', () => ({
  ApiError: class ApiError extends Error {},
  listIssues: vi.fn()
}))

import * as apiClient from '../utils/apiClient'

const vuetify = createVuetify()

function mountMainView() {
  return mount(MainView, {
    global: {
      plugins: [vuetify],
      stubs: {
        teleport: true,
        VDialog: { template: '<div><slot /></div>' },
        VPagination: {
          template: '<button data-testid="pagination" @click="$emit(\'update:modelValue\', 2)">page</button>'
        }
      }
    }
  })
}

describe('MainView', () => {
  it('sorts issues when header is clicked', async () => {
    // ソート操作で一覧取得が呼ばれることを確認する。
    setActivePinia(createPinia())
    const app = useAppStore()
    app.pageSize = 20
    const categories = useCategoriesStore()
    categories.selectedCategory = 'Cat'
    categories.items = [{ name: 'Cat' }]
    categories.loadCategories = vi.fn().mockResolvedValue(null)
    const issues = useIssuesStore()
    issues.issuesByCategory.Cat = { items: [], total: 0, lastLoadedAt: null, isLoading: false }

    apiClient.listIssues.mockResolvedValue({ issues: [], total: 0 })

    const wrapper = mountMainView()
    await wrapper.vm.$nextTick()

    await wrapper.find('[data-testid="sort-title"]').trigger('click')

    expect(apiClient.listIssues).toHaveBeenCalled()
  })

  it('filters issues by text', async () => {
    // テキスト検索で表示件数が絞られることを確認する。
    setActivePinia(createPinia())
    const categories = useCategoriesStore()
    categories.selectedCategory = 'Cat'
    categories.items = [{ name: 'Cat' }]
    categories.loadCategories = vi.fn().mockResolvedValue(null)
    const issues = useIssuesStore()
    issues.issuesByCategory.Cat = {
      items: [
        { issue_id: '1', title: 'Alpha', status: 'Open', priority: 'High', updated_at: '2024', due_date: '2024-01-01' },
        { issue_id: '2', title: 'Beta', status: 'Open', priority: 'High', updated_at: '2024', due_date: '2024-01-01' }
      ],
      total: 2,
      lastLoadedAt: null,
      isLoading: false
    }

    const wrapper = mountMainView()
    await wrapper.vm.$nextTick()

    const input = wrapper.find('[data-testid="filter-text"] input')
    await input.setValue('Al')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('Alpha')
    expect(wrapper.text()).not.toContain('Beta')
  })

  it('changes page via pagination', async () => {
    // ページ変更で一覧取得が呼ばれることを確認する。
    setActivePinia(createPinia())
    const app = useAppStore()
    app.pageSize = 20
    const categories = useCategoriesStore()
    categories.selectedCategory = 'Cat'
    categories.items = [{ name: 'Cat' }]
    categories.loadCategories = vi.fn().mockResolvedValue(null)
    const issues = useIssuesStore()
    issues.issuesByCategory.Cat = { items: [], total: 40, lastLoadedAt: null, isLoading: false }

    apiClient.listIssues.mockResolvedValue({ issues: [], total: 40 })

    const wrapper = mountMainView()
    await wrapper.vm.$nextTick()

    await wrapper.find('[data-testid="pagination"]').trigger('click')

    expect(apiClient.listIssues).toHaveBeenCalled()
  })
})
