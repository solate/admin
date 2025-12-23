<template>
  <div class="monitor-page">
    <div class="page-header">
      <h1 class="page-title">系统监控</h1>
      <el-button type="primary" @click="refreshData">
        <el-icon><Refresh /></el-icon>
        刷新数据
      </el-button>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div v-for="stat in statsData" :key="stat.title" class="stat-card">
        <div class="stat-icon" :style="{ backgroundColor: stat.color }">
          <el-icon :size="24">
            <component :is="stat.icon" />
          </el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-title">{{ stat.title }}</div>
          <div class="stat-value">{{ stat.value }}</div>
          <div class="stat-desc">{{ stat.description }}</div>
        </div>
      </div>
    </div>

    <!-- 监控图表 -->
    <div class="monitor-charts">
      <el-card class="chart-card">
        <template #header>
          <div class="chart-header">
            <span class="chart-title">CPU 使用率</span>
            <el-tag :type="cpuUsage > 80 ? 'danger' : cpuUsage > 50 ? 'warning' : 'success'">
              {{ cpuUsage }}%
            </el-tag>
          </div>
        </template>
        <v-chart :option="cpuChartOption" style="height: 200px" />
      </el-card>

      <el-card class="chart-card">
        <template #header>
          <div class="chart-header">
            <span class="chart-title">内存使用率</span>
            <el-tag :type="memoryUsage > 80 ? 'danger' : memoryUsage > 50 ? 'warning' : 'success'">
              {{ memoryUsage }}%
            </el-tag>
          </div>
        </template>
        <v-chart :option="memoryChartOption" style="height: 200px" />
      </el-card>

      <el-card class="chart-card">
        <template #header>
          <div class="chart-header">
            <span class="chart-title">磁盘使用率</span>
            <el-tag :type="diskUsage > 80 ? 'danger' : diskUsage > 50 ? 'warning' : 'success'">
              {{ diskUsage }}%
            </el-tag>
          </div>
        </template>
        <v-chart :option="diskChartOption" style="height: 200px" />
      </el-card>

      <el-card class="chart-card">
        <template #header>
          <div class="chart-header">
            <span class="chart-title">网络流量</span>
            <el-tag type="info">实时</el-tag>
          </div>
        </template>
        <v-chart :option="networkChartOption" style="height: 200px" />
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useThemeStore } from '../../stores/theme'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, GaugeChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, GridComponent } from 'echarts/components'
import VChart from 'vue-echarts'

use([CanvasRenderer, LineChart, GaugeChart, TitleComponent, TooltipComponent, GridComponent])

const themeStore = useThemeStore()

// 状态数据
const cpuUsage = ref(45)
const memoryUsage = ref(62)
const diskUsage = ref(58)

const statsData = [
  { title: '运行时间', value: '15天 8小时', description: '系统持续运行时间', icon: 'Clock', color: '#409eff' },
  { title: '在线用户', value: '128', description: '当前在线用户数', icon: 'User', color: '#67c23a' },
  { title: '请求总数', value: '1.2M', description: '累计处理请求数', icon: 'TrendCharts', color: '#e6a23c' },
  { title: '错误率', value: '0.02%', description: '最近24小时错误率', icon: 'Warning', color: '#f56c6c' }
]

// CPU 图表配置
const cpuChartOption = computed(() => {
  const isDark = themeStore.theme === 'dark'
  return {
    tooltip: {
      trigger: 'axis',
      backgroundColor: isDark ? '#1d1e1f' : '#fff',
      borderColor: isDark ? '#4c4d4f' : '#e4e7ed',
      textStyle: { color: isDark ? '#e5eaf3' : '#303133' }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: generateTimeLabels(20),
      axisLine: { lineStyle: { color: isDark ? '#4c4d4f' : '#e4e7ed' } },
      axisLabel: { color: isDark ? '#a3a6ad' : '#606266' }
    },
    yAxis: {
      type: 'value',
      max: 100,
      axisLine: { lineStyle: { color: isDark ? '#4c4d4f' : '#e4e7ed' } },
      axisLabel: { color: isDark ? '#a3a6ad' : '#606266', formatter: '{value}%' },
      splitLine: { lineStyle: { color: isDark ? '#2b2b2c' : '#ebeef5' } }
    },
    series: [{
      data: generateRandomData(20),
      type: 'line',
      smooth: true,
      itemStyle: { color: '#409eff' },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
            { offset: 1, color: 'rgba(64, 158, 255, 0.05)' }
          ]
        }
      }
    }]
  }
})

// 内存图表配置
const memoryChartOption = computed(() => {
  const isDark = themeStore.theme === 'dark'
  return {
    tooltip: {
      trigger: 'axis',
      backgroundColor: isDark ? '#1d1e1f' : '#fff',
      borderColor: isDark ? '#4c4d4f' : '#e4e7ed',
      textStyle: { color: isDark ? '#e5eaf3' : '#303133' }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: generateTimeLabels(20),
      axisLine: { lineStyle: { color: isDark ? '#4c4d4f' : '#e4e7ed' } },
      axisLabel: { color: isDark ? '#a3a6ad' : '#606266' }
    },
    yAxis: {
      type: 'value',
      max: 100,
      axisLine: { lineStyle: { color: isDark ? '#4c4d4f' : '#e4e7ed' } },
      axisLabel: { color: isDark ? '#a3a6ad' : '#606266', formatter: '{value}%' },
      splitLine: { lineStyle: { color: isDark ? '#2b2b2c' : '#ebeef5' } }
    },
    series: [{
      data: generateRandomData(20),
      type: 'line',
      smooth: true,
      itemStyle: { color: '#67c23a' },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(103, 194, 58, 0.3)' },
            { offset: 1, color: 'rgba(103, 194, 58, 0.05)' }
          ]
        }
      }
    }]
  }
})

