<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useTenantsStore } from '@/stores/modules/tenants'
import { useAuthStore } from '@/stores/modules/auth'
import { Building, ChevronDown, Check } from 'lucide-vue-next'

const tenantsStore = useTenantsStore()
const authStore = useAuthStore()

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
    t.domain?.toLowerCase().includes(query)
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

function selectTenant(tenantId: string) {
  tenantsStore.setCurrentTenant(tenantsStore.getTenantById(tenantId) || null)
  isOpen.value = false
}

onMounted(() => {
  tenantsStore.fetchTenants()
})
</script>

<template>
  <div class="relative">
    <!-- Trigger Button -->
    <button
      v-if="canSwitchTenant"
      @click="isOpen = !isOpen"
      class="flex items-center gap-2 px-3 py-2 bg-white/80 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-700 transition-all cursor-pointer"
      aria-label="Select tenant"
    >
      <el-icon :size="18" class="text-primary-600 dark:text-primary-400">
        <Building />
      </el-icon>
      <span class="text-sm font-medium text-slate-900 dark:text-slate-100">{{ currentTenantLabel }}</span>
      <el-icon :size="14" class="text-slate-500" :class="{ 'rotate-180': isOpen }">
        <ChevronDown />
      </el-icon>
    </button>

    <!-- Dropdown -->
    <Transition
      enter-active-class="transition ease-out duration-200"
      enter-from-class="opacity-0 scale-95 translate-y-2"
      enter-to-class="opacity-100 scale-100 translate-y-0"
      leave-active-class="transition ease-in duration-150"
      leave-from-class="opacity-100 scale-100 translate-y-0"
      leave-to-class="opacity-0 scale-95 translate-y-2"
    >
      <div
        v-if="isOpen && canSwitchTenant"
        @click.outside="isOpen = false"
        class="absolute right-0 mt-2 w-72 bg-white dark:bg-slate-800 rounded-xl shadow-xl border border-slate-200 dark:border-slate-700 z-50 overflow-hidden"
      >
        <!-- Search -->
        <div class="p-3 border-b border-slate-200 dark:border-slate-700">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="搜索租户..."
            class="w-full px-3 py-2 bg-slate-100 dark:bg-slate-700 border-0 rounded-lg text-sm text-slate-900 dark:text-slate-100 placeholder:text-slate-500 focus:ring-2 focus:ring-primary-500 outline-none"
          />
        </div>

        <!-- Tenant List -->
        <div class="max-h-64 overflow-y-auto p-2">
          <button
            v-for="tenant in filteredTenants"
            :key="tenant.id"
            @click="selectTenant(tenant.id)"
            class="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors cursor-pointer text-left group"
            :class="{
              'bg-primary-50 dark:bg-primary-900/20': tenantsStore.currentTenantId === tenant.id
            }"
          >
            <!-- Status Indicator -->
            <div
              class="w-2 h-2 rounded-full"
              :class="statusColorMap[tenant.status] || statusColorMap.unknown"
            />

            <!-- Tenant Info -->
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-slate-900 dark:text-slate-100 truncate">
                {{ tenant.name }}
              </p>
              <p class="text-xs text-slate-500 dark:text-slate-400 truncate">
                {{ tenant.domain }}
              </p>
            </div>

            <!-- Check Mark -->
            <el-icon
              v-if="tenantsStore.currentTenantId === tenant.id"
              :size="16"
              class="text-primary-600 dark:text-primary-400"
            >
              <Check />
            </el-icon>
          </button>

          <!-- Empty State -->
          <div
            v-if="filteredTenants.length === 0"
            class="px-3 py-8 text-center text-sm text-slate-500 dark:text-slate-400"
          >
            没有找到匹配的租户
          </div>
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
