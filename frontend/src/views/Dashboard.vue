<template>
  <div class="dashboard">
    <!-- 欢迎卡片 -->
    <div class="welcome-card">
      <div class="welcome-content">
        <h1 class="welcome-title">欢迎回来，{{ userInfo?.user_name || '管理员' }}！</h1>
        <p class="welcome-subtitle">今天是 {{ currentDate }}，系统运行正常</p>
        <div class="welcome-actions">
          <button class="primary-button" @click="goToUsers">
            <el-icon><User /></el-icon>
            管理用户
          </button>
          <button class="primary-button" @click="goToTenants">
            <el-icon><School /></el-icon>
            管理租户
          </button>
        </div>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div v-for="stat in statsData" :key="stat.title" class="stat-card">
        <div class="stat-header">
          <div class="stat-title">{{ stat.title }}</div>
          <div class="stat-icon" :style="{ background: stat.gradient }">
            <el-icon :size="20">
              <component :is="stat.icon" />
            </el-icon>
          </div>
        </div>
        <div class="stat-value">{{ formatNumber(stat.value) }}</div>
        <div class="stat-trend" :class="stat.trend > 0 ? 'positive' : 'negative'">
          <el-icon size="14">
            <TrendCharts v-if="stat.trend > 0" />
            <Bottom v-else />
          </el-icon>
          <span>{{ Math.abs(stat.trend) }}%</span>
          <span class="trend-text">较上月</span>
        </div>
      </div>
    </div>

    <!-- 图表区域 -->
    <div class="charts-section">
      <div class="charts-grid">
        <div class="chart-card glass-card">
          <div class="chart-header">
            <h3 class="chart-title">用户增长趋势</h3>
            <el-radio-group v-model="userGrowthPeriod" size="small">
              <el-radio-button label="7d">7天</el-radio-button>
              <el-radio-button label="30d">30天</el-radio-button>
              <el-radio-button label="90d">90天</el-radio-button>
            </el-radio-group>
          </div>
          <v-chart :option="userGrowthOption" style="height: 280px" />
        </div>

        <div class="chart-card glass-card">
          <div class="chart-header">
            <h3 class="chart-title">租户状态分布</h3>
          </div>
          <v-chart :option="tenantDistributionOption" style="height: 280px" />
        </div>
      </div>

      <div class="chart-card glass-card">
        <div class="chart-header">
          <h3 class="chart-title">系统性能监控</h3>
          <div class="chart-actions">
            <span class="status-indicator online">{{ systemHealth.text }}</span>
            <el-button text size="small" @click="toggleAutoRefresh">
              <el-icon>
                <VideoPlay v-if="!autoRefresh" />
                <VideoPause v-else />
              </el-icon>
            </el-button>
          </div>
        </div>
        <v-chart :option="performanceOption" style="height: 180px" />
      </div>
    </div>

    <!-- 快速操作和最近活动 -->
    <div class="bottom-section">
      <div class="quick-actions glass-card">
        <h3 class="section-title">快速操作</h3>
        <div class="actions-grid">
          <div
            v-for="action in quickActions"
            :key="action.title"
            class="action-card"
            @click="action.handler"
          >
            <div class="action-icon" :style="{ background: action.gradient }">
              <el-icon :size="20">
                <component :is="action.icon" />
              </el-icon>
            </div>
            <div class="action-content">
              <div class="action-title">{{ action.title }}</div>
              <div class="action-desc">{{ action.description }}</div>
            </div>
          </div>
        </div>
      </div>

      <div class="recent-activities glass-card">
        <h3 class="section-title">最近活动</h3>
        <div class="activities-list">
          <div
            v-for="activity in recentActivities"
            :key="activity.id"
            class="activity-item"
          >
            <el-avatar :size="32">
              {{ activity.user.charAt(0) }}
            </el-avatar>
            <div class="activity-content">
              <div class="activity-text">
                <span class="activity-user">{{ activity.user }}</span>
                <span class="activity-action">{{ activity.action }}</span>
              </div>
              <div class="activity-time">{{ activity.time }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useThemeStore } from '../stores/theme'
import { getUserInfo } from '../utils/token'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
} from 'echarts/components'
import VChart from 'vue-echarts'

use([
  CanvasRenderer,
  LineChart,
  PieChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
])

const router = useRouter()
const themeStore = useThemeStore()
const userInfo = getUserInfo()

const userGrowthPeriod = ref('30d')
const autoRefresh = ref(true)
let refreshTimer: NodeJS.Timeout | null = null

const currentDate = computed(() => {
  const now = new Date()
  return now.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    weekday: 'long'
  })
})

