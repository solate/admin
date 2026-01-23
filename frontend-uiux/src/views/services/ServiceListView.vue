<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '@/locales/composables'
import BaseBadge from '@/components/ui/BaseBadge.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import { Search, CirclePlus, Box, ChevronRight, Cloudy, Lock, BarChart3, Mail, Wallet, Coins, Settings, Zap } from 'lucide-vue-next'

const router = useRouter()
const { t } = useI18n()

// Mock Services Data
const services = ref([
  {
    id: 1,
    name: '云存储服务',
    description: '提供安全可靠的云存储解决方案，支持多租户数据隔离',
    category: 'storage',
    status: 'active',
    version: '2.0.0',
    usage: 156,
    maxCapacity: 1000
  },
  {
    id: 2,
    name: '消息队列',
    description: '高可用、可扩展的消息队列服务，支持异步处理',
    category: 'messaging',
    status: 'active',
    version: '1.5.2',
    usage: 89,
    maxCapacity: 500
  },
  {
    id: 3,
    name: '数据分析',
    description: '实时数据分析和可视化工具，支持自定义仪表板',
    category: 'analytics',
    status: 'active',
    version: '3.1.0',
    usage: 234,
    maxCapacity: 1000
  },
  {
    id: 4,
    name: '安全认证',
    description: '多因素认证、单点登录等安全服务',
    category: 'security',
    status: 'maintenance',
    version: '2.2.0',
    usage: 178,
    maxCapacity: 1000
  },
  {
    id: 5,
    name: '邮件服务',
    description: '企业级邮件发送和接收服务',
    category: 'communication',
    status: 'active',
    version: '1.8.0',
    usage: 67,
    maxCapacity: 500
  },
  {
    id: 6,
    name: '支付网关',
    description: '集成多种支付方式的支付处理服务',
    category: 'payment',
    status: 'active',
    version: '2.0.1',
    usage: 45,
    maxCapacity: 200
  },
  {
    id: 7,
    name: '数据库服务',
    description: '托管数据库服务，支持 MySQL、PostgreSQL 等多种数据库',
    category: 'database',
    status: 'active',
    version: '4.0.0',
    usage: 123,
    maxCapacity: 500
  },
  {
    id: 8,
    name: 'API 网关',
    description: '统一 API 管理和流量控制',
    category: 'infrastructure',
    status: 'active',
    version: '1.2.0',
    usage: 267,
    maxCapacity: 1000
  }
])

// Search and Filter
const searchQuery = ref('')
const selectedCategory = ref('all')

const categories = computed(() => [
  { value: 'all', label: t('service.category.all'), icon: Box },
  { value: 'storage', label: t('service.category.storage'), icon: Cloudy },
  { value: 'messaging', label: t('service.category.messaging'), icon: Box },
  { value: 'analytics', label: t('service.category.analytics'), icon: BarChart3 },
  { value: 'security', label: t('service.category.security'), icon: Lock },
  { value: 'communication', label: t('service.category.communication'), icon: Mail },
  { value: 'payment', label: t('service.category.payment'), icon: Wallet },
  { value: 'database', label: t('service.category.database'), icon: Coins }
])

// Filtered Data
const filteredServices = computed(() => {
  return services.value.filter(service => {
    const matchesSearch = service.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
                         service.description.toLowerCase().includes(searchQuery.value.toLowerCase())
    const matchesCategory = selectedCategory.value === 'all' || service.category === selectedCategory.value
    return matchesSearch && matchesCategory
  })
})

// Status Config
const statusConfig = computed(() => ({
  active: { variant: 'success', label: t('service.running') },
  maintenance: { variant: 'warning', label: t('service.maintenance') },
  deprecated: { variant: 'error', label: t('service.deprecated') }
}))

// Category Icons
const categoryIcons = {
  storage: Cloudy,
  messaging: Box,
  analytics: BarChart3,
  security: Lock,
  communication: Mail,
  payment: Wallet,
  database: Coins,
  infrastructure: Zap
}

// Navigation
const handleCreateService = () => {
  router.push('/dashboard/services/create')
}

const handleViewService = (service) => {
  router.push(`/dashboard/services/${service.id}`)
}

const getCategoryIcon = (category) => {
  return categoryIcons[category] || Settings
}

const getCategoryLabel = (category) => {
  const cat = categories.value.find(c => c.value === category)
  return cat ? cat.label : category
}

