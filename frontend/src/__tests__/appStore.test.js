import { createPinia, setActivePinia } from 'pinia'
import { describe, expect, it } from 'vitest'
import { useAppStore } from '../stores/app'

describe('app store', () => {
  it('tracks the current mode', () => {
    setActivePinia(createPinia())
    const store = useAppStore()

    expect(store.mode).toBe('vendor')
    store.setMode('contractor')
    expect(store.mode).toBe('contractor')
  })
})
