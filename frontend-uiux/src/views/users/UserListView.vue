<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useTenantsStore } from '@/stores/modules/tenants'
import BaseTable from '@/components/ui/BaseTable.vue'
import BaseBadge from '@/components/ui/BaseBadge.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseModal from '@/components/ui/BaseModal.vue'
import { Search, CirclePlus, Filter, User, Pencil, Trash2, Eye, Building, Clock, X } from 'lucide-vue-next'

const router = useRouter()
const tenantsStore = useTenantsStore()

// Mock Users Data
const users = ref([
  {
    id: 1,
    name: '张三',
    email: 'zhangsan@example.com',
    role: 'admin',
    tenantId: 'tenant-1',
    tenantName: '科技公司A',
    status: 'active',
    lastLoginAt: '2024-03-20T10:30:00Z',
    createdAt: '2024-01-15'
  },
  {
    id: 2,
    name: '李四',
    email: 'lisi@example.com',
    role: 'user',
    tenantId: 'tenant-1',
    tenantName: '科技公司A',
    status: 'active',
    lastLoginAt: '2024-03-19T15:45:00Z',
    createdAt: '2024-02-10'
  },
  {
    id: 3,
    name: '王五',
    email: 'wangwu@example.com',
    role: 'super_admin',
    tenantId: null,
    tenantName: '平台',
    status: 'active',
    lastLoginAt: '2024-03-20T08:00:00Z',
    createdAt: '2023-12-01'
  },
  {
    id: 4,
    name: '赵六',
    email: 'zhaoliu@example.com',
    role: 'auditor',
    tenantId: 'tenant-2',
    tenantName: '创业团队B',
    status: 'inactive',
    lastLoginAt: '2024-03-10T14:20:00Z',
    createdAt: '2024-01-20'
  },
  {
    id: 5,
    name: '孙七',
    email: 'sunqi@example.com',
    role: 'user',
    tenantId: 'tenant-2',
    tenantName: '创业团队B',
    status: 'suspended',
    lastLoginAt: '2024-02-28T09:15:00Z',
    createdAt: '2024-01-25'
  }
])

// Search and Filter
const searchQuery = ref('')
const selectedRole = ref('all')
const selectedStatus = ref('all')
const showFilters = ref(false)

const roleOptions = [
  { value: 'all', label: '全部角色' },
  { value: 'super_admin', label: '超级管理员' },
  { value: 'admin', label: '管理员' },
  { value: 'auditor', label: '审计员' },
  { value: 'user', label: '普通用户' }
]

const statusOptions = [
  { value: 'all', label: '全部状态' },
  { value: 'active', label: '活跃' },
  { value: 'inactive', label: '未激活' },
  { value: 'suspended', label: '已暂停' }
]

// Filtered Data
const filteredUsers = computed(() => {
  return users.value.filter(user => {
    const matchesSearch = !searchQuery.value ||
      user.name?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      user.email?.toLowerCase().includes(searchQuery.value.toLowerCase())

    const matchesRole = selectedRole.value === 'all' || user.role === selectedRole.value
    const matchesStatus = selectedStatus.value === 'all' || user.status === selectedStatus.value

    return matchesSearch && matchesRole && matchesStatus
  })
})

const activeFiltersCount = computed(() => {
  let count = 0
  if (selectedRole.value !== 'all') count++
  if (selectedStatus.value !== 'all') count++
  return count
})

// Status Config
const statusConfig = {
  active: { variant: 'success', label: '活跃' },
  inactive: { variant: 'default', label: '未激活' },
  suspended: { variant: 'error', label: '已暂停' }
}

// Role Config
const roleConfig = {
  super_admin: { variant: 'error', label: '超级管理员' },
  admin: { variant: 'primary', label: '管理员' },
  auditor: { variant: 'warning', label: '审计员' },
  user: { variant: 'default', label: '普通用户' }
}

// Table Columns
const columns = ref([
  { key: 'user', label: '用户', width: '30%' },
  { key: 'role', label: '角色', width: '12%' },
  { key: 'tenant', label: '租户', width: '18%' },
  { key: 'status', label: '状态', width: '12%' },
  { key: 'lastLogin', label: '最后登录', width: '14%' },
  { key: 'actions', label: '', width: '14%' }
])

// Delete Modal
const showDeleteModal = ref(false)
const userToDelete = ref(null)

const confirmDelete = (user) => {
  userToDelete.value = user
  showDeleteModal.value = true
}

const handleDelete = () => {
  if (userToDelete.value) {
    users.value = users.value.filter(u => u.id !== userToDelete.value.id)
    showDeleteModal.value = false
    userToDelete.value = null
  }
}

