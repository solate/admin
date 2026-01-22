import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n, { getCurrentLocale } from './locales'
import directives from './directives'
import { setupElementPlus, zhCn, en } from './plugins/element'
import './assets/main.css'

const app = createApp(App)

// 设置 Element Plus
setupElementPlus(app)

app.use(createPinia())
app.use(router)
app.use(i18n)
app.use(directives)

// 设置初始文档语言
document.documentElement.lang = getCurrentLocale()

// 导出 locale 供多语言切换使用
app.config.globalProperties.$elementLocales = { zhCn, en }

app.mount('#app')
