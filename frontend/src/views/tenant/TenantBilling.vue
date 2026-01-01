<template>
  <div class="tenant-billing-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>账单管理</span>
        </div>
      </template>

      <!-- 搜索筛选 -->
      <div class="search-bar">
        <el-input
          v-model="searchForm.tenant_name"
          placeholder="搜索租户名称"
          clearable
          style="width: 200px;"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-date-picker
          v-model="searchForm.date_range"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="YYYY-MM-DD"
          style="width: 240px;"
          @change="handleSearch"
        />
        <el-select
          v-model="searchForm.status"
          placeholder="全部状态"
          clearable
          style="width: 120px;"
          @change="handleSearch"
        >
          <el-option label="待支付" value="pending" />
          <el-option label="已支付" value="paid" />
          <el-option label="已取消" value="cancelled" />
        </el-select>
        <el-button type="primary" @click="handleSearch">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
        <el-button @click="handleReset">
          <el-icon><Refresh /></el-icon>
          重置
        </el-button>
      </div>

      <!-- 数据表格 -->
      <el-table v-loading="loading" :data="tableData" stripe style="width: 100%;">
        <el-table-column label="账单编号" prop="bill_no" width="180" />
        <el-table-column label="租户名称" prop="tenant_name" width="200" />
        <el-table-column label="账单金额" prop="amount" width="120">
          <template #default="{ row }">
            <span style="color: #f56c6c; font-weight: bold;">¥{{ row.amount }}</span>
          </template>
        </el-table-column>
        <el-table-column label="账单周期" prop="period" width="180" />
        <el-table-column label="账单日期" prop="bill_date" width="120" />
        <el-table-column label="支付时间" prop="paid_time" width="160" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="info" plain @click="handleViewDetail(row)">
                <el-icon><View /></el-icon>
                详情
              </el-button>
              <el-button
                v-if="row.status === 'pending'"
                size="small"
                type="success"
                plain
                @click="handleMarkPaid(row)"
              >
                <el-icon><Check /></el-icon>
                标记支付
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadData"
          @current-change="loadData"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, View, Check } from '@element-plus/icons-vue'

const loading = ref(false)

// 搜索表单
const searchForm = reactive({
  tenant_name: '',
  date_range: [] as string[],
  status: undefined as string | undefined
})

// 分页
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 账单数据
interface BillingInfo {
  id: string
  bill_no: string
  tenant_id: string
  tenant_name: string
  amount: number
  period: string
  bill_date: string
  paid_time: string
  status: string
}

const tableData = ref<BillingInfo[]>([
  {
    id: '1',
    bill_no: 'B202501010001',
    tenant_id: 't1',
    tenant_name: '示例企业A',
    amount: 99,
    period: '2025-01-01 至 2025-02-01',
    bill_date: '2025-01-01',
    paid_time: '2025-01-01 10:30:00',
    status: 'paid'
  },
  {
    id: '2',
    bill_no: 'B202501010002',
    tenant_id: 't2',
    tenant_name: '示例企业B',
    amount: 299,
    period: '2025-01-01 至 2025-04-01',
    bill_date: '2025-01-01',
    paid_time: '-',
    status: 'pending'
  }
])

// 事件处理
const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  Object.assign(searchForm, {
    tenant_name: '',
    date_range: [],
    status: undefined
  })
  handleSearch()
}

const handleViewDetail = (bill: BillingInfo) => {
  ElMessage.info('查看详情功能开发中...')
}

const handleMarkPaid = async (bill: BillingInfo) => {
  try {
    await ElMessageBox.confirm(
      `确定要将账单 "${bill.bill_no}" 标记为已支付吗？`,
      '标记支付',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // TODO: 调用 API 标记支付
    bill.status = 'paid'
    bill.paid_time = new Date().toLocaleString('zh-CN')
    ElMessage.success('标记成功')
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '标记失败')
    }
  }
}

const getStatusType = (status: string) => {
  switch (status) {
    case 'paid':
      return 'success'
    case 'pending':
      return 'warning'
    case 'cancelled':
      return 'info'
    default:
      return 'info'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'paid':
      return '已支付'
    case 'pending':
      return '待支付'
    case 'cancelled':
      return '已取消'
    default:
      return '未知'
  }
}

const loadData = async () => {
  loading.value = true
  try {
    // TODO: 调用 API 获取数据
    pagination.total = tableData.value.length
  } catch (error: any) {
    ElMessage.error(error.message || '获取账单列表失败')
  } finally {
    loading.value = false
  }
}

// 生命周期
onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.tenant-billing-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .search-bar {
    margin-bottom: 16px;
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .pagination-wrapper {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }

  .action-buttons {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    align-items: center;

    .el-button {
      margin: 0;
      padding: 4px 8px;
      font-size: 12px;
      border-radius: 4px;
      transition: all 0.3s ease;

      &:hover {
        transform: translateY(-1px);
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
      }

      .el-icon {
        margin-right: 2px;
        font-size: 12px;
      }
    }
  }
}
</style>
