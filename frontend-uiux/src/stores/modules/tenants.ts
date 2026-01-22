// Tenants store

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/utils/request'
import { tenantStorage } from '@/utils/storage'
import type { Tenant } from '@/types/models'

export const useTenantsStore = defineStore('tenants', () => {
  // State
  const tenants = ref<Tenant[]>([])
  const currentTenant = ref<Tenant | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Computed
  const activeTenants = computed(() =>
    tenants.value.filter((t) => t.status === 'active')
  )

  const tenantOptions = computed(() =>
    tenants.value.map((t) => ({
      label: t.name,
      value: t.id,
      status: t.status,
      domain: t.domain
    }))
  )

  const currentTenantId = computed(() => currentTenant.value?.id || null)

  // Actions
  async function fetchTenants() {
    isLoading.value = true
    error.value = null

    try {
      const response = await api.get('/tenants')
      tenants.value = response.data || []

      // Set current tenant from localStorage or first active tenant
      if (!currentTenant.value) {
        const savedTenantId = tenantStorage.getCurrentTenantId()
        if (
          savedTenantId &&
          tenants.value.find((t) => t.id === savedTenantId)
        ) {
          currentTenant.value = tenants.value.find((t) => t.id === savedTenantId) || null
        } else if (activeTenants.value.length > 0) {
          currentTenant.value = activeTenants.value[0]
        }
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch tenants'
      console.error('Error fetching tenants:', err)
    } finally {
      isLoading.value = false
    }
  }

  async function fetchActiveTenants() {
    isLoading.value = true
    error.value = null

    try {
      const response = await api.get('/tenants/all')
      tenants.value = response.data || []

      if (!currentTenant.value && tenants.value.length > 0) {
        currentTenant.value = tenants.value[0]
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch tenants'
      console.error('Error fetching active tenants:', err)
    } finally {
      isLoading.value = false
    }
  }

  function setCurrentTenant(tenant: Tenant | null) {
    currentTenant.value = tenant
    if (tenant) {
      tenantStorage.setCurrentTenantId(tenant.id)
      api.defaults.headers['X-Tenant-ID'] = tenant.id
    } else {
      tenantStorage.removeCurrentTenantId()
      delete api.defaults.headers['X-Tenant-ID']
    }
  }

  function switchTenant(tenantId: string) {
    const tenant = tenants.value.find((t) => t.id === tenantId)
    if (tenant) {
      setCurrentTenant(tenant)
    }
  }

  function getTenantById(id: string): Tenant | undefined {
    return tenants.value.find((t) => t.id === id)
  }

  function getTenantByDomain(domain: string): Tenant | undefined {
    return tenants.value.find((t) => t.domain === domain)
  }

  async function createTenant(tenantData: Partial<Tenant>): Promise<Tenant> {
    isLoading.value = true
    error.value = null

    try {
      const response = await api.post('/tenants', tenantData)
      const newTenant = response.data
      tenants.value.push(newTenant)
      return newTenant
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create tenant'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function updateTenant(
    id: string,
    updates: Partial<Tenant>
  ): Promise<void> {
    isLoading.value = true
    error.value = null

    try {
      const response = await api.put(`/tenants/${id}`, updates)
      const index = tenants.value.findIndex((t) => t.id === id)
      if (index !== -1) {
        tenants.value[index] = { ...tenants.value[index], ...response.data }
        // Update current tenant if it's the same
        if (currentTenant.value?.id === id) {
          currentTenant.value = tenants.value[index]
        }
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update tenant'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function deleteTenant(id: string): Promise<void> {
    isLoading.value = true
    error.value = null

    try {
      await api.delete(`/tenants/${id}`)
      const index = tenants.value.findIndex((t) => t.id === id)
      if (index !== -1) {
        tenants.value.splice(index, 1)
        // Reset current tenant if it was deleted
        if (currentTenant.value?.id === id && activeTenants.value.length > 0) {
          setCurrentTenant(activeTenants.value[0])
        }
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete tenant'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function updateTenantStatus(
    id: string,
    status: Tenant['status']
  ): Promise<void> {
    isLoading.value = true
    error.value = null

    try {
      const response = await api.put('/tenants/status', { id, status })
      const index = tenants.value.findIndex((t) => t.id === id)
      if (index !== -1) {
        tenants.value[index].status = status
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update tenant status'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  function getTenantsByPlan(plan: Tenant['plan']): Tenant[] {
    return tenants.value.filter((t) => t.plan === plan)
  }

  function getTenantsByStatus(status: Tenant['status']): Tenant[] {
    return tenants.value.filter((t) => t.status === status)
  }

  // Initialize
  function initialize() {
    const savedTenantId = tenantStorage.getCurrentTenantId()
    if (savedTenantId) {
      api.defaults.headers['X-Tenant-ID'] = savedTenantId
    }
  }

  return {
    // State
    tenants,
    currentTenant,
    isLoading,
    error,

    // Computed
    activeTenants,
    tenantOptions,
    currentTenantId,

    // Actions
    fetchTenants,
    fetchActiveTenants,
    setCurrentTenant,
    switchTenant,
    getTenantById,
    getTenantByDomain,
    createTenant,
    updateTenant,
    deleteTenant,
    updateTenantStatus,
    getTenantsByPlan,
    getTenantsByStatus,
    initialize
  }
})
