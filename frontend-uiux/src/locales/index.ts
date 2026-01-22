import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN'
import enUS from './en-US'

// Supported locales
export const SUPPORT_LOCALES = ['zh-CN', 'en-US'] as const
export type SupportedLocale = (typeof SUPPORT_LOCALES)[number]

// Get browser language or default language
function getDefaultLocale(): SupportedLocale {
  const saved = localStorage.getItem('locale')
  if (saved && SUPPORT_LOCALES.includes(saved as SupportedLocale)) {
    return saved as SupportedLocale
  }

  const browserLang = navigator.language || navigator.userLanguage
  if (browserLang.startsWith('zh')) {
    return 'zh-CN'
  }
  return 'en-US'
}

const i18n = createI18n({
  legacy: false, // Use Composition API mode
  locale: getDefaultLocale(),
  fallbackLocale: {
    zh: ['zh-CN'],
    'zh-CN': ['en-US'],
    en: ['en-US'],
    'en-US': 'en-US',
    default: 'en-US'
  },
  messages: {
    'zh-CN': zhCN,
    zh: zhCN,
    'en-US': enUS,
    en: enUS
  }
})

// Set locale method
export function setLocale(locale: SupportedLocale) {
  if (SUPPORT_LOCALES.includes(locale)) {
    i18n.global.locale.value = locale
    localStorage.setItem('locale', locale)
    document.documentElement.lang = locale
  }
}

// Get current locale
export function getCurrentLocale(): SupportedLocale {
  return i18n.global.locale.value as SupportedLocale
}

export default i18n
