// 用户偏好设置 store

import { defineStore } from 'pinia'
import { ref, watch, computed } from 'vue'
import type { UserPreferences } from '@/types/preferences'
import { DEFAULT_PREFERENCES, PREFERENCES_STORAGE_KEY } from '@/types/preferences'
import { useUiStore } from './ui'
import { useI18n } from 'vue-i18n'

export const usePreferencesStore = defineStore('preferences', () => {
  // 从 localStorage 读取初始状态
  const storedPreferences = localStorage.getItem(PREFERENCES_STORAGE_KEY)
  const initialState: UserPreferences = storedPreferences
    ? { ...DEFAULT_PREFERENCES, ...JSON.parse(storedPreferences) }
    : DEFAULT_PREFERENCES

  // State
  const preferences = ref<UserPreferences>(initialState)
  const isInitialized = ref(false)

  // 计算属性 - 便于访问
  const appearance = computed(() => preferences.value.appearance)
  const layout = computed(() => preferences.value.layout)
  const shortcuts = computed(() => preferences.value.shortcuts)
  const general = computed(() => preferences.value.general)

  /**
   * 初始化偏好设置
   * 同步到 useUiStore 和 DOM
   */
  function initialize() {
    if (isInitialized.value) return

    const uiStore = useUiStore()
    const { locale } = useI18n()

    // 同步主题模式
    syncThemeMode(appearance.value.themeMode)

    // 同步语言
    if (general.value.language !== locale.value) {
      uiStore.setLocale(general.value.language)
    }

    // 应用外观设置到 DOM
    applyAppearanceSettings()

    // 监听主题模式变化
    watch(
      () => appearance.value.themeMode,
      (newMode) => {
        syncThemeMode(newMode)
      }
    )

    // 监听语言变化
    watch(
      () => general.value.language,
      (newLanguage) => {
        uiStore.setLocale(newLanguage)
      }
    )

    // 监听外观设置变化，应用到 DOM
    watch(
      () => appearance.value,
      () => {
        applyAppearanceSettings()
      },
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

    // 应用主题色
    if (appearance.value.primaryColor) {
      // 将十六进制颜色转换为 RGB
      const hex = appearance.value.primaryColor.replace('#', '')
      const r = parseInt(hex.substring(0, 2), 16)
      const g = parseInt(hex.substring(2, 4), 16)
      const b = parseInt(hex.substring(4, 6), 16)
      root.style.setProperty('--color-primary', `${r} ${g} ${b}`)
    }

    // 应用圆角设置
    const borderRadiusMap = {
      none: '0',
      small: '0.125rem',
      medium: '0.5rem',
      large: '1rem'
    }
    root.style.setProperty('--border-radius', borderRadiusMap[appearance.value.borderRadius])
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

      // 监听系统主题变化
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      const handler = (e: MediaQueryListEvent) => uiStore.setDarkMode(e.matches)
      mediaQuery.addEventListener('change', handler)
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
  function updateShortcut(actionId: string, keyCombination: string) {
    preferences.value.shortcuts[actionId] = keyCombination
  }

  /**
   * 重置所有设置为默认值
   */
  function resetToDefaults() {
    preferences.value = { ...DEFAULT_PREFERENCES }
  }

  /**
   * 导出设置（用于备份或分享）
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
      // 验证导入的数据结构（简化验证）
      if (imported.appearance && imported.layout && imported.general) {
        preferences.value = { ...DEFAULT_PREFERENCES, ...imported }
        return true
      }
      return false
    } catch {
      return false
    }
  }

  return {
    // State
    preferences,
    isInitialized,
    appearance,
    layout,
    shortcuts,
    general,

    // Actions
    initialize,
    updateAppearance,
    updateLayout,
    updateGeneral,
    updateShortcut,
    resetToDefaults,
    exportSettings,
    importSettings
  }
})
