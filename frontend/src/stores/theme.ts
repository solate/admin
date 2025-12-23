import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export type Theme = 'light' | 'dark'

export const useThemeStore = defineStore('theme', () => {
  // 从localStorage读取主题设置，默认为light
  const savedTheme = (localStorage.getItem('theme') || 'light') as Theme
  const theme = ref<Theme>(savedTheme)

  // 切换主题
  const toggleTheme = () => {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
  }

  // 设置指定主题
  const setTheme = (newTheme: Theme) => {
    theme.value = newTheme
  }

  // 监听主题变化，更新DOM和localStorage
  watch(
    theme,
    (newTheme) => {
      document.documentElement.setAttribute('data-theme', newTheme)
      localStorage.setItem('theme', newTheme)
    },
    { immediate: true }
  )

  // 初始化主题
  const initTheme = () => {
    document.documentElement.setAttribute('data-theme', theme.value)
  }

  return {
    theme,
    toggleTheme,
    setTheme,
    initTheme
  }
})