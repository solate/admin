import http from './http'

// 岗位信息
export interface PositionInfo {
  position_id: string
  position_code: string
  position_name: string
  level: number
  description: string
  sort: number
  status: number
  created_at: number
  updated_at: number
}

// 岗位列表请求参数
export interface PositionListParams {
  page?: number
  page_size?: number
  keyword?: string
  status?: number
}

// 岗位列表响应
export interface PositionListResponse {
  list: PositionInfo[]
  page: number
  page_size: number
  total: number
  total_page: number
}

// 所有岗位列表响应
export interface AllPositionsResponse {
  list: PositionInfo[]
}

// 创建岗位请求
export interface CreatePositionRequest {
  position_code: string
  position_name: string
  level?: number
  description?: string
  sort?: number
  status?: number
}

// 创建岗位响应
export interface PositionResponse {
  position_id: string
  position_code: string
  position_name: string
  level: number
  description: string
  sort: number
  status: number
  created_at: number
  updated_at: number
}

// 更新岗位请求
export interface UpdatePositionRequest {
  position_code?: string
  position_name?: string
  level?: number
  description?: string
  sort?: number
  status?: number
}

export const positionApi = {
  // 获取岗位列表
  getList: (params: PositionListParams): Promise<PositionListResponse> => {
    return http.get('/api/v1/positions', { params })
  },

  // 获取所有岗位（不分页）
  getAll: (): Promise<AllPositionsResponse> => {
    return http.get('/api/v1/positions/all')
  },

  // 创建岗位
  create: (data: CreatePositionRequest): Promise<PositionResponse> => {
    return http.post('/api/v1/positions', data)
  },

  // 获取岗位详情
  getDetail: (positionId: string): Promise<PositionResponse> => {
    return http.get(`/api/v1/positions/${positionId}`)
  },

  // 更新岗位
  update: (positionId: string, data: UpdatePositionRequest): Promise<PositionResponse> => {
    return http.put(`/api/v1/positions/${positionId}`, data)
  },

  // 删除岗位
  delete: (positionId: string): Promise<boolean> => {
    return http.delete(`/api/v1/positions/${positionId}`)
  },

  // 更新岗位状态
  updateStatus: (positionId: string, status: number): Promise<boolean> => {
    return http.put(`/api/v1/positions/${positionId}/status/${status}`)
  }
}
