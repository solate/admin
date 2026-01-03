// 导出类型（有冲突的模块不使用 export *）
export * from './auth'
export * from './role'
export * from './menu'
export * from './tenant'
export * from './department'
export * from './position'

// user 和 auditLog 有类型冲突，需要分别导入
export type {
  User,
  UserListParams,
  UserListResponse,
  CreateUserRequest,
  UpdateUserRequest
} from './user'

export type {
  LoginLogInfo,
  LoginLogListParams,
  LoginLogListResponse,
  OperationLogInfo,
  OperationLogListParams,
  OperationLogListResponse
} from './auditLog'

// 导出API对象
export { authApi } from './auth'
export { userApi } from './user'
export { roleApi } from './role'
export { menuApi } from './menu'
export { tenantApi } from './tenant'
export { auditLogApi } from './auditLog'
export { departmentApi } from './department'
export { positionApi } from './position'
