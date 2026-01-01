<template>
  <div class="tenant-subscription-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>订阅管理</span>
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
        <el-select
          v-model="searchForm.status"
          placeholder="全部状态"
          clearable
          style="width: 120px;"
          @change="handleSearch"
        >
          <el-option label="正常" value="active" />
          <el-option label="即将到期" value="expiring" />
          <el-option label="已过期" value="expired" />
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
        <el-table-column label="租户名称" prop="tenant_name" width="200" />
        <el-table-column label="套餐名称" prop="package_name" width="150" />
        <el-table-column label="订阅时间" prop="start_date" width="120" />
        <el-table-column label="到期时间" prop="end_date" width="120" />
        <el-table-column label="剩余天数" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.days_left > 30" type="success">{{ row.days_left }}天</el-tag>
            <el-tag v-else-if="row.days_left > 0" type="warning">{{ row.days_left }}天</el-tag>
            <el-tag v-else type="danger">已过期</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'">
              {{ row.status === 'active' ? '正常' : '已过期' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" plain @click="handleRenew(row)">
                <el-icon><RefreshRight /></el-icon>
                续费
              </el-button>
              <el-button size="small" type="info" plain @click="handleViewDetail(row)">
                <el-icon><View /></el-icon>
                详情
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

    <!-- 续费对话框 -->
    <el-dialog
      v-model="renewDialogVisible"
      title="续费订阅"
      width="500px"
      :close-on-click-modal="false"
      @close="handleRenewDialogClose"
    >
      <el-form ref="renewFormRef" :model="renewForm" :rules="renewFormRules" label-width="100px">
        <el-form-item label="租户名称">
          <el-input :value="currentSubscription?.tenant_name" disabled />
        </el-form-item>
        <el-form-item label="当前套餐">
          <el-input :value="currentSubscription?.package_name" disabled />
        </el-form-item>
        <el-form-item label="到期时间">
          <el-input :value="currentSubscription?.end_date" disabled />
        </el-form-item>
        <el-form-item label="续费时长" prop="duration">
          <el-select v-model="renewForm.duration" placeholder="请选择续费时长">
            <el-option label="1个月" :value="1" />
            <el-option label="3个月" :value="3" />
            <el-option label="6个月" :value="6" />
            <el-option label="1年" :value="12" />
            <el-option label="2年" :value="24" />
            <el-option label="3年" :value="36" />
          </el-select>
        </el-form-item>
        <el-form-item label="新到期时间">
          <el-input :value="calculateNewEndDate()" disabled />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="renewDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleRenewSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Search, Refresh, RefreshRight, View } from '@element-plus/icons-vue'

const loading = ref(false)
const submitLoading = ref(false)
const renewDialogVisible = ref(false)

const renewFormRef = ref<FormInstance>()

// 搜索表单
const searchForm = reactive({
  tenant_name: '',
  status: undefined as string | undefined
})

// 分页
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 订阅数据
interface SubscriptionInfo {
  id: string
  tenant_id: string
  tenant_name: string
  package_id: string
  package_name: string
  start_date: string
  end_date: string
  days_left: number
  status: string
}

const tableData = ref<SubscriptionInfo[]>([
  {
    id: '1',
    tenant_id: 't1',
    tenant_name: '示例企业A',
    package_id: 'p1',
    package_name: '基础版',
    start_date: '2024-01-01',
    end_date: '2025-01-01',
    days_left: 365,
    status: 'active'
  },
  {
    id: '2',
    tenant_id: 't2',
    tenant_name: '示例企业B',
    package_id: 'p2',
    package_name: '专业版',
    start_date: '2024-06-01',
    end_date: '2024-12-01',
    days_left: 15,
    status: 'active'
  }
])

const currentSubscription = ref<SubscriptionInfo | null>(null)

// 续费表单
const renewForm = reactive({
  duration: 12
})

const renewFormRules: FormRules = {
  duration: [
    { required: true, message: '请选择续费时长', trigger: 'change' }
  ]
}

// 事件处理
const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  Object.assign(searchForm, {
    tenant_name: '',
    status: undefined
  })
  handleSearch()
}

const handleRenew = (subscription: SubscriptionInfo) => {
  currentSubscription.value = subscription
  renewForm.duration = 12
  renewDialogVisible.value = true
}

const handleRenewSubmit = async () => {
  if (!renewFormRef.value) return

  try {
    await renewFormRef.value.validate()
  } catch {
    return
  }

  submitLoading.value = true

  try {
    // TODO: 调用 API 续费
    ElMessage.success('续费成功')
    renewDialogVisible.value = false
    loadData()
  } catch (error: any) {
    ElMessage.error(error.message || '续费失败')
  } finally {
    submitLoading.value = false
  }
}

const handleRenewDialogClose = () => {
  renewFormRef.value?.resetFields()
  currentSubscription.value = null
}

const handleViewDetail = (subscription: SubscriptionInfo) => {
  ElMessage.info('查看详情功能开发中...')
}

const calculateNewEndDate = () => {
  if (!currentSubscription.value) return '-'
  const endDate = new Date(currentSubscription.value.end_date)
  endDate.setMonth(endDate.getMonth() + renewForm.duration)
  return endDate.toISOString().split('T')[0]
}

const loadData = async () => {
  loading.value = true
  try {
    // TODO: 调用 API 获取数据
    pagination.total = tableData.value.length
  } catch (error: any) {
    ElMessage.error(error.message || '获取订阅列表失败')
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
.tenant-subscription-page {
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
