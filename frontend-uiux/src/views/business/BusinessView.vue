<script setup>
import { ref, computed } from 'vue'
import { BarChart3, TrendingUp, Users, DollarSign } from 'lucide-vue-next'
import { useI18n } from '@/locales/composables'

const { t } = useI18n()

const businessMetrics = computed(() => [
  {
    name: t('business.metrics.totalRevenue'),
    value: '¥1,245,600',
    change: '+18.5%',
    changeType: 'positive',
    icon: DollarSign
  },
  {
    name: t('business.metrics.activeTenants'),
    value: '128',
    change: '+12',
    changeType: 'positive',
    icon: Users
  },
  {
    name: t('business.metrics.conversionRate'),
    value: '24.8%',
    change: '+3.2%',
    changeType: 'positive',
    icon: TrendingUp
  },
  {
    name: t('business.metrics.avgRevenue'),
    value: '¥9,730',
    change: '+6.1%',
    changeType: 'positive',
    icon: BarChart3
  }
])

const recentOrders = ref([
  { id: 'ORD-001', tenant: '科技公司A', amount: '¥12,500', status: 'completed', date: '2024-01-20' },
  { id: 'ORD-002', tenant: '创业团队B', amount: '¥8,900', status: 'pending', date: '2024-01-20' },
  { id: 'ORD-003', tenant: '贸易公司D', amount: '¥15,200', status: 'completed', date: '2024-01-19' },
  { id: 'ORD-004', tenant: '咨询公司E', amount: '¥6,800', status: 'processing', date: '2024-01-19' }
])

const statusColors = {
  completed: 'bg-green-100 text-green-700',
  pending: 'bg-yellow-100 text-yellow-700',
  processing: 'bg-blue-100 text-blue-700',
  failed: 'bg-red-100 text-red-700'
}

const statusLabels = computed(() => ({
  completed: t('business.orders.status.completed'),
  pending: t('business.orders.status.pending'),
  processing: t('business.orders.status.processing'),
  failed: t('business.orders.status.failed')
}))
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-display font-bold text-slate-900">{{ t('business.title') }}</h1>
      <p class="text-slate-600">{{ t('business.description') }}</p>
    </div>

    <!-- Metrics -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div
        v-for="metric in businessMetrics"
        :key="metric.name"
        class="glass-card p-6 rounded-2xl"
      >
        <div class="flex items-start justify-between">
          <div class="flex-1">
            <p class="text-sm text-slate-600 mb-1">{{ metric.name }}</p>
            <p class="text-2xl font-bold text-slate-900">{{ metric.value }}</p>
            <p
              class="text-sm mt-2"
              :class="metric.changeType === 'positive' ? 'text-green-600' : 'text-red-600'"
            >
              {{ metric.change }}
              <span class="text-slate-500">{{ t('business.metrics.vsLastMonth') }}</span>
            </p>
          </div>
          <div class="w-12 h-12 bg-primary-100 rounded-xl flex items-center justify-center">
            <component :is="metric.icon" class="w-6 h-6 text-primary-600" />
          </div>
        </div>
      </div>
    </div>

    <!-- Charts Section -->
    <div class="grid lg:grid-cols-2 gap-6">
      <!-- Revenue Chart -->
      <div class="glass-card rounded-2xl p-6">
        <h2 class="text-lg font-display font-semibold text-slate-900 mb-4">{{ t('business.charts.revenueTrend') }}</h2>
        <div class="h-64 flex items-center justify-center bg-slate-50 rounded-xl">
          <div class="text-center">
            <component :is="BarChart3" class="w-12 h-12 text-slate-400 mx-auto mb-2" />
            <p class="text-slate-500">{{ t('business.charts.placeholder') }}</p>
            <p class="text-sm text-slate-400">{{ t('business.charts.integrationNote') }}</p>
          </div>
        </div>
      </div>

      <!-- Tenant Distribution -->
      <div class="glass-card rounded-2xl p-6">
        <h2 class="text-lg font-display font-semibold text-slate-900 mb-4">{{ t('business.charts.tenantDistribution') }}</h2>
        <div class="h-64 flex items-center justify-center bg-slate-50 rounded-xl">
          <div class="text-center">
            <component :is="BarChart3" class="w-12 h-12 text-slate-400 mx-auto mb-2" />
            <p class="text-slate-500">{{ t('business.charts.placeholder') }}</p>
            <p class="text-sm text-slate-400">{{ t('business.charts.integrationNote') }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Recent Orders -->
    <div class="glass-card rounded-2xl p-6">
      <h2 class="text-lg font-display font-semibold text-slate-900 mb-4">{{ t('business.orders.title') }}</h2>
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-slate-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">{{ t('business.orders.table.id') }}</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">{{ t('business.orders.table.tenant') }}</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">{{ t('business.orders.table.amount') }}</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">{{ t('business.orders.table.status') }}</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">{{ t('business.orders.table.date') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr
              v-for="order in recentOrders"
              :key="order.id"
              class="hover:bg-slate-50 transition-colors cursor-pointer"
            >
              <td class="px-6 py-4 font-mono text-sm text-slate-900">{{ order.id }}</td>
              <td class="px-6 py-4 text-slate-600">{{ order.tenant }}</td>
              <td class="px-6 py-4 font-medium text-slate-900">{{ order.amount }}</td>
              <td class="px-6 py-4">
                <span
                  class="px-3 py-1 rounded-full text-xs font-medium"
                  :class="statusColors[order.status]"
                >
                  {{ statusLabels[order.status] }}
                </span>
              </td>
              <td class="px-6 py-4 text-slate-600 text-sm">{{ order.date }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
