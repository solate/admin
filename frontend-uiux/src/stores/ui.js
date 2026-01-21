import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

const STORAGE_KEY = 'ui-state'

export const useUiStore = defineStore('ui', () => {
  // 从 localStorage 读取初始状态
  const storedState = localStorage.getItem(STORAGE_KEY)
  const initialState = storedState ? JSON.parse(storedState) : {}

  const sidebarOpen = ref(initialState.sidebarOpen ?? true)
  const darkMode = ref(initialState.darkMode ?? false)
  const mobileMenuOpen = ref(false)

  // 初始化时同步主题状态到 DOM
  if (darkMode.value) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }

  // 监听状态变化并持久化到 localStorage
  watch(
    () => ({ sidebarOpen: sidebarOpen.value, darkMode: darkMode.value }),
    (state) => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
    },
    { deep: true }
  )

  function toggleSidebar() {
    sidebarOpen.value = !sidebarOpen.value
  }

  function toggleDarkMode() {
    darkMode.value = !darkMode.value
    if (darkMode.value) {
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

  return {
    sidebarOpen,
    darkMode,
    mobileMenuOpen,
    toggleSidebar,
    toggleDarkMode,
    toggleMobileMenu,
    closeMobileMenu
  }
})
