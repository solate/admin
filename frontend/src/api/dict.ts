import http from './http'

// 字典项信息
export interface DictItemInfo {
  item_id: string
  label: string
  value: string
  sort: number
  source: 'system' | 'custom'
}

// 字典信息（含字典项）
export interface DictInfo {
  type_id: string
  type_code: string
  type_name: string
  description?: string
  items: DictItemInfo[]
}

// 字典类型信息
export interface DictTypeInfo {
  type_id: string
  type_code: string
  type_name: string
  description?: string
  tenant_id: string
  created_at: number
  updated_at: number
}

// 字典类型列表请求参数
export interface DictTypeListParams {
  page?: number
  page_size?: number
  keyword?: string
}

// 字典类型列表响应
export interface DictTypeListResponse {
  list: DictTypeInfo[]
  page: number
  page_size: number
  total: number
  total_page: number
}

// 创建字典项请求
export interface CreateDictItemRequest {
  label: string
  value: string
  sort?: number
}

// 创建系统字典请求
export interface CreateSystemDictRequest {
  type_code: string
  type_name: string
  description?: string
  items: CreateDictItemRequest[]
}

// 更新字典项请求
export interface UpdateDictItemRequest {
  label: string
  sort?: number
}

// 批量更新字典项请求
export interface BatchUpdateDictItemsRequest {
  type_code: string
  items: UpdateDictItemRequest[]
}

// 更新系统字典请求
export interface UpdateSystemDictRequest {
  type_name?: string
  description?: string
  items?: CreateDictItemRequest[]
}

export const dictApi = {
  // 获取字典类型列表
  listTypes: (params: DictTypeListParams): Promise<DictTypeListResponse> => {
    return http.get('/api/v1/dict-types', { params })
  },

  // 获取系统字典类型列表（超管专用）
  listSystemTypes: (params: DictTypeListParams): Promise<DictTypeListResponse> => {
    return http.get('/api/v1/system/dict', { params })
  },

  // 获取字典详情（含字典项，已合并系统+覆盖）
  getDict: (typeCode: string): Promise<DictInfo> => {
    return http.get(`/api/v1/dict/${typeCode}`)
  },

  // 创建系统字典（超管专用）
  createSystemDict: (data: CreateSystemDictRequest): Promise<{ created: boolean }> => {
    return http.post('/api/v1/system/dict', data)
  },

  // 更新系统字典（超管专用）
  updateSystemDict: (
    typeCode: string,
    data: UpdateSystemDictRequest
  ): Promise<{ updated: boolean }> => {
    return http.put(`/api/v1/system/dict/${typeCode}`, data)
  },

  // 删除系统字典（超管专用）
  deleteSystemDict: (typeCode: string): Promise<{ deleted: boolean }> => {
    return http.delete(`/api/v1/system/dict/${typeCode}`)
  },

  // 批量更新字典项（租户覆盖）
  batchUpdateItems: (data: BatchUpdateDictItemsRequest): Promise<{ updated: boolean }> => {
    return http.put('/api/v1/dict/items', data)
  },

  // 恢复字典项系统默认值
  resetItem: (typeCode: string, value: string): Promise<{ reset: boolean }> => {
    return http.delete(`/api/v1/dict/${typeCode}/items/${value}`)
  },

  // 删除系统字典项（超管专用）
  deleteSystemDictItem: (typeCode: string, value: string): Promise<{ deleted: boolean }> => {
    return http.delete(`/api/v1/system/dict/${typeCode}/items/${value}`)
  }
}
