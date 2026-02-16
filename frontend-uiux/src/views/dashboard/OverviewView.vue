<script setup>
import { ref, computed, markRaw } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/modules/auth'
import { useI18n } from '@/locales/composables'
import KPICard from '@/components/ui/KPICard.vue'
import BaseTable from '@/components/tables/BaseTable.vue'
import BaseBadge from '@/components/ui/BaseBadge.vue'
import { Building, User, Box, BarChart3, TrendingUp, ChevronRight, Plus, Coins, Activity } from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()
const { t } = useI18n()

// KPI Stats - 原始数据
const statsData = ref([
  {
    titleKey: 'stats.totalTenants',
    value: '156',
    change: '+12',
    changeType: 'positive',
    trendKey: 'stats.vsLastMonth',
    icon: markRaw(Building),
    color: 'primary'
  },
  {
    titleKey: 'stats.activeServices',
    value: '24',
    change: '+3',
    changeType: 'positive',
    trendKey: 'stats.vsLastMonth',
    icon: markRaw(Box),
    color: 'success'
  },
  {
    titleKey: 'stats.totalUsers',
    value: '8,542',
    change: '+18',
    changeType: 'positive',
    trendKey: 'stats.vsLastMonth',
    icon: markRaw(User),
    color: 'info'
  },
  {
    titleKey: 'stats.monthlyRevenue',
    value: '¥89,500',
    change: '+24',
    changeType: 'positive',
    trendKey: 'stats.vsLastMonth',
    icon: markRaw(Coins),
    color: 'warning'
  }
])

// KPI Stats - 使用 computed 动态获取翻译
const stats = computed(() => {
  return statsData.value.map(stat => ({
    ...stat,
    title: t(`dashboard.${stat.titleKey}`),
    trend: t(`dashboard.${stat.trendKey}`)
  }))
})

// Quick Actions - 原始数据
const quickActionsData = ref([
  {
    nameKey: 'actions.addTenant',
    descKey: 'actions.addTenantDesc',
    path: '/dashboard/tenants',
    icon: markRaw(Building)
  },
  {
    nameKey: 'actions.configService',
    descKey: 'actions.configServiceDesc',
    path: '/dashboard/services',
    icon: markRaw(Box)
  },
  {
    nameKey: 'actions.manageUsers',
    descKey: 'actions.manageUsersDesc',
    path: '/dashboard/users',
    icon: markRaw(User)
  },
  {
    nameKey: 'actions.viewReport',
    descKey: 'actions.viewReportDesc',
    path: '/dashboard/analytics',
    icon: markRaw(BarChart3)
  }
])

// Quick Actions - 使用 computed 动态获取翻译
const quickActions = computed(() => {
  return quickActionsData.value.map(action => ({
    ...action,
    name: t(`dashboard.${action.nameKey}`),
    description: t(`dashboard.${action.descKey}`)
  }))
})

// Recent Activities - 原始数据
const recentActivitiesData = ref([
  {
    id: 1,
    type: 'tenant',
    titleKey: 'activities.newTenant',
    descKey: 'activities.newTenantDesc',
    descParams: { name: '科技公司A' },
    timeValue: 5,
    timeUnit: 'minutes',
    icon: markRaw(Building),
    color: 'primary'
  },
  {
    id: 2,
    type: 'service',
    titleKey: 'activities.serviceUpgrade',
    descKey: 'activities.serviceUpgradeDesc',
    descParams: { name: '云存储服务' },
    timeValue: 1,
    timeUnit: 'hours',
    icon: markRaw(Box),
    color: 'success'
  },
  {
    id: 3,
    type: 'user',
    titleKey: 'activities.newUser',
    descKey: 'activities.newUserDesc',
    descParams: { name: '张三', team: '创业团队B' },
    timeValue: 2,
    timeUnit: 'hours',
    icon: markRaw(User),
    color: 'info'
  },
  {
    id: 4,
    type: 'alert',
    titleKey: 'activities.systemAlert',
    descKey: 'activities.systemAlertDesc',
    descParams: { name: '消息队列服务' },
    timeValue: 3,
    timeUnit: 'hours',
    icon: markRaw(TrendingUp),
    color: 'warning'
  }
])

// Recent Activities - 使用 computed 动态获取翻译
const recentActivities = computed(() => {
  return recentActivitiesData.value.map(activity => ({
    ...activity,
    title: t(`dashboard.${activity.titleKey}`),
    description: t(`dashboard.${activity.descKey}`, activity.descParams),
    time: t(`dashboard.timeAgo.${activity.timeUnit}`, { n: activity.timeValue })
  }))
})

