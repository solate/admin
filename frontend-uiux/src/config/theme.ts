/**
 * 主题配置
 * 定义应用的主题相关配置
 */

import { appConfig } from './app'

/**
 * 主题配置
 */
export const themeConfig = {
  // 默认主题色
  defaultPrimaryColor: appConfig.theme.defaultPrimaryColor,

  // 主题色选项
  primaryColors: [
    '#2563eb', // blue
    '#8b5cf6', // violet
    '#a855f7', // purple
    '#d946ef', // fuchsia
    '#ec4899', // pink
    '#f43f5e', // rose
    '#f97316', // orange
    '#f59e0b', // amber
    '#22c55e', // green
    '#10b981', // emerald
    '#14b8a6', // teal
    '#06b6d4', // cyan
  ] as const,

  // 圆角选项（rem 单位）
  borderRadius: {
    none: '0',
    small: '0.125rem',   // 2px
    medium: '0.5rem',    // 8px
    large: '1rem',       // 16px
  } as const,

  // 主题模式选项
  themeModes: ['light', 'dark', 'auto'] as const,

  // 色盲模式选项
  colorBlindModes: ['none', 'protanopia', 'deuteranopia', 'tritanopia'] as const,

  // CSS 变量前缀
  cssVarPrefix: '--color-primary',

  // 主题切换动画时长（毫秒）
  transitionDuration: 200,
} as const

/**
 * 主题类型定义
 */
export type PrimaryColor = (typeof themeConfig.primaryColors)[number]
export type BorderRadiusOption = keyof typeof themeConfig.borderRadius
export type ThemeModeOption = (typeof themeConfig.themeModes)[number]
export type ColorBlindMode = (typeof themeConfig.colorBlindModes)[number]