const statsData = reactive([
  {
    title: '总用户数',
    value: 12847,
    trend: 12.5,
    icon: 'User',
    gradient: 'linear-gradient(135deg, #3b82f6 0%, #06b6d4 100%)'
  },
  {
    title: '活跃租户',
    value: 186,
    trend: 8.2,
    icon: 'School',
    gradient: 'linear-gradient(135deg, #10b981 0%, #059669 100%)'
  },
  {
    title: '今日访问',
    value: 3284,
    trend: -2.1,
    icon: 'TrendCharts',
    gradient: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)'
  },
  {
    title: '系统告警',
    value: 3,
    trend: -45.2,
    icon: 'Warning',
    gradient: 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)'
  }
])

const systemHealth = reactive({ text: '运行正常' })

const quickActions = [
  {
    title: '创建用户',
    description: '添加新的系统用户',
    icon: 'UserFilled',
    gradient: 'linear-gradient(135deg, #3b82f6 0%, #06b6d4 100%)',
    handler: () => router.push('/system/users?action=create')
  },
  {
    title: '创建租户',
    description: '添加新的租户',
    icon: 'School',
    gradient: 'linear-gradient(135deg, #10b981 0%, #059669 100%)',
    handler: () => router.push('/system/tenants?action=create')
  },
  {
    title: '系统配置',
    description: '管理系统设置',
    icon: 'Setting',
    gradient: 'linear-gradient(135deg, #6366f1 0%, #4f46e5 100%)',
    handler: () => router.push('/system')
  },
  {
    title: '数据备份',
    description: '备份系统数据',
    icon: 'Download',
    gradient: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)',
    handler: () => ElMessage.success('数据备份任务已启动')
  }
]

const recentActivities = [
  { id: 1, user: '张三', action: '创建了新租户 "科技公司"', time: '2分钟前' },
  { id: 2, user: '李四', action: '更新了用户权限设置', time: '15分钟前' },
  { id: 3, user: '王五', action: '删除了过期数据', time: '1小时前' },
  { id: 4, user: '赵六', action: '导出了用户报表', time: '2小时前' }
]

const userGrowthOption = computed(() => {
  const isDark = themeStore.theme === 'dark'
  return {
    tooltip: {
      trigger: 'axis',
      backgroundColor: isDark ? 'rgba(15, 23, 42, 0.9)' : 'rgba(255, 255, 255, 0.95)',
      borderColor: 'rgba(59, 130, 246, 0.2)',
      textStyle: { color: isDark ? '#f1f5f9' : '#1e293b' }
    },
    legend: {
      data: ['新增用户', '活跃用户'],
      textStyle: { color: isDark ? '#cbd5e1' : '#475569' }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: generateDateData(),
      axisLine: { lineStyle: { color: isDark ? 'rgba(148, 163, 184, 0.2)' : '#e2e8f0' } },
      axisLabel: { color: isDark ? '#94a3b8' : '#64748b' }
    },
    yAxis: {
      type: 'value',
      axisLine: { lineStyle: { color: isDark ? 'rgba(148, 163, 184, 0.2)' : '#e2e8f0' } },
      axisLabel: { color: isDark ? '#94a3b8' : '#64748b' },
      splitLine: { lineStyle: { color: isDark ? 'rgba(148, 163, 184, 0.1)' : '#f1f5f9' } }
    },
    series: [
      {
        name: '新增用户',
        type: 'line',
        smooth: true,
        data: generateRandomData(30, 10, 50),
        itemStyle: { color: '#3b82f6' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(59, 130, 246, 0.3)' },
              { offset: 1, color: 'rgba(59, 130, 246, 0.05)' }
            ]
          }
        }
      },
      {
        name: '活跃用户',
        type: 'line',
        smooth: true,
        data: generateRandomData(30, 100, 300),
        itemStyle: { color: '#06b6d4' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(6, 182, 212, 0.3)' },
              { offset: 1, color: 'rgba(6, 182, 212, 0.05)' }
            ]
          }
        }
      }
    ]
  }
})