// Top Tenants Table - 使用 computed 动态获取翻译
const tenantColumns = computed(() => [
  { prop: 'name', label: t('dashboard.table.tenantName') },
  { prop: 'users', label: t('dashboard.table.users') },
  { prop: 'revenue', label: t('dashboard.table.revenue') },
  { prop: 'status', label: t('dashboard.table.status') },
  { prop: 'action', label: '' }
])

const tenantsData = ref([
  { name: '科技公司A', users: 45, revenue: '¥12,500', status: 'active' },
  { name: '创业团队B', users: 12, revenue: '¥8,900', status: 'active' },
  { name: '贸易公司D', users: 28, revenue: '¥15,200', status: 'suspended' },
  { name: '咨询公司E', users: 18, revenue: '¥6,800', status: 'active' },
  { name: '设计工作室F', users: 8, revenue: '¥4,200', status: 'trial' }
])

// Status Config - 使用 computed 动态获取翻译
const statusConfig = computed(() => ({
  active: { variant: 'success', label: t('dashboard.status.active') },
  suspended: { variant: 'error', label: t('dashboard.status.suspended') },
  trial: { variant: 'warning', label: t('dashboard.status.trial') }
}))

const handleTenantClick = (tenant) => {
  router.push(`/dashboard/tenants/${tenant.id}`)
}

const getIconForColor = (color) => {
  const icons = {
    primary: Building,
    success: Box,
    info: User,
    warning: Coins
  }
  return icons[color] || Activity
}

