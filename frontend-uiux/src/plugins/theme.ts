/**
 * 主题初始化插件
 * 在应用启动时自动初始化主题设置
 *
 * @example
 * ```typescript
 * import themePlugin from './plugins/theme'
 * app.use(themePlugin)
 * ```
 */

import type { App } from 'vue'
import { usePreferencesStore } from '@/stores/modules/preferences'

/**
 * 将 HEX 颜色转换为 RGB 空格分隔格式
 */
function hexToRgb(hex: string): string {
  const cleanHex = hex.replace('#', '')
  const r = parseInt(cleanHex.substring(0, 2), 16)
  const g = parseInt(cleanHex.substring(2, 4), 16)
  const b = parseInt(cleanHex.substring(4, 6), 16)
  return `${r} ${g} ${b}`
}

/**
 * 将 HEX 颜色转换为 HSL 空格分隔格式
 */
function hexToHsl(hex: string): string {
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

export default {
  install: (app: App) => {
    // 等待 Pinia store 准备就绪
    app.mixin({
      beforeCreate() {
        const preferencesStore = usePreferencesStore()
        const { appearance } = preferencesStore
        const root = document.documentElement

        // ========== 应用主题色 ==========
        const primaryColor = appearance.primaryColor
        const rgb = hexToRgb(primaryColor)
        const hsl = hexToHsl(primaryColor)

        root.style.setProperty('--color-primary', rgb)
        root.style.setProperty('--color-primary-hsl', hsl)

        // ========== 应用圆角设置 ==========
        const borderRadiusMap = {
          none: '0',
          small: '0.125rem',
          medium: '0.5rem',
          large: '1rem'
        }
        root.style.setProperty('--border-radius', borderRadiusMap[appearance.borderRadius])

        // ========== 应用色盲模式 ==========
        if (appearance.colorBlindMode !== 'none') {
          root.setAttribute('data-color-blind', appearance.colorBlindMode)
        }

        // ========== 应用高对比度模式 ==========
        if (appearance.highContrast) {
          root.setAttribute('data-high-contrast', 'true')
        }

        // ========== 应用主题模式（深色/浅色） ==========
        if (appearance.themeMode === 'dark') {
          root.classList.add('dark')
        } else if (appearance.themeMode === 'light') {
          root.classList.remove('dark')
        } else {
          // auto 模式：跟随系统
          const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
          if (prefersDark) {
            root.classList.add('dark')
          } else {
            root.classList.remove('dark')
          }

          // 监听系统主题变化
          window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
            if (appearance.themeMode === 'auto') {
              if (e.matches) {
                root.classList.add('dark')
              } else {
                root.classList.remove('dark')
              }
            }
          })
        }

        // ========== 初始化完成标记 ==========
        // 可以用于后续的主题切换监听
        root.setAttribute('data-theme-initialized', 'true')
      }
    })
  }
}
