/**
 * 用户偏好设置 Store
 */

import { defineStore } from 'pinia'
import { ref, watch, computed } from 'vue'
import type { UserPreferences, BorderRadius } from '@/types/preferences'
import { DEFAULT_PREFERENCES, PREFERENCES_STORAGE_KEY } from '@/types/preferences'
import { useUiStore } from './ui'

/**
 * 旧的 borderRadius 值到新值的映射（数据迁移）
 */
const BORDER_RADIUS_MIGRATION: Record<string, BorderRadius> = {
  '0': 'none',
  '0.25': 'small',
  '0.5': 'medium',
  '0.75': 'medium',
  '1': 'large'
}

export const usePreferencesStore = defineStore('preferences', () => {
  // ============ State ============

  // 从 localStorage 读取初始状态
  const storedPreferences = localStorage.getItem(PREFERENCES_STORAGE_KEY)
  let initialState: UserPreferences

  if (storedPreferences) {
    try {
      const parsed = JSON.parse(storedPreferences)
      initialState = { ...DEFAULT_PREFERENCES, ...parsed }

      // 数据迁移：旧的 borderRadius 值转换为新值
      if (parsed.appearance?.borderRadius && BORDER_RADIUS_MIGRATION[parsed.appearance.borderRadius]) {
        initialState.appearance.borderRadius = BORDER_RADIUS_MIGRATION[parsed.appearance.borderRadius]
      }
    } catch {
      initialState = { ...DEFAULT_PREFERENCES }
    }
  } else {
    initialState = { ...DEFAULT_PREFERENCES }
  }

  const preferences = ref<UserPreferences>(initialState)
  const isInitialized = ref(false)

  // ============ Getters ============

  const appearance = computed(() => preferences.value.appearance)
  const layout = computed(() => preferences.value.layout)
  const shortcuts = computed(() => preferences.value.shortcuts)
  const general = computed(() => preferences.value.general)
  const widgets = computed(() => preferences.value.widgets)
  const copyright = computed(() => preferences.value.copyright)

  // ============ Actions ============

  /**
   * 初始化偏好设置
   * 同步到 useUiStore 和 DOM
   */
  function initialize() {
    if (isInitialized.value) return

    const uiStore = useUiStore()

    // 同步主题模式
    syncThemeMode(appearance.value.themeMode)

    // 同步语言
    uiStore.setLocale(general.value.language)

    // 应用外观设置到 DOM
    applyAppearanceSettings()

    // 监听主题模式变化
    watch(
      () => appearance.value.themeMode,
      (newMode) => syncThemeMode(newMode)
    )

    // 监听语言变化
    watch(
      () => general.value.language,
      (newLanguage) => uiStore.setLocale(newLanguage)
    )

    // 监听外观设置变化，应用到 DOM
    watch(
      () => appearance.value,
      () => applyAppearanceSettings(),
      { deep: true }
    )

    // 监听偏好设置变化，持久化到 localStorage
    watch(
      preferences,
      (newPreferences) => {
        localStorage.setItem(PREFERENCES_STORAGE_KEY, JSON.stringify(newPreferences))
      },
      { deep: true }
    )

    isInitialized.value = true
  }

  /**
   * 应用外观设置到 DOM
   */
  function applyAppearanceSettings() {
    const root = document.documentElement

    // 应用色盲模式
    if (appearance.value.colorBlindMode !== 'none') {
      root.setAttribute('data-color-blind', appearance.value.colorBlindMode)
    } else {
      root.removeAttribute('data-color-blind')
    }

    // 应用高对比度模式
    if (appearance.value.highContrast) {
      root.setAttribute('data-high-contrast', 'true')
    } else {
      root.removeAttribute('data-high-contrast')
    }

    // 应用灰色模式
    if (appearance.value.grayMode) {
      root.setAttribute('data-gray-mode', 'true')
    } else {
      root.removeAttribute('data-gray-mode')
    }

    // 应用深色侧边栏
    if (appearance.value.darkSidebar) {
      root.setAttribute('data-dark-sidebar', 'true')
    } else {
      root.removeAttribute('data-dark-sidebar')
    }

    // 应用深色顶栏
    if (appearance.value.darkHeader) {
      root.setAttribute('data-dark-header', 'true')
    } else {
      root.removeAttribute('data-dark-header')
    }

    // 应用主题色
    if (appearance.value.primaryColor) {
      const hex = appearance.value.primaryColor.replace('#', '')
      const r = parseInt(hex.substring(0, 2), 16)
      const g = parseInt(hex.substring(2, 4), 16)
      const b = parseInt(hex.substring(4, 6), 16)
      root.style.setProperty('--color-primary', `${r} ${g} ${b}`)
      root.style.setProperty('--el-color-primary', appearance.value.primaryColor)
    }

    // 应用圆角设置
    const borderRadiusMap: Record<BorderRadius, string> = {
      none: '0',
      small: '0.125rem',
      medium: '0.5rem',
      large: '1rem',
      custom: appearance.value.customBorderRadius || '0.5rem'
    }
    const radius = borderRadiusMap[appearance.value.borderRadius] || '0.5rem'
    root.style.setProperty('--border-radius', radius)
    root.style.setProperty('--el-border-radius-base', radius)
    root.style.setProperty('--el-border-radius-small', `calc(${radius} * 0.5)`)
    root.style.setProperty('--el-border-radius-round', `calc(${radius} * 2.5)`)
  }

  /**
   * 同步主题模式到 uiStore
   */
  function syncThemeMode(mode: 'light' | 'dark' | 'auto') {
    const uiStore = useUiStore()

    if (mode === 'dark') {
      uiStore.setDarkMode(true)
    } else if (mode === 'light') {
      uiStore.setDarkMode(false)
    } else {
      // auto 模式：根据系统偏好
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      uiStore.setDarkMode(prefersDark)
    }
  }

  /**
   * 更新外观设置
   */
  function updateAppearance<K extends keyof UserPreferences['appearance']>(
    key: K,
    value: UserPreferences['appearance'][K]
  ) {
    preferences.value.appearance[key] = value
  }

  /**
   * 更新布局设置
   */
  function updateLayout<K extends keyof UserPreferences['layout']>(
    key: K,
    value: UserPreferences['layout'][K]
  ) {
    preferences.value.layout[key] = value
  }

  /**
   * 更新通用设置
   */
  function updateGeneral<K extends keyof UserPreferences['general']>(
    key: K,
    value: UserPreferences['general'][K]
  ) {
    preferences.value.general[key] = value
  }

  /**
   * 更新快捷键
   */
  function updateShortcut<K extends keyof UserPreferences['shortcuts']>(
    key: K,
    value: UserPreferences['shortcuts'][K]
  ) {
    preferences.value.shortcuts[key] = value
  }

  /**
   * 更新小部件设置
   */
  function updateWidgets<K extends keyof UserPreferences['widgets']>(
    key: K,
    value: UserPreferences['widgets'][K]
  ) {
    preferences.value.widgets[key] = value
  }

  /**
   * 更新版权设置
   */
  function updateCopyright<K extends keyof UserPreferences['copyright']>(
    key: K,
    value: UserPreferences['copyright'][K]
  ) {
    preferences.value.copyright[key] = value
  }

  /**
   * 重置所有设置为默认值
   */
  function resetToDefaults() {
    preferences.value = { ...DEFAULT_PREFERENCES }
  }

  /**
   * 导出设置
   */
  function exportSettings(): string {
    return JSON.stringify(preferences.value, null, 2)
  }

  /**
   * 导入设置
   */
  function importSettings(jsonString: string): boolean {
    try {
      const imported = JSON.parse(jsonString)
      if (imported.appearance && imported.layout && imported.general) {
        preferences.value = { ...DEFAULT_PREFERENCES, ...imported }
        return true
      }
      return false
    } catch {
      return false
    }
  }

  // ============ Return ============

  return {
    // State
    preferences,
    isInitialized,
    appearance,
    layout,
    shortcuts,
    general,
    widgets,
    copyright,

    // Actions
    initialize,
    updateAppearance,
    updateLayout,
    updateGeneral,
    updateShortcut,
    updateWidgets,
    updateCopyright,
    resetToDefaults,
    exportSettings,
    importSettings
  }
})
