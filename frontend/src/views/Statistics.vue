<template>
  <div class="statistics">
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
              <div class="stat-desc">按采购价计算</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon">
              <el-icon :size="40" color="#f56c6c"><Money /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">¥{{ formatNumber(stats.total_sales_value) }}</div>
              <div class="stat-label">总销售价值</div>
              <div class="stat-desc">按销售价计算</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 低库存、入库、出库统计 -->
    <el-row :gutter="20" style="margin-top: 20px;" v-loading="loading">
      <el-col :span="8">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon">
              <el-icon :size="40" color="#f56c6c"><Warning /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.low_stock_products }}</div>
              <div class="stat-label">低库存商品数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon">
              <el-icon :size="40" color="#67c23a"><TopRight /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.total_in_quantity }}</div>
              <div class="stat-label">总入库数量</div>
              <div class="stat-desc">¥{{ formatNumber(stats.total_in_amount) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon">
              <el-icon :size="40" color="#e6a23c"><BottomRight /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.total_out_quantity }}</div>
              <div class="stat-label">总出库数量</div>
              <div class="stat-desc">¥{{ formatNumber(stats.total_out_amount) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 商品明细表格 -->
    <el-card style="margin-top: 20px;">
      <template #header>
        <div class="card-header">
          <span>商品库存明细</span>
          <el-button @click="loadStats" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>
      
      <el-table 
        :data="stats.product_detail_list" 
        v-loading="loading" 
        style="width: 100%;"
        :row-class-name="getRowClassName"
      >
        <el-table-column prop="product_name" label="商品名称" min-width="150" />
        <el-table-column prop="unit" label="单位" width="80" />
        <el-table-column prop="current_stock" label="当前库存" width="100" align="right">
          <template #default="{ row }">
            <el-tag :type="getStockType(row)">{{ row.current_stock }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="total_in_quantity" label="累计入库" width="100" align="right">
          <template #default="{ row }">
            <span class="text-success">+{{ row.total_in_quantity }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="total_out_quantity" label="累计出库" width="100" align="right">
          <template #default="{ row }">
            <span class="text-warning">-{{ row.total_out_quantity }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="purchase_price" label="采购价" width="100" align="right">
          <template #default="{ row }">
            ¥{{ row.purchase_price }}
          </template>
        </el-table-column>
        <el-table-column prop="sale_price" label="销售价" width="100" align="right">
          <template #default="{ row }">
            ¥{{ row.sale_price }}
          </template>
        </el-table-column>
        <el-table-column prop="stock_value" label="库存价值" width="120" align="right">
          <template #default="{ row }">
            <span class="text-primary">¥{{ formatNumber(row.stock_value) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Goods, Box, Money, Refresh, Warning, TopRight, BottomRight } from '@element-plus/icons-vue'
import { statsApi, type StatisticsResponse } from '../api'

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

function isLowStock(row: any) {
  return row.current_stock <= row.min_stock
}

function getStockType(row: any) {
  if (row.current_stock === 0) return 'danger'
  if (row.current_stock <= row.min_stock) return 'danger'
  if (row.current_stock <= row.min_stock * 1.5) return 'warning'
  return 'success'
}

function getRowClassName({ row }: { row: any }) {
  return isLowStock(row) ? 'low-stock-row' : ''
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stat-card {
  transition: all 0.3s ease;
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
  margin-bottom: 2px;
}

.stat-desc {
  font-size: 12px;
  color: #c0c4cc;
}

.text-success {
  color: #67c23a;
  font-weight: 500;
}

.text-warning {
  color: #e6a23c;
  font-weight: 500;
}

.text-primary {
  color: #409eff;
  font-weight: 600;
}

:deep(.el-table) {
  font-size: 14px;
}

:deep(.el-table th) {
  background-color: #f5f7fa;
  font-weight: 600;
}
</style>
