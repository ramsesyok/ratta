import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vuetify, { transformAssetUrls } from 'vite-plugin-vuetify'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue({
      template: { transformAssetUrls }
    }),
    vuetify({ autoImport: true })
  ],
  test: {
    environment: 'jsdom',
    server: {
      deps: {
        inline: ['vuetify']
      }
    },
    setupFiles: ['src/__tests__/setup.js']
  }
})