const tenantDistributionOption = computed(() => {
  const isDark = themeStore.theme === 'dark'
  return {
    tooltip: {
      trigger: 'item',
      backgroundColor: isDark ? 'rgba(15, 23, 42, 0.9)' : 'rgba(255, 255, 255, 0.95)',
      borderColor: 'rgba(59, 130, 246, 0.2)',
      textStyle: { color: isDark ? '#f1f5f9' : '#1e293b' }
    },
    legend: {
      orient: 'vertical',
      left: 'left',
      textStyle: { color: isDark ? '#cbd5e1' : '#475569' }
    },
    series: [{
      name: '租户状态',
      type: 'pie',
      radius: ['40%', '70%'],
      itemStyle: {
        borderRadius: 8,
        borderColor: isDark ? '#0f172a' : '#ffffff',
        borderWidth: 2
      },
      label: { show: false },
      emphasis: {
        label: {
          show: true,
          fontSize: 16,
          fontWeight: 'bold',
          color: isDark ? '#f1f5f9' : '#1e293b'
        }
      },
      data: [
        { value: 120, name: '正常运营', itemStyle: { color: '#10b981' } },
        { value: 45, name: '试用期', itemStyle: { color: '#f59e0b' } },
        { value: 18, name: '已暂停', itemStyle: { color: '#ef4444' } },
        { value: 3, name: '已过期', itemStyle: { color: '#6366f1' } }
      ]
    }]
  }
})

const performanceOption = computed(() => {
  const isDark = themeStore.theme === 'dark'
  return {
    tooltip: {
      trigger: 'axis',
      backgroundColor: isDark ? 'rgba(15, 23, 42, 0.9)' : 'rgba(255, 255, 255, 0.95)',
      borderColor: 'rgba(59, 130, 246, 0.2)',
      textStyle: { color: isDark ? '#f1f5f9' : '#1e293b' }
    },
    legend: {
      data: ['CPU', '内存', '磁盘'],
      textStyle: { color: isDark ? '#cbd5e1' : '#475569' }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: generateTimeData(),
      axisLine: { lineStyle: { color: isDark ? 'rgba(148, 163, 184, 0.2)' : '#e2e8f0' } },
      axisLabel: { color: isDark ? '#94a3b8' : '#64748b' }
    },
    yAxis: {
      type: 'value',
      max: 100,
      axisLine: { lineStyle: { color: isDark ? 'rgba(148, 163, 184, 0.2)' : '#e2e8f0' } },
      axisLabel: {
        color: isDark ? '#94a3b8' : '#64748b',
        formatter: '{value}%'
      },
      splitLine: { lineStyle: { color: isDark ? 'rgba(148, 163, 184, 0.1)' : '#f1f5f9' } }
    },
    series: [
      {
        name: 'CPU',
        type: 'line',
        smooth: true,
        data: generateRandomData(20, 20, 80),
        itemStyle: { color: '#3b82f6' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(59, 130, 246, 0.3)' },
              { offset: 1, color: 'rgba(59, 130, 246, 0.05)' }
            ]
          }
        }
      },
      {
        name: '内存',
        type: 'line',
        smooth: true,
        data: generateRandomData(20, 30, 70),
        itemStyle: { color: '#06b6d4' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(6, 182, 212, 0.3)' },
              { offset: 1, color: 'rgba(6, 182, 212, 0.05)' }
            ]
          }
        }
      },
      {
        name: '磁盘',
        type: 'line',
        smooth: true,
        data: generateRandomData(20, 10, 60),
        itemStyle: { color: '#10b981' }
      }
    ]
  }
})

function formatNumber(num: number): string {
  if (num >= 10000) return (num / 10000).toFixed(1) + 'w'
  return num.toLocaleString()
}

function generateDateData(): string[] {
  const dates = []
  const now = new Date()
  for (let i = 29; i >= 0; i--) {
    const date = new Date(now)
    date.setDate(date.getDate() - i)
    dates.push(date.toLocaleDateString('zh-CN', { month: 'numeric', day: 'numeric' }))
  }
  return dates
}

function generateTimeData(): string[] {
  const times = []
  const now = new Date()
  for (let i = 19; i >= 0; i--) {
    const time = new Date(now)
    time.setMinutes(time.getMinutes() - i * 3)
    times.push(time.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }))
  }
  return times
}

function generateRandomData(count: number, min: number, max: number): number[] {
  return Array.from({ length: count }, () =>
    Math.floor(Math.random() * (max - min + 1)) + min
  )
}

function goToUsers() {
  router.push('/system/users')
}

function goToTenants() {
  router.push('/system/tenants')
}

function toggleAutoRefresh() {
  autoRefresh.value = !autoRefresh.value
  if (autoRefresh.value) {
    startAutoRefresh()
    ElMessage.success('已开启自动刷新')
  } else {
    stopAutoRefresh()
    ElMessage.info('已暂停自动刷新')
  }
}

function startAutoRefresh() {
  refreshTimer = setInterval(() => {
    performanceOption.value.series.forEach((series: any) => {
      series.data.shift()
      series.data.push(Math.floor(Math.random() * 70) + 10)
    })
  }, 3000)
}

