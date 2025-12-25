import http from './http'
import type { RoleInfo } from './role'

// 租户成员信息
export interface TenantMemberInfo {
  user_id: string
  username: string
  name: string
  phone?: string
  email?: string
  status: number
  role_ids: string[]
  first_login: boolean
  last_login_time: number
  created_at: number
}

// 租户成员列表请求参数
export interface TenantMemberListParams {
  page?: number
  page_size?: number
  keyword?: string
  status?: number
}

// 租户成员列表响应
export interface TenantMemberListResponse {
  list: TenantMemberInfo[]
  page: number
  page_size: number
  total: number
  total_page: number
}

// 添加租户成员请求
export interface AddTenantMemberRequest {
  username: string
  name: string
  phone?: string
  email?: string
  role_ids: string[]
}

// 添加租户成员响应
export interface AddTenantMemberResponse {
  user_id: string
  username: string
  name: string
  initial_password: string // ⚠️ 只返回一次
  tenant_id: string
  role_ids: string[]
}

// 更新成员角色请求
export interface UpdateMemberRolesRequest {
  user_id: string
  role_ids: string[]
}

// 更新成员角色响应
export interface UpdateMemberRolesResponse {
  user_id: string
  role_ids: string[]
}

// 移除租户成员请求
export interface RemoveTenantMemberRequest {
  user_id: string
}

// 获取角色列表响应（用于角色选择）
export interface RolesForSelectionResponse {
  list: RoleInfo[]
}

export const tenantMemberApi = {
  // 获取租户成员列表
  getList: (params: TenantMemberListParams): Promise<TenantMemberListResponse> => {
    return http.get('/api/v1/tenant-members', { params })
  },

  // 添加租户成员
  add: (data: AddTenantMemberRequest): Promise<AddTenantMemberResponse> => {
    return http.post('/api/v1/tenant-members', data)
  },

  // 移除租户成员（请求体方式）
  remove: (data: RemoveTenantMemberRequest): Promise<boolean> => {
    return http.post('/api/v1/tenant-members/remove', data)
  },

  // 移除租户成员（路径参数方式）
  removeByPath: (userId: string): Promise<boolean> => {
    return http.delete(`/api/v1/tenant-members/${userId}`)
  },

  // 更新成员角色
  updateRoles: (data: UpdateMemberRolesRequest): Promise<UpdateMemberRolesResponse> => {
    return http.put('/api/v1/tenant-members/roles', data)
  }
}
