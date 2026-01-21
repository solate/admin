import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n, { getCurrentLocale } from './locales'
import directives from './directives'
import './assets/main.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(i18n)
app.use(directives)

// Set initial document language
document.documentElement.lang = getCurrentLocale()

app.mount('#app')
