import http from './http'
import type { PageResponse } from './factory'

// 字典类型信息
export interface DictTypeInfo {
  type_id: string
  name: string
  code: string
  description: string
  status: number
  created_at: number
}

// 字典类型列表请求参数
export interface DictTypeListParams {
  page: number
  page_size: number
  name?: string
  code?: string
  status?: number
}

// 字典类型列表响应
export interface DictTypeListResponse {
  page: PageResponse
  list: DictTypeInfo[]
}

// 创建字典类型请求
export interface CreateDictTypeRequest {
  name: string
  code: string
  description?: string
  status: number
}

// 创建字典类型响应
export interface CreateDictTypeResponse {
  type_id: string
}

// 更新字典类型请求
export interface UpdateDictTypeRequest {
  name?: string
  description?: string
  status?: number
}

// 字典项信息
export interface DictItemInfo {
  item_id: string
  type_code: string
  label: string
  value: string
  description: string
  sort: number
  status: number
  created_at: number
}

// 字典项列表请求参数
export interface DictItemListParams {
  page: number
  page_size: number
  label?: string
  status?: number
}

// 字典项列表响应
export interface DictItemListResponse {
  page: PageResponse
  list: DictItemInfo[]
}

// 字典项选项列表响应
export interface DictItemAllResponse {
  list: DictItemInfo[]
}

// 创建字典项请求
export interface CreateDictItemRequest {
  type_code: string
  label: string
  value: string
  description?: string
  sort?: number
  status: number
}

// 创建字典项响应
export interface CreateDictItemResponse {
  item_id: string
}

// 更新字典项请求
export interface UpdateDictItemRequest {
  type_code: string
  label?: string
  value?: string
  description?: string
  sort?: number
  status?: number
}

export const dictApi = {
  // 获取字典类型列表
  getDictTypeList: (params: DictTypeListParams): Promise<DictTypeListResponse> => {
    return http.get('/admin/v1/dict-types', { params })
  },

  // 创建字典类型
  createDictType: (data: CreateDictTypeRequest): Promise<CreateDictTypeResponse> => {
    return http.post('/admin/v1/dict-types', data)
  },

  // 获取字典类型详情
  getDictTypeDetail: (typeId: string): Promise<DictTypeInfo> => {
    return http.get(`/admin/v1/dict-types/${typeId}`)
  },

  // 更新字典类型
  updateDictType: (typeId: string, data: UpdateDictTypeRequest): Promise<boolean> => {
    return http.put(`/admin/v1/dict-types/${typeId}`, data)
  },

  // 删除字典类型
  deleteDictType: (typeId: string): Promise<boolean> => {
    return http.delete(`/admin/v1/dict-types/${typeId}`, {})
  },

  // 获取字典项列表
  getDictItemList: (typeCode: string, params: DictItemListParams): Promise<DictItemListResponse> => {
    return http.get(`/admin/v1/dict-types/${typeCode}/items`, { params })
  },

  // 创建字典项
  createDictItem: (typeCode: string, data: CreateDictItemRequest): Promise<CreateDictItemResponse> => {
    return http.post(`/admin/v1/dict-types/${typeCode}/items`, data)
  },

  // 获取字典项详情
  getDictItemDetail: (typeCode: string, itemId: string): Promise<DictItemInfo> => {
    return http.get(`/admin/v1/dict-types/${typeCode}/items/${itemId}`)
  },

  // 更新字典项
  updateDictItem: (typeCode: string, itemId: string, data: UpdateDictItemRequest): Promise<boolean> => {
    return http.put(`/admin/v1/dict-types/${typeCode}/items/${itemId}`, data)
  },

  // 删除字典项
  deleteDictItem: (typeCode: string, itemId: string): Promise<boolean> => {
    return http.delete(`/admin/v1/dict-types/${typeCode}/items/${itemId}`, {})
  },

  // 获取字典项选项列表
  getDictItemAll: (): Promise<DictItemAllResponse> => {
    return http.get('/admin/v1/items/all')
  }
}

