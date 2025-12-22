import http from './http'

// 统计响应
export interface ProductDetailStats {
  product_id: string
  product_name: string
  unit: string
  current_stock: number
  min_stock: number
  total_in_quantity: number
  total_out_quantity: number
  purchase_price: string
  sale_price: string
  stock_value: string
  status: number
}

export interface StatisticsResponse {
  total_products: number
  total_stock: number
  total_stock_value: string
  total_sales_value: string
  low_stock_products: number
  total_in_quantity: number
  total_in_amount: string
  total_out_quantity: number
  total_out_amount: string
  product_detail_list: ProductDetailStats[]
}

export const statsApi = {
  // 获取实时统计
  getStatistics: (): Promise<StatisticsResponse> => {
    return http.get('/business/v1/statistics')
  }
}

