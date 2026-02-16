// Mock data for development without backend

import type { User, Tenant, Service, DashboardStats } from '@/types/models'

// Mock users
export const mockUsers: User[] = [
  {
    id: '1',
    name: 'Admin User',
    email: 'admin@example.com',
    role: 'admin',
    tenantId: '1',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=admin',
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z'
  },
  {
    id: '2',
    name: 'Test User',
    email: 'user@example.com',
    role: 'user',
    tenantId: '2',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=user',
    createdAt: '2024-01-15T00:00:00Z',
    updatedAt: '2024-01-15T00:00:00Z'
  }
]

// Mock tenants
export const mockTenants: Tenant[] = [
  {
    id: '1',
    name: 'Acme Corporation',
    domain: 'acme.example.com',
    status: 'active',
    plan: 'enterprise',
    maxUsers: 100,
    currentUserCount: 45,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z'
  },
  {
    id: '2',
    name: 'Tech Startup',
    domain: 'tech.example.com',
    status: 'active',
    plan: 'pro',
    maxUsers: 25,
    currentUserCount: 12,
    createdAt: '2024-02-15T00:00:00Z',
    updatedAt: '2024-02-15T00:00:00Z'
  },
  {
    id: '3',
    name: 'Demo Company',
    domain: 'demo.example.com',
    status: 'pending',
    plan: 'free',
    maxUsers: 5,
    currentUserCount: 2,
    createdAt: '2024-03-01T00:00:00Z',
    updatedAt: '2024-03-01T00:00:00Z'
  }
]

// Mock services
export const mockServices: Service[] = [
  {
    id: '1',
    name: 'API Gateway',
    description: 'Main API gateway service',
    status: 'running',
    endpoint: 'https://api.example.com',
    config: { timeout: 30000, retries: 3 },
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z'
  },
  {
    id: '2',
    name: 'Auth Service',
    description: 'Authentication and authorization service',
    status: 'running',
    endpoint: 'https://auth.example.com',
    config: { jwtExpiry: '24h', refreshExpiry: '7d' },
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z'
  },
  {
    id: '3',
    name: 'Notification Service',
    description: 'Email and push notification service',
    status: 'stopped',
    endpoint: 'https://notify.example.com',
    config: { provider: 'sendgrid' },
    createdAt: '2024-01-15T00:00:00Z',
    updatedAt: '2024-02-01T00:00:00Z'
  }
]

// Mock dashboard stats
export const mockDashboardStats: DashboardStats = {
  totalTenants: 15,
  totalUsers: 234,
  totalServices: 8,
  activeServices: 6,
  revenueThisMonth: 45600,
  revenueGrowth: 12.5
}

// Mock tokens
export const mockTokens: Record<string, string> = {
  'admin@example.com': 'mock-admin-token-' + Date.now(),
  'user@example.com': 'mock-user-token-' + Date.now()
}
