import { createPinia, setActivePinia } from 'pinia'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { createVuetify } from 'vuetify'

import ContractorPasswordDialog from '../components/ContractorPasswordDialog.vue'
import { useAppStore } from '../stores/app'
import { Quit } from '../../wailsjs/runtime/runtime.js'

vi.mock('../../wailsjs/runtime/runtime.js', () => ({
  Quit: vi.fn()
}))

const vuetify = createVuetify()

describe('ContractorPasswordDialog', () => {
  it('shows failure message and quits on close', async () => {
    // 認証失敗時にメッセージが表示され、閉じるで終了することを確認する。
    setActivePinia(createPinia())
    const app = useAppStore()
    app.verifyContractorPassword = vi.fn().mockResolvedValue(null)

    const wrapper = mount(ContractorPasswordDialog, {
      global: {
        plugins: [vuetify],
        stubs: {
          teleport: true,
          VDialog: { template: '<div><slot /></div>' }
        }
      }
    })

    await wrapper.find('[data-testid="verify"]').trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('認証に失敗しました。')

    await wrapper.find('[data-testid="close"]').trigger('click')
    expect(Quit).toHaveBeenCalled()
  })

  it('emits verified on success', async () => {
    // 認証成功時に verified が発火することを確認する。
    setActivePinia(createPinia())
    const app = useAppStore()
    app.verifyContractorPassword = vi.fn().mockResolvedValue({ mode: 'Contractor' })

    const wrapper = mount(ContractorPasswordDialog, {
      global: {
        plugins: [vuetify],
        stubs: {
          teleport: true,
          VDialog: { template: '<div><slot /></div>' }
        }
      }
    })

    await wrapper.find('[data-testid="verify"]').trigger('click')
    await wrapper.vm.$nextTick()

    expect(wrapper.emitted().verified).toBeTruthy()
  })
})
