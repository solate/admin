<template>
  <div class="dashboard">
    <!-- 顶部汇总卡片 -->
    <el-row :gutter="20" v-loading="loading">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon">
              <el-icon :size="40" color="#409eff"><Goods /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.total_products }}</div>
              <div class="stat-label">商品总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon">
              <el-icon :size="40" color="#67c23a"><Box /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.total_stock }}</div>
              <div class="stat-label">总库存数量</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon">
              <el-icon :size="40" color="#e6a23c"><Money /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">¥{{ formatNumber(stats.total_stock_value) }}</div>
              <div class="stat-label">总库存价值</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon">
              <el-icon :size="40" color="#f56c6c"><Warning /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.low_stock_products }}</div>
              <div class="stat-label">低库存商品</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <!-- 入库出库统计 -->
    <el-row :gutter="20" style="margin-top: 20px;" v-loading="loading">
      <el-col :span="12">
        <el-card class="detail-card">
          <template #header>
            <div class="card-header">
              <el-icon color="#67c23a"><TopRight /></el-icon>
              <span>入库统计</span>
            </div>
          </template>
          <div class="detail-content">
            <div class="detail-item">
              <div class="detail-label">入库数量</div>
              <div class="detail-value text-success">{{ stats.total_in_quantity }}</div>
            </div>
            <el-divider />
            <div class="detail-item">
              <div class="detail-label">入库金额</div>
              <div class="detail-value text-success">¥{{ formatNumber(stats.total_in_amount) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card class="detail-card">
          <template #header>
            <div class="card-header">
              <el-icon color="#e6a23c"><BottomRight /></el-icon>
              <span>出库统计</span>
            </div>
          </template>
          <div class="detail-content">
            <div class="detail-item">
              <div class="detail-label">出库数量</div>
              <div class="detail-value text-warning">{{ stats.total_out_quantity }}</div>
            </div>
            <el-divider />
            <div class="detail-item">
              <div class="detail-label">出库金额</div>
              <div class="detail-value text-warning">¥{{ formatNumber(stats.total_out_amount) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 快捷入口 -->
    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-icon><Grid /></el-icon>
              <span>快捷入口</span>
            </div>
          </template>
          <div class="quick-links">
            <el-button type="primary" @click="$router.push('/products')">
              <el-icon><Goods /></el-icon>
              商品管理
            </el-button>
            <el-button type="success" @click="$router.push('/factories')">
              <el-icon><OfficeBuilding /></el-icon>
              工厂管理
            </el-button>
            <el-button type="warning" @click="$router.push('/statistics')">
              <el-icon><DataAnalysis /></el-icon>
              数据统计
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { 
  Goods, 
  Box, 
  Money, 
  Warning, 
  TopRight, 
  BottomRight, 
  Grid,
  OfficeBuilding,
  DataAnalysis
} from '@element-plus/icons-vue'
import { statsApi, type StatisticsResponse } from '../api'

const router = useRouter()
const loading = ref(false)
const stats = ref<StatisticsResponse>({
  total_products: 0,
  total_stock: 0,
  total_stock_value: '0',
  total_sales_value: '0',
  low_stock_products: 0,
  total_in_quantity: 0,
  total_in_amount: '0',
  total_out_quantity: 0,
  total_out_amount: '0',
  product_detail_list: []
})

onMounted(() => {
  loadStats()
})

async function loadStats() {
  loading.value = true
  try {
    const res = await statsApi.getStatistics()
    stats.value = res
  } catch (error) {
    ElMessage.error('加载统计数据失败')
  } finally {
    loading.value = false
  }
}

function formatNumber(value: string | number) {
  const num = typeof value === 'string' ? parseFloat(value) : value
  if (isNaN(num)) return '0.00'
  return num.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
</script>

<style scoped>
.dashboard {
  padding: 0;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.stat-card {
  transition: all 0.3s ease;
  height: 100%;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.stat-item {
  display: flex;
  align-items: center;
  padding: 20px;
  gap: 16px;
  min-height: 120px;
}

.stat-icon {
  flex-shrink: 0;
}

.stat-content {
  flex: 1;
  text-align: left;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 14px;
  color: #909399;
}

.detail-card {
  height: 100%;
}

.detail-content {
  padding: 20px;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
}

.detail-label {
  font-size: 14px;
  color: #606266;
}

.detail-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}

.detail-value.text-success {
  color: #67c23a;
}

.detail-value.text-warning {
  color: #e6a23c;
}

.quick-links {
  display: flex;
  gap: 16px;
  padding: 20px;
  justify-content: center;
}

.quick-links .el-button {
  flex: 1;
  max-width: 200px;
  height: 60px;
  font-size: 16px;
}

:deep(.el-divider--horizontal) {
  margin: 16px 0;
}
</style>
