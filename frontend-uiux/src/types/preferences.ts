// 用户偏好设置类型定义

/**
 * 用户偏好设置接口
 */
export interface UserPreferences {
  /** 外观设置 */
  appearance: AppearancePreferences
  /** 布局设置 */
  layout: LayoutPreferences
  /** 快捷键设置 */
  shortcuts: Record<string, string>
  /** 通用设置 */
  general: GeneralPreferences
}

/**
 * 外观偏好设置
 */
export interface AppearancePreferences {
  /** 主题模式 */
  themeMode: ThemeMode
  /** 主题色 */
  primaryColor: string
  /** 深色侧边栏 */
  darkSidebar: boolean
  /** 圆角大小 */
  borderRadius: BorderRadius
  /** 色盲模式 */
  colorBlindMode: ColorBlindMode
  /** 高对比度 */
  highContrast: boolean
}

/**
 * 主题模式
 */
export type ThemeMode = 'light' | 'dark' | 'auto'

/**
 * 圆角大小
 */
export type BorderRadius = 'none' | 'small' | 'medium' | 'large'

/**
 * 色盲模式
 */
export type ColorBlindMode = 'none' | 'protanopia' | 'deuteranopia' | 'tritanopia'

/**
 * 布局偏好设置
 */
export interface LayoutPreferences {
  /** 布局模式 */
  layoutMode: LayoutMode
  /** 侧边栏宽度 */
  sidebarWidth: SidebarWidth
  /** 导航样式 */
  navStyle: NavStyle
  /** 显示面包屑 */
  showBreadcrumbs: boolean
  /** 显示标签页 */
  showTabs: boolean
  /** 显示小部件 */
  showWidgets: boolean
  /** 显示页脚 */
  showFooter: boolean
  /** 显示版权信息 */
  showCopyright: boolean
}

/**
 * 布局模式
 */
export type LayoutMode = 'sidebar' | 'topbar'

/**
 * 侧边栏宽度
 */
export type SidebarWidth = 'narrow' | 'medium' | 'wide'

/**
 * 导航样式
 */
export type NavStyle = 'icon-text' | 'icon-only'

/**
 * 通用偏好设置
 */
export interface GeneralPreferences {
  /** 语言 */
  language: 'zh-CN' | 'en-US'
  /** 动态标题 */
  dynamicTitle: boolean
  /** 启用动画 */
  enableAnimations: boolean
  /** 页面过渡效果 */
  pageTransition: PageTransition
}

/**
 * 页面过渡效果
 */
export type PageTransition = 'fade' | 'slide' | 'scale'

/**
 * 默认偏好设置
 */
export const DEFAULT_PREFERENCES: UserPreferences = {
  appearance: {
    themeMode: 'auto',
    primaryColor: '#2563eb',
    darkSidebar: false,
    borderRadius: 'medium',
    colorBlindMode: 'none',
    highContrast: false
  },
  layout: {
    layoutMode: 'sidebar',
    sidebarWidth: 'medium',
    navStyle: 'icon-text',
    showBreadcrumbs: true,
    showTabs: false,
    showWidgets: true,
    showFooter: true,
    showCopyright: true
  },
  shortcuts: {},
  general: {
    language: 'zh-CN',
    dynamicTitle: true,
    enableAnimations: true,
    pageTransition: 'fade'
  }
}

/**
 * 主题色选项
 */
export const THEME_COLORS = [
  { name: 'blue', value: '#2563eb', label: '蓝色' },
  { name: 'purple', value: '#9333ea', label: '紫色' },
  { name: 'pink', value: '#ec4899', label: '粉色' },
  { name: 'red', value: '#ef4444', label: '红色' },
  { name: 'orange', value: '#f97316', label: '橙色' },
  { name: 'green', value: '#22c55e', label: '绿色' },
  { name: 'teal', value: '#14b8a6', label: '青色' },
  { name: 'cyan', value: '#06b6d4', label: '青蓝' }
]

/**
 * localStorage 存储键
 */
export const PREFERENCES_STORAGE_KEY = 'user-preferences'
