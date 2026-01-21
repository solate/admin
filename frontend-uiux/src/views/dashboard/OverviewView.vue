<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import KPICard from '@/components/ui/KPICard.vue'
import BaseTable from '@/components/ui/BaseTable.vue'
import BaseBadge from '@/components/ui/BaseBadge.vue'
import {
  Building,
  Users,
  Box,
  TrendingUp,
  TrendingDown,
  ChevronRight,
  Plus,
  Activity,
  DollarSign,
  BarChart
} from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()

// KPI Stats
const stats = ref([
  {
    title: '总租户数',
    value: '156',
    change: '+12',
    changeType: 'positive',
    trend: '较上月',
    icon: Building,
    color: 'primary'
  },
  {
    title: '活跃服务',
    value: '24',
    change: '+3',
    changeType: 'positive',
    trend: '较上月',
    icon: Box,
    color: 'success'
  },
  {
    title: '总用户数',
    value: '8,542',
    change: '+18',
    changeType: 'positive',
    trend: '较上月',
    icon: Users,
    color: 'info'
  },
  {
    title: '本月收入',
    value: '¥89,500',
    change: '+24',
    changeType: 'positive',
    trend: '较上月',
    icon: DollarSign,
    color: 'warning'
  }
])

// Quick Actions
const quickActions = ref([
  {
    name: '添加租户',
    description: '创建新的租户账户',
    path: '/dashboard/tenants',
    icon: Building
  },
  {
    name: '配置服务',
    description: '管理平台服务',
    path: '/dashboard/services',
    icon: Box
  },
  {
    name: '用户管理',
    description: '管理平台用户',
    path: '/dashboard/users',
    icon: Users
  },
  {
    name: '查看报表',
    description: '查看数据分析',
    path: '/dashboard/analytics',
    icon: BarChart
  }
])

// Recent Activities
const recentActivities = ref([
  {
    id: 1,
    type: 'tenant',
    title: '新租户注册',
    description: '科技公司A 已完成注册',
    time: '5 分钟前',
    icon: Building,
    color: 'primary'
  },
  {
    id: 2,
    type: 'service',
    title: '服务升级',
    description: '云存储服务已升级到 v2.0',
    time: '1 小时前',
    icon: Box,
    color: 'success'
  },
  {
    id: 3,
    type: 'user',
    title: '新用户加入',
    description: '张三 加入了创业团队B',
    time: '2 小时前',
    icon: Users,
    color: 'info'
  },
  {
    id: 4,
    type: 'alert',
    title: '系统提醒',
    description: '消息队列服务负载较高',
    time: '3 小时前',
    icon: Activity,
    color: 'warning'
  }
])

// Top Tenants Table
const tenantColumns = ref([
  { key: 'name', label: '租户名称', width: '30%' },
  { key: 'users', label: '用户数', width: '20%' },
  { key: 'revenue', label: '收入', width: '20%' },
  { key: 'status', label: '状态', width: '15%' },
  { key: 'action', label: '', width: '15%' }
])

const tenantsData = ref([
  { name: '科技公司A', users: 45, revenue: '¥12,500', status: 'active' },
  { name: '创业团队B', users: 12, revenue: '¥8,900', status: 'active' },
  { name: '贸易公司D', users: 28, revenue: '¥15,200', status: 'suspended' },
  { name: '咨询公司E', users: 18, revenue: '¥6,800', status: 'active' },
  { name: '设计工作室F', users: 8, revenue: '¥4,200', status: 'trial' }
])

const statusConfig = {
  active: { variant: 'success', label: '活跃' },
  suspended: { variant: 'error', label: '暂停' },
  trial: { variant: 'warning', label: '试用' }
}

const handleTenantClick = (tenant) => {
  router.push(`/dashboard/tenants/${tenant.id}`)
}

