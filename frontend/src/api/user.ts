import http from './http'
import type { PageResponse } from './factory'

// 角色列表信息
export interface RoleListInfo {
  role_id: string
  name: string
  code: string
  sort: number
}

// 用户信息
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

// 用户列表响应
export interface UserListResponse {
  page: PageResponse
  list: UserInfo[]
}

// 创建用户请求
export interface CreateUserRequest {
  user_name: string
  name: string
  password: string
  status: number
  phone: string
  email?: string
  sex?: number
  avatar?: string
  role_ids?: string[]
}

// 创建用户响应
export interface CreateUserResponse {
  user_id: string
}

// 更新用户请求
export interface UpdateUserRequest {
  name?: string
  email?: string
  status?: number
  sex?: number
  avatar?: string
  role_code_list?: string[]
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

export const userApi = {
  // 获取用户列表
  getList: (params: UserListParams): Promise<UserListResponse> => {
    return http.get('/admin/v1/users', { params })
  },

  // 创建用户
  create: (data: CreateUserRequest): Promise<CreateUserResponse> => {
    return http.post('/admin/v1/users', data)
  },

  // 获取用户详情
  getDetail: (userId: string): Promise<UserInfo> => {
    return http.get(`/admin/v1/users/${userId}`)
  },

  // 获取当前用户信息
  getCurrentUser: (): Promise<UserInfo> => {
    return http.get('/admin/v1/users/me')
  },

  // 更新用户
  update: (userId: string, data: UpdateUserRequest): Promise<boolean> => {
    return http.put(`/admin/v1/users/${userId}`, data)
  },

  // 删除用户
  delete: (userId: string): Promise<boolean> => {
    return http.delete(`/admin/v1/users/${userId}`, {})
  },

  // 查询登录记录
  getLoginLogs: (params: LoginLogListParams): Promise<LoginLogListResponse> => {
    return http.get('/admin/v1/login-logs', { params })
  }
}

