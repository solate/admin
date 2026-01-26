/**
 * 主题管理 Composable
 * 提供主题色切换、深色模式等主题相关功能
 *
 * @example
 * ```vue
 * <script setup>
 * const { primaryColor, setThemeColor, isDark } = useTheme()
 * </script>
 * ```
 */

import { computed, watch, onMounted } from 'vue'
import { usePreferencesStore } from '@/stores/modules/preferences'

export function useTheme() {
  const preferencesStore = usePreferencesStore()

  // ============ 状态 ============

  // 当前主题色
  const primaryColor = computed(() => preferencesStore.appearance.primaryColor)

  // 是否深色模式
  const isDark = computed(() => {
    const mode = preferencesStore.appearance.themeMode
    if (mode === 'dark') return true
    if (mode === 'light') return false
    // auto 模式：跟随系统
    return window.matchMedia('(prefers-color-scheme: dark)').matches
  })

  // 当前主题模式
  const themeMode = computed(() => preferencesStore.appearance.themeMode)

  // 圆角大小
  const borderRadius = computed(() => preferencesStore.appearance.borderRadius)

  // 色盲模式
  const colorBlindMode = computed(() => preferencesStore.appearance.colorBlindMode)

  // 高对比度模式
  const highContrast = computed(() => preferencesStore.appearance.highContrast)

  // ============ 方法 ============

  /**
   * 将 HEX 颜色转换为 RGB 空格分隔格式
   * @example '#2563eb' -> '37 99 235'
   */
  const hexToRgb = (hex: string): string => {
    const cleanHex = hex.replace('#', '')
    const r = parseInt(cleanHex.substring(0, 2), 16)
    const g = parseInt(cleanHex.substring(2, 4), 16)
    const b = parseInt(cleanHex.substring(4, 6), 16)
    return `${r} ${g} ${b}`
  }

  /**
   * 将 HEX 颜色转换为 HSL 空格分隔格式
   * @example '#2563eb' -> '221 83% 53%'
   */
  const hexToHsl = (hex: string): string => {
    const cleanHex = hex.replace('#', '')
    const r = parseInt(cleanHex.substring(0, 2), 16) / 255
    const g = parseInt(cleanHex.substring(2, 4), 16) / 255
    const b = parseInt(cleanHex.substring(4, 6), 16) / 255

    const max = Math.max(r, g, b)
    const min = Math.min(r, g, b)
    let h = 0
    let s = 0
    const l = (max + min) / 2

    if (max !== min) {
      const d = max - min
      s = l > 0.5 ? d / (2 - max - min) : d / (max + min)

      switch (max) {
        case r:
          h = ((g - b) / d + (g < b ? 6 : 0)) / 6
          break
        case g:
          h = ((b - r) / d + 2) / 6
          break
        case b:
          h = ((r - g) / d + 4) / 6
          break
      }
    }

    return `${Math.round(h * 360)} ${Math.round(s * 100)}% ${Math.round(l * 100)}%`
  }

  /**
   * 应用主题色到 DOM
   */
  const applyThemeColor = (color: string) => {
    const root = document.documentElement

    // 同时设置 RGB 和 HSL 格式，以支持不同的 Tailwind 配置
    const rgb = hexToRgb(color)
    const hsl = hexToHsl(color)

    root.style.setProperty('--color-primary', rgb)
    root.style.setProperty('--color-primary-hsl', hsl)
  }

  /**
   * 设置主题色
   */
  const setThemeColor = (color: string) => {
    preferencesStore.updateAppearance('primaryColor', color)
    applyThemeColor(color)
  }

  /**
   * 设置主题模式
   */
  const setThemeMode = (mode: 'light' | 'dark' | 'auto') => {
    preferencesStore.updateAppearance('themeMode', mode)
  }

  /**
   * 切换深色模式
   */
  const toggleDarkMode = () => {
    const current = isDark.value
    setThemeMode(current ? 'light' : 'dark')
  }

  /**
   * 设置圆角大小
   */
  const setBorderRadius = (radius: 'none' | 'small' | 'medium' | 'large') => {
    preferencesStore.updateAppearance('borderRadius', radius)
  }

  /**
   * 设置色盲模式
   */
  const setColorBlindMode = (mode: 'none' | 'protanopia' | 'deuteranopia' | 'tritanopia') => {
    preferencesStore.updateAppearance('colorBlindMode', mode)
  }

  /**
   * 切换高对比度模式
   */
  const toggleHighContrast = () => {
    preferencesStore.updateAppearance('highContrast', !highContrast.value)
  }

  /**
   * 初始化主题（应用所有设置到 DOM）
   */
  const initializeTheme = () => {
    const root = document.documentElement
    const { appearance } = preferencesStore

    // 应用主题色
    applyThemeColor(appearance.primaryColor)

    // 应用圆角
    const borderRadiusMap = {
      none: '0',
      small: '0.125rem',
      medium: '0.5rem',
      large: '1rem'
    }
    root.style.setProperty('--border-radius', borderRadiusMap[appearance.borderRadius])

    // 应用色盲模式
    if (appearance.colorBlindMode !== 'none') {
      root.setAttribute('data-color-blind', appearance.colorBlindMode)
    } else {
      root.removeAttribute('data-color-blind')
    }

    // 应用高对比度模式
    if (appearance.highContrast) {
      root.setAttribute('data-high-contrast', 'true')
    } else {
      root.removeAttribute('data-high-contrast')
    }

    // 应用主题模式（深色/浅色）
    if (appearance.themeMode === 'dark') {
      root.classList.add('dark')
    } else if (appearance.themeMode === 'light') {
      root.classList.remove('dark')
    } else {
      // auto 模式
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      if (prefersDark) {
        root.classList.add('dark')
      } else {
        root.classList.remove('dark')
      }
    }
  }

  // ============ 生命周期 ============

  // 组件挂载时初始化主题
  onMounted(() => {
    initializeTheme()
  })

  // 监听主题色变化
  watch(
    () => preferencesStore.appearance.primaryColor,
    (newColor) => {
      applyThemeColor(newColor)
    }
  )

  // 监听圆角变化
  watch(
    () => preferencesStore.appearance.borderRadius,
    (newRadius) => {
      const borderRadiusMap = {
        none: '0',
        small: '0.125rem',
        medium: '0.5rem',
        large: '1rem'
      }
      document.documentElement.style.setProperty('--border-radius', borderRadiusMap[newRadius])
    }
  )

  // 监听色盲模式变化
  watch(
    () => preferencesStore.appearance.colorBlindMode,
    (newMode) => {
      const root = document.documentElement
      if (newMode !== 'none') {
        root.setAttribute('data-color-blind', newMode)
      } else {
        root.removeAttribute('data-color-blind')
      }
    }
  )

  // 监听高对比度模式变化
  watch(
    () => preferencesStore.appearance.highContrast,
    (newValue) => {
      const root = document.documentElement
      if (newValue) {
        root.setAttribute('data-high-contrast', 'true')
      } else {
        root.removeAttribute('data-high-contrast')
      }
    }
  )

  // ============ 返回 ============

  return {
    // 状态
    primaryColor,
    isDark,
    themeMode,
    borderRadius,
    colorBlindMode,
    highContrast,

    // 方法
    setThemeColor,
    setThemeMode,
    toggleDarkMode,
    setBorderRadius,
    setColorBlindMode,
    toggleHighContrast,
    initializeTheme
  }
}
