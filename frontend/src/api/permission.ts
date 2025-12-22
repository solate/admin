import http from './http'

// 资源类型信息
export interface ResourceTypeInfo {
  type: string
  actions: string[]
  data_rules: string[]
}

// 获取资源类型列表响应
export interface GetResourceTypesResponse {
  list: ResourceTypeInfo[]
}

export const permissionApi = {
  // 获取资源类型列表
  getResourceTypes: (): Promise<GetResourceTypesResponse> => {
    return http.get('/admin/v1/rules/resource-types')
  }
}

