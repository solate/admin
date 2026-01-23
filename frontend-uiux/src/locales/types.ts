/**
 * 多语言类型定义
 * 提供类型安全的翻译键检查
 */

// 从语言文件推断类型
import type zhCN from './zh-CN'

// 定义支持的语言类型
export type SupportedLocale = 'zh-CN' | 'en-US'

export const SUPPORTED_LOCALES: SupportedLocale[] = ['zh-CN', 'en-US']

// 翻译消息类型 - 递归提取所有可能的翻译键路径
export type MessageSchema = typeof zhCN

// 将嵌套对象类型转换为点分隔的字符串联合类型
// 例如: { common: { search: '搜索' } } => 'common.search'
type Paths<T> = T extends object
  ? {
      [K in keyof T]: K extends string
        ? T[K] extends string
          ? K
          : K | `${K}.${Paths<T[K]>}`
        : never
    }[keyof T]
  : never

// 所有翻译键的类型
export type TranslationKey = Paths<MessageSchema>

// 翻译模块类型（用于命名空间导入）
export type TranslationModule = keyof MessageSchema

// 带参数的翻译函数类型
export type TranslateFunction = {
  // 基础翻译
  (key: TranslationKey): string
  // 带命名参数的翻译
  (key: TranslationKey, params: Record<string, unknown>): string
  // 带列表参数的翻译
  (key: TranslationKey, ...list: unknown[]): string
}

// 语言配置类型
export interface LocaleConfig {
  code: SupportedLocale
  name: string
  flag: string
}

// 支持的语言列表
export const LOCALE_CONFIGS: Record<SupportedLocale, LocaleConfig> = {
  'zh-CN': { code: 'zh-CN', name: '简体中文', flag: '🇨🇳' },
  'en-US': { code: 'en-US', name: 'English', flag: '🇺🇸' }
}

// 获取语言配置
export function getLocaleConfig(locale: SupportedLocale): LocaleConfig {
  return LOCALE_CONFIGS[locale]
}

// 获取浏览器语言或默认语言
export function getDefaultLocale(): SupportedLocale {
  const stored = localStorage.getItem('locale')
  if (stored && SUPPORTED_LOCALES.includes(stored as SupportedLocale)) {
    return stored as SupportedLocale
  }

  const browserLang = navigator.language || (navigator as any).userLanguage
  if (browserLang?.startsWith('zh')) {
    return 'zh-CN'
  }
  return 'en-US'
}

// 判断是否为 RTL 语言
export function isRTL(locale: SupportedLocale): boolean {
  const rtlLocales: SupportedLocale[] = []
  return rtlLocales.includes(locale)
}

// 获取语言显示名称
export function getLocaleName(locale: SupportedLocale, displayLocale?: SupportedLocale): string {
  const config = getLocaleConfig(locale)
  return config.name
}
