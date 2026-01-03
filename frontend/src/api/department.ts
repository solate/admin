import http from './http'

// 部门信息
export interface DepartmentInfo {
  department_id: string
  parent_id: string
  department_name: string
  description: string
  sort: number
  status: number
  created_at: number
  updated_at: number
}

// 部门树节点
export interface DepartmentTreeNode {
  department_id: string
  parent_id: string
  department_name: string
  description: string
  sort: number
  status: number
  created_at: number
  updated_at: number
  children?: DepartmentTreeNode[]
}

// 部门列表请求参数
export interface DepartmentListParams {
  page?: number
  page_size?: number
  keyword?: string
  status?: number
  parent_id?: string
}

// 部门列表响应
export interface DepartmentListResponse {
  list: DepartmentInfo[]
  page: number
  page_size: number
  total: number
  total_page: number
}

// 部门树响应
export interface DepartmentTreeResponse {
  tree: DepartmentTreeNode[]
}

// 创建部门请求
export interface CreateDepartmentRequest {
  parent_id?: string
  department_name: string
  description?: string
  sort?: number
  status?: number
}

// 创建部门响应
export interface DepartmentResponse {
  department_id: string
  parent_id: string
  department_name: string
  description: string
  sort: number
  status: number
  created_at: number
  updated_at: number
}

// 更新部门请求
export interface UpdateDepartmentRequest {
  parent_id?: string
  department_name?: string
  description?: string
  sort?: number
  status?: number
}

// 子部门列表响应
export interface DepartmentChildrenResponse {
  list: DepartmentInfo[]
}

export const departmentApi = {
  // 获取部门列表
  getList: (params: DepartmentListParams): Promise<DepartmentListResponse> => {
    return http.get('/api/v1/departments', { params })
  },

  // 获取部门树
  getTree: (): Promise<DepartmentTreeResponse> => {
    return http.get('/api/v1/departments/tree')
  },

  // 获取子部门
  getChildren: (departmentId: string): Promise<DepartmentChildrenResponse> => {
    return http.get(`/api/v1/departments/${departmentId}/children`)
  },

  // 创建部门
  create: (data: CreateDepartmentRequest): Promise<DepartmentResponse> => {
    return http.post('/api/v1/departments', data)
  },

  // 获取部门详情
  getDetail: (departmentId: string): Promise<DepartmentResponse> => {
    return http.get(`/api/v1/departments/${departmentId}`)
  },

  // 更新部门
  update: (departmentId: string, data: UpdateDepartmentRequest): Promise<DepartmentResponse> => {
    return http.put(`/api/v1/departments/${departmentId}`, data)
  },

  // 删除部门
  delete: (departmentId: string): Promise<boolean> => {
    return http.delete(`/api/v1/departments/${departmentId}`)
  },

  // 更新部门状态
  updateStatus: (departmentId: string, status: number): Promise<boolean> => {
    return http.put(`/api/v1/departments/${departmentId}/status/${status}`)
  }
}
