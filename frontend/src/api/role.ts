import http from './http'

// 角色信息
export interface RoleInfo {
  role_id: string
  name: string
  role_code: string
  description?: string
  status: number
  tenant_id: string
  created_at: number
  updated_at: number
}

// 角色列表请求参数
export interface RoleListParams {
  page?: number
  page_size?: number
  keyword?: string
  status?: number
}

// 角色列表响应
export interface RoleListResponse {
  list: RoleInfo[]
  page: number
  page_size: number
  total: number
  total_page: number
}

// 创建角色请求
export interface CreateRoleRequest {
  name: string
  role_code: string
  description?: string
  status?: number
}

// 创建角色响应
export interface RoleResponse {
  role_id: string
  name: string
  role_code: string
  description?: string
  status: number
  tenant_id: string
  created_at: number
  updated_at: number
}

// 更新角色请求
export interface UpdateRoleRequest {
  name?: string
  description?: string
  status?: number
}

// 获取所有角色响应
export interface GetAllRolesResponse {
  list: RoleInfo[]
}

// 权限信息
export interface Permission {
  resource: string
  action: string
  type: string
}

// 获取角色权限响应
export interface GetRolePermissionsResponse {
  list: Permission[]
}

// 设置角色权限请求
export interface SetRolePermissionsRequest {
  permission_list: Permission[]
}

// 获取用户角色响应
export interface GetUserRolesResponse {
  list: RoleInfo[]
}

// 设置用户角色请求
export interface SetUserRolesRequest {
  role_code_list: string[]
}

export const roleApi = {
  // 获取角色列表
  getList: (params: RoleListParams): Promise<RoleListResponse> => {
    return http.get('/api/v1/roles', { params })
  },

  // 创建角色
  create: (data: CreateRoleRequest): Promise<RoleResponse> => {
    return http.post('/api/v1/roles', data)
  },

  // 获取角色详情
  getDetail: (roleId: string): Promise<RoleResponse> => {
    return http.get(`/api/v1/roles/${roleId}`)
  },

  // 更新角色
  update: (roleId: string, data: UpdateRoleRequest): Promise<RoleResponse> => {
    return http.put(`/api/v1/roles/${roleId}`, data)
  },

  // 删除角色
  delete: (roleId: string): Promise<boolean> => {
    return http.delete(`/api/v1/roles/${roleId}`)
  },

  // 更新角色状态
  updateStatus: (roleId: string, status: number): Promise<boolean> => {
    return http.put(`/api/v1/roles/${roleId}/status/${status}`)
  },

  // 获取所有角色列表
  getAllRoles: (): Promise<GetAllRolesResponse> => {
    return http.get('/api/v1/roles')
  }
}

