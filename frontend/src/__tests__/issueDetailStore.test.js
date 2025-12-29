import { createPinia, setActivePinia } from 'pinia'
import { describe, expect, it, vi } from 'vitest'

import { useErrorsStore } from '../stores/errors'
import { useIssueDetailStore } from '../stores/issueDetail'

vi.mock('../utils/apiClient', () => ({
  ApiError: class ApiError extends Error {},
  getIssue: vi.fn(),
  updateIssue: vi.fn(),
  addComment: vi.fn()
}))

import * as apiClient from '../utils/apiClient'

describe('issue detail store', () => {
  it('captures error on schema invalid save', async () => {
    // スキーマ不整合の課題が保存されないことを確認する。
    setActivePinia(createPinia())
    const store = useIssueDetailStore()
    const errors = useErrorsStore()

    store.current = { issue_id: '1', category: 'Cat', is_schema_invalid: true }
    store.currentCategory = 'Cat'

    await store.saveIssue({ title: 'new' })

    expect(errors.items.length).toBe(1)
  })

  it('loads issue detail', async () => {
    // openIssue が詳細を反映することを確認する。
    setActivePinia(createPinia())
    const store = useIssueDetailStore()

    apiClient.getIssue.mockResolvedValue({ issue_id: '1', category: 'Cat' })

    await store.openIssue('Cat', '1')

    expect(store.current.issue_id).toBe('1')
  })
})
