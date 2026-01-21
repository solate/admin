<script setup>
import { ref, computed, onMounted } from 'vue'
import { useTenantsStore } from '@/stores/tenants'
import { useAuthStore } from '@/stores/auth'
import icons from '@/components/icons/index.js'

const tenantsStore = useTenantsStore()
const authStore = useAuthStore()

const { Building, ChevronDown, Check, Sparkles } = icons

const isOpen = ref(false)
const searchQuery = ref('')

// Check if user can switch tenants (super_admin or auditor)
const canSwitchTenant = computed(() => {
  const role = authStore.userRole
  return role === 'super_admin' || role === 'auditor'
})

const filteredTenants = computed(() => {
  if (!searchQuery.value) {
    return tenantsStore.tenants
  }
  const query = searchQuery.value.toLowerCase()
  return tenantsStore.tenants.filter(t =>
    t.name?.toLowerCase().includes(query) ||
    t.code?.toLowerCase().includes(query)
  )
})

const currentTenantLabel = computed(() => {
  return tenantsStore.currentTenant?.name || 'Select Tenant'
})

const currentTenantStatus = computed(() => {
  return tenantsStore.currentTenant?.status || 'unknown'
})

const statusColorMap = {
  active: 'bg-emerald-500',
  trial: 'bg-amber-500',
  suspended: 'bg-red-500',
  unknown: 'bg-slate-400'
}

function selectTenant(tenantId) {
  tenantsStore.setCurrentTenant(tenantId)
  isOpen.value = false

  // Optionally refresh page data with new tenant context
  window.location.reload()
}

function getInitials(name) {
  if (!name) return '?'
  return name
    .split(' ')
    .map(n => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)
}

onMounted(() => {
  if (canSwitchTenant.value) {
    tenantsStore.fetchActiveTenants()
  }
})
</script>

<template>
  <div class="relative">
    <!-- Trigger Button -->
    <button
      v-if="canSwitchTenant"
      @click="isOpen = !isOpen"
      class="flex items-center gap-2 px-3 py-2 bg-white/80 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-700 transition-all cursor-pointer"
      aria-label="Switch tenant"
      aria-haspopup="listbox"
      :aria-expanded="isOpen"
    >
      <div class="w-8 h-8 bg-gradient-to-br from-primary-500 to-primary-600 rounded-lg flex items-center justify-center">
        <Building class="w-4 h-4 text-white" />
      </div>
      <div class="text-left hidden sm:block">
        <p class="text-xs font-medium text-slate-900 dark:text-slate-100">{{ currentTenantLabel }}</p>
        <p class="text-xs text-slate-500 dark:text-slate-400 capitalize">{{ currentTenantStatus }}</p>
      </div>
      <div class="flex items-center gap-1">
        <span :class="['w-2 h-2 rounded-full', statusColorMap[currentTenantStatus]]"></span>
        <ChevronDown :class="['w-4 h-4 text-slate-500 transition-transform', isOpen ? 'rotate-180' : '']" />
      </div>
    </button>

    <!-- Current Tenant Display (for non-admin users) -->
    <div
      v-else
      class="flex items-center gap-2 px-3 py-2 bg-slate-100 dark:bg-slate-800 rounded-xl"
    >
      <div class="w-8 h-8 bg-gradient-to-br from-primary-500 to-primary-600 rounded-lg flex items-center justify-center">
        <Building class="w-4 h-4 text-white" />
      </div>
      <div class="text-left">
        <p class="text-xs font-medium text-slate-900 dark:text-slate-100">{{ currentTenantLabel }}</p>
        <p class="text-xs text-slate-500 dark:text-slate-400 capitalize">{{ currentTenantStatus }}</p>
      </div>
      <span :class="['w-2 h-2 rounded-full', statusColorMap[currentTenantStatus]]"></span>
    </div>

    <!-- Dropdown Panel -->
    <Transition
      enter-active-class="transition ease-out duration-200"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition ease-in duration-150"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-if="isOpen && canSwitchTenant"
        class="absolute right-0 mt-2 w-72 bg-white dark:bg-slate-800 rounded-2xl shadow-xl border border-slate-200 dark:border-slate-700 z-50 overflow-hidden"
      >
        <!-- Header -->
        <div class="p-4 border-b border-slate-200 dark:border-slate-700">
          <div class="flex items-center gap-2 mb-3">
            <Sparkles class="w-5 h-5 text-primary-500" />
            <h3 class="font-semibold text-slate-900 dark:text-slate-100">Switch Tenant</h3>
          </div>
          <!-- Search Input -->
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search tenants..."
            class="w-full px-3 py-2 bg-slate-100 dark:bg-slate-700 border-0 rounded-lg text-sm text-slate-900 dark:text-slate-100 placeholder-slate-500 focus:ring-2 focus:ring-primary-500 outline-none"
          >
        </div>

        <!-- Tenant List -->
        <div
          class="max-h-64 overflow-y-auto"
          role="listbox"
        >
          <div
            v-for="tenant in filteredTenants"
            :key="tenant.id"
            @click="selectTenant(tenant.id)"
            role="option"
            :aria-selected="tenantsStore.currentTenant?.id === tenant.id"
            class="flex items-center gap-3 px-4 py-3 hover:bg-slate-50 dark:hover:bg-slate-700 cursor-pointer transition-colors"
            :class="{
              'bg-primary-50 dark:bg-primary-900/30': tenantsStore.currentTenant?.id === tenant.id
            }"
          >
            <!-- Avatar -->
            <div class="w-10 h-10 bg-gradient-to-br from-primary-500 to-primary-600 rounded-xl flex items-center justify-center flex-shrink-0">
              <span class="text-sm font-semibold text-white">{{ getInitials(tenant.name) }}</span>
            </div>

            <!-- Info -->
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-slate-900 dark:text-slate-100 truncate">{{ tenant.name }}</p>
              <p class="text-xs text-slate-500 dark:text-slate-400">{{ tenant.code }}</p>
            </div>

            <!-- Status & Check -->
            <div class="flex items-center gap-2">
              <span :class="['w-2 h-2 rounded-full', statusColorMap[tenant.status]]"></span>
              <Check
                v-if="tenantsStore.currentTenant?.id === tenant.id"
                class="w-5 h-5 text-primary-600 dark:text-primary-400"
              />
            </div>
          </div>

          <!-- Empty State -->
          <div
            v-if="filteredTenants.length === 0"
            class="px-4 py-8 text-center"
          >
            <Building class="w-12 h-12 text-slate-300 dark:text-slate-600 mx-auto mb-3" />
            <p class="text-sm text-slate-500 dark:text-slate-400">No tenants found</p>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-3 border-t border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-900/50">
          <p class="text-xs text-slate-500 dark:text-slate-400 text-center">
            Switching tenant will reload the page
          </p>
        </div>
      </div>
    </Transition>

    <!-- Overlay -->
    <div
      v-if="isOpen"
      @click="isOpen = false"
      class="fixed inset-0 z-40"
    ></div>
  </div>
</template>

<style scoped>
/* Custom scrollbar for tenant list */
.overflow-y-auto::-webkit-scrollbar {
  width: 4px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: transparent;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 2px;
}

.dark .overflow-y-auto::-webkit-scrollbar-thumb {
  background: #475569;
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}
</style>
