// UI state store

import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { elementLocales } from '@/locales'

const STORAGE_KEY = 'ui-state'
// Locale storage key - 必须与 locales/index.ts 保持一致
const LOCALE_KEY = 'locale'

export const useUiStore = defineStore('ui', () => {
  // Read initial state from localStorage
  const storedState = localStorage.getItem(STORAGE_KEY)
  const initialState = storedState ? JSON.parse(storedState) : {}

  // 读取 locale 时使用独立的 key（与 locales/index.ts 保持一致）
  const storedLocale = localStorage.getItem(LOCALE_KEY)

  // State
  const sidebarOpen = ref(initialState.sidebarOpen ?? true)
  const darkMode = ref(initialState.darkMode ?? false)
  const mobileMenuOpen = ref(false)
  const locale = ref(storedLocale || initialState.locale || 'zh-CN')

  // Element Plus locale - 根据 locale 初始化
  const elementLocale = ref(elementLocales[locale.value as keyof typeof elementLocales])

  // Initialize theme on DOM
  if (darkMode.value) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }

  // Watch state changes and persist to localStorage
  watch(
    () => ({
      sidebarOpen: sidebarOpen.value,
      darkMode: darkMode.value,
      locale: locale.value
    }),
    (state) => {
      // 将 locale 单独存储到独立的 key（与 locales/index.ts 保持一致）
      if (state.locale) {
        localStorage.setItem(LOCALE_KEY, state.locale)
      }
      // 其他状态存储到 ui-state
      localStorage.setItem(STORAGE_KEY, JSON.stringify({
        sidebarOpen: state.sidebarOpen,
        darkMode: state.darkMode
      }))
    },
    { deep: true }
  )

  // Actions
  function toggleSidebar() {
    sidebarOpen.value = !sidebarOpen.value
  }

  function setSidebarOpen(open: boolean) {
    sidebarOpen.value = open
  }

  function toggleDarkMode() {
    darkMode.value = !darkMode.value
    if (darkMode.value) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }

  function setDarkMode(enabled: boolean) {
    darkMode.value = enabled
    if (enabled) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }

  function toggleMobileMenu() {
    mobileMenuOpen.value = !mobileMenuOpen.value
  }

  function closeMobileMenu() {
    mobileMenuOpen.value = false
  }

  function setLocale(newLocale: string) {
    locale.value = newLocale
    elementLocale.value = elementLocales[newLocale as keyof typeof elementLocales]
    document.documentElement.lang = newLocale
    // 存储 locale 到独立的 key
    localStorage.setItem(LOCALE_KEY, newLocale)
  }

  return {
    // State
    sidebarOpen,
    darkMode,
    mobileMenuOpen,
    locale,
    elementLocale,

    // Actions
    toggleSidebar,
    setSidebarOpen,
    toggleDarkMode,
    setDarkMode,
    toggleMobileMenu,
    closeMobileMenu,
    setLocale
  }
})
