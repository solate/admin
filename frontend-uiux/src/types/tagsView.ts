// 标签页系统类型定义

/**
 * 标签页项
 */
export interface TagView {
  /** 路由路径 */
  path: string
  /** 路由名称 */
  name: string
  /** 标签标题 */
  title: string
  /** 图标名称 */
  icon?: string
  /** 查询参数 */
  query?: Record<string, string>
  /** 完整路径（包含查询参数） */
  fullPath: string
  /** 是否固定（不可关闭） */
  affix?: boolean
}

/**
 * 标签页状态
 */
export interface TagsViewState {
  /** 已访问的标签列表 */
  visitedTags: TagView[]
  /** 缓存的组件名称（用于 keep-alive） */
  cachedTags: string[]
  /** 当前激活的标签路径 */
  activeTagPath: string
}

/**
 * localStorage 存储键
 */
export const TAGS_VIEW_STORAGE_KEY = 'tags-view-state'
