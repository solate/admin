import http from './http'
import type { PageResponse } from './factory'
import type { RoleListInfo } from './user'

// 角色信息
export interface RoleInfo {
  role_id: string
  name: string
  code: string
  description: string
  status: number
  sort: number
  created_at: number
}

// 角色列表请求参数
export interface RoleListParams {
  page: number
  page_size: number
  name?: string
  code?: string
  status?: number
}

// 角色列表响应
export interface RoleListResponse {
  page: PageResponse
  list: RoleInfo[]
}

// 创建角色请求
export interface CreateRoleRequest {
  name: string
  code: string
  description?: string
  status: number
  sort?: number
}

// 创建角色响应
export interface CreateRoleResponse {
  role_id: string
}

// 更新角色请求
export interface UpdateRoleRequest {
  name?: string
  description?: string
  status?: number
  sort?: number
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
    return http.get('/admin/v1/roles', { params })
  },

  // 创建角色
  create: (data: CreateRoleRequest): Promise<CreateRoleResponse> => {
    return http.post('/admin/v1/roles', data)
  },

  // 获取角色详情
  getDetail: (roleId: string): Promise<RoleInfo> => {
    return http.get(`/admin/v1/roles/${roleId}`)
  },

  // 更新角色
  update: (roleId: string, data: UpdateRoleRequest): Promise<boolean> => {
    return http.put(`/admin/v1/roles/${roleId}`, data)
  },

  // 删除角色
  delete: (roleId: string): Promise<boolean> => {
    return http.delete(`/admin/v1/roles/${roleId}`, {})
  },

  // 获取所有角色列表
  getAllRoles: (): Promise<GetAllRolesResponse> => {
    return http.get('/admin/v1/roles/all')
  },

  // 获取角色权限列表
  getRolePermissions: (roleCode: string): Promise<GetRolePermissionsResponse> => {
    return http.get(`/admin/v1/roles/${roleCode}/permissions`)
  },

  // 设置角色权限
  setRolePermissions: (roleCode: string, data: SetRolePermissionsRequest): Promise<boolean> => {
    return http.post(`/admin/v1/roles/${roleCode}/permissions`, data)
  },

  // 获取用户角色列表
  getUserRoles: (userId: string): Promise<GetUserRolesResponse> => {
    return http.get(`/admin/v1/users/${userId}/roles`)
  },

  // 设置用户角色
  setUserRoles: (userId: string, data: SetUserRolesRequest): Promise<boolean> => {
    return http.post(`/admin/v1/users/${userId}/roles`, data)
  }
}

