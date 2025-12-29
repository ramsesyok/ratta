import { describe, expect, it, vi } from 'vitest'

import { ApiError, getAppBootstrap, unwrapResponse } from '../utils/apiClient'

vi.mock('../../wailsjs/go/main/App.js', () => ({
  GetAppBootstrap: vi.fn()
}))

import * as App from '../../wailsjs/go/main/App.js'

describe('api client', () => {
  it('unwraps ok response', () => {
    // ok=true の場合に data を返すことを確認する。
    const result = unwrapResponse({ ok: true, data: { value: 1 } }, 'Test')
    expect(result).toEqual({ value: 1 })
  })

  it('throws ApiError on error response', () => {
    // ok=false の場合に ApiError が送出されることを確認する。
    expect(() =>
      unwrapResponse({ ok: false, error: { error_code: 'E_TEST', message: 'ng' } }, 'Test')
    ).toThrow(ApiError)
  })

  it('wraps GetAppBootstrap', async () => {
    // Wails バインディングを呼び出して結果を返すことを確認する。
    App.GetAppBootstrap.mockResolvedValue({ ok: true, data: { has_config: true } })

    const result = await getAppBootstrap()

    expect(result).toEqual({ has_config: true })
    expect(App.GetAppBootstrap).toHaveBeenCalled()
  })
})
