/**
 * 租户相关验证器
 */

import { REGEX_PATTERNS } from '@/config/constants'
import { TENANT_PLANS, TENANT_STATUS } from '@/config/constants'

/**
 * 验证租户名称
 */
export function validateTenantName(name: string): boolean {
  return name.length >= 2 && name.length <= 50
}

/**
 * 验证租户域名（子域名格式）
 */
export function validateTenantDomain(domain: string): boolean {
  // 子域名规则：小写字母、数字、连字符，2-63位
  const subdomainRegex = /^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$|^[a-z0-9]$/
  return subdomainRegex.test(domain)
}

/**
 * 验证租户计划
 */
export function validateTenantPlan(plan: string): boolean {
  return Object.values(TENANT_PLANS).includes(plan as any)
}

/**
 * 验证租户状态
 */
export function validateTenantStatus(status: string): boolean {
  return Object.values(TENANT_STATUS).includes(status as any)
}

/**
 * 验证租户 URL
 */
export function validateTenantUrl(url: string): boolean {
  return REGEX_PATTERNS.URL.test(url)
}

/**
 * 验证租户 Logo URL
 */
export function validateTenantLogoUrl(url: string): boolean {
  return REGEX_PATTERNS.URL.test(url) || url.startsWith('/')
}

/**
 * 验证租户配额
 */
export function validateTenantQuota(
  quota: {
    users?: number
    storage?: number
    apiCalls?: number
  }
): { isValid: boolean; errors: string[] } {
  const errors: string[] = []

  if (quota.users !== undefined) {
    if (quota.users < 0) {
      errors.push('用户配额不能为负数')
    }
    if (quota.users > 10000) {
      errors.push('用户配额不能超过 10000')
    }
  }

  if (quota.storage !== undefined) {
    if (quota.storage < 0) {
      errors.push('存储配额不能为负数')
    }
    // 转换为 GB
    const storageGB = quota.storage / (1024 * 1024 * 1024)
    if (storageGB > 1000) {
      errors.push('存储配额不能超过 1000GB')
    }
  }

  if (quota.apiCalls !== undefined) {
    if (quota.apiCalls < 0) {
      errors.push('API 调用配额不能为负数')
    }
    if (quota.apiCalls > 10000000) {
      errors.push('API 调用配额不能超过 1000万次')
    }
  }

  return {
    isValid: errors.length === 0,
    errors,
  }
}

/**
 * 验证租户数据
 */
export function validateTenantData(data: {
  name?: string
  domain?: string
  plan?: string
  status?: string
  url?: string
  logo?: string
}): {
  isValid: boolean
  errors: Record<string, string>
} {
  const errors: Record<string, string> = {}

  if (data.name && !validateTenantName(data.name)) {
    errors.name = '租户名称长度必须在 2-50 位之间'
  }

  if (data.domain && !validateTenantDomain(data.domain)) {
    errors.domain = '租户域名格式不正确（小写字母、数字、连字符，2-63位）'
  }

  if (data.plan && !validateTenantPlan(data.plan)) {
    errors.plan = '无效的租户计划'
  }

  if (data.status && !validateTenantStatus(data.status)) {
    errors.status = '无效的租户状态'
  }

  if (data.url && !validateTenantUrl(data.url)) {
    errors.url = 'URL 格式不正确'
  }

  if (data.logo && !validateTenantLogoUrl(data.logo)) {
    errors.logo = 'Logo URL 格式不正确'
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  }
}

/**
 * 生成租户域名
 * 基于租户名称生成可用的子域名
 */
export function generateTenantDomain(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
    .substring(0, 63)
}

/**
 * 计算租户计划的配额限制
 */
export function getTenantPlanQuota(plan: string) {
  const quotas = {
    free: {
      users: 3,
      storage: 1 * 1024 * 1024 * 1024, // 1GB
      apiCalls: 10000,
    },
    basic: {
      users: 10,
      storage: 10 * 1024 * 1024 * 1024, // 10GB
      apiCalls: 100000,
    },
    pro: {
      users: 50,
      storage: 100 * 1024 * 1024 * 1024, // 100GB
      apiCalls: 1000000,
    },
    enterprise: {
      users: -1, // 无限制
      storage: -1, // 无限制
      apiCalls: -1, // 无限制
    },
  }

  return quotas[plan as keyof typeof quotas] || quotas.free
}
