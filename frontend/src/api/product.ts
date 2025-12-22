import http from './http'
import type { PageResponse } from './factory'

// 商品信息
export interface ProductInfo {
  product_id: string
  product_name: string
  unit: string
  purchase_price: string
  sale_price: string
  current_stock: number
  min_stock: number
  status: number
  factory_id: string
  factory_name: string
  created_at: number
  updated_at: number
}

// 商品列表请求参数
export interface ProductListParams {
  page: number
  page_size: number
  product_name?: string
  factory_id?: string
  status?: number
}

// 商品列表响应
export interface ProductListResponse {
  page: PageResponse
  list: ProductInfo[]
}

// 创建商品请求
export interface CreateProductRequest {
  product_name: string
  unit: string
  purchase_price: string
  sale_price: string
  current_stock: number
  min_stock: number
  status: number
  factory_id?: string
}

// 创建商品响应
export interface CreateProductResponse {
  product_id: string
}

// 更新商品请求
export interface UpdateProductRequest {
  product_name?: string
  unit?: string
  purchase_price?: string
  sale_price?: string
  current_stock?: number
  min_stock?: number
  status?: number
  factory_id?: string
}

export const productApi = {
  // 获取商品列表
  getList: (params: ProductListParams): Promise<ProductListResponse> => {
    return http.get('/business/v1/products', { params })
  },

  // 创建商品
  create: (data: CreateProductRequest): Promise<CreateProductResponse> => {
    return http.post('/business/v1/products', data)
  },

  // 获取商品详情
  getDetail: (productId: string): Promise<ProductInfo> => {
    return http.get(`/business/v1/products/${productId}`)
  },

  // 更新商品
  update: (productId: string, data: UpdateProductRequest): Promise<boolean> => {
    return http.put(`/business/v1/products/${productId}`, data)
  },

  // 删除商品
  delete: (productId: string): Promise<boolean> => {
    return http.delete(`/business/v1/products/${productId}`, {})
  }
}
