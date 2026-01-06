import http from './http'
import type { PageResponse } from './factory'

// 角色列表信息
export interface RoleListInfo {
  role_id: string
  name: string
  code: string
  sort: number
}

// 用户信息（后端返回的原始结构）
export interface UserResponse {
  user_id: string
  username: string
  nickname: string
  avatar: string
  phone: string
  email: string
  status: number
  tenant_id: string
  last_login_time: number
  created_at: number
  updated_at: number
}

// 用户包装结构
export interface UserWrapper {
  user: UserResponse
}

// 用户信息（前端使用的简化结构）
export interface UserInfo {
  user_id: string
  user_name: string
  name: string
  phone: string
  email: string
  avatar: string
  status: number
  created_at: number
  role_list: RoleListInfo[]
}

// 用户列表请求参数
export interface UserListParams {
  page: number
  page_size: number
  name?: string
  phone?: string
  status?: number
}

// 用户列表响应（后端返回的原始结构）
export interface UserListResponse {
  page: PageResponse
  list: UserWrapper[]
}

// 创建用户请求
export interface CreateUserRequest {
  username: string
  nickname: string
  password: string
  status: number
  phone: string
  email?: string
  sex?: number
  avatar?: string
}

// 创建用户响应
export interface CreateUserResponse {
  user: UserResponse
}

// 更新用户请求
export interface UpdateUserRequest {
  nickname?: string
  email?: string
  status?: number
  phone?: string
  remark?: string
}

// 登录日志信息
export interface LoginLogInfo {
  log_id: string
  user_id: string
  user_name: string
  ip: string
  user_agent: string
  status: number
  message: string
  created_at: number
}

// 登录日志列表请求参数
export interface LoginLogListParams {
  page: number
  page_size: number
  user_name?: string
  ip?: string
  status?: number
  start_time?: string
  end_time?: string
}

// 登录日志列表响应
export interface LoginLogListResponse {
  page: PageResponse
  list: LoginLogInfo[]
}

// 重置密码请求
export interface ResetPasswordRequest {
  password?: string // 可选，为空则自动生成
}

// 重置密码响应
export interface ResetPasswordResponse {
  password?: string // 重置后的密码（仅显示一次）
  auto_generated: boolean // 是否自动生成
  message: string
}

// 角色信息（用户角色）
export interface UserRoleInfo {
  role_id: string
  role_code: string
  name: string
  description?: string
}

// 用户角色响应
export interface UserRolesResponse {
  user_id: string
  username: string
  roles: UserRoleInfo[]
}

// 分配角色请求
export interface AssignRolesRequest {
  role_codes: string[]
}

export const userApi = {
  // 获取用户列表
  getList: (params: UserListParams): Promise<UserListResponse> => {
    return http.get('/api/v1/users', { params })
  },

  // 创建用户
  create: (data: CreateUserRequest): Promise<CreateUserResponse> => {
    return http.post('/api/v1/users', data)
  },

  // 获取用户详情
  getDetail: (userId: string): Promise<UserInfo> => {
    return http.get(`/api/v1/users/${userId}`)
  },

  // 获取当前用户信息
  getCurrentUser: (): Promise<UserInfo> => {
    return http.get('/api/v1/users/me')
  },

  // 更新用户
  update: (userId: string, data: UpdateUserRequest): Promise<boolean> => {
    return http.put(`/api/v1/users/${userId}`, data)
  },

  // 删除用户
  delete: (userId: string): Promise<boolean> => {
    return http.delete(`/api/v1/users/${userId}`, {})
  },

  // 查询登录记录
  getLoginLogs: (params: LoginLogListParams): Promise<LoginLogListResponse> => {
    return http.get('/api/v1/login-logs', { params })
  },

  // 重置用户密码（超管功能）
  resetPassword: (userId: string, data: ResetPasswordRequest): Promise<ResetPasswordResponse> => {
    return http.post(`/api/v1/users/${userId}/password/reset`, data)
  },

  // 获取用户角色列表
  getUserRoles: (userId: string): Promise<UserRolesResponse> => {
    return http.get(`/api/v1/users/${userId}/roles`)
  },

  // 为用户分配角色（覆盖式）
  assignRoles: (userId: string, data: AssignRolesRequest): Promise<{ assigned: boolean }> => {
    return http.put(`/api/v1/users/${userId}/roles`, data)
  }
}

