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
  shortcuts: ShortcutPreferences
  /** 通用设置 */
  general: GeneralPreferences
  /** 小部件设置 */
  widgets: WidgetPreferences
  /** 版权设置 */
  copyright: CopyrightPreferences
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
  /** 深色顶栏 */
  darkHeader: boolean
  /** 圆角大小 */
  borderRadius: BorderRadius
  /** 自定义圆角值（当 borderRadius 为 'custom' 时使用） */
  customBorderRadius?: string
  /** 色盲模式 */
  colorBlindMode: ColorBlindMode
  /** 高对比度 */
  highContrast: boolean
  /** 灰色模式 */
  grayMode: boolean
}

/**
 * 主题模式
 */
export type ThemeMode = 'light' | 'dark' | 'auto'

/**
 * 圆角大小
 * - none: 无圆角 (0px)
 * - small: 小圆角 (2px)
 * - medium: 中圆角 (8px) - 默认
 * - large: 大圆角 (16px)
 * - custom: 自定义值
 */
export type BorderRadius = 'none' | 'small' | 'medium' | 'large' | 'custom'

/**
 * 色彩视觉模式
 * - none: 正常视觉
 * - protanopia: 红色盲 (完全无法感知红色)
 * - deuteranopia: 绿色盲 (完全无法感知绿色)
 * - tritanopia: 蓝色盲 (完全无法感知蓝色)
 * - protanomaly: 红色弱 (红色敏感度降低)
 * - deuteranomaly: 绿色弱 (绿色敏感度降低，最常见)
 * - tritanomaly: 蓝色弱 (蓝色敏感度降低)
 * - achromatopsia: 全色盲 (只能看到灰度)
 */
export type ColorBlindMode = 'none' | 'protanopia' | 'deuteranopia' | 'tritanopia' | 'protanomaly' | 'deuteranomaly' | 'tritanomaly' | 'achromatopsia'

/**
 * 布局偏好设置
 */
export interface LayoutPreferences {
  /** 布局模式 */
  layoutMode: LayoutMode
  /** 侧边栏宽度（精确像素值） */
  sidebarWidth: number
  /** 侧边栏折叠宽度 */
  sidebarCollapsedWidth: number
  /** 侧边栏是否可折叠 */
  sidebarCollapsible: boolean
  /** 侧边栏默认折叠状态 */
  sidebarCollapsed: boolean
  /** 导航样式 */
  navStyle: NavStyle
  /** 导航菜单手风琴模式 */
  navAccordion: boolean
  /** 顶栏模式 */
  headerMode: HeaderMode
  /** 顶栏高度 */
  headerHeight: number
  /** 内容区域宽度模式 */
  contentWidthMode: ContentWidthMode
  /** 内容区域定宽值 */
  contentFixedWidth?: number
  /** 显示面包屑 */
  showBreadcrumbs: boolean
  /** 仅有一个面包屑时隐藏 */
  hideSingleBreadcrumb: boolean
  /** 面包屑样式 */
  breadcrumbStyle: BreadcrumbStyle
  /** 面包屑显示首页 */
  breadcrumbShowHome: boolean
  /** 显示标签页 */
  showTabs: boolean
  /** 标签页持久化 */
  tabsPersistent: boolean
  /** 标签页可拖拽 */
  tabsDraggable: boolean
  /** 标签页样式 */
  tabsStyle: TabsStyle
  /** 显示小部件 */
  showWidgets: boolean
  /** 小部件位置 */
  widgetsPosition: WidgetsPosition
  /** 显示页脚 */
  showFooter: boolean
  /** 页脚固定在底部 */
  footerFixed: boolean
  /** 显示版权信息 */
  showCopyright: boolean
}

/**
 * 布局模式
 */
export type LayoutMode = 'sidebar' | 'topbar' | 'mixed' | 'horizontal'

/**
 * 顶栏模式
 */
export type HeaderMode = 'static' | 'fixed' | 'auto-hide'

/**
 * 内容宽度模式
 */
export type ContentWidthMode = 'fluid' | 'fixed'

/**
 * 导航样式
 */
export type NavStyle = 'icon-text' | 'icon-only'

/**
 * 面包屑样式
 */
export type BreadcrumbStyle = 'normal' | 'background'

/**
 * 标签页样式
 */
export type TabsStyle = 'chrome' | 'plain' | 'card' | 'smart'

/**
 * 小部件位置
 */
export type WidgetsPosition = 'header' | 'sidebar' | 'auto'

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
  /** 页面切换进度条 */
  pageProgress: boolean
  /** 页面切换 Loading */
  pageLoading: boolean
  /** 检查更新间隔（天） */
  checkUpdatesInterval: number
  /** 启用检查更新 */
  enableCheckUpdates: boolean
}

