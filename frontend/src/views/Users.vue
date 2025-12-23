<template>
  <div class="users-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">用户管理</h1>
        <p class="page-subtitle">管理系统中的所有用户账户</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          创建用户
        </el-button>
        <el-button @click="handleImport">
          <el-icon><Upload /></el-icon>
          批量导入
        </el-button>
        <el-button @click="handleExport">
          <el-icon><Download /></el-icon>
          导出数据
        </el-button>
      </div>
    </div>

    <!-- 搜索筛选 -->
    <el-card class="search-card">
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="关键词">
          <el-input
            v-model="searchForm.keyword"
            placeholder="用户名/邮箱/手机号"
            clearable
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部状态" clearable>
            <el-option label="正常" value="active" />
            <el-option label="禁用" value="disabled" />
            <el-option label="待激活" value="pending" />
          </el-select>
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="searchForm.roleId" placeholder="全部角色" clearable>
            <el-option
              v-for="role in roleOptions"
              :key="role.id"
              :label="role.name"
              :value="role.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="创建时间">
          <el-date-picker
            v-model="searchForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
          />
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
    </el-card>

    <!-- 数据表格 -->
    <el-card class="table-card">
      <div class="table-header">
        <div class="table-title">
          <span>用户列表</span>
          <el-tag type="info" size="small">{{ pagination.total }} 条记录</el-tag>
        </div>
        <div class="table-actions">
          <el-button-group>
            <el-button
              :type="viewMode === 'table' ? 'primary' : ''"
              @click="viewMode = 'table'"
            >
              <el-icon><Grid /></el-icon>
            </el-button>
            <el-button
              :type="viewMode === 'card' ? 'primary' : ''"
              @click="viewMode = 'card'"
            >
              <el-icon><List /></el-icon>
            </el-button>
          </el-button-group>
        </div>
      </div>

      <!-- 表格视图 -->
      <el-table
        v-if="viewMode === 'table'"
        v-loading="loading"
        :data="tableData"
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column label="用户信息" min-width="200">
          <template #default="{ row }">
            <div class="user-info">
              <el-avatar :size="40" :src="row.avatar">
                {{ row.username.charAt(0).toUpperCase() }}
              </el-avatar>
              <div class="user-details">
                <div class="user-name">{{ row.username }}</div>
                <div class="user-email">{{ row.email }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="角色" width="120">
          <template #default="{ row }">
            <el-tag v-for="role in row.roles" :key="role" size="small">
              {{ getRoleName(role) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最后登录" width="150">
          <template #default="{ row }">
            {{ formatDate(row.lastLoginTime) }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="150">
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" text size="small" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button type="warning" text size="small" @click="handleResetPassword(row)">
              重置密码
            </el-button>
            <el-dropdown @command="(command) => handleMoreAction(command, row)">
              <el-button type="info" text size="small">
                更多<el-icon class="el-icon--right"><arrow-down /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="view">查看详情</el-dropdown-item>
                  <el-dropdown-item command="permissions">权限设置</el-dropdown-item>
                  <el-dropdown-item command="logs">操作日志</el-dropdown-item>
                  <el-dropdown-item
                    :command="row.status === 'active' ? 'disable' : 'enable'"
                    :divided="row.status === 'active'"
                  >
                    {{ row.status === 'active' ? '禁用账户' : '启用账户' }}
                  </el-dropdown-item>
                  <el-dropdown-item command="delete" divided class="danger-item">
                    删除用户
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>

      <!-- 卡片视图 -->
      <div v-else class="user-cards">
        <el-row :gutter="20">
          <el-col
            v-for="user in tableData"
            :key="user.id"
            :xs="24"
            :sm="12"
            :md="8"
            :lg="6"
          >
            <div class="user-card">
              <div class="user-card-header">
                <el-avatar :size="60" :src="user.avatar">
                  {{ user.username.charAt(0).toUpperCase() }}
                </el-avatar>
                <el-tag :type="getStatusType(user.status)" size="small">
                  {{ getStatusText(user.status) }}
                </el-tag>
              </div>
              <div class="user-card-body">
                <h4 class="user-card-name">{{ user.username }}</h4>
                <p class="user-card-email">{{ user.email }}</p>
                <div class="user-card-roles">
                  <el-tag
                    v-for="role in user.roles"
                    :key="role"
                    size="small"
                    class="role-tag"
                  >
                    {{ getRoleName(role) }}
                  </el-tag>
                </div>
                <div class="user-card-meta">
                  <div class="meta-item">
                    <span class="meta-label">最后登录</span>
                    <span class="meta-value">{{ formatDate(user.lastLoginTime) }}</span>
                  </div>
                  <div class="meta-item">
                    <span class="meta-label">创建时间</span>
                    <span class="meta-value">{{ formatDate(user.createdAt) }}</span>
                  </div>
                </div>
              </div>
              <div class="user-card-footer">
                <el-button type="primary" size="small" @click="handleEdit(user)">
                  编辑
                </el-button>
                <el-button size="small" @click="handleResetPassword(user)">
                  重置密码
                </el-button>
                <el-dropdown @command="(command) => handleMoreAction(command, user)">
                  <el-button size="small">
                    更多<el-icon class="el-icon--right"><arrow-down /></el-icon>
                  </el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="view">查看详情</el-dropdown-item>
                      <el-dropdown-item command="permissions">权限设置</el-dropdown-item>
                      <el-dropdown-item command="logs">操作日志</el-dropdown-item>
                      <el-dropdown-item
                        :command="user.status === 'active' ? 'disable' : 'enable'"
                        :divided="user.status === 'active'"
                      >
                        {{ user.status === 'active' ? '禁用账户' : '启用账户' }}
                      </el-dropdown-item>
                      <el-dropdown-item command="delete" divided class="danger-item">
                        删除用户
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 用户表单对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'create' ? '创建用户' : '编辑用户'"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="用户名" prop="username">
              <el-input
                v-model="formData.username"
                placeholder="请输入用户名"
                :disabled="dialogType === 'edit'"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="邮箱" prop="email">
              <el-input v-model="formData.email" placeholder="请输入邮箱" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="手机号" prop="phone">
              <el-input v-model="formData.phone" placeholder="请输入手机号" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="真实姓名" prop="realName">
              <el-input v-model="formData.realName" placeholder="请输入真实姓名" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="角色" prop="roleIds">
              <el-select
                v-model="formData.roleIds"
                placeholder="请选择角色"
                multiple
                style="width: 100%"
              >
                <el-option
                  v-for="role in roleOptions"
                  :key="role.id"
                  :label="role.name"
                  :value="role.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态" prop="status">
              <el-radio-group v-model="formData.status">
                <el-radio label="active">正常</el-radio>
                <el-radio label="disabled">禁用</el-radio>
                <el-radio label="pending">待激活</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="备注" prop="remark">
          <el-input
            v-model="formData.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入备注信息"
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import dayjs from 'dayjs'

// 接口定义
interface User {
  id: number
  username: string
  email: string
  phone: string
  realName: string
  avatar?: string
  roles: string[]
  status: 'active' | 'disabled' | 'pending'
  lastLoginTime: string
  createdAt: string
  remark?: string
}

interface Role {
  id: string
  name: string
  description?: string
}

// 响应式数据
const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const dialogType = ref<'create' | 'edit'>('create')
const viewMode = ref<'table' | 'card'>('table')
const selectedUsers = ref<User[]>([])

// 搜索表单
const searchForm = reactive({
  keyword: '',
  status: '',
  roleId: '',
  dateRange: [] as string[]
})

// 分页
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 用户表单
const formData = reactive({
  id: 0,
  username: '',
  email: '',
  phone: '',
  realName: '',
  roleIds: [] as string[],
  status: 'active' as 'active' | 'disabled' | 'pending',
  remark: ''
})

// 表单验证规则
const formRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  phone: [
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ],
  roleIds: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
}

// 角色选项
const roleOptions = ref<Role[]>([
  { id: '1', name: '超级管理员', description: '系统超级管理员' },
  { id: '2', name: '管理员', description: '系统管理员' },
  { id: '3', name: '运营', description: '运营人员' },
  { id: '4', name: '普通用户', description: '普通用户' }
])

// 用户数据
const tableData = ref<User[]>([
  {
    id: 1,
    username: 'admin',
    email: 'admin@example.com',
    phone: '13800138000',
    realName: '系统管理员',
    roles: ['1'],
    status: 'active',
    lastLoginTime: '2025-12-22 10:30:00',
    createdAt: '2025-01-01 09:00:00'
  },
  {
    id: 2,
    username: 'zhangsan',
    email: 'zhangsan@example.com',
    phone: '13800138001',
    realName: '张三',
    roles: ['2', '3'],
    status: 'active',
    lastLoginTime: '2025-12-22 09:15:00',
    createdAt: '2025-01-15 14:20:00'
  },
  {
    id: 3,
    username: 'lisi',
    email: 'lisi@example.com',
    phone: '13800138002',
    realName: '李四',
    roles: ['4'],
    status: 'disabled',
    lastLoginTime: '2025-12-20 16:45:00',
    createdAt: '2025-02-01 11:30:00'
  }
])

const formRef = ref<FormInstance>()

// 计算属性
const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    active: 'success',
    disabled: 'danger',
    pending: 'warning'
  }
  return types[status] || 'info'
}

const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    active: '正常',
    disabled: '禁用',
    pending: '待激活'
  }
  return texts[status] || '未知'
}

