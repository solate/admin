import http from './http'

// 租户信息
export interface TenantInfo {
  tenant_id: string
  name: string
  code: string
  description?: string
  status: number
  created_at: number
  updated_at: number
}

// 租户列表请求参数
export interface TenantListParams {
  page?: number
  page_size?: number
  code?: string
  name?: string
  status?: number
}

// 租户列表响应
export interface TenantListResponse {
  list: TenantInfo[]
  page: number
  page_size: number
  total: number
  total_page: number
}

// 创建租户请求
export interface CreateTenantRequest {
  name: string
  code: string
  description?: string
}

// 创建租户响应
export interface CreateTenantResponse {
  tenant_id: string
  name: string
  code: string
  description?: string
  status: number
}

// 租户详情响应
export interface TenantDetailResponse {
  tenant_id: string
  name: string
  code: string
  description?: string
  status: number
  created_at: number
  updated_at: number
}

// 更新租户请求
export interface UpdateTenantRequest {
  name?: string
  description?: string
  status?: number
}

// 更新租户状态响应
export interface UpdateTenantStatusResponse {
  tenant_id: string
  status: number
}

export const tenantApi = {
  // 获取租户列表
  getList: (params: TenantListParams): Promise<TenantListResponse> => {
    return http.get('/api/v1/tenants', { params })
  },

  // 创建租户
  create: (data: CreateTenantRequest): Promise<CreateTenantResponse> => {
    return http.post('/api/v1/tenants', data)
  },

  // 获取租户详情
  getDetail: (tenantId: string): Promise<TenantDetailResponse> => {
    return http.get(`/api/v1/tenants/${tenantId}`)
  },

  // 更新租户
  update: (tenantId: string, data: UpdateTenantRequest): Promise<TenantDetailResponse> => {
    return http.put(`/api/v1/tenants/${tenantId}`, data)
  },

  // 删除租户
  delete: (tenantId: string): Promise<boolean> => {
    return http.delete(`/api/v1/tenants/${tenantId}`)
  },

  // 更新租户状态
  updateStatus: (tenantId: string, status: number): Promise<UpdateTenantStatusResponse> => {
    return http.put(`/api/v1/tenants/${tenantId}/status/${status}`)
  }
}

