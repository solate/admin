import { createI18n } from 'vue-i18n'
import zh from './zh-CN'
import en from './en-US'

export type LocaleType = 'zh-CN' | 'en-US'

export const locales: { key: LocaleType; label: string }[] = [
  { key: 'zh-CN', label: '简体中文' },
  { key: 'en-US', label: 'English' }
]

export const defaultLocale: LocaleType = 'zh-CN'

const i18n = createI18n({
  legacy: false,
  locale: defaultLocale,
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zh,
    'en-US': en
  }
})

export default i18n
