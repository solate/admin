import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN.js'
import enUS from './en-US.js'

// 支持的语言列表
export const SUPPORT_LOCALES = ['zh-CN', 'en-US']

// 获取浏览器语言或默认语言
function getDefaultLocale() {
  const saved = localStorage.getItem('locale')
  if (saved && SUPPORT_LOCALES.includes(saved)) {
    return saved
  }

  const browserLang = navigator.language || navigator.userLanguage
  if (browserLang.startsWith('zh')) {
    return 'zh-CN'
  }
  return 'en-US'
}

const i18n = createI18n({
  legacy: false, // 使用 Composition API 模式
  locale: getDefaultLocale(),
  fallbackLocale: {
    'zh': ['zh-CN'],
    'zh-CN': ['en-US'],
    'en': ['en-US'],
    'en-US': 'en-US',
    default: 'en-US'
  },
  messages: {
    'zh-CN': zhCN,
    'zh': zhCN,
    'en-US': enUS,
    'en': enUS
  }
})

// 设置语言的方法
export function setLocale(locale) {
  if (SUPPORT_LOCALES.includes(locale)) {
    i18n.global.locale.value = locale
    localStorage.setItem('locale', locale)
    document.documentElement.lang = locale
  }
}

// 获取当前语言
export function getCurrentLocale() {
  return i18n.global.locale.value
}

export default i18n
