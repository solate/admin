<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useTenantsStore } from '@/stores/tenants'
import BaseTable from '@/components/ui/BaseTable.vue'
import BaseBadge from '@/components/ui/BaseBadge.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseModal from '@/components/ui/BaseModal.vue'
import {
  Search,
  CirclePlus,
  OfficeBuilding,
  Edit,
  Delete,
  View,
  ArrowRight
} from '@element-plus/icons-vue'

const router = useRouter()
const tenantsStore = useTenantsStore()

// Search and Filter
const searchQuery = ref('')
const selectedPlan = ref('all')

const plans = [
  { value: 'all', label: '全部套餐' },
  { value: 'basic', label: '基础版' },
  { value: 'pro', label: '专业版' },
  { value: 'enterprise', label: '企业版' }
]

// Filtered Data
const filteredTenants = computed(() => {
  return tenantsStore.tenants.filter(tenant => {
    const matchesSearch = tenant.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
                         tenant.domain.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesPlan = selectedPlan.value === 'all' || tenant.plan === selectedPlan.value
    return matchesSearch && matchesPlan
  })
})

// Status Config
const statusConfig = {
  active: { variant: 'success', label: '活跃' },
  suspended: { variant: 'error', label: '暂停' },
  trial: { variant: 'warning', label: '试用' }
}

// Plan Config
const planConfig = {
  basic: { variant: 'default', label: '基础版' },
  pro: { variant: 'primary', label: '专业版' },
  enterprise: { variant: 'info', label: '企业版' }
}

// Table Columns
const columns = ref([
  { key: 'name', label: '租户', width: '25%' },
  { key: 'domain', label: '域名', width: '20%' },
  { key: 'plan', label: '套餐', width: '12%' },
  { key: 'users', label: '用户数', width: '13%' },
  { key: 'status', label: '状态', width: '12%' },
  { key: 'actions', label: '', width: '18%' }
])

// Delete Modal
const showDeleteModal = ref(false)
const tenantToDelete = ref(null)

const confirmDelete = (tenant) => {
  tenantToDelete.value = tenant
  showDeleteModal.value = true
}

const handleDelete = () => {
  if (tenantToDelete.value) {
    tenantsStore.deleteTenant(tenantToDelete.value.id)
    showDeleteModal.value = false
    tenantToDelete.value = null
  }
}

// Navigation
const handleCreateTenant = () => {
  router.push('/dashboard/tenants/create')
}

const handleViewTenant = (tenant) => {
  router.push(`/dashboard/tenants/${tenant.id}`)
}

const handleEditTenant = (tenant) => {
  router.push(`/dashboard/tenants/${tenant.id}`)
}

