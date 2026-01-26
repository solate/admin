/**
 * 租户上下文管理
 * 提供租户相关的业务逻辑和上下文管理
 */

import { computed } from 'vue'
import { useTenantsStore } from '@/stores/modules/tenants'
import { appConfig } from '@/config/app'

/**
 * 租户上下文管理类
 */
export class TenantContext {
  private static currentTenantId: string | null = null

  /**
   * 获取当前租户 ID
   */
  static getCurrentTenantId(): string | null {
    return (
      this.currentTenantId ||
      localStorage.getItem(appConfig.storage.tenantKey)
    )
  }

  /**
   * 设置当前租户 ID
   */
  static setCurrentTenantId(tenantId: string): void {
    this.currentTenantId = tenantId
    localStorage.setItem(appConfig.storage.tenantKey, tenantId)
  }

  /**
   * 清除当前租户 ID
   */
  static clearCurrentTenantId(): void {
    this.currentTenantId = null
    localStorage.removeItem(appConfig.storage.tenantKey)
  }

  /**
   * 获取当前租户信息
   */
  static getCurrentTenant() {
    const store = useTenantsStore()
    const tenantId = this.getCurrentTenantId()
    return tenantId ? store.tenants.find((t) => t.id === tenantId) : null
  }

  /**
   * 检查是否在租户上下文中
   */
  static isInTenantContext(): boolean {
    return !!this.getCurrentTenantId()
  }

  /**
   * 获取租户特定配置
   */
  static getTenantConfig() {
    const tenant = this.getCurrentTenant()
    if (!tenant) return null

    return {
      id: tenant.id,
      name: tenant.name,
      logo: tenant.logo,
      theme: tenant.theme,
      settings: tenant.settings || {},
    }
  }
}

/**
 * 租户上下文 Composable
 * 用于在 Vue 组件中访问租户上下文
 */
export function useTenantContext() {
  const tenantsStore = useTenantsStore()

  // 当前租户 ID
  const currentTenantId = computed(() => TenantContext.getCurrentTenantId())

  // 当前租户信息
  const currentTenant = computed(() => tenantsStore.currentTenant)

  // 是否在租户上下文中
  const isInContext = computed(() => TenantContext.isInTenantContext())

  // 所有可用租户列表
  const availableTenants = computed(() => tenantsStore.tenants)

  /**
   * 切换租户
   */
  const switchTenant = (tenantId: string) => {
    const tenant = availableTenants.value.find((t) => t.id === tenantId)
    if (tenant) {
      tenantsStore.setCurrentTenant(tenant)
      TenantContext.setCurrentTenantId(tenantId)
    }
  }

  /**
   * 退出租户上下文
   */
  const exitContext = () => {
    tenantsStore.setCurrentTenant(null)
    TenantContext.clearCurrentTenantId()
  }

  /**
   * 获取租户特定配置
   */
  const getTenantConfig = computed(() => {
    return currentTenant.value
      ? {
          id: currentTenant.value.id,
          name: currentTenant.value.name,
          logo: currentTenant.value.logo,
          theme: currentTenant.value.theme,
          settings: currentTenant.value.settings || {},
        }
      : null
  })

  return {
    currentTenantId,
    currentTenant,
    isInContext,
    availableTenants,
    switchTenant,
    exitContext,
    tenantConfig: getTenantConfig,
  }
}
