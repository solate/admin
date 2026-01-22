// Domain models

export interface User {
  id: string
  name: string
  email: string
  role: UserRole
  tenantId?: string | null
  avatar?: string
  createdAt?: string
  updatedAt?: string
}

export type UserRole = 'admin' | 'user' | 'moderator'

export interface Tenant {
  id: string
  name: string
  domain: string
  status: TenantStatus
  plan: TenantPlan
  maxUsers: number
  currentUserCount: number
  createdAt: string
  updatedAt: string
}

export type TenantStatus = 'active' | 'suspended' | 'pending'
export type TenantPlan = 'free' | 'pro' | 'enterprise'

export interface Service {
  id: string
  name: string
  description: string
  status: ServiceStatus
  endpoint: string
  config: Record<string, any>
  createdAt: string
  updatedAt: string
}

export type ServiceStatus = 'running' | 'stopped' | 'error' | 'deploying'

export interface Role {
  id: string
  name: string
  description: string
  permissions: Permission[]
  createdAt: string
  updatedAt: string
}

export interface Permission {
  id: string
  name: string
  resource: string
  action: string
  description?: string
}

export interface Notification {
  id: string
  title: string
  message: string
  type: NotificationType
  isRead: boolean
  createdAt: string
}

export type NotificationType = 'info' | 'success' | 'warning' | 'error'

export interface AuditLog {
  id: string
  userId: string
  userName: string
  action: string
  resource: string
  resourceId?: string
  details: Record<string, any>
  ipAddress?: string
  userAgent?: string
  createdAt: string
}

export interface DashboardStats {
  totalTenants: number
  totalUsers: number
  totalServices: number
  activeServices: number
  revenueThisMonth: number
  revenueGrowth: number
}

export interface ChartData {
  labels: string[]
  datasets: {
    label: string
    data: number[]
    backgroundColor?: string | string[]
    borderColor?: string | string[]
  }[]
}