const clearFilters = () => {
  selectedRole.value = 'all'
  selectedStatus.value = 'all'
}

// Navigation
const handleCreateUser = () => {
  router.push('/dashboard/users/create')
}

const handleViewUser = (user) => {
  router.push(`/dashboard/users/${user.id}`)
}

const handleEditUser = (user) => {
  router.push(`/dashboard/users/${user.id}`)
}

const handleRowClick = ({ row }) => {
  handleViewUser(row)
}

// Utilities
const getInitials = (name) => {
  if (!name) return '??'
  return name
    .split(' ')
    .map(n => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)
}

const formatDate = (dateString) => {
  if (!dateString) return '从未'
  const date = new Date(dateString)
  const now = new Date()
  const diff = now - date

  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
  if (diff < 604800000) return `${Math.floor(diff / 86400000)} 天前`

  return date.toLocaleDateString('zh-CN')
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">用户管理</h1>
        <p class="text-slate-600 dark:text-slate-400 mt-1">
          管理所有租户的用户账户
        </p>
      </div>
      <BaseButton
        variant="primary"
        @click="handleCreateUser"
      >
        <CirclePlus  :size="20"  />
        新建用户
      </BaseButton>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="card p-5">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center">
            <User  :size="24"  class="text-primary-600 dark:text-primary-400" />
          </div>
          <div>
            <p class="text-sm text-slate-600 dark:text-slate-400">总用户数</p>
            <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">{{ users.length }}</p>
          </div>
        </div>
      </div>

      <div class="card p-5">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-success-100 dark:bg-success-900/30 flex items-center justify-center">
            <span class="text-2xl font-bold text-success-600 dark:text-success-400">
              {{ users.filter(u => u.status === 'active').length }}
            </span>
          </div>
          <div>
            <p class="text-sm text-slate-600 dark:text-slate-400">活跃用户</p>
            <p class="text-lg font-semibold text-slate-900 dark:text-slate-100">
              {{ Math.round(users.filter(u => u.status === 'active').length / users.length * 100) || 0 }}%
            </p>
          </div>
        </div>
      </div>

      <div class="card p-5">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-warning-100 dark:bg-warning-900/30 flex items-center justify-center">
            <span class="text-lg font-bold text-warning-600 dark:text-warning-400">👑</span>
          </div>
          <div>
            <p class="text-sm text-slate-600 dark:text-slate-400">管理员</p>
            <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">
              {{ users.filter(u => u.role === 'admin' || u.role === 'super_admin').length }}
            </p>
          </div>
        </div>
      </div>

      <div class="card p-5">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-info-100 dark:bg-info-900/30 flex items-center justify-center">
            <Building  :size="24"  class="text-info-600 dark:text-info-400" />
          </div>
          <div>
            <p class="text-sm text-slate-600 dark:text-slate-400">租户数</p>
            <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">
              {{ new Set(users.map(u => u.tenantId)).size }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Search and Filters -->
    <div class="card p-4">
      <div class="flex flex-col sm:flex-row gap-4">
        <!-- Search -->
        <div class="relative flex-1">
          <Search  :size="20"  class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input
            v-model="searchQuery"
            type="search"
            placeholder="搜索用户名或邮箱..."
            class="w-full pl-10 pr-4 py-2 bg-slate-100 dark:bg-slate-700 border-0 rounded-lg text-sm text-slate-900 dark:text-slate-100 placeholder:text-slate-400 focus:ring-2 focus:ring-primary-500 outline-none transition-all"
          />
        </div>

        <!-- Filter Toggle -->
        <button
          @click="showFilters = !showFilters"
          class="flex items-center gap-2 px-4 py-2 bg-slate-100 dark:bg-slate-700 rounded-lg hover:bg-slate-200 dark:hover:bg-slate-600 transition-colors cursor-pointer"
        >
          <Filter  :size="16"  />
          <span class="text-sm font-medium">筛选</span>
          <span
            v-if="activeFiltersCount > 0"
            class="w-5 h-5 bg-primary-600 text-white text-xs font-semibold rounded-full flex items-center justify-center"
          >
            {{ activeFiltersCount }}
          </span>
        </button>
      </div>

      <!-- Filters Panel -->
      <div
        v-if="showFilters"
        class="mt-4 pt-4 border-t border-slate-200 dark:border-slate-700"
      >
        <div class="flex flex-col sm:flex-row gap-4">
          <!-- Role Filter -->
          <div class="flex-1">
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">角色</label>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="option in roleOptions"
                :key="option.value"
                class="px-3 py-1.5 text-xs font-medium rounded-lg transition-all cursor-pointer"
                :class="selectedRole === option.value
                  ? 'bg-primary-600 text-white'
                  : 'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-600'"
                @click="selectedRole = option.value"
              >
                {{ option.label }}
              </button>
            </div>
          </div>

          <!-- Status Filter -->
          <div class="flex-1">
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">状态</label>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="option in statusOptions"
                :key="option.value"
                class="px-3 py-1.5 text-xs font-medium rounded-lg transition-all cursor-pointer"
                :class="selectedStatus === option.value
                  ? 'bg-primary-600 text-white'
                  : 'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-600'"
                @click="selectedStatus = option.value"
              >
                {{ option.label }}
              </button>
            </div>
          </div>

          <!-- Clear Filters -->
          <div class="flex items-end">
            <button
              v-if="activeFiltersCount > 0"
              class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 transition-colors cursor-pointer"
              @click="clearFilters"
            >
              <X  :size="16"  />
              清除筛选
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Users Table -->
    <BaseTable
      :columns="columns"
      :data="filteredUsers"
      :striped="true"
      :hoverable="true"
      @row-click="handleRowClick"
    >
      <template #cell-user="{ row }">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-full bg-gradient-to-br from-primary-500 to-primary-600 flex items-center justify-center flex-shrink-0">
            <span class="text-sm font-semibold text-white">{{ getInitials(row.name) }}</span>
          </div>
          <div>
            <p class="font-medium text-slate-900 dark:text-slate-100">{{ row.name }}</p>
            <p class="text-xs text-slate-500 dark:text-slate-400">{{ row.email }}</p>
          </div>
        </div>
      </template>

      <template #cell-role="{ row }">
        <BaseBadge
          :variant="roleConfig[row.role]?.variant || 'default'"
          :size="'sm'"
        >
          {{ roleConfig[row.role]?.label || row.role }}
        </BaseBadge>
      </template>

      <template #cell-tenant="{ row }">
        <div class="flex items-center gap-2">
          <Building  :size="16"  class="text-slate-400" />
          <span class="text-sm text-slate-700 dark:text-slate-300">{{ row.tenantName || '平台' }}</span>
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

      <template #cell-lastLogin="{ row }">
        <div class="flex items-center gap-1.5 text-sm text-slate-600 dark:text-slate-400">
          <Clock  :size="14"  />
          <span>{{ formatDate(row.lastLoginAt) }}</span>
        </div>
      </template>

      <template #cell-actions="{ row }">
        <div class="flex items-center justify-end gap-1">
          <button
            class="p-1.5 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
            :title="'查看 ' + row.name"
            @click.stop="handleViewUser(row)"
          >
            <Eye  :size="16"  class="text-slate-400" />
          </button>
          <button
            class="p-1.5 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
            :title="'编辑 ' + row.name"
            @click.stop="handleEditUser(row)"
          >
            <Pencil  :size="16"  class="text-slate-400" />
          </button>
          <button
            class="p-1.5 hover:bg-error-50 dark:hover:bg-error-900/30 rounded-lg transition-colors cursor-pointer"
            :title="'删除 ' + row.name"
            @click.stop="confirmDelete(row)"
          >
            <Trash2  :size="16"  class="text-error-400" />
          </button>
        </div>
      </template>
    </BaseTable>

    <!-- Empty State -->
    <div
      v-if="filteredUsers.length === 0"
      class="card p-12 text-center"
    >
      <User  :size="64"  class="text-slate-300 dark:text-slate-600 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-2">
        没有找到用户
      </h3>
      <p class="text-slate-500 dark:text-slate-400 mb-6">
        {{ searchQuery || activeFiltersCount > 0 ? '尝试调整搜索条件或筛选器' : '开始创建第一个用户' }}
      </p>
      <BaseButton
        v-if="!searchQuery && activeFiltersCount === 0"
        variant="primary"
        @click="handleCreateUser"
      >
        <CirclePlus  :size="20"  />
        新建用户
      </BaseButton>
    </div>

    <!-- Delete Confirmation Modal -->
    <BaseModal
      v-model:open="showDeleteModal"
      title="确认删除"
      size="sm"
    >
      <p class="text-slate-600 dark:text-slate-400">
        确定要删除用户 <span class="font-semibold text-slate-900 dark:text-slate-100">{{ userToDelete?.name }}</span> 吗？
        <br>
        <span class="text-error-600 dark:text-error-400">此操作不可撤销，将永久删除该用户的所有数据。</span>
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
