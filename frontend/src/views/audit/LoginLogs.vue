<template>
  <div class="login-logs-page">
    <div class="page-header">
      <h1 class="page-title">登录日志</h1>
    </div>

    <!-- 筛选条件 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filterForm">
        <el-form-item label="用户名">
          <el-input v-model="filterForm.user_name" placeholder="输入用户名" clearable style="width: 180px" />
        </el-form-item>
        <el-form-item label="登录类型">
          <el-select v-model="filterForm.login_type" placeholder="全部" clearable style="width: 120px">
            <el-option label="密码登录" value="PASSWORD" />
            <el-option label="SSO登录" value="SSO" />
            <el-option label="OAuth登录" value="OAUTH" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filterForm.status" placeholder="全部" clearable style="width: 100px">
            <el-option label="成功" :value="1" />
            <el-option label="失败" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="IP地址">
          <el-input v-model="filterForm.ip_address" placeholder="输入IP地址" clearable style="width: 150px" />
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="X"
            style="width: 260px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            查询
          </el-button>
          <el-button @click="handleReset">
            <el-icon><RefreshLeft /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 日志列表 -->
    <el-card class="table-card">
      <el-table :data="logsData" style="width: 100%" v-loading="loading" stripe>
        <el-table-column prop="user_name" label="用户名" width="120" />
        <el-table-column prop="login_type" label="登录类型" width="100">
          <template #default="{ row }">
            <el-tag :type="getLoginTypeTagType(row.login_type)" size="small">
              {{ getLoginTypeLabel(row.login_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="login_ip" label="IP地址" width="140" />
        <el-table-column prop="login_location" label="登录位置" min-width="150" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="Number(row.status) === 1 ? 'success' : 'danger'" size="small">
              {{ Number(row.status) === 1 ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="fail_reason" label="失败原因" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">
            <span :class="{ 'text-danger': row.status === 0 }">{{ row.fail_reason || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="登录时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" size="small" @click="viewDetail(row)">
              <el-icon><View /></el-icon>
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="登录日志详情" width="600px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="日志ID">{{ currentLog?.log_id }}</el-descriptions-item>
        <el-descriptions-item label="用户名">{{ currentLog?.user_name }}</el-descriptions-item>
        <el-descriptions-item label="登录类型">
          <el-tag :type="getLoginTypeTagType(currentLog?.login_type)" size="small">
            {{ getLoginTypeLabel(currentLog?.login_type) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="IP地址">{{ currentLog?.login_ip }}</el-descriptions-item>
        <el-descriptions-item label="登录位置">{{ currentLog?.login_location || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="currentLog?.status === 1 ? 'success' : 'danger'" size="small">
            {{ currentLog?.status === 1 ? '成功' : '失败' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item v-if="currentLog?.status === 0" label="失败原因">
          <span class="text-danger">{{ currentLog?.fail_reason }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="用户代理" :span="2">{{ currentLog?.user_agent }}</el-descriptions-item>
        <el-descriptions-item label="登录时间">{{ formatTime(currentLog?.created_at) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { auditLogApi, type LoginLogInfo, type LoginLogListParams } from '@/api/auditLog'
import { formatTime } from '@/utils/date'

const filterForm = reactive<LoginLogListParams>({
  user_name: '',
  login_type: '',
  status: undefined,
  ip_address: ''
})

const dateRange = ref<number[]>()

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const loading = ref(false)
const detailVisible = ref(false)
const currentLog = ref<LoginLogInfo | null>(null)
const logsData = ref<LoginLogInfo[]>([])

function getLoginTypeLabel(type: string): string {
  const typeMap: Record<string, string> = {
    PASSWORD: '密码',
    SSO: 'SSO',
    OAUTH: 'OAuth'
  }
  return typeMap[type] || type
}

function getLoginTypeTagType(type: string): string {
  const typeMap: Record<string, string> = {
    PASSWORD: 'primary',
    SSO: 'success',
    OAUTH: 'warning'
  }
  return typeMap[type] || ''
}

async function loadLogs() {
  loading.value = true
  try {
    const params: LoginLogListParams = {
      page: pagination.page,
      page_size: pagination.page_size,
      ...filterForm
    }

    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }

    const response = await auditLogApi.getLoginLogs(params)
    logsData.value = response.list || []
    pagination.total = response.total
  } catch (error) {
    console.error('加载登录日志失败:', error)
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  loadLogs()
}

function handleReset() {
  filterForm.user_name = ''
  filterForm.login_type = ''
  filterForm.status = undefined
  filterForm.ip_address = ''
  dateRange.value = undefined
  pagination.page = 1
  loadLogs()
}

function handleSizeChange(size: number) {
  pagination.page_size = size
  loadLogs()
}

function handleCurrentChange(page: number) {
  pagination.page = page
  loadLogs()
}

function viewDetail(row: LoginLogInfo) {
  currentLog.value = row
  detailVisible.value = true
}

onMounted(() => {
  loadLogs()
})
</script>

<style scoped lang="scss">
.login-logs-page {
  .page-header {
    margin-bottom: 24px;

    .page-title {
      font-size: var(--font-size-extra-large);
      font-weight: 600;
      margin: 0;
      color: var(--text-primary);
    }
  }

  .filter-card {
    margin-bottom: 16px;
  }

  .table-card {
    .pagination-container {
      margin-top: 16px;
      display: flex;
      justify-content: flex-end;
    }

    .text-danger {
      color: #f56c6c;
    }
  }
}
</style>
