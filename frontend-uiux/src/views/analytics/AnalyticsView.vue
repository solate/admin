<script setup>
import { ref } from 'vue'
import { DataAnalysis } from '@element-plus/icons-vue'

const analyticsData = ref({
  totalRequests: 2458900,
  activeUsers: 8542,
  avgResponseTime: 120,
  errorRate: 0.12
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-display font-bold text-slate-900">数据分析</h1>
      <p class="text-slate-600">平台运营数据和分析报告</p>
    </div>

    <!-- Stats -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="glass-card p-6 rounded-2xl">
        <p class="text-sm text-slate-600 mb-1">总请求数</p>
        <p class="text-2xl font-bold text-slate-900">{{ analyticsData.totalRequests.toLocaleString() }}</p>
      </div>
      <div class="glass-card p-6 rounded-2xl">
        <p class="text-sm text-slate-600 mb-1">活跃用户</p>
        <p class="text-2xl font-bold text-slate-900">{{ analyticsData.activeUsers.toLocaleString() }}</p>
      </div>
      <div class="glass-card p-6 rounded-2xl">
        <p class="text-sm text-slate-600 mb-1">平均响应时间</p>
        <p class="text-2xl font-bold text-slate-900">{{ analyticsData.avgResponseTime }}ms</p>
      </div>
      <div class="glass-card p-6 rounded-2xl">
        <p class="text-sm text-slate-600 mb-1">错误率</p>
        <p class="text-2xl font-bold text-slate-900">{{ analyticsData.errorRate }}%</p>
      </div>
    </div>

    <!-- Charts Grid -->
    <div class="grid lg:grid-cols-2 gap-6">
      <div class="glass-card rounded-2xl p-6">
        <h2 class="text-lg font-display font-semibold text-slate-900 mb-4">请求量趋势</h2>
        <div class="h-80 flex items-center justify-center bg-slate-50 rounded-xl">
          <div class="text-center">
            <el-icon :size="64" class="text-slate-400 mx-auto mb-4"><DataAnalysis /></el-icon>
            <p class="text-slate-500">数据可视化图表</p>
            <p class="text-sm text-slate-400 mt-2">可集成 Chart.js、ECharts 或 ApexCharts</p>
          </div>
        </div>
      </div>

      <div class="glass-card rounded-2xl p-6">
        <h2 class="text-lg font-display font-semibold text-slate-900 mb-4">用户活跃度</h2>
        <div class="h-80 flex items-center justify-center bg-slate-50 rounded-xl">
          <div class="text-center">
            <el-icon :size="64" class="text-slate-400 mx-auto mb-4"><DataAnalysis /></el-icon>
            <p class="text-slate-500">数据可视化图表</p>
            <p class="text-sm text-slate-400 mt-2">可集成 Chart.js、ECharts 或 ApexCharts</p>
          </div>
        </div>
      </div>

      <div class="glass-card rounded-2xl p-6">
        <h2 class="text-lg font-display font-semibold text-slate-900 mb-4">服务使用分布</h2>
        <div class="h-80 flex items-center justify-center bg-slate-50 rounded-xl">
          <div class="text-center">
            <el-icon :size="64" class="text-slate-400 mx-auto mb-4"><DataAnalysis /></el-icon>
            <p class="text-slate-500">数据可视化图表</p>
            <p class="text-sm text-slate-400 mt-2">可集成 Chart.js、ECharts 或 ApexCharts</p>
          </div>
        </div>
      </div>

      <div class="glass-card rounded-2xl p-6">
        <h2 class="text-lg font-display font-semibold text-slate-900 mb-4">错误分析</h2>
        <div class="h-80 flex items-center justify-center bg-slate-50 rounded-xl">
          <div class="text-center">
            <el-icon :size="64" class="text-slate-400 mx-auto mb-4"><DataAnalysis /></el-icon>
            <p class="text-slate-500">数据可视化图表</p>
            <p class="text-sm text-slate-400 mt-2">可集成 Chart.js、ECharts 或 ApexCharts</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
