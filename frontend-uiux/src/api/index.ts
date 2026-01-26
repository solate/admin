// API module exports

export { api, request } from '@/utils/request'
export * from './modules/auth'
export * from './modules/tenants'
export * from './modules/users'
export * from './modules/services'
export * from './modules/dashboard'

export { api as default } from '@/utils/request'
