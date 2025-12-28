import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createVuetify } from 'vuetify'
import 'vuetify/styles'
import App from './App.vue'
import './style.css'

const app = createApp(App)
app.use(createPinia())
app.use(createVuetify())
app.mount('#app')
