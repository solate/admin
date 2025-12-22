import http from './http'
import type { PageResponse } from './factory'

// Casbin规则信息
export interface CasbinRuleInfo {
  type: string
  resource: string
  action: string
}

// 菜单信息
export interface MenuInfo {
  menu_id: string
  code: string
  parent_id: string
  name: string
  path: string
  component: string
  redirect: string
  icon: string
  sort: number
  type: string
  status: number
  created_at: number
}

// 菜单树
export interface MenuTree {
  menu_id: string
  code: string
  parent_id: string
  name: string
  path: string
  component: string
  redirect: string
  icon: string
  sort: number
  type: string
  status: number
  children: MenuTree[]
  rule: CasbinRuleInfo
}

// 菜单列表请求参数
export interface MenuListParams {
  page: number
  page_size: number
  name?: string
  type?: string
  status?: number
}

// 菜单列表响应
export interface MenuListResponse {
  page: PageResponse
  list: MenuInfo[]
}

// 菜单树响应
export interface MenuTreeResponse {
  list: MenuTree[]
}

// 创建菜单请求
export interface CreateMenuRequest {
  code: string
  action: string
  type: string
  parent_id: string
  name: string
  path?: string
  component?: string
  redirect?: string
  icon?: string
  sort?: number
  status?: number
}

// 创建菜单响应
export interface CreateMenuResponse {
  menu_id: string
}

// 更新菜单请求
export interface UpdateMenuRequest {
  name?: string
  path?: string
  component?: string
  redirect?: string
  icon?: string
  sort?: number
  status?: number
  parent_id: string
}

export const menuApi = {
  // 获取菜单列表
  getList: (params: MenuListParams): Promise<MenuListResponse> => {
    return http.get('/admin/v1/menus', { params })
  },

  // 创建菜单
  create: (data: CreateMenuRequest): Promise<CreateMenuResponse> => {
    return http.post('/admin/v1/menus', data)
  },

  // 获取菜单详情
  getDetail: (menuId: string): Promise<MenuInfo> => {
    return http.get(`/admin/v1/menus/${menuId}`)
  },

  // 更新菜单
  update: (menuId: string, data: UpdateMenuRequest): Promise<boolean> => {
    return http.put(`/admin/v1/menus/${menuId}`, data)
  },

  // 删除菜单
  delete: (menuId: string): Promise<boolean> => {
    return http.delete(`/admin/v1/menus/${menuId}`, {})
  },

  // 获取所有菜单
  getAllMenu: (): Promise<MenuTreeResponse> => {
    return http.get('/admin/v1/menus/all')
  },

  // 获取菜单树
  getMenuTree: (): Promise<MenuTreeResponse> => {
    return http.get('/admin/v1/menus/tree')
  }
}

