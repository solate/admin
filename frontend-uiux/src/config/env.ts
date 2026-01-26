/**
 * 环境变量配置
 * 统一管理所有环境相关的配置
 */
export const env = {
  // API 配置
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',

  // 应用配置
  appTitle: import.meta.env.VITE_APP_TITLE || 'Multi-Tenant SaaS',
  appVersion: import.meta.env.VITE_APP_VERSION || '1.0.0',

  // 环境判断
  isDev: import.meta.env.DEV,
  isProd: import.meta.env.PROD,
  mode: import.meta.env.MODE
} as const

export type Env = typeof env
