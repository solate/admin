import http from './http'

// 验证码响应
export interface CaptchaResponse {
  captcha_id: string
  captcha_data: string // Base64图片数据
}

// 租户信息
export interface TenantInfo {
  tenant_id: string
  tenant_code: string
  name: string
  description: string
}

// 角色信息
export interface RoleInfo {
  role_id: string
  role_code: string
  name: string
  description: string
}

// 登录请求
export interface LoginRequest {
  username: string
  password: string
  captcha_id: string
  captcha: string
  last_tenant_id?: string // 上次选择的租户ID
}

// 用户信息
export interface User {
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

// 登录响应（只返回 token）
export interface LoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
}

// 用户档案响应
export interface ProfileResponse {
  user: User
  tenant: TenantInfo
  roles: RoleInfo[]
}

// 选择租户请求
export interface SelectTenantRequest {
  tenant_id: string
}

// 选择租户响应
export interface SelectTenantResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  tenant: TenantInfo
  roles: RoleInfo[]
}

// 注册请求
export interface RegisterRequest {
  user_name: string
  password: string
  nick_name: string
  phone: string
  email?: string
  sex?: number
  avatar?: string
}

// 注册响应
export interface RegisterResponse {
  user_id: string
}

// 刷新Token请求
export interface RefreshTokenRequest {
  refresh_token: string
}

// 刷新Token响应
export interface RefreshTokenResponse {
  access_token: string
  refresh_token?: string
  tenant?: TenantInfo
  roles?: RoleInfo[]
}

// 切换租户请求
export interface SwitchTenantRequest {
  tenant_id: string
}

// 切换租户响应（只返回 token，需要调用 getProfile 获取完整信息）
export interface SwitchTenantResponse {
  access_token: string
  refresh_token: string
  expires_in: number
}

// 可用租户列表响应
export interface AvailableTenantsResponse {
  tenants: TenantInfo[]
}

// 修改密码请求
export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

// 重置密码请求
export interface ResetPasswordRequest {
  user_id: string
  new_password: string
}

// 活跃设备响应
export interface ActiveDevicesResponse {
  active_devices: number
}

export const authApi = {
  // 获取验证码(添加时间戳防止缓存)
  getCaptcha: (): Promise<CaptchaResponse> => {
    return http.get('/api/v1/auth/captcha', {
      params: { t: Date.now() }
    })
  },

  // 用户登录（需要提供租户编码）
  login: (tenantCode: string, data: LoginRequest): Promise<LoginResponse> => {
    return http.post(`/api/v1/auth/${tenantCode}/login`, data)
  },

  // 获取当前用户信息（支持传入 token 用于登录后立即获取用户信息）
  getProfile: (token?: string): Promise<ProfileResponse> => {
    if (token) {
      // 如果传入了 token，直接使用 axios 调用，绕过 http 实例的拦截器
      return http.get('/api/v1/profile', {
        headers: { Authorization: `Bearer ${token}` }
      })
    }
    // 否则使用默认的 http 实例（会自动从 localStorage 获取 token）
    return http.get('/api/v1/profile')
  },

  // 选择租户
  selectTenant: (userId: string, data: SelectTenantRequest): Promise<SelectTenantResponse> => {
    return http.post('/api/v1/auth/select-tenant', data, {
      headers: {
        'X-User-ID': userId
      }
    })
  },

  // 用户注册
  register: (data: RegisterRequest): Promise<RegisterResponse> => {
    return http.post('/admin/v1/auth/register', data)
  },

  // 刷新访问令牌
  refreshToken: (data: RefreshTokenRequest): Promise<RefreshTokenResponse> => {
    return http.post('/api/v1/auth/refresh', data)
  },

  // 用户登出
  logout: (): Promise<boolean> => {
    return http.post('/api/v1/auth/logout')
  },

  // 用户登出（所有设备）
  logoutAll: (): Promise<boolean> => {
    return http.post('/admin/v1/auth/logout-all')
  },

  // 修改密码
  changePassword: (data: ChangePasswordRequest): Promise<boolean> => {
    return http.post('/admin/v1/auth/change-password', data)
  },

  // 重置密码
  resetPassword: (data: ResetPasswordRequest): Promise<boolean> => {
    return http.post('/admin/v1/auth/reset-password', data)
  },

  // 获取当前用户活跃设备数量
  getActiveDevices: (): Promise<ActiveDevicesResponse> => {
    return http.get('/admin/v1/auth/devices/active')
  },

  // 切换租户
  switchTenant: (data: SwitchTenantRequest): Promise<SwitchTenantResponse> => {
    return http.post('/api/v1/auth/switch-tenant', data)
  },

  // 获取可切换的租户列表
  getAvailableTenants: (): Promise<AvailableTenantsResponse> => {
    return http.get('/api/v1/auth/available-tenants')
  }
}