const getRoleName = (roleId: string) => {
  const role = roleOptions.value.find(r => r.id === roleId)
  return role?.name || roleId
}

const formatDate = (date: string) => {
  return dayjs(date).format('YYYY-MM-DD HH:mm')
}

// 事件处理
const handleSearch = () => {
  pagination.page = 1
  fetchUsers()
}

const handleReset = () => {
  Object.assign(searchForm, {
    keyword: '',
    status: '',
    roleId: '',
    dateRange: []
  })
  handleSearch()
}

const handleCreate = () => {
  dialogType.value = 'create'
  resetFormData()
  dialogVisible.value = true
}

const handleEdit = (user: User) => {
  dialogType.value = 'edit'
  Object.assign(formData, {
    id: user.id,
    username: user.username,
    email: user.email,
    phone: user.phone,
    realName: user.realName,
    roleIds: user.roles,
    status: user.status,
    remark: user.remark || ''
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate()
  submitLoading.value = true

  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))

    if (dialogType.value === 'create') {
      // 创建用户
      const newUser: User = {
        id: Date.now(),
        username: formData.username,
        email: formData.email,
        phone: formData.phone,
        realName: formData.realName,
        roles: formData.roleIds,
        status: formData.status,
        lastLoginTime: '',
        createdAt: dayjs().format('YYYY-MM-DD HH:mm:ss'),
        remark: formData.remark
      }
      tableData.value.unshift(newUser)
      ElMessage.success('用户创建成功')
    } else {
      // 编辑用户
      const index = tableData.value.findIndex(u => u.id === formData.id)
      if (index !== -1) {
        Object.assign(tableData.value[index], {
          email: formData.email,
          phone: formData.phone,
          realName: formData.realName,
          roles: formData.roleIds,
          status: formData.status,
          remark: formData.remark
        })
        ElMessage.success('用户更新成功')
      }
    }

    dialogVisible.value = false
    fetchUsers()
  } catch (error) {
    ElMessage.error('操作失败')
  } finally {
    submitLoading.value = false
  }
}

