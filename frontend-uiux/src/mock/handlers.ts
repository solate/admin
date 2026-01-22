/**
 * Mock 数据处理器
 * 统一管理所有 Mock 数据和模拟 API 响应
 */

import type { LoginRequest, RegisterRequest, ListParams } from '@/types/api'
import type { User, Tenant, Service, DashboardStats, AuditLog } from '@/types/models'

// 模拟延迟
const delay = (ms: number = 500) => new Promise(resolve => setTimeout(resolve, ms))

// 生成 Mock Token
const generateToken = () => `mock-jwt-token-${Date.now()}`

// Mock 用户数据
const mockUsers: User[] = [
  {
    id: '1',
    name: 'Admin User',
    email: 'admin@example.com',
    role: 'admin',
    tenantId: 'tenant-1'
  },
  {
    id: '2',
    name: 'Regular User',
    email: 'user@example.com',
    role: 'user',
    tenantId: 'tenant-2'
  }
]

// Mock 租户数据
const mockTenants: Tenant[] = [
  {
    id: 'tenant-1',
    name: '企业 A',
    domain: 'company-a.example.com',
    status: 'active',
    plan: 'enterprise',
    maxUsers: 100,
    currentUserCount: 45,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-15T00:00:00Z'
  },
  {
    id: 'tenant-2',
    name: '企业 B',
    domain: 'company-b.example.com',
    status: 'active',
    plan: 'pro',
    maxUsers: 50,
    currentUserCount: 23,
    createdAt: '2024-01-05T00:00:00Z',
    updatedAt: '2024-01-10T00:00:00Z'
  },
  {
    id: 'tenant-3',
    name: '企业 C',
    domain: 'company-c.example.com',
    status: 'pending',
    plan: 'free',
    maxUsers: 10,
    currentUserCount: 3,
    createdAt: '2024-01-20T00:00:00Z',
    updatedAt: '2024-01-20T00:00:00Z'
  }
]

// Mock 服务数据
const mockServices: Service[] = [
  {
    id: 'svc-1',
    name: 'API 服务',
    description: '核心 API 服务',
    status: 'running',
    endpoint: 'https://api.example.com',
    config: { timeout: 30000, retries: 3 },
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-15T00:00:00Z'
  },
  {
    id: 'svc-2',
    name: '数据库服务',
    description: '主数据库服务',
    status: 'running',
    endpoint: 'postgresql://db.example.com:5432',
    config: { poolSize: 20, backup: true },
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-10T00:00:00Z'
  },
  {
    id: 'svc-3',
    name: '缓存服务',
    description: 'Redis 缓存',
    status: 'running',
    endpoint: 'redis://cache.example.com:6379',
    config: { ttl: 3600, maxSize: 1000 },
    createdAt: '2024-01-05T00:00:00Z',
    updatedAt: '2024-01-12T00:00:00Z'
  }
]

/**
 * 认证相关 Mock API
 */
export const authMock = {
  async login(credentials: LoginRequest) {
    await delay()

    const user = mockUsers.find(u => u.email === credentials.email)
    if (!user) {
      throw new Error('用户不存在')
    }

    if (credentials.password !== 'password') {
      throw new Error('密码错误')
    }

    return {
      user,
      token: generateToken(),
      refreshToken: `refresh-${generateToken()}`
    }
  },

  async register(data: RegisterRequest) {
    await delay()

    const existingUser = mockUsers.find(u => u.email === data.email)
    if (existingUser) {
      throw new Error('邮箱已被注册')
    }

    const newUser: User = {
      id: Date.now().toString(),
      name: data.name,
      email: data.email,
      role: 'user',
      tenantId: null
    }

    mockUsers.push(newUser)

    return {
      user: newUser,
      token: generateToken(),
      refreshToken: `refresh-${generateToken()}`
    }
  },

  async logout() {
    await delay(200)
    return { success: true }
  },

  async me(token: string) {
    await delay()

    if (!token || !token.startsWith('mock-jwt-token')) {
      throw new Error('无效的 token')
    }

    return mockUsers[0]
  },

  async refreshToken(refreshToken: string) {
    await delay()

    if (!refreshToken) {
      throw new Error('无效的 refresh token')
    }

    return {
      token: generateToken()
    }
  }
}

/**
 * 租户相关 Mock API
 */