const getColorClass = (color) => {
  const classes = {
    primary: 'text-primary-600',
    success: 'text-success-600',
    info: 'text-info-600',
    warning: 'text-warning-600'
  }
  return classes[color] || 'text-slate-600'
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">{{ t('dashboard.title') }}</h1>
        <p class="text-slate-600 dark:text-slate-400 mt-1">
          {{ t('dashboard.welcome') }}，{{ authStore.user?.name || 'Admin' }}
        </p>
      </div>
      <button
        class="inline-flex items-center gap-2 px-4 py-2.5 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-all focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 cursor-pointer"
        @click="router.push('/dashboard/tenants/create')"
      >
        <Plus :size="20" />
        {{ t('dashboard.createTenant') }}
      </button>
    </div>

    <!-- KPI Stats Grid -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <KPICard
        v-for="stat in stats"
        :key="stat.title"
        :title="stat.title"
        :value="stat.value"
        :change="stat.change"
        :change-type="stat.changeType"
        :trend="stat.trend"
        :icon="stat.icon"
      />
    </div>

    <!-- Quick Actions & Recent Activities -->
    <div class="grid lg:grid-cols-3 gap-6">
      <!-- Quick Actions -->
      <div class="lg:col-span-2">
        <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-4">{{ t('dashboard.quickActions') }}</h2>
        <div class="grid sm:grid-cols-2 gap-4">
          <router-link
            v-for="action in quickActions"
            :key="action.nameKey"
            :to="action.path"
            class="card p-5 hover:shadow-card-hover transition-all cursor-pointer group"
          >
            <div class="flex items-start justify-between">
              <div class="flex-1">
                <h3 class="font-semibold text-slate-900 dark:text-slate-100 mb-1 group-hover:text-primary-600 dark:group-hover:text-primary-400 transition-colors">
                  {{ action.name }}
                </h3>
                <p class="text-sm text-slate-600 dark:text-slate-400">{{ action.description }}</p>
              </div>
              <div class="w-10 h-10 rounded-lg bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center group-hover:bg-primary-600 transition-colors">
                <component :is="action.icon" class="text-primary-600 dark:text-primary-400 group-hover:text-white transition-colors"  :size="20"  />
              </div>
            </div>
          </router-link>
        </div>
      </div>

      <!-- Recent Activities -->
      <div>
        <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-4">{{ t('dashboard.recentActivities') }}</h2>
        <div class="card p-4 space-y-3">
          <div
            v-for="activity in recentActivities"
            :key="activity.id"
            class="flex items-start gap-3 p-3 hover:bg-slate-50 dark:hover:bg-slate-700/50 rounded-lg transition-colors cursor-pointer"
          >
            <div
              :class="[
                'w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0',
                activity.color === 'primary' && 'bg-primary-100 dark:bg-primary-900/30',
                activity.color === 'success' && 'bg-success-100 dark:bg-success-900/30',
                activity.color === 'info' && 'bg-info-100 dark:bg-info-900/30',
                activity.color === 'warning' && 'bg-warning-100 dark:bg-warning-900/30'
              ]"
            >
              <el-icon :size="20">
                <component
                  :is="activity.icon"
                  :class="[
                    activity.color === 'primary' && 'text-primary-600 dark:text-primary-400',
                    activity.color === 'success' && 'text-success-600 dark:text-success-400',
                    activity.color === 'info' && 'text-info-600 dark:text-info-400',
                    activity.color === 'warning' && 'text-warning-600 dark:text-warning-400'
                  ]"
                />
              </el-icon>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-slate-900 dark:text-slate-100">{{ activity.title }}</p>
              <p class="text-xs text-slate-500 dark:text-slate-400 truncate mt-0.5">{{ activity.description }}</p>
              <p class="text-xs text-slate-400 dark:text-slate-500 mt-1">{{ activity.time }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Top Tenants Table -->
    <div>
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100">{{ t('dashboard.topTenants') }}</h2>
        <router-link
          to="/dashboard/tenants"
          class="inline-flex items-center gap-1 text-sm font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300 transition-colors cursor-pointer"
        >
          {{ t('dashboard.viewAll') }}
          <ChevronRight :size="16" />
        </router-link>
      </div>

      <BaseTable
        :columns="tenantColumns"
        :data="tenantsData"
        :striped="true"
        @row-click="handleTenantClick"
      >
        <template #cell-name="{ row }">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center">
              <Building :size="20" class="text-primary-600 dark:text-primary-400" />
            </div>
            <span class="font-medium text-slate-900 dark:text-slate-100">{{ row.name }}</span>
          </div>
        </template>

        <template #cell-users="{ row }">
          <div class="flex items-center gap-2">
            <User  :size="16"  class="text-slate-400" />
            <span class="text-slate-700 dark:text-slate-300">{{ row.users }}</span>
          </div>
        </template>

        <template #cell-revenue="{ row }">
          <span class="font-medium text-slate-900 dark:text-slate-100">{{ row.revenue }}</span>
        </template>

        <template #cell-status="{ row }">
          <BaseBadge
            :variant="statusConfig[row.status]?.variant || 'default'"
            :size="'sm'"
          >
            {{ statusConfig[row.status]?.label || row.status }}
          </BaseBadge>
        </template>

        <template #cell-action>
          <ChevronRight :size="20" class="text-slate-400" />
        </template>
      </BaseTable>
    </div>

    <!-- Revenue Chart Placeholder (for future chart.js integration) -->
    <div class="grid lg:grid-cols-2 gap-6">
      <div class="card p-6">
        <div class="flex items-center justify-between mb-6">
          <div>
            <h3 class="text-lg font-semibold text-slate-900 dark:text-slate-100">{{ t('dashboard.revenueTrend') }}</h3>
            <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">{{ t('dashboard.past6Months') }}</p>
          </div>
          <div class="flex items-center gap-2 text-success-600">
            <TrendingUp :size="20" />
            <span class="text-sm font-medium">+24%</span>
          </div>
        </div>

        <!-- Chart Placeholder -->
        <div class="h-64 flex items-center justify-center bg-slate-50 dark:bg-slate-700/30 rounded-lg">
          <div class="text-center">
            <BarChart3 :size="48" class="text-slate-300 dark:text-slate-600 mx-auto mb-3" />
            <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('dashboard.chartComingSoon') }}</p>
            <p class="text-xs text-slate-400 dark:text-slate-500 mt-1">{{ t('dashboard.chartIntegration') }}</p>
          </div>
        </div>
      </div>

      <div class="card p-6">
        <div class="flex items-center justify-between mb-6">
          <div>
            <h3 class="text-lg font-semibold text-slate-900 dark:text-slate-100">{{ t('dashboard.userGrowth') }}</h3>
            <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">{{ t('dashboard.past6Months') }}</p>
          </div>
          <div class="flex items-center gap-2 text-success-600">
            <TrendingUp :size="20" />
            <span class="text-sm font-medium">+18%</span>
          </div>
        </div>

        <!-- Chart Placeholder -->
        <div class="h-64 flex items-center justify-center bg-slate-50 dark:bg-slate-700/30 rounded-lg">
          <div class="text-center">
            <TrendingUp :size="48" class="text-slate-300 dark:text-slate-600 mx-auto mb-3" />
            <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('dashboard.chartComingSoon') }}</p>
            <p class="text-xs text-slate-400 dark:text-slate-500 mt-1">{{ t('dashboard.chartIntegration') }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
