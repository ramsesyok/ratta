import { createPinia, setActivePinia } from 'pinia'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { createVuetify } from 'vuetify'

import ProjectSelectDialog from '../components/ProjectSelectDialog.vue'
import { useAppStore } from '../stores/app'

vi.mock('../utils/apiClient', () => ({
  ApiError: class ApiError extends Error {},
  getAppBootstrap: vi.fn(),
  validateProjectRoot: vi.fn(),
  saveLastProjectRoot: vi.fn(),
  createProjectRoot: vi.fn(),
  detectMode: vi.fn(),
  verifyContractorPassword: vi.fn()
}))

vi.mock('../../wailsjs/runtime/runtime.js', () => ({
  Quit: vi.fn()
}))

const vuetify = createVuetify()

describe('ProjectSelectDialog', () => {
  it('shows validation error when invalid path', async () => {
    // 無効なパスが入力された場合にエラーメッセージが表示されることを確認する。
    setActivePinia(createPinia())
    const app = useAppStore()
    app.bootstrapLoaded = true
    app.lastProjectRootPath = ''
    app.selectProjectRoot = vi.fn().mockResolvedValue({
      is_valid: false,
      message: 'Path does not exist.'
    })

    const wrapper = mount(ProjectSelectDialog, {
      global: {
        plugins: [vuetify],
        stubs: {
          teleport: true,
          VDialog: { template: '<div><slot /></div>' }
        }
      }
    })

    await wrapper.vm.$nextTick()
    await wrapper.find('[data-testid="validate"]').trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('Path does not exist.')
  })

  it('emits selected on success', async () => {
    // 正常に選択された場合に selected が発火することを確認する。
    setActivePinia(createPinia())
    const app = useAppStore()
    app.bootstrapLoaded = true
    app.lastProjectRootPath = 'C:/proj'
    app.selectProjectRoot = vi.fn().mockResolvedValue({
      is_valid: true,
      normalized_path: 'C:/proj'
    })

    const wrapper = mount(ProjectSelectDialog, {
      global: {
        plugins: [vuetify],
        stubs: {
          teleport: true,
          VDialog: { template: '<div><slot /></div>' }
        }
      }
    })

    await wrapper.vm.$nextTick()
    await wrapper.find('[data-testid="validate"]').trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.emitted().selected).toBeTruthy()
  })
})
