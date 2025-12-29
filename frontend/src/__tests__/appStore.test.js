import { createPinia, setActivePinia } from 'pinia'
import { describe, expect, it, vi } from 'vitest'

import { useAppStore } from '../stores/app'
import { useErrorsStore } from '../stores/errors'

vi.mock('../utils/apiClient', () => ({
  ApiError: class ApiError extends Error {},
  getAppBootstrap: vi.fn(),
  validateProjectRoot: vi.fn(),
  saveLastProjectRoot: vi.fn(),
  createProjectRoot: vi.fn(),
  detectMode: vi.fn(),
  verifyContractorPassword: vi.fn()
}))

import * as apiClient from '../utils/apiClient'

describe('app store', () => {
  it('bootstraps and updates state', async () => {
    // 起動時情報が state に反映されることを確認する。
    setActivePinia(createPinia())
    const store = useAppStore()

    apiClient.getAppBootstrap.mockResolvedValue({
      ui_page_size: 50,
      last_project_root_path: 'C:/proj',
      has_contractor_auth_file: true
    })

    await store.bootstrap()

    expect(store.pageSize).toBe(50)
    expect(store.projectRoot).toBe('C:/proj')
    expect(store.contractorAuthRequired).toBe(true)
    expect(store.bootstrapLoaded).toBe(true)
  })

  it('captures errors on bootstrap failure', async () => {
    // 取得失敗時に errors ストアへ登録されることを確認する。
    setActivePinia(createPinia())
    const store = useAppStore()
    const errors = useErrorsStore()

    apiClient.getAppBootstrap.mockRejectedValue(new Error('failed'))

    await store.bootstrap()

    expect(errors.items.length).toBe(1)
  })
})