const getUsagePercent = (service) => {
  return Math.round((service.usage / service.maxCapacity) * 100)
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">{{ t('service.title') }}</h1>
        <p class="text-slate-600 dark:text-slate-400 mt-1">
          {{ t('service.manageServices') }}
        </p>
      </div>
      <BaseButton
        variant="primary"
        @click="handleCreateService"
      >
        <CirclePlus  :size="20"  />
        {{ t('service.addService') }}
      </BaseButton>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="card p-5">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center">
            <Box  :size="24"  class="text-primary-600 dark:text-primary-400" />
          </div>
          <div>
            <p class="text-sm text-slate-600 dark:text-slate-400">{{ t('service.totalServices') }}</p>
            <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">{{ services.length }}</p>
          </div>
        </div>
      </div>

      <div class="card p-5">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-success-100 dark:bg-success-900/30 flex items-center justify-center">
            <Zap  :size="24"  class="text-success-600 dark:text-success-400" />
          </div>
          <div>
            <p class="text-sm text-slate-600 dark:text-slate-400">{{ t('service.running') }}</p>
            <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">
              {{ services.filter(s => s.status === 'active').length }}
            </p>
          </div>
        </div>
      </div>

      <div class="card p-5">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-warning-100 dark:bg-warning-900/30 flex items-center justify-center">
            <Settings  :size="24"  class="text-warning-600 dark:text-warning-400" />
          </div>
          <div>
            <p class="text-sm text-slate-600 dark:text-slate-400">{{ t('service.maintenance') }}</p>
            <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">
              {{ services.filter(s => s.status === 'maintenance').length }}
            </p>
          </div>
        </div>
      </div>

      <div class="card p-5">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-info-100 dark:bg-info-900/30 flex items-center justify-center">
            <span class="text-lg font-bold text-info-600 dark:text-info-400">📁</span>
          </div>
          <div>
            <p class="text-sm text-slate-600 dark:text-slate-400">{{ t('service.categories') }}</p>
            <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">{{ categories.length - 1 }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Category Filters -->
    <div class="flex flex-wrap gap-2">
      <button
        v-for="category in categories"
        :key="category.value"
        class="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg font-medium transition-all cursor-pointer"
        :class="selectedCategory === category.value
          ? 'bg-primary-600 text-white'
          : 'bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700 border border-slate-200 dark:border-slate-700'"
        @click="selectedCategory = category.value"
      >
        <component :is="category.icon || Box"  :size="16"  />
        {{ category.label }}
      </button>
    </div>

    <!-- Search -->
    <div class="card p-4">
      <div class="relative">
        <Search  :size="20"  class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
        <input
          v-model="searchQuery"
          type="search"
          :placeholder="t('service.searchByNameOrDesc')"
          class="w-full pl-10 pr-4 py-2.5 bg-slate-100 dark:bg-slate-700 border-0 rounded-lg text-sm text-slate-900 dark:text-slate-100 placeholder:text-slate-400 focus:ring-2 focus:ring-primary-500 outline-none transition-all"
        />
      </div>
    </div>

    <!-- Services Grid -->
    <div class="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div
        v-for="service in filteredServices"
        :key="service.id"
        class="card p-6 hover:shadow-card-hover transition-all cursor-pointer group"
        @click="handleViewService(service)"
      >
        <div class="flex items-start justify-between mb-4">
          <div
            :class="[
              'w-14 h-14 rounded-xl flex items-center justify-center transition-colors',
              service.status === 'active'
                ? 'bg-primary-100 dark:bg-primary-900/30 group-hover:bg-primary-600'
                : 'bg-warning-100 dark:bg-warning-900/30'
            ]"
          >
            <el-icon :size="28">
              <component
                :is="getCategoryIcon(service.category)"
                :class="[
                  'transition-colors',
                  service.status === 'active'
                    ? 'text-primary-600 dark:text-primary-400 group-hover:text-white'
                    : 'text-warning-600 dark:text-warning-400'
                ]"
              />
            </el-icon>
          </div>
          <BaseBadge
            :variant="statusConfig[service.status]?.variant || 'default'"
            :size="'sm'"
          >
            {{ statusConfig[service.status]?.label || service.status }}
          </BaseBadge>
        </div>

        <h3 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-2 group-hover:text-primary-600 dark:group-hover:text-primary-400 transition-colors">
          {{ service.name }}
        </h3>

        <p class="text-slate-600 dark:text-slate-400 text-sm mb-4 line-clamp-2">
          {{ service.description }}
        </p>

        <div class="space-y-3">
          <!-- Version -->
          <div class="flex items-center justify-between text-sm">
            <span class="text-slate-500 dark:text-slate-400">{{ t('service.version') }}</span>
            <span class="font-mono text-slate-900 dark:text-slate-100">{{ service.version }}</span>
          </div>

          <!-- Usage -->
          <div>
            <div class="flex items-center justify-between text-sm mb-1">
              <span class="text-slate-500 dark:text-slate-400">{{ t('service.usage') }}</span>
              <span class="text-slate-900 dark:text-slate-100">{{ service.usage }} / {{ service.maxCapacity }}</span>
            </div>
            <div class="w-full h-2 bg-slate-100 dark:bg-slate-700 rounded-full overflow-hidden">
              <div
                class="h-full transition-all duration-500 rounded-full"
                :class="getUsagePercent(service) > 80 ? 'bg-error-500' : getUsagePercent(service) > 60 ? 'bg-warning-500' : 'bg-success-500'"
                :style="{ width: getUsagePercent(service) + '%' }"
              />
            </div>
          </div>

          <!-- Category -->
          <div class="flex items-center justify-between text-sm">
            <span class="text-slate-500 dark:text-slate-400">{{ t('service.categoryLabel') }}</span>
            <span class="text-slate-900 dark:text-slate-100">{{ getCategoryLabel(service.category) }}</span>
          </div>
        </div>

        <div class="mt-4 pt-4 border-t border-slate-200 dark:border-slate-700 flex items-center justify-between">
          <span class="text-xs text-slate-400 dark:text-slate-500">{{ t('service.clickToViewDetails') }}</span>
          <ChevronRight  :size="20"  class="text-slate-400 group-hover:text-primary-600 dark:group-hover:text-primary-400 transition-colors" />
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div
      v-if="filteredServices.length === 0"
      class="card p-12 text-center"
    >
      <Box  :size="64"  class="text-slate-300 dark:text-slate-600 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-2">
        {{ t('service.noServiceFound') }}
      </h3>
      <p class="text-slate-500 dark:text-slate-400">
        {{ searchQuery || selectedCategory !== 'all' ? t('service.tryOtherFilters') : t('service.addFirstService') }}
      </p>
    </div>
  </div>
</template>
