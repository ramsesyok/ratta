import { createPinia, setActivePinia } from 'pinia'
import { describe, expect, it } from 'vitest'

import { ApiError } from '../utils/apiClient'
import { useErrorsStore } from '../stores/errors'

describe('errors store', () => {
  it('captures generic errors', () => {
    // 例外が UiErrorItem に変換されることを確認する。
    setActivePinia(createPinia())
    const store = useErrorsStore()

    store.capture(new Error('boom'), { source: 'app', action: 'test' })

    expect(store.items.length).toBe(1)
    expect(store.items[0].user_message).toBe('boom')
  })

  it('captures api errors', () => {
    // ApiError が API 情報付きで登録されることを確認する。
    setActivePinia(createPinia())
    const store = useErrorsStore()

    const apiError = new ApiError('ng', { error_code: 'E_VALIDATION' })
    store.captureApiError(apiError, { source: 'backend', action: 'call' })

    expect(store.items[0].api.error_code).toBe('E_VALIDATION')
  })

  it('marks items as read', () => {
    // 既読操作が反映されることを確認する。
    setActivePinia(createPinia())
    const store = useErrorsStore()

    const item = store.capture(new Error('boom'), { source: 'app', action: 'test' })
    store.markRead(item.id)

    expect(store.items[0].is_read).toBe(true)
  })
})
