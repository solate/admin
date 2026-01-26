/**
 * 应用常量定义
 * 集中管理应用中使用的各种常量
 */

// ============ 主题相关 ============

/** 预设主题色选项 */
export const THEME_COLORS = [
  { name: 'blue', label: '蓝色', value: '#2563eb' },
  { name: 'violet', label: '紫色', value: '#8b5cf6' },
  { name: 'purple', label: '紫罗兰', value: '#a855f7' },
  { name: 'fuchsia', label: '品红', value: '#d946ef' },
  { name: 'pink', label: '粉色', value: '#ec4899' },
  { name: 'rose', label: '玫瑰', value: '#f43f5e' },
  { name: 'orange', label: '橙色', value: '#f97316' },
  { name: 'amber', label: '琥珀', value: '#f59e0b' },
  { name: 'green', label: '绿色', value: '#22c55e' },
  { name: 'emerald', label: '翡翠', value: '#10b981' },
  { name: 'teal', label: '青色', value: '#14b8a6' },
  { name: 'cyan', label: '青蓝', value: '#06b6d4' },
] as const

/** 主题模式选项 */
export const THEME_MODES = {
  LIGHT: 'light',
  DARK: 'dark',
  AUTO: 'auto',
} as const

/** 圆角大小选项 */
export const BORDER_RADIUS = {
  NONE: 'none',
  SMALL: 'small',
  MEDIUM: 'medium',
  LARGE: 'large',
} as const

/** 色盲模式选项 */
export const COLOR_BLIND_MODES = {
  NONE: 'none',
  PROTANOPIA: 'protanopia',   // 红色盲
  DEUTERANOPIA: 'deuteranopia', // 绿色盲
  TRITANOPIA: 'tritanopia',    // 蓝色盲
} as const

// ============ 用户相关 ============

/** 用户状态 */
export const USER_STATUS = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
  SUSPENDED: 'suspended',
  PENDING: 'pending',
} as const

/** 用户角色 */
export const USER_ROLES = {
  ADMIN: 'admin',
  USER: 'user',
  GUEST: 'guest',
} as const

// ============ 租户相关 ============

/** 租户状态 */
export const TENANT_STATUS = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
  SUSPENDED: 'suspended',
  PENDING: 'pending',
} as const

/** 租户计划类型 */
export const TENANT_PLANS = {
  FREE: 'free',
  BASIC: 'basic',
  PRO: 'pro',
  ENTERPRISE: 'enterprise',
} as const

// ============ 服务相关 ============

/** 服务状态 */
export const SERVICE_STATUS = {
  RUNNING: 'running',
  STOPPED: 'stopped',
  DEPLOYING: 'deploying',
  ERROR: 'error',
  PENDING: 'pending',
} as const

/** 服务类型 */
export const SERVICE_TYPES = {
  WEB: 'web',
  API: 'api',
  WORKER: 'worker',
  DATABASE: 'database',
  CACHE: 'cache',
} as const

// ============ 通知相关 ============

/** 通知类型 */
export const NOTIFICATION_TYPES = {
  INFO: 'info',
  SUCCESS: 'success',
  WARNING: 'warning',
  ERROR: 'error',
} as const

/** 通知优先级 */
export const NOTIFICATION_PRIORITIES = {
  LOW: 'low',
  MEDIUM: 'medium',
  HIGH: 'high',
  URGENT: 'urgent',
} as const

// ============ HTTP 相关 ============

/** HTTP 状态码 */
export const HTTP_STATUS = {
  OK: 200,
  CREATED: 201,
  NO_CONTENT: 204,
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  METHOD_NOT_ALLOWED: 405,
  CONFLICT: 409,
  UNPROCESSABLE_ENTITY: 422,
  INTERNAL_SERVER_ERROR: 500,
  SERVICE_UNAVAILABLE: 503,
} as const

/** 请求方法 */
export const HTTP_METHODS = {
  GET: 'GET',
  POST: 'POST',
  PUT: 'PUT',
  PATCH: 'PATCH',
  DELETE: 'DELETE',
} as const

// ============ 日期时间相关 ============

/** 日期格式 */
export const DATE_FORMATS = {
  DATE: 'YYYY-MM-DD',
  TIME: 'HH:mm:ss',
  DATETIME: 'YYYY-MM-DD HH:mm:ss',
  DATETIME_SHORT: 'YYYY-MM-DD HH:mm',
  MONTH: 'YYYY-MM',
  YEAR: 'YYYY',
} as const

/** 时间范围选项 */
export const DATE_RANGES = {
  TODAY: 'today',
  YESTERDAY: 'yesterday',
  LAST_7_DAYS: 'last_7_days',
  LAST_30_DAYS: 'last_30_days',
  THIS_MONTH: 'this_month',
  LAST_MONTH: 'last_month',
  THIS_YEAR: 'this_year',
  CUSTOM: 'custom',
} as const

// ============ 正则表达式 ============