export const tenantMock = {
  async list(params?: ListParams) {
    await delay()

    const page = params?.page || 1
    const pageSize = params?.pageSize || 10
    const start = (page - 1) * pageSize
    const end = start + pageSize

    return {
      items: mockTenants.slice(start, end),
      total: mockTenants.length,
      page,
      pageSize,
      totalPages: Math.ceil(mockTenants.length / pageSize)
    }
  },

  async getAll() {
    await delay()
    return mockTenants
  },

  async detail(id: string) {
    await delay()

    const tenant = mockTenants.find(t => t.id === id)
    if (!tenant) {
      throw new Error('租户不存在')
    }

    return tenant
  },

  async create(data: Partial<Tenant>) {
    await delay()

    const newTenant: Tenant = {
      id: `tenant-${Date.now()}`,
      name: data.name || '新租户',
      domain: data.domain || 'new.example.com',
      status: 'pending',
      plan: 'free',
      maxUsers: 10,
      currentUserCount: 0,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    }

    mockTenants.push(newTenant)
    return newTenant
  },

  async update(id: string, data: Partial<Tenant>) {
    await delay()

    const index = mockTenants.findIndex(t => t.id === id)
    if (index === -1) {
      throw new Error('租户不存在')
    }

    mockTenants[index] = { ...mockTenants[index], ...data, updatedAt: new Date().toISOString() }
    return mockTenants[index]
  },

  async delete(id: string) {
    await delay()

    const index = mockTenants.findIndex(t => t.id === id)
    if (index === -1) {
      throw new Error('租户不存在')
    }

    mockTenants.splice(index, 1)
  },

  async updateStatus(id: string, status: Tenant['status']) {
    await delay()

    const tenant = mockTenants.find(t => t.id === id)
    if (!tenant) {
      throw new Error('租户不存在')
    }

    tenant.status = status
    tenant.updatedAt = new Date().toISOString()
    return tenant
  }
}

/**
 * 用户相关 Mock API
 */
export const userMock = {
  async list(params?: ListParams) {
    await delay()

    const page = params?.page || 1
    const pageSize = params?.pageSize || 10
    const start = (page - 1) * pageSize
    const end = start + pageSize

    return {
      items: mockUsers.slice(start, end),
      total: mockUsers.length,
      page,
      pageSize,
      totalPages: Math.ceil(mockUsers.length / pageSize)
    }
  },

  async detail(id: string) {
    await delay()

    const user = mockUsers.find(u => u.id === id)
    if (!user) {
      throw new Error('用户不存在')
    }

    return user
  },

  async create(data: Partial<User>) {
    await delay()

    const newUser: User = {
      id: Date.now().toString(),
      name: data.name || '新用户',
      email: data.email || '',
      role: data.role || 'user',
      tenantId: data.tenantId || null,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    }

    mockUsers.push(newUser)
    return newUser
  },

  async update(id: string, data: Partial<User>) {
    await delay()

    const index = mockUsers.findIndex(u => u.id === id)
    if (index === -1) {
      throw new Error('用户不存在')
    }

    mockUsers[index] = { ...mockUsers[index], ...data, updatedAt: new Date().toISOString() }
    return mockUsers[index]
  },

  async delete(id: string) {
    await delay()

    const index = mockUsers.findIndex(u => u.id === id)
    if (index === -1) {
      throw new Error('用户不存在')
    }

    mockUsers.splice(index, 1)
  },

  async assignRoles(id: string, roles: string[]) {
    await delay()

    const user = mockUsers.find(u => u.id === id)
    if (!user) {
      throw new Error('用户不存在')
    }

    // 这里简化处理，实际应该更新用户角色
    return user
  },

  async changePassword(data: { oldPassword: string; newPassword: string }) {
    await delay()
    // Mock 实现，实际应该验证旧密码
    return { success: true }
  },

  async profile() {
    await delay()
    return mockUsers[0]
  }
}

/**
 * 服务相关 Mock API
 */
export const serviceMock = {
  async list(params?: ListParams) {
    await delay()

    const page = params?.page || 1
    const pageSize = params?.pageSize || 10
    const start = (page - 1) * pageSize
    const end = start + pageSize

    return {
      items: mockServices.slice(start, end),
      total: mockServices.length,
      page,
      pageSize,
      totalPages: Math.ceil(mockServices.length / pageSize)
    }
  },

  async detail(id: string) {
    await delay()

    const service = mockServices.find(s => s.id === id)
    if (!service) {
      throw new Error('服务不存在')
    }

    return service
  },

  async create(data: Partial<Service>) {
    await delay()

    const newService: Service = {
      id: `svc-${Date.now()}`,
      name: data.name || '新服务',
      description: data.description || '',
      status: 'stopped',
      endpoint: data.endpoint || '',
      config: data.config || {},
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    }

    mockServices.push(newService)
    return newService
  },

  async update(id: string, data: Partial<Service>) {
    await delay()

    const index = mockServices.findIndex(s => s.id === id)
    if (index === -1) {
      throw new Error('服务不存在')
    }

    mockServices[index] = { ...mockServices[index], ...data, updatedAt: new Date().toISOString() }
    return mockServices[index]
  },

  async delete(id: string) {
    await delay()

    const index = mockServices.findIndex(s => s.id === id)
    if (index === -1) {
      throw new Error('服务不存在')
    }

    mockServices.splice(index, 1)
  },

  async toggle(id: string, enabled: boolean) {
    await delay()

    const service = mockServices.find(s => s.id === id)
    if (!service) {
      throw new Error('服务不存在')
    }

    service.status = enabled ? 'running' : 'stopped'
    service.updatedAt = new Date().toISOString()
    return service
  }
}

