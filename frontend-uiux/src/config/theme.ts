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

  // 主题色选项（包含标签和分组）
  primaryColorOptions: [
    // 基础色系
    { name: 'default', value: '#2563eb', label: '默认蓝', group: 'basic' },
    { name: 'violet', value: '#8b5cf6', label: '紫罗兰', group: 'basic' },
    { name: 'cherry-blossom', value: '#ec4899', label: '樱花粉', group: 'basic' },
    { name: 'rose', value: '#f43f5e', label: '玫瑰红', group: 'basic' },

    // 活力色系
    { name: 'orange', value: '#f97316', label: '橙黄色', group: 'vibrant' },
    { name: 'lemon', value: '#facc15', label: '柠檬黄', group: 'vibrant' },
    { name: 'amber', value: '#f59e0b', label: '琥珀金', group: 'vibrant' },

    // 自然色系
    { name: 'sky', value: '#0ea5e9', label: '天蓝色', group: 'nature' },
    { name: 'cyan', value: '#06b6d4', label: '青碧色', group: 'nature' },
    { name: 'light-green', value: '#84cc16', label: '浅绿色', group: 'nature' },
    { name: 'emerald', value: '#10b981', label: '翡翠绿', group: 'nature' },
    { name: 'teal', value: '#14b8a6', label: '蓝绿色', group: 'nature' },
    { name: 'deep-green', value: '#059669', label: '深绿色', group: 'nature' },
    { name: 'deep-blue', value: '#1d4ed8', label: '深蓝色', group: 'nature' },

    // 中性色系
    { name: 'zinc', value: '#71717a', label: '锌色灰', group: 'neutral' },
    { name: 'slate', value: '#64748b', label: '石板灰', group: 'neutral' },
    { name: 'neutral', value: '#737373', label: '中灰色', group: 'neutral' },
  ] as const,

  // 主题色选项（仅色值，向后兼容）
  primaryColors: [
    '#2563eb', // default blue
    '#8b5cf6', // violet
    '#ec4899', // cherry-blossom pink
    '#f43f5e', // rose
    '#f97316', // orange
    '#facc15', // lemon
    '#f59e0b', // amber
    '#0ea5e9', // sky
    '#06b6d4', // cyan
    '#84cc16', // light-green
    '#10b981', // emerald
    '#14b8a6', // teal
    '#059669', // deep-green
    '#1d4ed8', // deep-blue
    '#71717a', // zinc
    '#64748b', // slate
    '#737373', // neutral
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