const getIconForColor = (color) => {
  const icons = {
    primary: Building,
    success: Box,
    info: Users,
    warning: DollarSign
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
        <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">概览</h1>
        <p class="text-slate-600 dark:text-slate-400 mt-1">
          欢迎回来，{{ authStore.user?.name || 'Admin' }}
        </p>
      </div>
      <button
        class="inline-flex items-center gap-2 px-4 py-2.5 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-all focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 cursor-pointer"
        @click="router.push('/dashboard/tenants/create')"
      >
        <Plus class="w-5 h-5" />
        新建租户
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
        <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-4">快捷操作</h2>
        <div class="grid sm:grid-cols-2 gap-4">
          <router-link
            v-for="action in quickActions"
            :key="action.name"
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
                <component :is="action.icon" class="w-5 h-5 text-primary-600 dark:text-primary-400 group-hover:text-white transition-colors" />
              </div>
            </div>
          </router-link>
        </div>
      </div>

      <!-- Recent Activities -->
      <div>
        <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-4">最近活动</h2>
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
              <component
                :is="activity.icon"
                :class="[
                  'w-5 h-5',
                  activity.color === 'primary' && 'text-primary-600 dark:text-primary-400',
                  activity.color === 'success' && 'text-success-600 dark:text-success-400',
                  activity.color === 'info' && 'text-info-600 dark:text-info-400',
                  activity.color === 'warning' && 'text-warning-600 dark:text-warning-400'
                ]"
              />
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
        <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100">热门租户</h2>
        <router-link
          to="/dashboard/tenants"
          class="inline-flex items-center gap-1 text-sm font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300 transition-colors cursor-pointer"
        >
          查看全部
          <ChevronRight class="w-4 h-4" />
        </router-link>
      </div>

      <BaseTable
        :columns="tenantColumns"
        :data="tenantsData"
        :striped="true"
        :hoverable="true"
        @row-click="handleTenantClick"
      >
        <template #cell-name="{ row }">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center">
              <Building class="w-5 h-5 text-primary-600 dark:text-primary-400" />
            </div>
            <span class="font-medium text-slate-900 dark:text-slate-100">{{ row.name }}</span>
          </div>
        </template>

        <template #cell-users="{ row }">
          <div class="flex items-center gap-2">
            <Users class="w-4 h-4 text-slate-400" />
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
          <ChevronRight class="w-5 h-5 text-slate-400" />
        </template>
      </BaseTable>
    </div>

    <!-- Revenue Chart Placeholder (for future chart.js integration) -->
    <div class="grid lg:grid-cols-2 gap-6">
      <div class="card p-6">
        <div class="flex items-center justify-between mb-6">
          <div>
            <h3 class="text-lg font-semibold text-slate-900 dark:text-slate-100">收入趋势</h3>
            <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">过去6个月</p>
          </div>
          <div class="flex items-center gap-2 text-success-600">
            <TrendingUp class="w-5 h-5" />
            <span class="text-sm font-medium">+24%</span>
          </div>
        </div>

        <!-- Chart Placeholder -->
        <div class="h-64 flex items-center justify-center bg-slate-50 dark:bg-slate-700/30 rounded-lg">
          <div class="text-center">
            <BarChart class="w-12 h-12 text-slate-300 dark:text-slate-600 mx-auto mb-3" />
            <p class="text-sm text-slate-500 dark:text-slate-400">图表组件即将推出</p>
            <p class="text-xs text-slate-400 dark:text-slate-500 mt-1">将集成 Chart.js</p>
          </div>
        </div>
      </div>

      <div class="card p-6">
        <div class="flex items-center justify-between mb-6">
          <div>
            <h3 class="text-lg font-semibold text-slate-900 dark:text-slate-100">用户增长</h3>
            <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">过去6个月</p>
          </div>
          <div class="flex items-center gap-2 text-success-600">
            <TrendingUp class="w-5 h-5" />
            <span class="text-sm font-medium">+18%</span>
          </div>
        </div>

        <!-- Chart Placeholder -->
        <div class="h-64 flex items-center justify-center bg-slate-50 dark:bg-slate-700/30 rounded-lg">
          <div class="text-center">
            <Activity class="w-12 h-12 text-slate-300 dark:text-slate-600 mx-auto mb-3" />
            <p class="text-sm text-slate-500 dark:text-slate-400">图表组件即将推出</p>
            <p class="text-xs text-slate-400 dark:text-slate-500 mt-1">将集成 Chart.js</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
