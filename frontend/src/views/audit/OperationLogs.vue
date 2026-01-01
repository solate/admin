<template>
  <div class="operation-logs-page">
    <div class="page-header">
      <h1 class="page-title">操作日志</h1>
    </div>

    <!-- 筛选条件 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filterForm">
        <el-form-item label="模块">
          <el-select v-model="filterForm.module" placeholder="全部" clearable style="width: 140px">
            <el-option label="用户管理" value="user" />
            <el-option label="角色管理" value="role" />
            <el-option label="租户管理" value="tenant" />
            <el-option label="菜单管理" value="menu" />
            <el-option label="权限管理" value="permission" />
          </el-select>
        </el-form-item>
        <el-form-item label="操作类型">
          <el-select v-model="filterForm.operation_type" placeholder="全部" clearable style="width: 120px">
            <el-option label="创建" value="CREATE" />
            <el-option label="更新" value="UPDATE" />
            <el-option label="删除" value="DELETE" />
            <el-option label="查询" value="QUERY" />
          </el-select>
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="filterForm.user_name" placeholder="输入用户名" clearable style="width: 150px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filterForm.status" placeholder="全部" clearable style="width: 100px">
            <el-option label="成功" :value="1" />
            <el-option label="失败" :value="2" />
          </el-select>
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
        <el-table-column prop="user_name" label="操作人" width="100" />
        <el-table-column prop="module" label="模块" width="100">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ getModuleLabel(row.module) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="operation_type" label="操作类型" width="90">
          <template #default="{ row }">
            <el-tag :type="getOperationTypeTagType(row.operation_type)" size="small">
              {{ getOperationTypeLabel(row.operation_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="resource_type" label="资源类型" width="100" />
        <el-table-column prop="resource_name" label="资源名称" min-width="120" show-overflow-tooltip />
        <el-table-column prop="request_method" label="请求方式" width="80">
          <template #default="{ row }">
            <el-tag :type="getMethodTagType(row.request_method)" size="small">
              {{ row.request_method }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ip_address" label="IP地址" width="130" />
        <el-table-column prop="status" label="状态" width="70">
          <template #default="{ row }">
            <el-tag :type="Number(row.status) === 1 ? 'success' : 'danger'" size="small">
              {{ Number(row.status) === 1 ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="操作时间" width="180">
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
    <el-dialog v-model="detailVisible" title="操作日志详情" width="700px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="日志ID" :span="2">{{ currentLog?.log_id }}</el-descriptions-item>
        <el-descriptions-item label="操作人">{{ currentLog?.user_name }}</el-descriptions-item>
        <el-descriptions-item label="模块">
          <el-tag size="small" type="info">{{ getModuleLabel(currentLog?.module) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="操作类型">
          <el-tag :type="getOperationTypeTagType(currentLog?.operation_type)" size="small">
            {{ getOperationTypeLabel(currentLog?.operation_type) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="资源类型">{{ currentLog?.resource_type || '-' }}</el-descriptions-item>
        <el-descriptions-item label="资源ID">{{ currentLog?.resource_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="资源名称" :span="2">{{ currentLog?.resource_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="请求方式">
          <el-tag :type="getMethodTagType(currentLog?.request_method)" size="small">
            {{ currentLog?.request_method }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="currentLog?.status === 1 ? 'success' : 'danger'" size="small">
            {{ currentLog?.status === 1 ? '成功' : '失败' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="请求路径" :span="2">{{ currentLog?.request_path }}</el-descriptions-item>
        <el-descriptions-item label="IP地址">{{ currentLog?.ip_address }}</el-descriptions-item>
        <el-descriptions-item label="操作时间">{{ formatTime(currentLog?.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="请求参数" :span="2">
          <pre class="json-content">{{ formatJson(currentLog?.request_params) }}</pre>
        </el-descriptions-item>
        <el-descriptions-item v-if="currentLog?.old_value" label="旧值" :span="2">
          <pre class="json-content">{{ formatJson(currentLog?.old_value) }}</pre>
        </el-descriptions-item>
        <el-descriptions-item v-if="currentLog?.new_value" label="新值" :span="2">
          <pre class="json-content">{{ formatJson(currentLog?.new_value) }}</pre>
        </el-descriptions-item>
        <el-descriptions-item v-if="currentLog?.status === 2 && currentLog?.error_message" label="错误信息" :span="2">
          <span class="text-danger">{{ currentLog.error_message }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="用户代理" :span="2">{{ currentLog?.user_agent }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { auditLogApi, type OperationLogInfo, type OperationLogListParams } from '@/api/auditLog'
import { formatTime } from '@/utils/date'

const filterForm = reactive<OperationLogListParams>({
  module: '',
  operation_type: '',
  user_name: '',
  status: undefined
})

const dateRange = ref<number[]>()

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const loading = ref(false)
const detailVisible = ref(false)
const currentLog = ref<OperationLogInfo | null>(null)
const logsData = ref<OperationLogInfo[]>([])

function getModuleLabel(module: string | undefined): string {
  const moduleMap: Record<string, string> = {
    user: '用户管理',
    role: '角色管理',
    tenant: '租户管理',
    menu: '菜单管理',
    permission: '权限管理'
  }
  return moduleMap[module || ''] || module || '-'
}

function getOperationTypeLabel(type: string | undefined): string {
  const typeMap: Record<string, string> = {
    CREATE: '创建',
    UPDATE: '更新',
    DELETE: '删除',
    QUERY: '查询',
    EXPORT: '导出',
    LOGIN: '登录',
    LOGOUT: '登出'
  }
  return typeMap[type || ''] || type || '-'
}

function getOperationTypeTagType(type: string | undefined): string {
  const typeMap: Record<string, string> = {
    CREATE: 'success',
    UPDATE: 'warning',
    DELETE: 'danger',
    QUERY: 'info',
    EXPORT: 'primary'
  }
  return typeMap[type || ''] || ''
}

function getMethodTagType(method: string | undefined): string {
  const methodMap: Record<string, string> = {
    GET: 'info',
    POST: 'success',
    PUT: 'warning',
    DELETE: 'danger'
  }
  return methodMap[method || ''] || ''
}

function formatJson(json: string | undefined): string {
  if (!json) return '-'
  try {
    const parsed = JSON.parse(json)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return json
  }
}

async function loadLogs() {
  loading.value = true
  try {
    const params: OperationLogListParams = {
      page: pagination.page,
      page_size: pagination.page_size,
      ...filterForm
    }

    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }

    const response = await auditLogApi.getOperationLogs(params)
    logsData.value = response.list || []
    pagination.total = response.total
  } catch (error) {
    console.error('加载操作日志失败:', error)
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.page = 1
  loadLogs()
}

function handleReset() {
  filterForm.module = ''
  filterForm.operation_type = ''
  filterForm.user_name = ''
  filterForm.status = undefined
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

function viewDetail(row: OperationLogInfo) {
  currentLog.value = row
  detailVisible.value = true
}

onMounted(() => {
  loadLogs()
})
</script>

<style scoped lang="scss">
.operation-logs-page {
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

    .json-content {
      margin: 0;
      padding: 8px;
      background: #f5f7fa;
      border-radius: 4px;
      font-size: 12px;
      max-height: 200px;
      overflow: auto;
    }

    .text-danger {
      color: #f56c6c;
    }
  }
}
</style>
