/**
 * 应用配置
 * 集中管理应用的全局配置项
 */

export const appConfig = {
  // 应用基础信息
  name: 'Admin Dashboard',
  version: '1.0.0',
  description: '多租户 SaaS 管理平台',

  // 功能开关
  features: {
    enableNotifications: true,  // 是否启用通知功能
    enableSearch: true,          // 是否启用全局搜索
    enablePreferences: true,     // 是否启用用户偏好设置
    enableMultiTenant: true,     // 是否启用多租户功能
    enableDarkMode: true,        // 是否启用深色模式
    enableI18n: true,            // 是否启用国际化
  } as const,

  // 分页配置
  pagination: {
    defaultPageSize: 20,
    pageSizeOptions: [10, 20, 50, 100] as const,
  },

  // 表格配置
  table: {
    defaultPageSize: 20,
    showSizeChanger: true,
    showQuickJumper: true,
  },

  // API 配置
  api: {
    baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
    timeout: 30000, // 30秒
  },

  // 存储配置
  storage: {
    tokenKey: 'auth_token',
    refreshTokenKey: 'refresh_token',
    tenantKey: 'current_tenant_id',
    preferencesKey: 'user_preferences',
  },

  // 路由配置
  router: {
    base: '/',
    mode: 'history' as const, // 'hash' | 'history'
  },

  // 主题配置
  theme: {
    defaultThemeMode: 'auto' as const, // 'light' | 'dark' | 'auto'
    defaultPrimaryColor: '#2563eb',    // 默认主题色
    defaultBorderRadius: 'medium' as const, // 'none' | 'small' | 'medium' | 'large'
  },

  // 语言配置
  i18n: {
    defaultLocale: 'zh-CN' as const,
    fallbackLocale: 'en-US' as const,
    availableLocales: ['zh-CN', 'en-US'] as const,
  },
} as const

/**
 * 导出类型
 */
export type AppConfig = typeof appConfig
export type FeatureFlags = typeof appConfig.features
export type ThemeMode = typeof appConfig.theme.defaultThemeMode
export type BorderRadius = typeof appConfig.theme.defaultBorderRadius