/** 常用正则表达式 */
export const REGEX_PATTERNS = {
  /** 邮箱 */
  EMAIL: /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/,
  /** 手机号（中国大陆） */
  PHONE_CN: /^1[3-9]\d{9}$/,
  /** URL */
  URL: /^https?:\/\/.+/,
  /** IP 地址 */
  IP: /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/,
  /** 十六进制颜色 */
  HEX_COLOR: /^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$/,
  /** 用户名（字母开头，允许字母数字下划线，4-16位） */
  USERNAME: /^[a-zA-Z][a-zA-Z0-9_]{3,15}$/,
  /** 密码（至少8位，包含字母和数字） */
  PASSWORD: /^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d@$!%*#?&]{8,}$/,
} as const

// ============ 文件相关 ============

/** 文件大小单位 */
export const FILE_SIZE_UNITS = ['B', 'KB', 'MB', 'GB', 'TB'] as const

/** 常用 MIME 类型 */
export const MIME_TYPES = {
  // 图片
  JPEG: 'image/jpeg',
  PNG: 'image/png',
  GIF: 'image/gif',
  SVG: 'image/svg+xml',
  WEBP: 'image/webp',
  // 文档
  PDF: 'application/pdf',
  DOC: 'application/msword',
  DOCX: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  XLS: 'application/vnd.ms-excel',
  XLSX: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  // 文本
  TXT: 'text/plain',
  CSV: 'text/csv',
  JSON: 'application/json',
  // 压缩
  ZIP: 'application/zip',
  RAR: 'application/x-rar-compressed',
} as const

/** 最大文件上传大小（字节） */
export const MAX_FILE_SIZES = {
  IMAGE: 5 * 1024 * 1024,      // 5MB
  DOCUMENT: 10 * 1024 * 1024,  // 10MB
  VIDEO: 100 * 1024 * 1024,    // 100MB
  AVATAR: 2 * 1024 * 1024,     // 2MB
} as const

// ============ 分页相关 ============

/** 默认分页配置 */
export const PAGINATION = {
  DEFAULT_PAGE: 1,
  DEFAULT_PAGE_SIZE: 20,
  PAGE_SIZE_OPTIONS: [10, 20, 50, 100],
  MAX_PAGE_SIZE: 1000,
} as const

// ============ 动画相关 ============

/** 动画持续时间（毫秒） */
export const ANIMATION_DURATION = {
  FAST: 150,
  NORMAL: 200,
  SLOW: 300,
  EXTRA_SLOW: 500,
} as const

/** 缓动函数 */
export const EASING = {
  LINEAR: 'linear',
  EASE: 'ease',
  EASE_IN: 'ease-in',
  EASE_OUT: 'ease-out',
  EASE_IN_OUT: 'ease-in-out',
} as const

// ============ 键盘快捷键 ============

/** 默认快捷键配置 */
export const DEFAULT_SHORTCUTS = {
  // 全局
  TOGGLE_DARK_MODE: 'Cmd+D',
  TOGGLE_SIDEBAR: 'Cmd+B',
  OPEN_SEARCH: 'Cmd+K',
  OPEN_NOTIFICATIONS: 'Cmd+N',
  OPEN_SETTINGS: 'Cmd+,',
  // 导航
  GO_DASHBOARD: 'G+D',
  GO_TENANTS: 'G+T',
  GO_USERS: 'G+U',
  GO_SERVICES: 'G+S',
  // 操作
  SAVE: 'Cmd+S',
  CANCEL: 'Esc',
  DELETE: 'Cmd+Backspace',
} as const

// ============ 错误消息 ============

/** 通用错误消息 */
export const ERROR_MESSAGES = {
  NETWORK_ERROR: '网络连接失败，请检查您的网络设置',
  UNAUTHORIZED: '未授权，请先登录',
  FORBIDDEN: '您没有权限执行此操作',
  NOT_FOUND: '请求的资源不存在',
  SERVER_ERROR: '服务器错误，请稍后再试',
  VALIDATION_ERROR: '表单验证失败，请检查输入',
  UNKNOWN_ERROR: '发生未知错误，请稍后再试',
} as const

// ============ 成功消息 ============

/** 通用成功消息 */
export const SUCCESS_MESSAGES = {
  CREATE_SUCCESS: '创建成功',
  UPDATE_SUCCESS: '更新成功',
  DELETE_SUCCESS: '删除成功',
  SAVE_SUCCESS: '保存成功',
  COPY_SUCCESS: '已复制到剪贴板',
  UPLOAD_SUCCESS: '上传成功',
} as const

/**
 * 导出类型
 */
export type ThemeColorName = (typeof THEME_COLORS)[number]['name']
export type ThemeMode = typeof THEME_MODES[keyof typeof THEME_MODES]
export type BorderRadius = typeof BORDER_RADIUS[keyof typeof BORDER_RADIUS]
export type UserStatus = typeof USER_STATUS[keyof typeof USER_STATUS]
export type TenantStatus = typeof TENANT_STATUS[keyof typeof TENANT_STATUS]
export type ServiceStatus = typeof SERVICE_STATUS[keyof typeof SERVICE_STATUS]
export type NotificationType = typeof NOTIFICATION_TYPES[keyof typeof NOTIFICATION_TYPES]
