import http from './http'
import type { PageResponse } from './factory'

// 租户信息
export interface TenantInfo {
  tenant_id: string
  name: string
  code: string
  description: string
  status: number
}

// 租户列表请求参数
export interface TenantListParams {
  page: number
  page_size: number
  status: number
}

// 租户列表响应
export interface TenantListResponse {
  page: PageResponse
  list: TenantInfo[]
}

// 创建租户请求
export interface CreateTenantRequest {
  name: string
  code: string
  description?: string
  status?: number
}

// 创建租户响应
export interface CreateTenantResponse {
  tenant_id: string
}

// 租户详情响应
export interface GetTenantResponse {
  tenant_id: string
  name: string
  code: string
  description: string
  status: number
}

// 更新租户请求
export interface UpdateTenantRequest {
  name: string
  description: string
  status: number
}

export const tenantApi = {
  // 获取租户列表
  getList: (params: TenantListParams): Promise<TenantListResponse> => {
    return http.get('/admin/v1/tenants', { params })
  },

  // 创建租户
  create: (data: CreateTenantRequest): Promise<CreateTenantResponse> => {
    return http.post('/admin/v1/tenants', data)
  },

  // 获取租户详情
  getDetail: (tenantId: string): Promise<GetTenantResponse> => {
    return http.get(`/admin/v1/tenants/${tenantId}`)
  },

  // 更新租户
  update: (tenantId: string, data: UpdateTenantRequest): Promise<boolean> => {
    return http.put(`/admin/v1/tenants/${tenantId}`, data)
  },

  // 删除租户
  delete: (tenantId: string): Promise<boolean> => {
    return http.delete(`/admin/v1/tenants/${tenantId}`, {})
  }
}

