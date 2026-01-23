/**
 * 类型安全的 i18n composable
 * 提供带类型检查的翻译函数
 */

import { computed } from 'vue'
import { useI18n as useVueI18n } from 'vue-i18n'
import type { TranslateFunction, TranslationKey, SupportedLocale } from './types'

/**
 * 类型安全的翻译函数
 * 自动补全翻译键，编译时检查键是否存在
 */
export function useI18n() {
  const i18n = useVueI18n()

  // 类型安全的 t 函数
  const t: TranslateFunction = (key: any, ...args: any[]) => {
    return i18n.t(key, ...args) as string
  }

  // 获取当前语言
  const locale = computed(() => i18n.locale.value as SupportedLocale)

  // 设置语言
  const setLocale = (newLocale: SupportedLocale) => {
    i18n.locale.value = newLocale
  }

  // 复数形式
  const tc = (key: TranslationKey, choice?: number) => {
    return i18n.tc(key, choice)
  }

  // 日期格式化
  const d = (date: Date | number, format?: string) => {
    return i18n.d(date, format)
  }

  // 数字格式化
  const n = (number: number, format?: string) => {
    return i18n.n(number, format)
  }

  // 检查键是否存在
  const te = (key: TranslationKey): boolean => {
    return i18n.te(key)
  }

  // 可用语言列表
  const availableLocales: SupportedLocale[] = ['zh-CN', 'en-US']

  return {
    // 翻译函数
    t,
    tc,
    d,
    n,
    te,

    // 语言相关
    locale,
    setLocale,
    availableLocales,

    // 原始 i18n 实例（如需要更复杂的功能）
    i18n
  }
}

// 导出类型
export type UseI18nReturn = ReturnType<typeof useI18n>
