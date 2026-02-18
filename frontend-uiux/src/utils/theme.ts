/**
 * 主题工具函数
 * 统一管理所有主题相关的颜色转换和 DOM 操作
 */

import type { AppearancePreferences, BorderRadius, ThemeMode, ColorBlindMode } from '@/types/preferences'

// ============================================================================
// 常量定义
// ============================================================================

/**
 * 圆角值映射表 (rem 单位)
 */
export const BORDER_RADIUS_MAP: Record<BorderRadius, string> = {
  none: '0',
  small: '0.125rem',    // 2px
  medium: '0.5rem',     // 8px
  large: '1rem',        // 16px
  custom: ''            // 使用 customBorderRadius
}

/**
 * localStorage 键名统一管理
 */
export const THEME_STORAGE_KEYS = {
  PREFERENCES: 'user-preferences',
  UI_STATE: 'ui-state',
  LOCALE: 'locale'
} as const

// ============================================================================
// 颜色转换函数
// ============================================================================

/**
 * 将十六进制颜色转换为 RGB 值
 * @param hex 十六进制颜色 (如 #2563eb)
 * @returns RGB 字符串 (如 "37 99 235")
 */
export function hexToRgb(hex: string): string {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex)
  if (!result) return '0 0 0'

  const r = parseInt(result[1], 16)
  const g = parseInt(result[2], 16)
  const b = parseInt(result[3], 16)

  return `${r} ${g} ${b}`
}

/**
 * 将十六进制颜色转换为 HSL 值
 * @param hex 十六进制颜色 (如 #2563eb)
 * @returns HSL 字符串 (如 "220 100% 54%")
 */
export function hexToHsl(hex: string): string {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex)
  if (!result) return '0 0% 0%'

  let r = parseInt(result[1], 16) / 255
  let g = parseInt(result[2], 16) / 255
  let b = parseInt(result[3], 16) / 255

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
 * 验证十六进制颜色格式
 * @param color 颜色字符串
 * @returns 是否为有效的十六进制颜色
 */
export function isValidHexColor(color: string): boolean {
  return /^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$/.test(color)
}

// ============================================================================
// DOM 操作函数
// ============================================================================

/**
 * 应用主题色到 DOM
 * @param primaryColor 主题色 (十六进制)
 */
export function applyPrimaryColor(primaryColor: string): void {
  const root = document.documentElement
  const rgb = hexToRgb(primaryColor)
  const hsl = hexToHsl(primaryColor)

  // 设置 CSS 变量 (RGB 格式，用于 tailwind)
  root.style.setProperty('--color-primary', rgb)
  // 设置 CSS 变量 (HSL 格式，用于自定义属性)
  root.style.setProperty('--color-primary-hsl', hsl)
  // 设置 Element Plus 主题色
  root.style.setProperty('--el-color-primary', primaryColor)
}

/**
 * 获取圆角 CSS 值
 * @param borderRadius 圆角设置
 * @param customValue 自定义圆角值 (当 borderRadius 为 'custom' 时使用)
 * @returns CSS 圆角值
 */
export function getBorderRadiusValue(borderRadius: BorderRadius, customValue?: string): string {
  if (borderRadius === 'custom') {
    // 验证自定义值格式
    const value = customValue?.trim() || '0.5rem'
    // 支持 rem 和 px 单位，无单位默认为 px
    if (/^\d+(\.\d+)?(rem|px|em|vw|vh|%)?$/.test(value)) {
      return value.endsWith('rem') || value.endsWith('px') || value.endsWith('em') ||
             value.endsWith('vw') || value.endsWith('vh') || value.endsWith('%')
        ? value
        : `${value}px`
    }
    return '0.5rem'
  }
  return BORDER_RADIUS_MAP[borderRadius]
}

/**
 * 应用圆角设置到 DOM
 * @param borderRadius 圆角设置
 * @param customValue 自定义圆角值
 */
export function applyBorderRadius(borderRadius: BorderRadius, customValue?: string): void {
  const root = document.documentElement
  const value = getBorderRadiusValue(borderRadius, customValue)

  // 设置全局圆角变量
  root.style.setProperty('--border-radius', value)

  // 同步更新 Element Plus 圆角变量
  root.style.setProperty('--el-border-radius-base', value)
  root.style.setProperty('--el-border-radius-small', `calc(${value} * 0.5)`)
  root.style.setProperty('--el-border-radius-round', `calc(${value} * 2.5)`)
}

