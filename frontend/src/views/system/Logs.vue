<template>
  <div class="logs-page">
    <div class="page-header">
      <h1 class="page-title">操作日志</h1>
      <el-button type="primary" @click="exportLogs">
        <el-icon><Download /></el-icon>
        导出日志
      </el-button>
    </div>

    <!-- 筛选条件 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filterForm">
        <el-form-item label="操作类型">
          <el-select v-model="filterForm.type" placeholder="全部" clearable style="width: 150px">
            <el-option label="全部" value="" />
            <el-option label="登录" value="login" />
            <el-option label="创建" value="create" />
            <el-option label="更新" value="update" />
            <el-option label="删除" value="delete" />
            <el-option label="导出" value="export" />
          </el-select>
        </el-form-item>
        <el-form-item label="操作人">
          <el-input v-model="filterForm.operator" placeholder="输入操作人" clearable style="width: 200px" />
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filterForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            style="width: 280px"
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
      <el-table :data="logsData" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="operator" label="操作人" width="120">
          <template #default="{ row }">
            <div class="operator-cell">
              <el-avatar :size="32">{{ row.operator.charAt(0) }}</el-avatar>
              <span>{{ row.operator }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="module" label="操作模块" width="120" />
        <el-table-column prop="action" label="操作类型" width="100">
          <template #default="{ row }">
            <el-tag :type="getActionTagType(row.actionType)" size="small">
              {{ row.action }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="操作描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="ip" label="IP地址" width="140" />
        <el-table-column prop="createdAt" label="操作时间" width="180" />
        <el-table-column label="操作" width="100" fixed="right">
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
          v-model:current-page="pagination.currentPage"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="日志详情" width="600px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="操作人">{{ currentLog?.operator }}</el-descriptions-item>
        <el-descriptions-item label="操作模块">{{ currentLog?.module }}</el-descriptions-item>
        <el-descriptions-item label="操作类型">
          <el-tag :type="getActionTagType(currentLog?.actionType)" size="small">
            {{ currentLog?.action }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="操作描述">{{ currentLog?.description }}</el-descriptions-item>
        <el-descriptions-item label="IP地址">{{ currentLog?.ip }}</el-descriptions-item>
        <el-descriptions-item label="用户代理">{{ currentLog?.userAgent }}</el-descriptions-item>
        <el-descriptions-item label="操作时间">{{ currentLog?.createdAt }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'

interface LogItem {
  id: number
  operator: string
  module: string
  action: string
  actionType: string
  description: string
  ip: string
  userAgent: string
  createdAt: string
}

const filterForm = reactive({
  type: '',
  operator: '',
  dateRange: null as any
})

const pagination = reactive({
  currentPage: 1,
  pageSize: 20,
  total: 0
})

const loading = ref(false)
const detailVisible = ref(false)
const currentLog = ref<LogItem | null>(null)

// 模拟日志数据
const logsData = ref<LogItem[]>([
  {
    id: 1,
    operator: '张三',
    module: '用户管理',
    action: '创建',
    actionType: 'create',
    description: '创建了用户 "李四"',
    ip: '192.168.1.100',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    createdAt: '2024-01-15 14:32:18'
  },
  {
    id: 2,
    operator: '李四',
    module: '租户管理',
    action: '更新',
    actionType: 'update',
    description: '更新了租户 "科技公司" 的配置信息',
    ip: '192.168.1.101',
    userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
    createdAt: '2024-01-15 14:28:45'
  },
  {
    id: 3,
    operator: '王五',
    module: '角色管理',
    action: '删除',
    actionType: 'delete',
    description: '删除了角色 "临时角色"',
    ip: '192.168.1.102',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    createdAt: '2024-01-15 14:15:22'
  },
  {
    id: 4,
    operator: '赵六',
    module: '权限管理',
    action: '导出',
    actionType: 'export',
    description: '导出了权限报表',
    ip: '192.168.1.103',
    userAgent: 'Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36',
    createdAt: '2024-01-15 13:58:10'
  },
  {
    id: 5,
    operator: '张三',
    module: '系统设置',
    action: '更新',
    actionType: 'update',
    description: '修改了系统配置 "最大上传文件大小"',
    ip: '192.168.1.100',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    createdAt: '2024-01-15 13:42:55'
  }
])

pagination.total = 156

function getActionTagType(type: string | undefined): string {
  const typeMap: Record<string, string> = {
    login: 'success',
    create: 'primary',
    update: 'warning',
    delete: 'danger',
    export: 'info'
  }
  return typeMap[type || ''] || ''
}

function handleSearch() {
  loading.value = true
  setTimeout(() => {
    loading.value = false
    ElMessage.success('查询成功')
  }, 500)
}

function handleReset() {
  filterForm.type = ''
  filterForm.operator = ''
  filterForm.dateRange = null
  ElMessage.info('已重置筛选条件')
}

function handleSizeChange(size: number) {
  pagination.pageSize = size
  handleSearch()
}

function handleCurrentChange(page: number) {
  pagination.currentPage = page
  handleSearch()
}

function viewDetail(row: LogItem) {
  currentLog.value = row
  detailVisible.value = true
}

function exportLogs() {
  ElMessage.success('日志导出任务已启动')
}
</script>

<style scoped lang="scss">
.logs-page {
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

  .filter-card {
    margin-bottom: 16px;
  }

  .table-card {
    .operator-cell {
      display: flex;
      align-items: center;
      gap: 8px;

      span {
        font-size: var(--font-size-small);
      }
    }

    .pagination-container {
      margin-top: 16px;
      display: flex;
      justify-content: flex-end;
    }
  }
}
</style>