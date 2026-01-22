// Tenant management composable

import { computed } from 'vue'
import { useTenantsStore } from '@/stores/modules/tenants'
import { tenantStorage } from '@/utils/storage'
import type { Tenant } from '@/types/models'

export function useTenant() {
  const tenantsStore = useTenantsStore()

  const tenants = computed(() => tenantsStore.tenants)
  const currentTenant = computed(() => tenantsStore.currentTenant)
  const currentTenantId = computed(() => tenantsStore.currentTenantId)
  const isLoading = computed(() => tenantsStore.isLoading)

  const initialize = async () => {
    await tenantsStore.initialize()
  }

  const fetchTenants = async () => {
    await tenantsStore.fetchTenants()
  }

  const setCurrentTenant = (tenant: Tenant | null) => {
    tenantsStore.setCurrentTenant(tenant)
  }

  const switchTenant = (tenantId: string) => {
    tenantsStore.switchTenant(tenantId)
  }

  const createTenant = async (data: Partial<Tenant>): Promise<Tenant> => {
    return await tenantsStore.createTenant(data)
  }

  const updateTenant = async (id: string, data: Partial<Tenant>): Promise<void> => {
    await tenantsStore.updateTenant(id, data)
  }

  const deleteTenant = async (id: string): Promise<void> => {
    await tenantsStore.deleteTenant(id)
  }

  const getTenantById = (id: string): Tenant | undefined => {
    return tenants.value.find((t) => t.id === id)
  }

  // Initialize from storage on composable creation
  const storedTenantId = tenantStorage.getCurrentTenantId()
  if (storedTenantId && !currentTenant.value) {
    initialize()
  }

  return {
    tenants,
    currentTenant,
    currentTenantId,
    isLoading,
    initialize,
    fetchTenants,
    setCurrentTenant,
    switchTenant,
    createTenant,
    updateTenant,
    deleteTenant,
    getTenantById
  }
}
