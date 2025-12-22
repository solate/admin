import http from './http'
import type { PageResponse } from './factory'

// 库存信息
export interface InventoryInfo {
  inventory_id: string
  product_id: string
  product_name: string
  operation_type: string
  quantity: number
  unit_price: string
  total_amount: string
  operator_id: string
  operator_name: string
  remark: string
  operation_time: number
  before_stock: number
  after_stock: number
}

// 库存记录列表请求参数
export interface InventoryListParams {
  page: number
  page_size: number
  product_id?: string
  operation_type?: string
  operator_id?: string
  start_time?: string
  end_time?: string
}

// 库存记录列表响应
export interface InventoryListResponse {
  page: PageResponse
  list: InventoryInfo[]
}

// 商品入库请求
export interface ProductInRequest {
  product_id: string
  quantity: number
  unit_price: string
  operator_id: string
  remark: string
}

// 商品出库请求
export interface ProductOutRequest {
  product_id: string
  quantity: number
  unit_price: string
  operator_id: string
  remark: string
}

// 库存操作响应
export interface InventoryOperationResponse {
  inventory_id: string
  message: string
}

// 库存信息
export interface StockInfo {
  product_id: string
  product_name: string
  factory_id: string
  factory_name: string
  current_stock: number
  min_stock: number
  unit: string
  purchase_price: string
  sale_price: string
  status: number
  is_low_stock: boolean
}

// 商品库存请求参数
export interface ProductStockParams {
  product_id?: string
  factory_id?: string
}

// 商品库存响应
export interface ProductStockResponse {
  list: StockInfo[]
}

export const inventoryApi = {
  // 商品入库
  productIn: (data: ProductInRequest): Promise<InventoryOperationResponse> => {
    return http.post('/business/v1/inventory/in', data)
  },

  // 商品出库
  productOut: (data: ProductOutRequest): Promise<InventoryOperationResponse> => {
    return http.post('/business/v1/inventory/out', data)
  },

  // 获取库存记录列表
  getList: (params: InventoryListParams): Promise<InventoryListResponse> => {
    return http.get('/business/v1/inventory/list', { params })
  },

  // 获取商品库存信息
  getStock: (params: ProductStockParams): Promise<ProductStockResponse> => {
    return http.get('/business/v1/inventory/stock', { params })
  },

  // 获取库存操作历史
  getHistory: (params: InventoryListParams): Promise<InventoryListResponse> => {
    return http.get('/business/v1/inventory/history', { params })
  }
}

