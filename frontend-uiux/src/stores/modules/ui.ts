/**
 * UI 状态 Store
 *
 * 职责：
 * 1. 管理侧边栏、菜单等 UI 状态
 * 2. 管理 darkMode 状态（由 preferencesStore 同步）
 * 3. 管理语言设置
 *
 * 注意：DOM 操作由 plugins/theme.ts 统一处理
 */

import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { elementLocales } from '@/locales'

// 存储键名
const UI_STATE_KEY = 'ui-state'
const LOCALE_KEY = 'locale'

export const useUiStore = defineStore('ui', () => {
  // ============ 初始化状态 ============

  const storedState = localStorage.getItem(UI_STATE_KEY)
  const initialState = storedState ? JSON.parse(storedState) : {}

  const storedLocale = localStorage.getItem(LOCALE_KEY)

  // ============ State ============

  /** 侧边栏是否展开 */
  const sidebarOpen = ref(initialState.sidebarOpen ?? true)

  /** 是否深色模式（由 preferencesStore 同步） */
  const darkMode = ref(initialState.darkMode ?? false)

  /** 移动端菜单是否展开 */
  const mobileMenuOpen = ref(false)

  /** 当前语言 */
  const locale = ref(storedLocale || initialState.locale || 'zh-CN')

  /** Element Plus 语言包 */
  const elementLocale = ref(elementLocales[locale.value as keyof typeof elementLocales])

  // ============ 持久化 ============

  watch(
    [sidebarOpen, darkMode],
    () => {
      localStorage.setItem(UI_STATE_KEY, JSON.stringify({
        sidebarOpen: sidebarOpen.value,
        darkMode: darkMode.value
      }))
    }
  )

  // ============ Actions ============

  /**
   * 切换侧边栏展开/折叠
   */
  function toggleSidebar(): void {
    sidebarOpen.value = !sidebarOpen.value
  }

  /**
   * 设置侧边栏展开状态
   */
  function setSidebarOpen(open: boolean): void {
    sidebarOpen.value = open
  }

  /**
   * 切换深色模式
   */
  function toggleDarkMode(): void {
    darkMode.value = !darkMode.value
    if (darkMode.value) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }

  /**
   * 设置深色模式
   */
  function setDarkMode(enabled: boolean): void {
    darkMode.value = enabled
    if (enabled) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }

  /**
   * 切换移动端菜单
   */
  function toggleMobileMenu(): void {
    mobileMenuOpen.value = !mobileMenuOpen.value
  }

  /**
   * 关闭移动端菜单
   */
  function closeMobileMenu(): void {
    mobileMenuOpen.value = false
  }

  /**
   * 设置语言
   */
  function setLocale(newLocale: string): void {
    locale.value = newLocale
    elementLocale.value = elementLocales[newLocale as keyof typeof elementLocales]
    document.documentElement.lang = newLocale
    localStorage.setItem(LOCALE_KEY, newLocale)
  }

  // ============ Return ============

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
