<template>
  <div class="tenants-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>租户管理</span>
          <div class="header-actions">
            <el-button type="primary" @click="handleCreate">
              <el-icon><Plus /></el-icon>
              新建租户
            </el-button>
            <el-button @click="handleExport">
              <el-icon><Download /></el-icon>
              导出数据
            </el-button>
          </div>
        </div>
      </template>

      <!-- 搜索筛选 -->
      <div class="search-bar">
        <el-form :model="searchForm" inline class="search-form">
          <el-form-item label="租户名称">
            <el-input
              v-model="searchForm.name"
              placeholder="请输入租户名称"
              clearable
              style="width: 200px;"
              @keyup.enter="handleSearch"
            />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="searchForm.status" placeholder="全部状态" clearable style="width: 120px;">
              <el-option label="正常运营" value="active" />
              <el-option label="试用期" value="trial" />
              <el-option label="已暂停" value="suspended" />
              <el-option label="已过期" value="expired" />
            </el-select>
          </el-form-item>
          <el-form-item label="套餐类型">
            <el-select v-model="searchForm.plan" placeholder="全部套餐" clearable style="width: 120px;">
              <el-option label="基础版" value="basic" />
              <el-option label="专业版" value="professional" />
              <el-option label="企业版" value="enterprise" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">
              <el-icon><Search /></el-icon>
              搜索
            </el-button>
            <el-button @click="handleReset">
              <el-icon><Refresh /></el-icon>
              重置
            </el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 数据表格 -->
      <el-table
        v-loading="loading"
        :data="tableData"
        stripe
        @selection-change="handleSelectionChange"
        style="width: 100%;"
      >
      <el-table-column type="selection" width="55" />
      <el-table-column label="租户信息" min-width="280">
        <template #default="{ row }">
          <div class="tenant-info">
            <el-avatar :size="36" :style="{ backgroundColor: row.color }">
              {{ row.name.charAt(0).toUpperCase() }}
            </el-avatar>
            <span class="tenant-name">{{ row.name }}</span>
            <span class="tenant-domain">{{ row.domain }}</span>
            <span class="tenant-admin">管理员：{{ row.adminName }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="套餐" width="120">
        <template #default="{ row }">
          <el-tag :type="getPlanType(row.plan)">
            {{ getPlanName(row.plan) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)">
            {{ getStatusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="用户数" width="80">
        <template #default="{ row }">
          {{ row.userCount }}
        </template>
      </el-table-column>
      <el-table-column label="到期时间" width="120">
        <template #default="{ row }">
          <span :class="{ 'text-danger': isExpiringSoon(row.expiryDate) }">
            {{ formatDate(row.expiryDate) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="120">
        <template #default="{ row }">
          {{ formatDate(row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <div class="action-buttons">
            <el-button size="small" type="primary" plain @click="handleEdit(row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button size="small" type="success" plain @click="handleManageUsers(row)">
              <el-icon><User /></el-icon>
              用户
            </el-button>
            <el-button size="small" type="warning" plain @click="handleRenew(row)">
              <el-icon><RefreshRight /></el-icon>
              续费
            </el-button>
            <el-dropdown @command="(command) => handleMoreAction(command, row)">
              <el-button size="small" type="info" plain>
                更多<el-icon class="el-icon--right"><arrow-down /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="view">查看详情</el-dropdown-item>
                  <el-dropdown-item command="statistics">数据统计</el-dropdown-item>
                  <el-dropdown-item command="settings">租户设置</el-dropdown-item>
                  <el-dropdown-item
                    :command="row.status === 'active' ? 'suspend' : 'activate'"
                    :divided="row.status === 'active'"
                  >
                    {{ row.status === 'active' ? '暂停服务' : '激活服务' }}
                  </el-dropdown-item>
                  <el-dropdown-item command="delete" divided class="danger-item">
                    删除租户
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <Pagination
      v-model:current-page="pagination.page"
      v-model:page-size="pagination.size"
      :total="pagination.total"
      @change="fetchTenants"
    />
    </el-card>

    <!-- 租户表单对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'create' ? '创建租户' : '编辑租户'"
      width="700px"
      @close="handleDialogClose"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="120px"
      >
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="租户名称" prop="name">
              <el-input
                v-model="formData.name"
                placeholder="请输入租户名称"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="域名" prop="domain">
              <el-input
                v-model="formData.domain"
                placeholder="请输入域名"
                :disabled="dialogType === 'edit'"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="管理员姓名" prop="adminName">
              <el-input
                v-model="formData.adminName"
                placeholder="请输入管理员姓名"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="管理员邮箱" prop="adminEmail">
              <el-input
                v-model="formData.adminEmail"
                placeholder="请输入管理员邮箱"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="套餐类型" prop="plan">
              <el-select v-model="formData.plan" placeholder="请选择套餐">
                <el-option label="基础版" value="basic" />
                <el-option label="专业版" value="professional" />
                <el-option label="企业版" value="enterprise" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态" prop="status">
              <el-radio-group v-model="formData.status">
                <el-radio label="active">正常运营</el-radio>
                <el-radio label="trial">试用期</el-radio>
                <el-radio label="suspended">已暂停</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="用户限制" prop="userLimit">
              <el-input-number
                v-model="formData.userLimit"
                :min="1"
                :max="10000"
                controls-position="right"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="到期时间" prop="expiryDate">
              <el-date-picker
                v-model="formData.expiryDate"
                type="date"
                placeholder="请选择到期时间"
                value-format="YYYY-MM-DD"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入租户描述"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 续费对话框 -->
    <el-dialog v-model="renewDialogVisible" title="租户续费" width="500px">
      <el-form :model="renewForm" label-width="100px">
        <el-form-item label="当前套餐">
          <el-tag :type="getPlanType(currentTenant?.plan || '')">
            {{ getPlanName(currentTenant?.plan || '') }}
          </el-tag>
        </el-form-item>
        <el-form-item label="续费时长">
          <el-radio-group v-model="renewForm.duration">
            <el-radio :label="1">1个月</el-radio>
            <el-radio :label="3">3个月</el-radio>
            <el-radio :label="6">6个月</el-radio>
            <el-radio :label="12">12个月</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="新到期时间">
          <el-date-picker
            v-model="renewForm.newExpiryDate"
            type="date"
            placeholder="请选择到期时间"
            value-format="YYYY-MM-DD"
            style="width: 100%"
            disabled
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="renewDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="renewLoading" @click="handleRenewSubmit">
          确认续费
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { Edit, User, RefreshRight } from '@element-plus/icons-vue'
import Pagination from '../../components/Pagination.vue'
import dayjs from 'dayjs'

// 接口定义
interface Tenant {
  id: number
  name: string
  domain: string
  adminName: string
  adminEmail: string
  color: string
  plan: 'basic' | 'professional' | 'enterprise'
  status: 'active' | 'trial' | 'suspended' | 'expired'
  userCount: number
  userLimit: number
  expiryDate: string
  createdAt: string
  description?: string
}

const router = useRouter()

// 响应式数据
const loading = ref(false)
const submitLoading = ref(false)
const renewLoading = ref(false)
const dialogVisible = ref(false)
const renewDialogVisible = ref(false)
const dialogType = ref<'create' | 'edit'>('create')
const selectedTenants = ref<Tenant[]>([])
const currentTenant = ref<Tenant | null>(null)

// 搜索表单
const searchForm = reactive({
  name: '',
  status: '',
  plan: '',
  dateRange: [] as string[]
})

// 分页
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 租户统计
const tenantStats = reactive({
  active: 0,
  trial: 0,
  totalUsers: 0
})

// 租户表单
const formData = reactive({
  id: 0,
  name: '',
  domain: '',
  adminName: '',
  adminEmail: '',
  plan: 'basic' as 'basic' | 'professional' | 'enterprise',
  status: 'active' as 'active' | 'trial' | 'suspended',
  userLimit: 10,
  expiryDate: '',
  description: ''
})

// 续费表单
const renewForm = reactive({
  duration: 12,
  newExpiryDate: ''
})

// 表单验证规则
const formRules = {
  name: [
    { required: true, message: '请输入租户名称', trigger: 'blur' },
    { min: 2, max: 50, message: '租户名称长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  domain: [
    { required: true, message: '请输入域名', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9-]+$/, message: '域名只能包含字母、数字和连字符', trigger: 'blur' }
  ],
  adminName: [
    { required: true, message: '请输入管理员姓名', trigger: 'blur' }
  ],
  adminEmail: [
    { required: true, message: '请输入管理员邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  plan: [
    { required: true, message: '请选择套餐类型', trigger: 'change' }
  ],
  userLimit: [
    { required: true, message: '请输入用户限制', trigger: 'blur' }
  ],
  expiryDate: [
    { required: true, message: '请选择到期时间', trigger: 'change' }
  ]
}

// 租户数据
const tableData = ref<Tenant[]>([
  {
    id: 1,
    name: '科技有限公司',
    domain: 'tech-company',
    adminName: '张三',
    adminEmail: 'zhangsan@techcompany.com',
    color: '#409eff',
    plan: 'professional',
    status: 'active',
    userCount: 45,
    userLimit: 100,
    expiryDate: '2025-12-31',
    createdAt: '2025-01-01',
    description: '专注于企业级软件开发'
  },
  {
    id: 2,
    name: '电商集团',
    domain: 'ecommerce-group',
    adminName: '李四',
    adminEmail: 'lisi@ecommerce.com',
    color: '#67c23a',
    plan: 'enterprise',
    status: 'active',
    userCount: 120,
    userLimit: 500,
    expiryDate: '2026-06-30',
    createdAt: '2025-02-15',
    description: '大型电商平台'
  },
  {
    id: 3,
    name: '创业团队',
    domain: 'startup-team',
    adminName: '王五',
    adminEmail: 'wangwu@startup.com',
    color: '#e6a23c',
    plan: 'basic',
    status: 'trial',
    userCount: 8,
    userLimit: 10,
    expiryDate: '2025-12-30',
    createdAt: '2025-11-01',
    description: '初创互联网公司'
  }
])

const formRef = ref<FormInstance>()

// 计算属性
const getPlanType = (plan: string) => {
  const types: Record<string, string> = {
    basic: 'info',
    professional: 'warning',
    enterprise: 'danger'
  }
  return types[plan] || 'info'
}

const getPlanName = (plan: string) => {
  const names: Record<string, string> = {
    basic: '基础版',
    professional: '专业版',
    enterprise: '企业版'
  }
  return names[plan] || plan
}

const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    active: 'success',
    trial: 'warning',
    suspended: 'danger',
    expired: 'info'
  }
  return types[status] || 'info'
}

const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    active: '正常运营',
    trial: '试用期',
    suspended: '已暂停',
    expired: '已过期'
  }
  return texts[status] || status
}

const formatDate = (date: string) => {
  return dayjs(date).format('YYYY-MM-DD')
}

const isExpiringSoon = (date: string) => {
  const days = dayjs(date).diff(dayjs(), 'days')
  return days < 30 && days >= 0
}

// 监听续费时长变化
watch(
  () => renewForm.duration,
  (duration) => {
    if (currentTenant.value) {
      const newDate = dayjs(currentTenant.value.expiryDate).add(duration, 'month')
      renewForm.newExpiryDate = newDate.format('YYYY-MM-DD')
    }
  }
)

// 事件处理
const handleSearch = () => {
  pagination.page = 1
  fetchTenants()
}

const handleReset = () => {
  Object.assign(searchForm, {
    name: '',
    status: '',
    plan: '',
    dateRange: []
  })
  handleSearch()
}

const handleCreate = () => {
  dialogType.value = 'create'
  resetFormData()
  dialogVisible.value = true
}

const handleEdit = (tenant: Tenant) => {
  dialogType.value = 'edit'
  Object.assign(formData, {
    id: tenant.id,
    name: tenant.name,
    domain: tenant.domain,
    adminName: tenant.adminName,
    adminEmail: tenant.adminEmail,
    plan: tenant.plan,
    status: tenant.status,
    userLimit: tenant.userLimit,
    expiryDate: tenant.expiryDate,
    description: tenant.description || ''
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate()
  submitLoading.value = true

  try {
    await new Promise(resolve => setTimeout(resolve, 1000))

    if (dialogType.value === 'create') {
      const newTenant: Tenant = {
        id: Date.now(),
        name: formData.name,
        domain: formData.domain,
        adminName: formData.adminName,
        adminEmail: formData.adminEmail,
        color: `#${Math.floor(Math.random()*16777215).toString(16)}`,
        plan: formData.plan,
        status: formData.status,
        userCount: 0,
        userLimit: formData.userLimit,
        expiryDate: formData.expiryDate,
        createdAt: dayjs().format('YYYY-MM-DD'),
        description: formData.description
      }
      tableData.value.unshift(newTenant)
      ElMessage.success('租户创建成功')
    } else {
      const index = tableData.value.findIndex(t => t.id === formData.id)
      if (index !== -1) {
        Object.assign(tableData.value[index], {
          name: formData.name,
          adminName: formData.adminName,
          adminEmail: formData.adminEmail,
          plan: formData.plan,
          status: formData.status,
          userLimit: formData.userLimit,
          expiryDate: formData.expiryDate,
          description: formData.description
        })
        ElMessage.success('租户更新成功')
      }
    }

    dialogVisible.value = false
    fetchTenants()
  } catch (error) {
    ElMessage.error('操作失败')
  } finally {
    submitLoading.value = false
  }
}

const handleRenew = (tenant: Tenant) => {
  currentTenant.value = tenant
  renewForm.duration = 12
  const newDate = dayjs(tenant.expiryDate).add(12, 'month')
  renewForm.newExpiryDate = newDate.format('YYYY-MM-DD')
  renewDialogVisible.value = true
}

const handleRenewSubmit = async () => {
  if (!currentTenant.value) return

  renewLoading.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 1000))

    const index = tableData.value.findIndex(t => t.id === currentTenant.value!.id)
    if (index !== -1) {
      tableData.value[index].expiryDate = renewForm.newExpiryDate
      if (tableData.value[index].status === 'expired') {
        tableData.value[index].status = 'active'
      }
    }

    ElMessage.success('续费成功')
    renewDialogVisible.value = false
    fetchTenants()
  } catch (error) {
    ElMessage.error('续费失败')
  } finally {
    renewLoading.value = false
  }
}

const handleManageUsers = (tenant: Tenant) => {
  router.push(`/system/users?tenantId=${tenant.id}`)
}

const handleMoreAction = async (command: string, tenant: Tenant) => {
  switch (command) {
    case 'view':
      ElMessage.info('查看详情功能开发中...')
      break
    case 'statistics':
      ElMessage.info('数据统计功能开发中...')
      break
    case 'settings':
      ElMessage.info('租户设置功能开发中...')
      break
    case 'activate':
    case 'suspend':
      try {
        const action = command === 'activate' ? '激活' : '暂停'
        await ElMessageBox.confirm(
          `确定要${action}租户 "${tenant.name}" 吗？`,
          `${action}租户`,
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )

        tenant.status = command === 'activate' ? 'active' : 'suspended'
        ElMessage.success(`租户${action}成功`)
      } catch {
        // 用户取消
      }
      break
    case 'delete':
      try {
        await ElMessageBox.confirm(
          `确定要删除租户 "${tenant.name}" 吗？此操作不可恢复！`,
          '删除租户',
          {
            confirmButtonText: '确定删除',
            cancelButtonText: '取消',
            type: 'error'
          }
        )

        const index = tableData.value.findIndex(t => t.id === tenant.id)
        if (index !== -1) {
          tableData.value.splice(index, 1)
          ElMessage.success('租户删除成功')
          fetchTenants()
        }
      } catch {
        // 用户取消
      }
      break
  }
}

const handleExport = () => {
  ElMessage.info('导出数据功能开发中...')
}

const handleSelectionChange = (selection: Tenant[]) => {
  selectedTenants.value = selection
}

const handleSizeChange = (size: number) => {
  pagination.size = size
  fetchTenants()
}

const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchTenants()
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}

const resetFormData = () => {
  Object.assign(formData, {
    id: 0,
    name: '',
    domain: '',
    adminName: '',
    adminEmail: '',
    plan: 'basic',
    status: 'active',
    userLimit: 10,
    expiryDate: dayjs().add(1, 'year').format('YYYY-MM-DD'),
    description: ''
  })
}

const fetchTenants = async () => {
  loading.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 500))

    // 更新统计
    tenantStats.active = tableData.value.filter(t => t.status === 'active').length
    tenantStats.trial = tableData.value.filter(t => t.status === 'trial').length
    tenantStats.totalUsers = tableData.value.reduce((sum, t) => sum + t.userCount, 0)

    pagination.total = tableData.value.length
  } catch (error) {
    ElMessage.error('获取租户列表失败')
  } finally {
    loading.value = false
  }
}

// 生命周期
onMounted(() => {
  fetchTenants()
})
</script>

<style scoped lang="scss">
.tenants-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .header-actions {
      display: flex;
      gap: 10px;
    }
  }

  .search-bar {
    margin-bottom: 16px;

    .search-form {
      .el-form-item {
        margin-bottom: 0;
      }
    }
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

  .tenant-info {
    display: flex;
    align-items: center;
    gap: 12px;

    .tenant-details {
      display: flex;
      align-items: center;
      gap: 16px;

      .tenant-name {
        font-size: 14px;
        font-weight: 600;
        color: var(--text-primary);
      }

      .tenant-domain {
        font-size: 13px;
        color: var(--text-secondary);
      }

      .tenant-admin {
        font-size: 12px;
        color: var(--text-regular);
      }
    }
  }

  .text-danger {
    color: var(--danger-color);
    font-weight: 500;
  }
}

// 危险操作样式
:deep(.danger-item) {
  color: var(--danger-color);
}
</style>