import { defineStore } from 'pinia'
import { computed } from 'vue'
import { useDark } from '@vueuse/core'

export type Theme = 'light' | 'dark'

export const useThemeStore = defineStore('theme', () => {
  // 使用 VueUse 的 useDark - 自动处理 localStorage 和系统偏好
  // Element Plus 官方暗黑模式使用 html.dark class
  const isDark = useDark({
    storageKey: 'theme',
    valueDark: 'dark',
    valueLight: '',
    disableTransition: false,
  })

  // 获取当前主题值
  const theme = computed<Theme>(() => isDark.value ? 'dark' : 'light')

  // 切换主题 - 直接修改 isDark 的值
  const toggleTheme = () => {
    isDark.value = !isDark.value
  }

  // 设置指定主题
  const setTheme = (newTheme: Theme) => {
    isDark.value = newTheme === 'dark'
  }

  // 初始化主题 - 确保 useDark 正确应用了初始状态
  const initTheme = () => {
    const htmlElement = document.documentElement
    const currentValue = localStorage.getItem('theme') || 'light'

    // 确保 HTML class 与 localStorage 同步
    if (currentValue === 'dark' && !htmlElement.classList.contains('dark')) {
      htmlElement.classList.add('dark')
    } else if (currentValue !== 'dark' && htmlElement.classList.contains('dark')) {
      htmlElement.classList.remove('dark')
    }
  }

  return {
    theme,
    isDark,
    toggleTheme,
    setTheme,
    initTheme
  }
})