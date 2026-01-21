import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import type { LocaleType } from '@/locales'

const LOCALE_STORAGE_KEY = 'locale'
const DEFAULT_LOCALE: LocaleType = 'zh-CN'

export const useLocaleStore = defineStore('locale', () => {
  // 从 localStorage 读取保存的语言设置
  const storedLocale = localStorage.getItem(LOCALE_STORAGE_KEY) as LocaleType
  const currentLocale = ref<LocaleType>(
    storedLocale && ['zh-CN', 'en-US'].includes(storedLocale) ? storedLocale : DEFAULT_LOCALE
  )

  // 切换语言
  const setLocale = (locale: LocaleType) => {
    currentLocale.value = locale
    localStorage.setItem(LOCALE_STORAGE_KEY, locale)
  }

  // 监听语言变化,可以在这里添加额外的逻辑,比如重新加载某些数据
  watch(currentLocale, (newLocale) => {
    // 设置 HTML lang 属性
    document.documentElement.lang = newLocale
  })

  // 初始化时设置 lang 属性
  document.documentElement.lang = currentLocale.value

  return {
    currentLocale,
    setLocale
  }
})
