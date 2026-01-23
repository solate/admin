/**
 * Vue I18n 配置
 * 参考最佳实践：Vitesse, Element Plus, Nuxt i18n
 */

import { createI18n } from 'vue-i18n'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import type { App } from 'vue'
import type { I18n, I18nOptions } from 'vue-i18n'
import type { SupportedLocale, MessageSchema } from './types'

// 导入语言包
import zhCNMessages from './zh-CN'
import enUSMessages from './en-US'

// LocalStorage key
const LOCALE_STORAGE_KEY = 'locale'

// Element Plus locale 映射
export const elementLocales = {
  'zh-CN': zhCn,
  'en-US': en
} as const

// 支持的语言
export const SUPPORT_LOCALES = ['zh-CN', 'en-US'] as const
export type SupportedLocale = (typeof SUPPORT_LOCALES)[number]

/**
 * 获取浏览器语言或默认语言
 */
function getInitialLocale(): SupportedLocale {
  // 1. 检查 localStorage
  const stored = localStorage.getItem(LOCALE_STORAGE_KEY)
  if (stored && SUPPORT_LOCALES.includes(stored as SupportedLocale)) {
    return stored as SupportedLocale
  }

  // 2. 检查浏览器语言
  const browserLang = navigator.language || (navigator as any).userLanguage
  if (browserLang?.startsWith('zh')) {
    return 'zh-CN'
  }

  // 3. 默认英文
  return 'en-US'
}

/**
 * 创建 i18n 实例
 */
export function createI18nInstance(options?: Partial<I18nOptions>): I18n {
  const initialLocale = getInitialLocale()

  const i18n = createI18n({
    // 使用 Composition API 模式
    legacy: false,

    // 当前语言
    locale: initialLocale,

    // 回退语言
    fallbackLocale: {
      'zh-CN': ['en-US'],
      'en-US': 'en-US',
      default: 'en-US'
    },

    // 翻译消息
    messages: {
      'zh-CN': zhCNMessages,
      'en-US': enUSMessages
    },

    // 缺失翻译处理
    missing: (locale, key) => {
      if (import.meta.env.DEV) {
        console.warn(`[i18n] Missing translation: "${key}" for locale: "${locale}"`)
      }
      // 返回 key 本身作为兜底
      return key
    },

    // 翻译失败时的回退
    fallbackWarn: import.meta.env.DEV,
    silentTranslationWarn: !import.meta.env.DEV,

    // HTML 消息支持
    escapeParameterHtml: false,

    // 保持 HTML 标签
    preserveDirectiveContent: true,

    // 全局注入 $t
    globalInjection: true,

    // 自定义选项
    ...options
  }) as I18n<MessageSchema>

  return i18n
}

// 创建全局 i18n 实例
const i18n = createI18nInstance()

/**
 * 设置语言
 * 同步更新 Vue I18n、Element Plus locale 和 localStorage
 */
export async function setLocale(locale: SupportedLocale) {
  if (!SUPPORT_LOCALES.includes(locale)) {
    console.warn(`[i18n] Unsupported locale: ${locale}`)
    return
  }

  // 1. 更新 Vue I18n
  i18n.global.locale.value = locale

  // 2. 更新 localStorage
  localStorage.setItem(LOCALE_STORAGE_KEY, locale)

  // 3. 更新 document lang
  document.documentElement.lang = locale

  // 4. 同步更新 uiStore 的 elementLocale (延迟导入避免循环依赖)
  try {
    const { useUiStore } = await import('@/stores/modules/ui')
    const uiStore = useUiStore()
    uiStore.setLocale(locale)
  } catch (error) {
    console.warn('[i18n] Failed to sync with uiStore:', error)
  }
}

/**
 * 获取当前语言
 */
export function getCurrentLocale(): SupportedLocale {
  return i18n.global.locale.value as SupportedLocale
}

/**
 * 切换到下一个语言
 */
export function toggleLocale() {
  const currentIndex = SUPPORT_LOCALES.indexOf(getCurrentLocale())
  const nextIndex = (currentIndex + 1) % SUPPORT_LOCALES.length
  setLocale(SUPPORT_LOCALES[nextIndex])
}

/**
 * 检查是否为 RTL 语言
 */
export function isRTL(locale?: SupportedLocale): boolean {
  const targetLocale = locale || getCurrentLocale()
  // 目前没有 RTL 语言，预留接口
  return false
}

/**
 * 按需加载语言包（懒加载）
 * 用于支持未来更多语言
 */
export async function loadLocaleMessages(locale: SupportedLocale): Promise<void> {
  // 如果已经加载，直接返回
  if (i18n.global.availableLocales.includes(locale)) {
    return
  }

  // 动态导入语言包
  try {
    // 动态导入用于懒加载语言包
    const messages = await import(/* @vite-ignore */ `./${locale}.ts`)
    i18n.global.setLocaleMessage(locale, messages.default)
  } catch (error) {
    console.error(`[i18n] Failed to load locale: ${locale}`, error)
    throw error
  }
}

/**
 * Vue 插件安装
 */
export function installI18n(app: App) {
  app.use(i18n)

  // 设置初始 document lang (仅在浏览器环境)
  if (typeof window !== 'undefined') {
    document.documentElement.lang = getCurrentLocale()
  }
}

export default i18n