const handleRowClick = ({ row }) => {
  handleViewTenant(row)
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">租户管理</h1>
        <p class="text-slate-600 dark:text-slate-400 mt-1">
          管理平台上的所有租户账户
        </p>
      </div>
      <BaseButton
        variant="primary"
        @click="handleCreateTenant"
      >
        <el-icon :size="20"><CirclePlus /></el-icon>
        新建租户
      </BaseButton>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="card p-5">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center">
            <el-icon :size="24" class="text-primary-600 dark:text-primary-400"><OfficeBuilding /></el-icon>
          </div>
          <div>
            <p class="text-sm text-slate-600 dark:text-slate-400">总租户数</p>
            <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">{{ tenantsStore.tenants.length }}</p>
          </div>
        </div>
      </div>

      <div class="card p-5">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-success-100 dark:bg-success-900/30 flex items-center justify-center">
            <span class="text-2xl font-bold text-success-600 dark:text-success-400">
              {{ tenantsStore.getActiveTenants().length }}
            </span>
          </div>
          <div>
            <p class="text-sm text-slate-600 dark:text-slate-400">活跃租户</p>
            <p class="text-lg font-semibold text-slate-900 dark:text-slate-100">
              {{ Math.round(tenantsStore.getActiveTenants().length / tenantsStore.tenants.length * 100) || 0 }}%
            </p>
          </div>
        </div>
      </div>

      <div class="card p-5">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-info-100 dark:bg-info-900/30 flex items-center justify-center">
            <span class="text-lg font-bold text-info-600 dark:text-info-400">👥</span>
          </div>
          <div>
            <p class="text-sm text-slate-600 dark:text-slate-400">总用户数</p>
            <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">
              {{ tenantsStore.tenants.reduce((sum, t) => sum + t.users, 0) }}
            </p>
          </div>
        </div>
      </div>

      <div class="card p-5">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-warning-100 dark:bg-warning-900/30 flex items-center justify-center">
            <span class="text-lg font-bold text-warning-600 dark:text-warning-400">💎</span>
          </div>
          <div>
            <p class="text-sm text-slate-600 dark:text-slate-400">企业版</p>
            <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">
              {{ tenantsStore.getTenantsByPlan('enterprise').length }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="card p-4">
      <div class="flex flex-col sm:flex-row gap-4">
        <!-- Search -->
        <div class="relative flex-1">
          <el-icon :size="20" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"><Search /></el-icon>
          <input
            v-model="searchQuery"
            type="search"
            placeholder="搜索租户名称或域名..."
            class="w-full pl-10 pr-4 py-2 bg-slate-100 dark:bg-slate-700 border-0 rounded-lg text-sm text-slate-900 dark:text-slate-100 placeholder:text-slate-400 focus:ring-2 focus:ring-primary-500 outline-none transition-all"
          />
        </div>

        <!-- Plan Filter -->
        <div class="flex items-center gap-2">
          <button
            v-for="plan in plans"
            :key="plan.value"
            class="px-4 py-2 text-sm font-medium rounded-lg transition-all cursor-pointer"
            :class="selectedPlan === plan.value
              ? 'bg-primary-600 text-white'
              : 'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-600'"
            @click="selectedPlan = plan.value"
          >
            {{ plan.label }}
          </button>
        </div>
      </div>
    </div>

    <!-- Tenants Table -->
    <BaseTable
      :columns="columns"
      :data="filteredTenants"
      :striped="true"
      :hoverable="true"
      @row-click="handleRowClick"
    >
      <template #cell-name="{ row }">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-lg bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center">
            <el-icon :size="20" class="text-primary-600 dark:text-primary-400"><OfficeBuilding /></el-icon>
          </div>
          <div>
            <p class="font-medium text-slate-900 dark:text-slate-100">{{ row.name }}</p>
            <p class="text-xs text-slate-500 dark:text-slate-400">ID: {{ row.id }}</p>
          </div>
        </div>
      </template>

      <template #cell-domain="{ row }">
        <span class="text-sm text-slate-700 dark:text-slate-300 font-mono">
          {{ row.domain }}
        </span>
      </template>

      <template #cell-plan="{ row }">
        <BaseBadge
          :variant="planConfig[row.plan]?.variant || 'default'"
          :size="'sm'"
        >
          {{ planConfig[row.plan]?.label || row.plan }}
        </BaseBadge>
      </template>

      <template #cell-users="{ row }">
        <div class="flex items-center gap-2">
          <span class="font-medium text-slate-900 dark:text-slate-100">{{ row.users }}</span>
          <span class="text-slate-400">/</span>
          <span class="text-sm text-slate-500 dark:text-slate-400">{{ row.maxUsers }}</span>
        </div>
      </template>

      <template #cell-status="{ row }">
        <BaseBadge
          :variant="statusConfig[row.status]?.variant || 'default'"
          :size="'sm'"
        >
          {{ statusConfig[row.status]?.label || row.status }}
        </BaseBadge>
      </template>

      <template #cell-actions="{ row }">
        <div class="flex items-center justify-end gap-1">
          <button
            class="p-1.5 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
            :title="'查看 ' + row.name"
            @click.stop="handleViewTenant(row)"
          >
            <el-icon :size="16" class="text-slate-400"><View /></el-icon>
          </button>
          <button
            class="p-1.5 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
            :title="'编辑 ' + row.name"
            @click.stop="handleEditTenant(row)"
          >
            <el-icon :size="16" class="text-slate-400"><Edit /></el-icon>
          </button>
          <button
            class="p-1.5 hover:bg-error-50 dark:hover:bg-error-900/30 rounded-lg transition-colors cursor-pointer"
            :title="'删除 ' + row.name"
            @click.stop="confirmDelete(row)"
          >
            <el-icon :size="16" class="text-error-400"><Delete /></el-icon>
          </button>
        </div>
      </template>
    </BaseTable>

    <!-- Empty State -->
    <div
      v-if="filteredTenants.length === 0"
      class="card p-12 text-center"
    >
      <el-icon :size="64" class="text-slate-300 dark:text-slate-600 mx-auto mb-4"><OfficeBuilding /></el-icon>
      <h3 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-2">
        没有找到租户
      </h3>
      <p class="text-slate-500 dark:text-slate-400 mb-6">
        {{ searchQuery || selectedPlan !== 'all' ? '尝试调整搜索条件或筛选器' : '开始创建第一个租户' }}
      </p>
      <BaseButton
        v-if="!searchQuery && selectedPlan === 'all'"
        variant="primary"
        @click="handleCreateTenant"
      >
        <el-icon :size="20"><CirclePlus /></el-icon>
        新建租户
      </BaseButton>
    </div>

    <!-- Delete Confirmation Modal -->
    <BaseModal
      v-model:open="showDeleteModal"
      title="确认删除"
      size="sm"
    >
      <p class="text-slate-600 dark:text-slate-400">
        确定要删除租户 <span class="font-semibold text-slate-900 dark:text-slate-100">{{ tenantToDelete?.name }}</span> 吗？
        <br>
        <span class="text-error-600 dark:text-error-400">此操作不可撤销。</span>
      </p>
      <template #footer>
        <BaseButton
          variant="ghost"
          @click="showDeleteModal = false"
        >
          取消
        </BaseButton>
        <BaseButton
          variant="danger"
          @click="handleDelete"
        >
          删除
        </BaseButton>
      </template>
    </BaseModal>
  </div>
</template>
