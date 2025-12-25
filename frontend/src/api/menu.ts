import http from './http'

// 菜单信息（与后端 DTO 对应）
export interface MenuInfo {
  permission_id: string
  tenant_id: string
  name: string
  type: string
  parent_id?: string
  resource?: string
  action?: string
  path?: string
  component?: string
  redirect?: string
  icon?: string
  sort?: number
  status: number
  description?: string
  created_at: number
  updated_at: number
}

// 菜单树节点
export interface MenuTreeNode {
  permission_id: string
  tenant_id: string
  name: string
  type: string
  parent_id?: string
  resource?: string
  action?: string
  path?: string
  component?: string
  redirect?: string
  icon?: string
  sort?: number
  status: number
  description?: string
  created_at: number
  updated_at: number
  children?: MenuTreeNode[]
}

// 菜单树响应
export interface MenuTreeResponse {
  list: MenuTreeNode[]
}

// 菜单列表请求参数
export interface MenuListParams {
  page?: number
  page_size?: number
  name?: string
  type?: string
  status?: number
}

// 分页响应
export interface PageResponse {
  total: number
  page: number
  page_size: number
}

// 菜单列表响应
export interface MenuListResponse {
  total: number
  page: number
  page_size: number
  list: MenuInfo[]
}

// 创建菜单请求
export interface CreateMenuRequest {
  name: string
  type: 'MENU' | 'BUTTON'
  parent_id?: string
  path?: string
  component?: string
  redirect?: string
  icon?: string
  sort?: number
  status?: number
  action?: string
  resource?: string
}

// 更新菜单请求
export interface UpdateMenuRequest {
  name?: string
  parent_id?: string
  path?: string
  component?: string
  redirect?: string
  icon?: string
  sort?: number
  status?: number
  action?: string
  resource?: string
}

export const menuApi = {
  // 获取菜单列表（分页）
  getList: (params: MenuListParams): Promise<MenuListResponse> => {
    return http.get('/api/v1/menus', { params })
  },

  // 创建菜单
  create: (data: CreateMenuRequest): Promise<MenuInfo> => {
    return http.post('/api/v1/menus', data)
  },

  // 获取菜单详情
  getDetail: (menuId: string): Promise<MenuInfo> => {
    return http.get(`/api/v1/menus/${menuId}`)
  },

  // 更新菜单
  update: (menuId: string, data: UpdateMenuRequest): Promise<MenuInfo> => {
    return http.put(`/api/v1/menus/${menuId}`, data)
  },

  // 删除菜单
  delete: (menuId: string): Promise<{ deleted: boolean }> => {
    return http.delete(`/api/v1/menus/${menuId}`)
  },

  // 获取所有菜单（平铺）
  getAllMenu: (): Promise<{ list: MenuInfo[] }> => {
    return http.get('/api/v1/menus/all')
  },

  // 获取菜单树
  getMenuTree: (): Promise<MenuTreeResponse> => {
    return http.get('/api/v1/menus/tree')
  },

  // 更新菜单状态
  updateStatus: (menuId: string, status: number): Promise<{ updated: boolean }> => {
    return http.put(`/api/v1/menus/${menuId}/status/${status}`)
  }
}
