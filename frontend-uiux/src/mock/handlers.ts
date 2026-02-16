// Mock API handlers for development without backend

import type { LoginRequest, RegisterRequest, AuthResponse, ApiResponse, ListResponse } from '@/types/api'
import type { User, Tenant, Service, DashboardStats } from '@/types/models'
import { mockUsers, mockTenants, mockServices, mockDashboardStats, mockTokens } from './data'

// Simulate network delay
const delay = (ms: number = 300) => new Promise(resolve => setTimeout(resolve, ms))

// Generate a mock token
const generateToken = (email: string): string => {
  return 'mock-token-' + btoa(email) + '-' + Date.now()
}

// Find user by credentials
const findUserByCredentials = (email: string, password: string): User | null => {
  // For mock, accept any password for known emails, or any email with password "password"
  const user = mockUsers.find(u => u.email === email)
  if (user && (password === 'password' || password === '123456')) {
    return user
  }
  // Allow login with any email for demo purposes
  if (password === 'password' || password === '123456') {
    return {
      id: String(Date.now()),
      name: email.split('@')[0],
      email,
      role: email.includes('admin') ? 'admin' : 'user',
      tenantId: '1'
    }
  }
  return null
}

// Auth handlers
export const mockAuthHandlers = {
  async login(credentials: LoginRequest): Promise<AuthResponse> {
    await delay()

    const user = findUserByCredentials(credentials.email, credentials.password)
    if (!user) {
      throw new Error('Invalid email or password')
    }

    const token = generateToken(credentials.email)
    return {
      user,
      token,
      refreshToken: 'mock-refresh-token-' + Date.now()
    }
  },

  async logout(): Promise<void> {
    await delay(100)
  },

  async register(data: RegisterRequest): Promise<AuthResponse> {
    await delay()

    const existingUser = mockUsers.find(u => u.email === data.email)
    if (existingUser) {
      throw new Error('Email already exists')
    }

    const newUser: User = {
      id: String(Date.now()),
      name: data.name,
      email: data.email,
      role: 'user',
      tenantId: null
    }

    return {
      user: newUser,
      token: generateToken(data.email),
      refreshToken: 'mock-refresh-token-' + Date.now()
    }
  },

  async me(token: string): Promise<User> {
    await delay(100)

    // Extract email from mock token
    const match = token.match(/mock-token-([^-]+)-/)
    if (!match) {
      throw new Error('Invalid token')
    }

    const email = atob(match[1])
    const user = mockUsers.find(u => u.email === email)
    if (!user) {
      // Return a default user for any valid mock token
      return {
        id: '1',
        name: email.split('@')[0],
        email,
        role: email.includes('admin') ? 'admin' : 'user',
        tenantId: '1'
      }
    }
    return user
  }
}

// Tenant handlers
export const mockTenantHandlers = {
  async list(): Promise<ListResponse<Tenant>> {
    await delay()
    return {
      items: mockTenants,
      total: mockTenants.length,
      page: 1,
      pageSize: 10,
      totalPages: 1
    }
  },

  async get(id: string): Promise<Tenant> {
    await delay()
    const tenant = mockTenants.find(t => t.id === id)
    if (!tenant) {
      throw new Error('Tenant not found')
    }
    return tenant
  },

  async create(data: Partial<Tenant>): Promise<Tenant> {
    await delay()
    const newTenant: Tenant = {
      id: String(Date.now()),
      name: data.name || 'New Tenant',
      domain: data.domain || 'new.example.com',
      status: 'pending',
      plan: 'free',
      maxUsers: 5,
      currentUserCount: 0,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    }
    return newTenant
  }
}

// Service handlers
export const mockServiceHandlers = {
  async list(): Promise<ListResponse<Service>> {
    await delay()
    return {
      items: mockServices,
      total: mockServices.length,
      page: 1,
      pageSize: 10,
      totalPages: 1
    }
  },

  async get(id: string): Promise<Service> {
    await delay()
    const service = mockServices.find(s => s.id === id)
    if (!service) {
      throw new Error('Service not found')
    }
    return service
  }
}

// Dashboard handlers
export const mockDashboardHandlers = {
  async stats(): Promise<DashboardStats> {
    await delay()
    return mockDashboardStats
  }
}