// 磁盘图表配置
const diskChartOption = computed(() => {
  const isDark = themeStore.theme === 'dark'
  return {
    tooltip: {
      trigger: 'axis',
      backgroundColor: isDark ? '#1d1e1f' : '#fff',
      borderColor: isDark ? '#4c4d4f' : '#e4e7ed',
      textStyle: { color: isDark ? '#e5eaf3' : '#303133' }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: generateTimeLabels(20),
      axisLine: { lineStyle: { color: isDark ? '#4c4d4f' : '#e4e7ed' } },
      axisLabel: { color: isDark ? '#a3a6ad' : '#606266' }
    },
    yAxis: {
      type: 'value',
      max: 100,
      axisLine: { lineStyle: { color: isDark ? '#4c4d4f' : '#e4e7ed' } },
      axisLabel: { color: isDark ? '#a3a6ad' : '#606266', formatter: '{value}%' },
      splitLine: { lineStyle: { color: isDark ? '#2b2b2c' : '#ebeef5' } }
    },
    series: [{
      data: generateRandomData(20),
      type: 'line',
      smooth: true,
      itemStyle: { color: '#e6a23c' },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(230, 162, 60, 0.3)' },
            { offset: 1, color: 'rgba(230, 162, 60, 0.05)' }
          ]
        }
      }
    }]
  }
})

// 网络流量图表配置
const networkChartOption = computed(() => {
  const isDark = themeStore.theme === 'dark'
  return {
    tooltip: {
      trigger: 'axis',
      backgroundColor: isDark ? '#1d1e1f' : '#fff',
      borderColor: isDark ? '#4c4d4f' : '#e4e7ed',
      textStyle: { color: isDark ? '#e5eaf3' : '#303133' }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: generateTimeLabels(20),
      axisLine: { lineStyle: { color: isDark ? '#4c4d4f' : '#e4e7ed' } },
      axisLabel: { color: isDark ? '#a3a6ad' : '#606266' }
    },
    yAxis: {
      type: 'value',
      axisLine: { lineStyle: { color: isDark ? '#4c4d4f' : '#e4e7ed' } },
      axisLabel: { color: isDark ? '#a3a6ad' : '#606266', formatter: '{value} MB/s' },
      splitLine: { lineStyle: { color: isDark ? '#2b2b2c' : '#ebeef5' } }
    },
    series: [
      {
        name: '上行',
        data: generateRandomData(20, 5, 50),
        type: 'line',
        smooth: true,
        itemStyle: { color: '#409eff' }
      },
      {
        name: '下行',
        data: generateRandomData(20, 10, 80),
        type: 'line',
        smooth: true,
        itemStyle: { color: '#67c23a' }
      }
    ]
  }
})

function generateTimeLabels(count: number): string[] {
  const labels = []
  const now = new Date()
  for (let i = count - 1; i >= 0; i--) {
    const time = new Date(now.getTime() - i * 3000)
    labels.push(time.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' }))
  }
  return labels
}

function generateRandomData(count: number, min = 20, max = 80): number[] {
  return Array.from({ length: count }, () => Math.floor(Math.random() * (max - min + 1)) + min)
}

function refreshData() {
  cpuUsage.value = Math.floor(Math.random() * 60) + 20
  memoryUsage.value = Math.floor(Math.random() * 50) + 30
  diskUsage.value = Math.floor(Math.random() * 40) + 40
  ElMessage.success('数据已刷新')
}

let refreshTimer: NodeJS.Timeout | null = null

onMounted(() => {
  refreshTimer = setInterval(() => {
    cpuUsage.value = Math.floor(Math.random() * 60) + 20
    memoryUsage.value = Math.floor(Math.random() * 50) + 30
  }, 5000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped lang="scss">
.monitor-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;

    .page-title {
      font-size: var(--font-size-extra-large);
      font-weight: 600;
      margin: 0;
      color: var(--text-primary);
    }
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 16px;
    margin-bottom: 24px;

    .stat-card {
      background: var(--bg-white);
      border-radius: 8px;
      padding: 20px;
      border: 1px solid var(--border-lighter);
      display: flex;
      align-items: center;
      gap: 16px;
      transition: all 0.3s ease;

      &:hover {
        transform: translateY(-2px);
        box-shadow: var(--box-shadow-light);
      }

      .stat-icon {
        width: 56px;
        height: 56px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: white;
        flex-shrink: 0;
      }

      .stat-content {
        flex: 1;

        .stat-title {
          font-size: var(--font-size-small);
          color: var(--text-secondary);
          margin-bottom: 8px;
        }

        .stat-value {
          font-size: var(--font-size-extra-large);
          font-weight: 600;
          color: var(--text-primary);
          margin-bottom: 4px;
        }

        .stat-desc {
          font-size: var(--font-size-small);
          color: var(--text-placeholder);
        }
      }
    }
  }

  .monitor-charts {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 20px;

    @media (max-width: 1200px) {
      grid-template-columns: 1fr;
    }

    .chart-card {
      .chart-header {
        display: flex;
        justify-content: space-between;
        align-items: center;

        .chart-title {
          font-size: var(--font-size-large);
          font-weight: 600;
          color: var(--text-primary);
        }
      }
    }
  }
}
</style>