/**
 * 页面过渡效果
 */
export type PageTransition = 'fade' | 'slide' | 'scale' | 'zoom' | 'fade-slide'

/**
 * 快捷键偏好设置
 */
export interface ShortcutPreferences {
  /** 是否启用快捷键 */
  enable: boolean
  /** 全局搜索快捷键 */
  globalSearch: boolean
  /** 退出登录快捷键 */
  globalLogout: boolean
  /** 锁定屏幕快捷键 */
  globalLockScreen: boolean
  /** 打开设置快捷键 */
  globalPreferences: boolean
  /** 刷新页面快捷键 */
  refresh: boolean
  /** 全屏快捷键 */
  fullscreen: boolean
}

/**
 * 小部件偏好设置
 */
export interface WidgetPreferences {
  /** 全局搜索 */
  globalSearch: boolean
  /** 主题切换 */
  themeToggle: boolean
  /** 语言切换 */
  languageToggle: boolean
  /** 全屏 */
  fullscreen: boolean
  /** 通知 */
  notification: boolean
  /** 锁屏 */
  lockScreen: boolean
  /** 侧边栏切换 */
  sidebarToggle: boolean
  /** 刷新 */
  refresh: boolean
  /** 小部件位置 */
  position: WidgetsPosition
}

/**
 * 版权偏好设置
 */
export interface CopyrightPreferences {
  /** 是否启用 */
  enable: boolean
  /** 公司名称 */
  companyName: string
  /** 公司网站 */
  companySiteLink: string
  /** 日期 */
  date: string
  /** ICP 备案号 */
  icp: string
  /** ICP 链接 */
  icpLink: string
  /** 在设置中显示 */
  settingShow: boolean
}

/**
 * 默认偏好设置
 */
export const DEFAULT_PREFERENCES: UserPreferences = {
  appearance: {
    themeMode: 'auto',
    primaryColor: '#2563eb',
    darkSidebar: false,
    darkHeader: false,
    borderRadius: 'medium',
    customBorderRadius: '0.5rem',
    colorBlindMode: 'none',
    highContrast: false,
    grayMode: false
  },
  layout: {
    layoutMode: 'sidebar',
    sidebarWidth: 256,
    sidebarCollapsedWidth: 64,
    sidebarCollapsible: true,
    sidebarCollapsed: false,
    navStyle: 'icon-text',
    navAccordion: true,
    headerMode: 'fixed',
    headerHeight: 60,
    contentWidthMode: 'fluid',
    showBreadcrumbs: true,
    hideSingleBreadcrumb: false,
    breadcrumbStyle: 'normal',
    breadcrumbShowHome: true,
    showTabs: false,
    tabsPersistent: true,
    tabsDraggable: true,
    tabsStyle: 'chrome',
    showWidgets: true,
    widgetsPosition: 'auto',
    showFooter: true,
    footerFixed: false,
    showCopyright: true
  },
  shortcuts: {
    enable: true,
    globalSearch: true,
    globalLogout: true,
    globalLockScreen: true,
    globalPreferences: true,
    refresh: true,
    fullscreen: true
  },
  general: {
    language: 'zh-CN',
    dynamicTitle: true,
    enableAnimations: true,
    pageTransition: 'fade-slide',
    pageProgress: true,
    pageLoading: true,
    checkUpdatesInterval: 7,
    enableCheckUpdates: false
  },
  widgets: {
    globalSearch: true,
    themeToggle: true,
    languageToggle: true,
    fullscreen: true,
    notification: true,
    lockScreen: true,
    sidebarToggle: true,
    refresh: true,
    position: 'auto'
  },
  copyright: {
    enable: true,
    companyName: 'Your Company',
    companySiteLink: 'https://example.com',
    date: new Date().getFullYear().toString(),
    icp: '',
    icpLink: '',
    settingShow: true
  }
}

/**
 * 主题色选项
 * 分组：基础色、活力色、自然色、中性色
 */
export const THEME_COLORS = [
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
]

/**
 * 主题色分组
 */
export const THEME_COLOR_GROUPS = {
  basic: { name: 'basic', label: '基础色系' },
  vibrant: { name: 'vibrant', label: '活力色系' },
  nature: { name: 'nature', label: '自然色系' },
  neutral: { name: 'neutral', label: '中性色系' },
} as const

export type ThemeColorGroup = keyof typeof THEME_COLOR_GROUPS

/**
 * localStorage 存储键
 */
export const PREFERENCES_STORAGE_KEY = 'user-preferences'
