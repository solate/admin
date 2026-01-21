import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/services/api'

export const useTenantsStore = defineStore('tenants', () => {
  // State
  const tenants = ref([])
  const currentTenant = ref(null)
  const loading = ref(false)
  const error = ref(null)

  // Computed
  const activeTenants = computed(() =>
    tenants.value.filter(t => t.status === 'active')
  )

  const tenantOptions = computed(() =>
    tenants.value.map(t => ({
      label: t.name,
      value: t.id,
      status: t.status,
      code: t.code
    }))
  )

  // Actions
  async function fetchTenants() {
    loading.value = true
    error.value = null
    try {
      const response = await api.get('/tenants')
      tenants.value = response.data || []

      // Set current tenant from localStorage or first active tenant
      if (!currentTenant.value) {
        const savedTenantId = localStorage.getItem('currentTenantId')
        if (savedTenantId && tenants.value.find(t => t.id === savedTenantId)) {
          currentTenant.value = tenants.value.find(t => t.id === savedTenantId)
        } else if (activeTenants.value.length > 0) {
          currentTenant.value = activeTenants.value[0]
        }
      }
    } catch (err) {
      error.value = err.message || 'Failed to fetch tenants'
      console.error('Error fetching tenants:', err)
    } finally {
      loading.value = false
    }
  }

  async function fetchActiveTenants() {
    loading.value = true
    error.value = null
    try {
      const response = await api.get('/tenants/all')
      tenants.value = response.data || []

      if (!currentTenant.value && tenants.value.length > 0) {
        currentTenant.value = tenants.value[0]
      }
    } catch (err) {
      error.value = err.message || 'Failed to fetch tenants'
      console.error('Error fetching active tenants:', err)
    } finally {
      loading.value = false
    }
  }

  function setCurrentTenant(tenantId) {
    const tenant = tenants.value.find(t => t.id === tenantId)
    if (tenant) {
      currentTenant.value = tenant
      localStorage.setItem('currentTenantId', tenantId)

      // Update API headers with new tenant context
      api.defaults.headers['X-Tenant-ID'] = tenantId
    }
  }

  function getTenantById(id) {
    return tenants.value.find(t => t.id === id)
  }

  function getTenantByCode(code) {
    return tenants.value.find(t => t.code === code)
  }

  async function createTenant(tenantData) {
    loading.value = true
    error.value = null
    try {
      const response = await api.post('/tenants', tenantData)
      const newTenant = response.data
      tenants.value.push(newTenant)
      return newTenant
    } catch (err) {
      error.value = err.message || 'Failed to create tenant'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateTenant(id, updates) {
    loading.value = true
    error.value = null
    try {
      const response = await api.put(`/tenants/${id}`, updates)
      const index = tenants.value.findIndex(t => t.id === id)
      if (index !== -1) {
        tenants.value[index] = { ...tenants.value[index], ...response.data }
        // Update current tenant if it's the same
        if (currentTenant.value?.id === id) {
          currentTenant.value = tenants.value[index]
        }
      }
      return response.data
    } catch (err) {
      error.value = err.message || 'Failed to update tenant'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteTenant(id) {
    loading.value = true
    error.value = null
    try {
      await api.delete(`/tenants/${id}`)
      const index = tenants.value.findIndex(t => t.id === id)
      if (index !== -1) {
        tenants.value.splice(index, 1)
        // Reset current tenant if it was deleted
        if (currentTenant.value?.id === id && activeTenants.value.length > 0) {
          setCurrentTenant(activeTenants.value[0].id)
        }
      }
    } catch (err) {
      error.value = err.message || 'Failed to delete tenant'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateTenantStatus(id, status) {
    loading.value = true
    error.value = null
    try {
      const response = await api.put('/tenants/status', { id, status })
      const index = tenants.value.findIndex(t => t.id === id)
      if (index !== -1) {
        tenants.value[index].status = status
      }
      return response.data
    } catch (err) {
      error.value = err.message || 'Failed to update tenant status'
      throw err
    } finally {
      loading.value = false
    }
  }

  function getTenantsByPlan(plan) {
    return tenants.value.filter(t => t.plan === plan)
  }

  function getTenantsByStatus(status) {
    return tenants.value.filter(t => t.status === status)
  }

  // Initialize
  function initialize() {
    const savedTenantId = localStorage.getItem('currentTenantId')
    if (savedTenantId) {
      api.defaults.headers['X-Tenant-ID'] = savedTenantId
    }
  }

  return {
    // State
    tenants,
    currentTenant,
    loading,
    error,

    // Computed
    activeTenants,
    tenantOptions,

    // Actions
    fetchTenants,
    fetchActiveTenants,
    setCurrentTenant,
    getTenantById,
    getTenantByCode,
    createTenant,
    updateTenant,
    deleteTenant,
    updateTenantStatus,
    getTenantsByPlan,
    getTenantsByStatus,
    initialize
  }
})
