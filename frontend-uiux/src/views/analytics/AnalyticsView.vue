<script setup>
import { ref, computed } from 'vue'
import { BarChart3 } from 'lucide-vue-next'
import { useI18n } from '@/locales/composables'

const { t } = useI18n()

const analyticsData = ref({
  totalRequests: 2458900,
  activeUsers: 8542,
  avgResponseTime: 120,
  errorRate: 0.12
})

// 动态翻译图表配置
const chartLabels = computed(() => ({
  requestTrends: t('analytics.requestTrends'),
  userActivity: t('analytics.userActivity'),
  serviceDistribution: t('analytics.serviceDistribution'),
  errorAnalysis: t('analytics.errorAnalysis'),
  chartPlaceholder: t('analytics.chartPlaceholder'),
  integrationHint: t('analytics.integrationHint')
}))
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-display font-bold text-slate-900">{{ t('analytics.title') }}</h1>
      <p class="text-slate-600">{{ t('analytics.description') }}</p>
    </div>

    <!-- Stats -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="glass-card p-6 rounded-2xl">
        <p class="text-sm text-slate-600 mb-1">{{ t('analytics.totalRequests') }}</p>
        <p class="text-2xl font-bold text-slate-900">{{ analyticsData.totalRequests.toLocaleString() }}</p>
      </div>
      <div class="glass-card p-6 rounded-2xl">
        <p class="text-sm text-slate-600 mb-1">{{ t('analytics.activeUsers') }}</p>
        <p class="text-2xl font-bold text-slate-900">{{ analyticsData.activeUsers.toLocaleString() }}</p>
      </div>
      <div class="glass-card p-6 rounded-2xl">
        <p class="text-sm text-slate-600 mb-1">{{ t('analytics.avgResponseTime') }}</p>
        <p class="text-2xl font-bold text-slate-900">{{ analyticsData.avgResponseTime }}ms</p>
      </div>
      <div class="glass-card p-6 rounded-2xl">
        <p class="text-sm text-slate-600 mb-1">{{ t('analytics.errorRate') }}</p>
        <p class="text-2xl font-bold text-slate-900">{{ analyticsData.errorRate }}%</p>
      </div>
    </div>

    <!-- Charts Grid -->
    <div class="grid lg:grid-cols-2 gap-6">
      <div class="glass-card rounded-2xl p-6">
        <h2 class="text-lg font-display font-semibold text-slate-900 mb-4">{{ chartLabels.requestTrends }}</h2>
        <div class="h-80 flex items-center justify-center bg-slate-50 rounded-xl">
          <div class="text-center">
            <BarChart3  :size="64"  class="text-slate-400 mx-auto mb-4" />
            <p class="text-slate-500">{{ chartLabels.chartPlaceholder }}</p>
            <p class="text-sm text-slate-400 mt-2">{{ chartLabels.integrationHint }}</p>
          </div>
        </div>
      </div>

      <div class="glass-card rounded-2xl p-6">
        <h2 class="text-lg font-display font-semibold text-slate-900 mb-4">{{ chartLabels.userActivity }}</h2>
        <div class="h-80 flex items-center justify-center bg-slate-50 rounded-xl">
          <div class="text-center">
            <BarChart3  :size="64"  class="text-slate-400 mx-auto mb-4" />
            <p class="text-slate-500">{{ chartLabels.chartPlaceholder }}</p>
            <p class="text-sm text-slate-400 mt-2">{{ chartLabels.integrationHint }}</p>
          </div>
        </div>
      </div>

      <div class="glass-card rounded-2xl p-6">
        <h2 class="text-lg font-display font-semibold text-slate-900 mb-4">{{ chartLabels.serviceDistribution }}</h2>
        <div class="h-80 flex items-center justify-center bg-slate-50 rounded-xl">
          <div class="text-center">
            <BarChart3  :size="64"  class="text-slate-400 mx-auto mb-4" />
            <p class="text-slate-500">{{ chartLabels.chartPlaceholder }}</p>
            <p class="text-sm text-slate-400 mt-2">{{ chartLabels.integrationHint }}</p>
          </div>
        </div>
      </div>

      <div class="glass-card rounded-2xl p-6">
        <h2 class="text-lg font-display font-semibold text-slate-900 mb-4">{{ chartLabels.errorAnalysis }}</h2>
        <div class="h-80 flex items-center justify-center bg-slate-50 rounded-xl">
          <div class="text-center">
            <BarChart3  :size="64"  class="text-slate-400 mx-auto mb-4" />
            <p class="text-slate-500">{{ chartLabels.chartPlaceholder }}</p>
            <p class="text-sm text-slate-400 mt-2">{{ chartLabels.integrationHint }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
