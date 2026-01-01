// 导出类型（有冲突的模块不使用 export *）
export * from './auth'
export * from './role'
export * from './menu'
export * from './permission'
export * from './dict'
export * from './tenant'
export * from './factory'
export * from './product'
export * from './inventory'
export * from './stats'

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
export { permissionApi } from './permission'
export { dictApi } from './dict'
export { tenantApi } from './tenant'
export { factoryApi } from './factory'
export { productApi } from './product'
export { inventoryApi } from './inventory'
export { statsApi } from './stats'
export { auditLogApi } from './auditLog'

