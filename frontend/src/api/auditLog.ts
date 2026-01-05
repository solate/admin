import http from './http'

// ==================== 登录日志 ====================

// 登录日志信息
export interface LoginLogInfo {
  log_id: string
  tenant_id: string
  user_id: string
  user_name: string
  operation_type: string // 操作类型: LOGIN=登录, LOGOUT=登出
  login_type: string // 登录方式: PASSWORD=密码, SSO=单点登录, OAUTH=第三方登录
  login_ip: string
  login_location: string
  user_agent: string
  status: number
  fail_reason: string
  created_at: number
}

// 登录日志列表请求参数
export interface LoginLogListParams {
  page?: number
  page_size?: number
  user_id?: string
  user_name?: string
  operation_type?: string // 操作类型筛选: LOGIN, LOGOUT
  login_type?: string // 登录方式筛选: PASSWORD, SSO, OAUTH
  status?: number
  start_date?: number
  end_date?: number
  ip_address?: string
}

// 登录日志列表响应
export interface LoginLogListResponse {
  page: number
  page_size: number
  total: number
  total_page: number
  list: LoginLogInfo[]
}

// ==================== 操作日志 ====================

// 操作日志信息
export interface OperationLogInfo {
  log_id: string
  tenant_id: string
  user_id: string
  user_name: string
  module: string
  operation_type: string
  resource_type: string
  resource_id: string
  resource_name: string
  request_method: string
  request_path: string
  request_params: string
  old_value: string
  new_value: string
  status: number
  error_message: string
  ip_address: string
  user_agent: string
  created_at: number
}

// 操作日志列表请求参数
export interface OperationLogListParams {
  page?: number
  page_size?: number
  module?: string
  operation_type?: string
  resource_type?: string
  user_name?: string
  status?: number
  start_date?: number
  end_date?: number
}

// 操作日志列表响应
export interface OperationLogListResponse {
  page: number
  page_size: number
  total: number
  total_page: number
  list: OperationLogInfo[]
}

export const auditLogApi = {
  // ==================== 登录日志 ====================

  // 获取登录日志列表
  getLoginLogs: (params: LoginLogListParams): Promise<LoginLogListResponse> => {
    return http.get('/api/v1/login-logs', { params })
  },

  // 获取登录日志详情
  getLoginLogDetail: (logId: string): Promise<LoginLogInfo> => {
    return http.get(`/api/v1/login-logs/${logId}`)
  },

  // ==================== 操作日志 ====================

  // 获取操作日志列表
  getOperationLogs: (params: OperationLogListParams): Promise<OperationLogListResponse> => {
    return http.get('/api/v1/operation-logs', { params })
  },

  // 获取操作日志详情
  getOperationLogDetail: (logId: string): Promise<OperationLogInfo> => {
    return http.get(`/api/v1/operation-logs/${logId}`)
  }
}