function stopAutoRefresh() {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

onMounted(() => {
  if (autoRefresh.value) {
    startAutoRefresh()
  }
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped lang="scss">
.dashboard {
  .welcome-card {
    background: var(--gradient-primary);
    border-radius: var(--border-radius-xl);
    padding: 32px;
    color: white;
    margin-bottom: 24px;
    box-shadow: var(--box-shadow-medium);

    .welcome-content {
      .welcome-title {
        font-size: 28px;
        font-weight: 700;
        margin: 0 0 8px 0;
      }

      .welcome-subtitle {
        font-size: 14px;
        opacity: 0.9;
        margin: 0 0 24px 0;
      }

      .welcome-actions {
        display: flex;
        gap: 12px;

        button {
          background: white;
          color: var(--primary-color);
        }
      }
    }
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 16px;
    margin-bottom: 24px;

    .stat-card {
      background: var(--bg-white);
      border-radius: var(--border-radius-large);
      padding: 20px;
      border: 1px solid var(--border-base);
      box-shadow: var(--box-shadow-light);
      transition: var(--transition-base);

      &:hover {
        box-shadow: var(--glow-primary);
        transform: translateY(-2px);
      }

      .stat-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 16px;

        .stat-title {
          font-size: 14px;
          color: var(--text-secondary);
        }

        .stat-icon {
          width: 40px;
          height: 40px;
          border-radius: 10px;
          display: flex;
          align-items: center;
          justify-content: center;
          color: white;
        }
      }

      .stat-value {
        font-size: 28px;
        font-weight: 700;
        color: var(--text-primary);
        margin-bottom: 8px;
      }

      .stat-trend {
        display: flex;
        align-items: center;
        gap: 4px;
        font-size: 13px;

        &.positive {
          color: var(--success-color);
        }

        &.negative {
          color: var(--danger-color);
        }

        .trend-text {
          color: var(--text-secondary);
          margin-left: 4px;
        }
      }
    }
  }

  .charts-section {
    .charts-grid {
      display: grid;
      grid-template-columns: 2fr 1fr;
      gap: 16px;
      margin-bottom: 16px;

      @media (max-width: 1200px) {
        grid-template-columns: 1fr;
      }
    }

    .chart-card {
      padding: 20px;

      .chart-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 16px;

        .chart-title {
          font-size: 16px;
          font-weight: 600;
          color: var(--text-primary);
          margin: 0;
        }

        .chart-actions {
          display: flex;
          align-items: center;
          gap: 12px;
        }
      }
    }
  }

  .bottom-section {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;

    @media (max-width: 768px) {
      grid-template-columns: 1fr;
    }
  }

  .quick-actions, .recent-activities {
    padding: 20px;

    .section-title {
      font-size: 16px;
      font-weight: 600;
      color: var(--text-primary);
      margin: 0 0 16px 0;
    }
  }

  .quick-actions {
    .actions-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
      gap: 12px;

      .action-card {
        display: flex;
        align-items: center;
        gap: 12px;
        padding: 12px;
        border-radius: var(--border-radius);
        cursor: pointer;
        transition: var(--transition-base);
        border: 1px solid var(--border-base);

        &:hover {
          border-color: var(--primary-color);
          box-shadow: var(--glow-primary);
        }

        .action-icon {
          width: 40px;
          height: 40px;
          border-radius: 10px;
          display: flex;
          align-items: center;
          justify-content: center;
          color: white;
          flex-shrink: 0;
        }

        .action-content {
          .action-title {
            font-size: 14px;
            font-weight: 600;
            color: var(--text-primary);
            margin-bottom: 2px;
          }

          .action-desc {
            font-size: 12px;
            color: var(--text-secondary);
          }
        }
      }
    }
  }

  .recent-activities {
    .activities-list {
      .activity-item {
        display: flex;
        align-items: center;
        gap: 12px;
        padding: 12px 0;
        border-bottom: 1px solid var(--border-base);

        &:last-child {
          border-bottom: none;
        }

        &:hover {
          background: var(--bg-light);
          margin: 0 -12px;
          padding: 12px;
          border-radius: var(--border-radius);
        }

        .el-avatar {
          background: var(--gradient-primary);
          color: white;
          flex-shrink: 0;
        }

        .activity-content {
          flex: 1;

          .activity-text {
            font-size: 14px;
            color: var(--text-primary);
            margin-bottom: 4px;

            .activity-user {
              font-weight: 600;
            }
          }

          .activity-time {
            font-size: 12px;
            color: var(--text-secondary);
          }
        }
      }
    }
  }
}
</style>