const handleResetPassword = async (user: User) => {
  try {
    await ElMessageBox.confirm(
      `确定要重置用户 "${user.username}" 的密码吗？新密码将发送到用户邮箱。`,
      '重置密码',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))
    ElMessage.success('密码重置成功，新密码已发送到用户邮箱')
  } catch {
    // 用户取消
  }
}

const handleMoreAction = async (command: string, user: User) => {
  switch (command) {
    case 'view':
      ElMessage.info('查看详情功能开发中...')
      break
    case 'permissions':
      ElMessage.info('权限设置功能开发中...')
      break
    case 'logs':
      ElMessage.info('操作日志功能开发中...')
      break
    case 'enable':
    case 'disable':
      try {
        await ElMessageBox.confirm(
          `确定要${command === 'enable' ? '启用' : '禁用'}用户 "${user.username}" 吗？`,
          `${command === 'enable' ? '启用' : '禁用'}用户`,
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )

        user.status = command === 'enable' ? 'active' : 'disabled'
        ElMessage.success(`用户${command === 'enable' ? '启用' : '禁用'}成功`)
      } catch {
        // 用户取消
      }
      break
    case 'delete':
      try {
        await ElMessageBox.confirm(
          `确定要删除用户 "${user.username}" 吗？此操作不可恢复！`,
          '删除用户',
          {
            confirmButtonText: '确定删除',
            cancelButtonText: '取消',
            type: 'error'
          }
        )

        const index = tableData.value.findIndex(u => u.id === user.id)
        if (index !== -1) {
          tableData.value.splice(index, 1)
          ElMessage.success('用户删除成功')
          fetchUsers()
        }
      } catch {
        // 用户取消
      }
      break
  }
}