/**
 * 应用主题模式到 DOM
 * @param themeMode 主题模式
 */
export function applyThemeMode(themeMode: ThemeMode): void {
  const root = document.documentElement

  if (themeMode === 'dark') {
    root.classList.add('dark')
    root.classList.remove('light')
  } else if (themeMode === 'light') {
    root.classList.remove('dark')
    root.classList.add('light')
  } else {
    // auto 模式：跟随系统
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    if (prefersDark) {
      root.classList.add('dark')
      root.classList.remove('light')
    } else {
      root.classList.remove('dark')
      root.classList.add('light')
    }
  }
}

/**
 * 应用色彩视觉模式到 DOM
 * @param mode 色彩视觉模式
 */
export function applyColorBlindMode(mode: ColorBlindMode): void {
  const root = document.documentElement

  if (mode === 'none') {
    root.removeAttribute('data-color-blind')
  } else {
    root.setAttribute('data-color-blind', mode)
  }
}

/**
 * 应用高对比度模式到 DOM
 * @param enabled 是否启用
 */
export function applyHighContrast(enabled: boolean): void {
  const root = document.documentElement

  if (enabled) {
    root.setAttribute('data-high-contrast', 'true')
  } else {
    root.removeAttribute('data-high-contrast')
  }
}

/**
 * 应用灰色模式到 DOM
 * @param enabled 是否启用
 */
export function applyGrayMode(enabled: boolean): void {
  const root = document.documentElement

  if (enabled) {
    root.setAttribute('data-gray-mode', 'true')
  } else {
    root.removeAttribute('data-gray-mode')
  }
}

/**
 * 应用深色侧边栏到 DOM
 * @param enabled 是否启用
 */
export function applyDarkSidebar(enabled: boolean): void {
  const root = document.documentElement

  if (enabled) {
    root.setAttribute('data-dark-sidebar', 'true')
  } else {
    root.removeAttribute('data-dark-sidebar')
  }
}

/**
 * 应用深色顶栏到 DOM
 * @param enabled 是否启用
 */
export function applyDarkHeader(enabled: boolean): void {
  const root = document.documentElement

  if (enabled) {
    root.setAttribute('data-dark-header', 'true')
  } else {
    root.removeAttribute('data-dark-header')
  }
}

/**
 * 批量应用所有外观设置到 DOM
 * @param appearance 外观设置对象
 */
export function applyAppearanceToDOM(appearance: AppearancePreferences): void {
  // 主题色
  applyPrimaryColor(appearance.primaryColor)

  // 圆角
  applyBorderRadius(appearance.borderRadius, appearance.customBorderRadius)

  // 主题模式
  applyThemeMode(appearance.themeMode)

  // 色彩视觉模式
  applyColorBlindMode(appearance.colorBlindMode)

  // 高对比度
  applyHighContrast(appearance.highContrast)

  // 灰色模式
  applyGrayMode(appearance.grayMode)

  // 深色侧边栏
  applyDarkSidebar(appearance.darkSidebar)

  // 深色顶栏
  applyDarkHeader(appearance.darkHeader)
}

// ============================================================================
// 系统主题监听
// ============================================================================

let systemThemeMediaQuery: MediaQueryList | null = null
let systemThemeCallback: ((e: MediaQueryListEvent) => void) | null = null

/**
 * 获取系统主题偏好
 * @returns 是否偏好深色模式
 */
export function getSystemThemePreference(): boolean {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

/**
 * 监听系统主题变化
 * @param callback 主题变化时的回调函数
 * @returns 清理函数
 */
export function watchSystemTheme(callback: (isDark: boolean) => void): () => void {
  // 清理之前的监听器
  if (systemThemeCallback && systemThemeMediaQuery) {
    systemThemeMediaQuery.removeEventListener('change', systemThemeCallback)
  }

  systemThemeMediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  systemThemeCallback = (e: MediaQueryListEvent) => {
    callback(e.matches)
  }

  systemThemeMediaQuery.addEventListener('change', systemThemeCallback)

  // 返回清理函数
  return () => {
    if (systemThemeMediaQuery && systemThemeCallback) {
      systemThemeMediaQuery.removeEventListener('change', systemThemeCallback)
      systemThemeMediaQuery = null
      systemThemeCallback = null
    }
  }
}
