/**
 * 主题管理 Composable
 *
 * 提供主题相关的计算属性和操作方法。
 * 注意：主题的 DOM 更新由 plugins/theme.ts 统一处理，
 * 这里只负责提供便捷的访问和修改接口。
 *
 * @example
 * ```vue
 * <script setup>
 * const { primaryColor, setThemeColor, isDark, setThemeMode } = useTheme()
 * </script>
 * ```
 */

import { computed } from 'vue'
import { usePreferencesStore } from '@/stores/modules/preferences'
import type { ThemeMode, BorderRadius, ColorBlindMode } from '@/types/preferences'

export function useTheme() {
  const preferencesStore = usePreferencesStore()

  // ============ 计算属性 ============

  /** 当前主题色 */
  const primaryColor = computed(() => preferencesStore.appearance.primaryColor)

  /** 是否深色模式 */
  const isDark = computed(() => {
    const mode = preferencesStore.appearance.themeMode
    if (mode === 'dark') return true
    if (mode === 'light') return false
    // auto 模式：跟随系统
    return window.matchMedia('(prefers-color-scheme: dark)').matches
  })

  /** 当前主题模式 */
  const themeMode = computed(() => preferencesStore.appearance.themeMode)

  /** 圆角大小 */
  const borderRadius = computed(() => preferencesStore.appearance.borderRadius)

  /** 自定义圆角值 */
  const customBorderRadius = computed(() => preferencesStore.appearance.customBorderRadius)

  /** 色彩视觉模式 */
  const colorBlindMode = computed(() => preferencesStore.appearance.colorBlindMode)

  /** 高对比度模式 */
  const highContrast = computed(() => preferencesStore.appearance.highContrast)

  /** 灰色模式 */
  const grayMode = computed(() => preferencesStore.appearance.grayMode)

  /** 深色侧边栏 */
  const darkSidebar = computed(() => preferencesStore.appearance.darkSidebar)

  /** 深色顶栏 */
  const darkHeader = computed(() => preferencesStore.appearance.darkHeader)

  // ============ 方法 ============

  /**
   * 设置主题色
   * @param color 十六进制颜色值 (如 #2563eb)
   */
  function setThemeColor(color: string): void {
    preferencesStore.updateAppearance('primaryColor', color)
  }

  /**
   * 设置主题模式
   * @param mode 主题模式
   */
  function setThemeMode(mode: ThemeMode): void {
    preferencesStore.updateAppearance('themeMode', mode)
  }

  /**
   * 切换深色模式
   */
  function toggleDarkMode(): void {
    const newMode = isDark.value ? 'light' : 'dark'
    setThemeMode(newMode)
  }

  /**
   * 设置圆角大小
   * @param radius 圆角设置
   */
  function setBorderRadius(radius: BorderRadius): void {
    preferencesStore.updateAppearance('borderRadius', radius)
  }

  /**
   * 设置自定义圆角值
   * @param value CSS 圆角值 (如 0.5rem 或 8px)
   */
  function setCustomBorderRadius(value: string): void {
    preferencesStore.updateAppearance('customBorderRadius', value)
    // 如果当前不是 custom 模式，自动切换
    if (preferencesStore.appearance.borderRadius !== 'custom') {
      preferencesStore.updateAppearance('borderRadius', 'custom')
    }
  }

  /**
   * 设置色彩视觉模式
   * @param mode 色彩视觉模式
   */
  function setColorBlindMode(mode: ColorBlindMode): void {
    preferencesStore.updateAppearance('colorBlindMode', mode)
  }

  /**
   * 切换高对比度模式
   */
  function toggleHighContrast(): void {
    preferencesStore.updateAppearance('highContrast', !highContrast.value)
  }

  /**
   * 切换灰色模式
   */
  function toggleGrayMode(): void {
    preferencesStore.updateAppearance('grayMode', !grayMode.value)
  }

  /**
   * 切换深色侧边栏
   */
  function toggleDarkSidebar(): void {
    preferencesStore.updateAppearance('darkSidebar', !darkSidebar.value)
  }

  /**
   * 切换深色顶栏
   */
  function toggleDarkHeader(): void {
    preferencesStore.updateAppearance('darkHeader', !darkHeader.value)
  }

  // ============ 返回 ============

  return {
    // 计算属性
    primaryColor,
    isDark,
    themeMode,
    borderRadius,
    customBorderRadius,
    colorBlindMode,
    highContrast,
    grayMode,
    darkSidebar,
    darkHeader,

    // 方法
    setThemeColor,
    setThemeMode,
    toggleDarkMode,
    setBorderRadius,
    setCustomBorderRadius,
    setColorBlindMode,
    toggleHighContrast,
    toggleGrayMode,
    toggleDarkSidebar,
    toggleDarkHeader
  }
}