const handleImport = () => {
  ElMessage.info('批量导入功能开发中...')
}

const handleExport = () => {
  ElMessage.info('导出数据功能开发中...')
}

const handleSelectionChange = (selection: User[]) => {
  selectedUsers.value = selection
}

const handleSizeChange = (size: number) => {
  pagination.size = size
  fetchUsers()
}

const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchUsers()
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}

const resetFormData = () => {
  Object.assign(formData, {
    id: 0,
    username: '',
    email: '',
    phone: '',
    realName: '',
    roleIds: [],
    status: 'active',
    remark: ''
  })
}

const fetchUsers = async () => {
  loading.value = true
  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 500))
    pagination.total = tableData.value.length
  } catch (error) {
    ElMessage.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

// 生命周期
onMounted(() => {
  fetchUsers()
})
</script>

<style scoped lang="scss">
.users-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 24px;

    .header-content {
      .page-title {
        font-size: 28px;
        font-weight: 700;
        color: var(--text-primary);
        margin: 0 0 8px 0;
      }

      .page-subtitle {
        color: var(--text-secondary);
        margin: 0;
      }
    }

    .header-actions {
      display: flex;
      gap: 12px;
    }
  }

  .search-card {
    margin-bottom: 24px;

    .search-form {
      .el-form-item {
        margin-bottom: 0;
      }
    }
  }

  .table-card {
    .table-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;

      .table-title {
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 16px;
        font-weight: 600;
        color: var(--text-primary);
      }
    }

    .user-info {
      display: flex;
      align-items: center;
      gap: 12px;

      .user-details {
        .user-name {
          font-weight: 600;
          color: var(--text-primary);
          margin-bottom: 4px;
        }

        .user-email {
          font-size: 13px;
          color: var(--text-secondary);
        }
      }
    }

    .user-cards {
      margin-top: 16px;

      .user-card {
        background: var(--bg-white);
        border: 1px solid var(--border-lighter);
        border-radius: 12px;
        padding: 20px;
        transition: all 0.3s ease;
        height: 100%;

        &:hover {
          box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
          transform: translateY(-2px);
        }

        .user-card-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 16px;
        }

        .user-card-body {
          .user-card-name {
            font-size: 18px;
            font-weight: 600;
            color: var(--text-primary);
            margin: 0 0 8px 0;
          }

          .user-card-email {
            color: var(--text-secondary);
            margin: 0 0 12px 0;
            font-size: 14px;
          }

          .user-card-roles {
            margin-bottom: 16px;

            .role-tag {
              margin-right: 4px;
              margin-bottom: 4px;
            }
          }

          .user-card-meta {
            .meta-item {
              display: flex;
              justify-content: space-between;
              margin-bottom: 4px;
              font-size: 13px;

              .meta-label {
                color: var(--text-secondary);
              }

              .meta-value {
                color: var(--text-primary);
              }
            }
          }
        }

        .user-card-footer {
          display: flex;
          gap: 8px;
          margin-top: 16px;
          padding-top: 16px;
          border-top: 1px solid var(--border-lighter);
        }
      }
    }

    .pagination-wrapper {
      display: flex;
      justify-content: center;
      margin-top: 24px;
    }
  }
}

// 危险操作样式
:deep(.danger-item) {
  color: var(--danger-color);
}
</style>