/**
 * Dashboard 相关 Mock API
 */
export const dashboardMock = {
  async stats(): Promise<DashboardStats> {
    await delay()

    return {
      totalTenants: mockTenants.length,
      totalUsers: mockUsers.length,
      totalServices: mockServices.length,
      activeServices: mockServices.filter(s => s.status === 'running').length,
      revenueThisMonth: 123456,
      revenueGrowth: 15.5
    }
  },

  async charts() {
    await delay()

    return {
      labels: ['2024-01-01', '2024-01-02', '2024-01-03', '2024-01-04', '2024-01-05', '2024-01-06', '2024-01-07'],
      datasets: [
        {
          label: '收入',
          data: [1200, 1900, 1500, 2100, 1800, 2400, 2200],
          backgroundColor: 'rgba(37, 99, 235, 0.2)',
          borderColor: 'rgb(37, 99, 235)'
        }
      ]
    }
  },

  async activity(params?: { limit?: number }): Promise<AuditLog[]> {
    await delay()

    const logs: AuditLog[] = [
      {
        id: '1',
        userId: '1',
        userName: 'Admin User',
        action: 'login',
        resource: 'auth',
        details: { ip: '192.168.1.1' },
        ipAddress: '192.168.1.1',
        userAgent: 'Mozilla/5.0',
        createdAt: '2024-01-15T10:30:00Z'
      },
      {
        id: '2',
        userId: '1',
        userName: 'Admin User',
        action: 'create',
        resource: 'tenant',
        resourceId: 'tenant-3',
        details: { name: '企业 C' },
        ipAddress: '192.168.1.1',
        userAgent: 'Mozilla/5.0',
        createdAt: '2024-01-15T09:15:00Z'
      }
    ]

    return params?.limit ? logs.slice(0, params.limit) : logs
  }
}

/**
 * 其他 Mock API
 */
export const roleMock = {
  async list(params?: ListParams) {
    await delay()
    return {
      items: [
        { id: '1', name: '管理员', description: '系统管理员', permissions: [], createdAt: '2024-01-01', updatedAt: '2024-01-01' },
        { id: '2', name: '普通用户', description: '普通用户角色', permissions: [], createdAt: '2024-01-01', updatedAt: '2024-01-01' }
      ],
      total: 2,
      page: 1,
      pageSize: 10,
      totalPages: 1
    }
  },

  async detail(id: string) {
    await delay()
    return { id, name: '管理员', description: '系统管理员', permissions: [], createdAt: '2024-01-01', updatedAt: '2024-01-01' }
  }
}

export const permissionMock = {
  async list() {
    await delay()
    return [
      { id: '1', name: '用户管理', resource: 'users', action: 'read', description: '查看用户' },
      { id: '2', name: '用户创建', resource: 'users', action: 'create', description: '创建用户' }
    ]
  }
}

export const auditLogMock = {
  async list(params?: ListParams) {
    await delay()
    return {
      items: [],
      total: 0,
      page: 1,
      pageSize: 10,
      totalPages: 0
    }
  }
}

export const notificationMock = {
  async list(params?: ListParams) {
    await delay()
    return {
      items: [],
      total: 0,
      page: 1,
      pageSize: 10,
      totalPages: 0
    }
  },

  async markAsRead(id: string) {
    await delay()
    return { success: true }
  },

  async markAllAsRead() {
    await delay()
    return { success: true }
  },

  async unreadCount() {
    await delay()
    return { count: 0 }
  }
}

export const settingMock = {
  async get(key: string) {
    await delay()
    return { key, value: '' }
  },

  async set(key: string, value: any) {
    await delay()
    return { key, value }
  },

  async getAll() {
    await delay()
    return {}
  }
}
