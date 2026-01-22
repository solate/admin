import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n, { getCurrentLocale } from './locales'
import directives from './directives'
import { setupElementPlus, zhCn, en } from './plugins/element'
import './styles/index.css'

const app = createApp(App)

// Setup Element Plus
setupElementPlus(app)

app.use(createPinia())
app.use(router)
app.use(i18n)
app.use(directives)

// Set initial document language
document.documentElement.lang = getCurrentLocale()

// Export locales for multi-language switching
app.config.globalProperties.$elementLocales = { zhCn, en }

app.mount('#app')
