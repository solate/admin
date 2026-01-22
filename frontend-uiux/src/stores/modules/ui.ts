// UI state store

import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'

const STORAGE_KEY = 'ui-state'

export const useUiStore = defineStore('ui', () => {
  // Read initial state from localStorage
  const storedState = localStorage.getItem(STORAGE_KEY)
  const initialState = storedState ? JSON.parse(storedState) : {}

  // State
  const sidebarOpen = ref(initialState.sidebarOpen ?? true)
  const darkMode = ref(initialState.darkMode ?? false)
  const mobileMenuOpen = ref(false)
  const locale = ref(initialState.locale ?? 'zh-CN')

  // Element Plus locale
  const elementLocale = ref(zhCn)

  // Initialize theme on DOM
  if (darkMode.value) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }

  // Initialize Element Plus locale
  if (locale.value) {
    elementLocale.value = locale.value === 'zh-CN' ? zhCn : en
  }

  // Watch state changes and persist to localStorage
  watch(
    () => ({
      sidebarOpen: sidebarOpen.value,
      darkMode: darkMode.value,
      locale: locale.value
    }),
    (state) => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
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
    elementLocale.value = newLocale === 'zh-CN' ? zhCn : en
    document.documentElement.lang = newLocale
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
