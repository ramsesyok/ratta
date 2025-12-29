import { createPinia, setActivePinia } from 'pinia'
import { describe, expect, it, vi } from 'vitest'

import { useErrorsStore } from '../stores/errors'
import { useIssuesStore } from '../stores/issues'

vi.mock('../utils/apiClient', () => ({
  ApiError: class ApiError extends Error {},
  listIssues: vi.fn()
}))

import * as apiClient from '../utils/apiClient'

describe('issues store', () => {
  it('captures errors on loadIssues failure', async () => {
    // 読み込み失敗時に errors ストアへ登録されることを確認する。
    setActivePinia(createPinia())
    const store = useIssuesStore()
    const errors = useErrorsStore()

    apiClient.listIssues.mockRejectedValue(new Error('failed'))

    await store.loadIssues('Cat')

    expect(errors.items.length).toBe(1)
  })

  it('updates cache on successful load', async () => {
    // 正常応答がキャッシュに反映されることを確認する。
    setActivePinia(createPinia())
    const store = useIssuesStore()

    apiClient.listIssues.mockResolvedValue({ issues: [{ issue_id: '1' }], total: 1 })

    await store.loadIssues('Cat')

    expect(store.issuesByCategory.Cat.items.length).toBe(1)
  })
})
