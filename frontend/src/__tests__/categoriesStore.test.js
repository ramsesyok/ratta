import { createPinia, setActivePinia } from 'pinia'
import { describe, expect, it, vi } from 'vitest'

import { useCategoriesStore } from '../stores/categories'
import { useErrorsStore } from '../stores/errors'

vi.mock('../utils/apiClient', () => ({
  listCategories: vi.fn(),
  createCategory: vi.fn(),
  renameCategory: vi.fn(),
  deleteCategory: vi.fn()
}))

import * as apiClient from '../utils/apiClient'

describe('categories store', () => {
  it('loads categories', async () => {
    // カテゴリ一覧が state に反映されることを確認する。
    setActivePinia(createPinia())
    const store = useCategoriesStore()

    apiClient.listCategories.mockResolvedValue({ categories: [{ name: 'Cat' }] })

    await store.loadCategories()

    expect(store.items.length).toBe(1)
  })

  it('captures permission error on create', async () => {
    // Vendor モードで作成するとエラーが登録されることを確認する。
    setActivePinia(createPinia())
    const store = useCategoriesStore()
    const errors = useErrorsStore()

    await store.createCategory('Cat')

    expect(errors.items.length).toBe(1)
  })
})
