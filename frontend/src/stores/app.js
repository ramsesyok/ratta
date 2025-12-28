import { defineStore } from 'pinia'

export const useAppStore = defineStore('app', {
  state: () => ({
    mode: 'vendor'
  }),
  actions: {
    setMode(mode) {
      this.mode = mode
    }
  }
})
