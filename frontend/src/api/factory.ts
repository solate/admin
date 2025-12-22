import http from './http'

// 分页信息
export interface PageResponse {
  total: number
  page_size: number
  request_page_size: number
  current: number
}

// 工厂信息
export interface FactoryInfo {
  factory_id: string
  factory_name: string
  address: string
  contact_phone: string
  status: number
  created_at: number
  updated_at: number
}

// 工厂列表请求参数
export interface FactoryListParams {
  page: number
  page_size: number
  factory_name?: string
  status?: number
}

// 工厂列表响应
export interface FactoryListResponse {
  page: PageResponse
  list: FactoryInfo[]
}

// 创建工厂请求
export interface CreateFactoryRequest {
  factory_name: string
  address: string
  contact_phone: string
  status: number
}

// 创建工厂响应
export interface CreateFactoryResponse {
  factory_id: string
}

// 更新工厂请求
export interface UpdateFactoryRequest {
  factory_name?: string
  address?: string
  contact_phone?: string
  status?: number
}

export const factoryApi = {
  // 获取工厂列表
  getList: (params: FactoryListParams): Promise<FactoryListResponse> => {
    return http.get('/business/v1/factories', { params })
  },

  // 创建工厂
  create: (data: CreateFactoryRequest): Promise<CreateFactoryResponse> => {
    return http.post('/business/v1/factories', data)
  },

  // 获取工厂详情
  getDetail: (factoryId: string): Promise<FactoryInfo> => {
    return http.get(`/business/v1/factories/${factoryId}`)
  },

  // 更新工厂
  update: (factoryId: string, data: UpdateFactoryRequest): Promise<boolean> => {
    return http.put(`/business/v1/factories/${factoryId}`, data)
  },

  // 删除工厂
  delete: (factoryId: string): Promise<boolean> => {
    return http.delete(`/business/v1/factories/${factoryId}`, {})
  }
}
