import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createVuetify } from 'vuetify'
import 'vuetify/styles'
import App from './App.vue'
import './style.css'
import '@mdi/font/css/materialdesignicons.css'

const app = createApp(App)
app.use(createPinia())
app.use(createVuetify({
    theme: {
        defaultTheme: 'dark',
    },
}))
app.mount('#app')
