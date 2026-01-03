<template>
  <div class="users-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span class="card-title">用户管理</span>
            <el-tag type="info" size="small">{{ pagination.total }} 条记录</el-tag>
          </div>
          <div class="header-actions">
            <el-button type="primary" @click="handleCreate">
              <el-icon><Plus /></el-icon>
              新建用户
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
      </template>

      <!-- 搜索筛选 -->
      <div class="search-bar">
        <el-form :model="searchForm" inline class="search-form">
          <el-form-item label="关键词">
            <el-input
              v-model="searchForm.keyword"
              placeholder="用户名/邮箱/手机号"
              clearable
              style="width: 200px;"
              @keyup.enter="handleSearch"
            />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="searchForm.status" placeholder="全部状态" clearable style="width: 120px;">
              <el-option label="正常" value="active" />
              <el-option label="禁用" value="disabled" />
            </el-select>
          </el-form-item>
          <el-form-item label="角色">
            <el-select v-model="searchForm.roleId" placeholder="全部角色" clearable style="width: 150px;">
              <el-option
                v-for="role in roleOptions"
                :key="role.role_id"
                :label="role.name"
                :value="role.role_id"
              />
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
        <div class="view-toggle">
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
        style="width: 100%;"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column label="用户信息" min-width="200">
          <template #default="{ row }">
            <div class="user-info">
              <el-avatar :size="40">
                <template v-if="row.avatar">
                  <img :src="row.avatar" />
                </template>
                <template v-else>
                  {{ row.user_name?.charAt(0)?.toUpperCase() || '?' }}
                </template>
              </el-avatar>
              <div class="user-details">
                <div class="user-name">{{ row.user_name || '-' }}</div>
                <div class="user-email">{{ row.email || '-' }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="真实姓名" width="120">
          <template #default="{ row }">
            {{ row.name }}
          </template>
        </el-table-column>
        <el-table-column label="角色" width="150">
          <template #default="{ row }">
            <el-tag v-for="role in row.role_list" :key="role.role_id" size="small">
              {{ role.name }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="StatusUtils.getTagType(row.status)" size="small">
              {{ StatusUtils.getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="150">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
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
                    :command="StatusUtils.isActive(row.status) ? 'disable' : 'enable'"
                    :divided="StatusUtils.isActive(row.status)"
                  >
                    {{ StatusUtils.getToggleActionText(row.status) + '账户' }}
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
                <el-avatar :size="60">
                  <template v-if="user.avatar">
                    <img :src="user.avatar" />
                  </template>
                  <template v-else>
                    {{ user.user_name?.charAt(0)?.toUpperCase() || '?' }}
                  </template>
                </el-avatar>
                <el-tag :type="StatusUtils.getTagType(user.status)" size="small">
                  {{ StatusUtils.getStatusText(user.status) }}
                </el-tag>
              </div>
              <div class="user-card-body">
                <h4 class="user-card-name">{{ user.user_name }}</h4>
                <p class="user-card-email">{{ user.email }}</p>
                <div class="user-card-roles">
                  <el-tag
                    v-for="role in user.role_list"
                    :key="role.role_id"
                    size="small"
                    class="role-tag"
                  >
                    {{ role.name }}
                  </el-tag>
                </div>
                <div class="user-card-meta">
                  <div class="meta-item">
                    <span class="meta-label">真实姓名</span>
                    <span class="meta-value">{{ user.name }}</span>
                  </div>
                  <div class="meta-item">
                    <span class="meta-label">创建时间</span>
                    <span class="meta-value">{{ formatDate(user.created_at) }}</span>
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
                        :command="StatusUtils.isActive(user.status) ? 'disable' : 'enable'"
                        :divided="StatusUtils.isActive(user.status)"
                      >
                        {{ StatusUtils.getToggleActionText(user.status) + '账户' }}
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
        :rules="dialogType === 'create' ? formRules : editFormRules"
        label-width="100px"
      >
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="用户名" prop="user_name">
              <el-input
                v-model="formData.user_name"
                placeholder="请输入用户名"
                :disabled="dialogType === 'edit'"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="真实姓名" prop="name">
              <el-input v-model="formData.name" placeholder="请输入真实姓名" />
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
            <el-form-item label="邮箱" prop="email">
              <el-input v-model="formData.email" placeholder="请输入邮箱" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="密码" prop="password">
              <el-input
                v-model="formData.password"
                type="password"
                placeholder="请输入密码"
                :disabled="dialogType === 'edit'"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="角色" prop="role_ids">
              <el-select
                v-model="formData.role_ids"
                placeholder="请选择角色"
                multiple
                style="width: 100%"
              >
                <el-option
                  v-for="role in roleOptions"
                  :key="role.role_id"
                  :label="role.name"
                  :value="role.role_id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="状态" prop="status">
              <el-radio-group v-model="formData.status">
                <el-radio :value="1">正常</el-radio>
                <el-radio :value="2">禁用</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import dayjs from 'dayjs'
import { userApi, type UserInfo, type UserListParams } from '@/api/user'
import { roleApi, type RoleInfo } from '@/api/role'
import { StatusUtils } from '@/utils/status'

// 接口定义
interface User {
  user_id: string
  user_name: string
  name: string
  phone: string
  email: string
  avatar?: string
  role_list: RoleListInfo[]
  status: number
  created_at: number
}

interface RoleListInfo {
  role_id: string
  name: string
  code: string
  sort: number
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
  user_id: '',
  user_name: '',
  email: '',
  phone: '',
  name: '',
  password: '',
  role_ids: [] as string[],
  status: 1,
  avatar: '',
  sex: 0
})

// 表单验证规则
const formRules = {
  user_name: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  name: [
    { required: true, message: '请输入真实姓名', trigger: 'blur' }
  ],
  email: [
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  phone: [
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' }
  ],
  role_ids: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
}

// 编辑时的表单验证规则（密码不需要）
const editFormRules = {
  name: [
    { required: true, message: '请输入真实姓名', trigger: 'blur' }
  ],
  email: [
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  phone: [
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ]
}

// 角色选项
const roleOptions = ref<RoleInfo[]>([])

// 用户数据
const tableData = ref<User[]>([])

const formRef = ref<FormInstance>()

// 工具函数
const formatDate = (timestamp: number) => {
  return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm')
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
    user_id: user.user_id,
    user_name: user.user_name,
    email: user.email || '',
    phone: user.phone,
    name: user.name,
    password: '',
    role_ids: user.role_list.map(r => r.role_id),
    status: user.status,
    avatar: user.avatar || '',
    sex: 0
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate()
  submitLoading.value = true

  try {
    if (dialogType.value === 'create') {
      // 创建用户
      await userApi.create({
        user_name: formData.user_name,
        name: formData.name,
        password: formData.password,
        phone: formData.phone,
        email: formData.email || undefined,
        status: formData.status,
        role_ids: formData.role_ids.length > 0 ? formData.role_ids : undefined
      })
      ElMessage.success('用户创建成功')
    } else {
      // 编辑用户
      await userApi.update(formData.user_id, {
        nickname: formData.name,
        email: formData.email || undefined,
        status: formData.status,
        phone: formData.phone || undefined
      })
      ElMessage.success('用户更新成功')
    }

    dialogVisible.value = false
    fetchUsers()
  } catch (error: any) {
    ElMessage.error(error?.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

const handleResetPassword = async (user: User) => {
  try {
    await ElMessageBox.confirm(
      `确定要重置用户 "${user.user_name}" 的密码吗？新密码将发送到用户邮箱。`,
      '重置密码',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // TODO: 调用重置密码API
    ElMessage.info('重置密码功能开发中...')
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
          `确定要${command === 'enable' ? '启用' : '禁用'}用户 "${user.user_name}" 吗？`,
          `${command === 'enable' ? '启用' : '禁用'}用户`,
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )

        const newStatus = StatusUtils.toggleStatus(user.status)
        await userApi.update(user.user_id, { status: newStatus })
        ElMessage.success(`用户${StatusUtils.getToggleActionText(user.status)}成功`)
        fetchUsers()
      } catch (error: any) {
        if (error !== 'cancel') {
          ElMessage.error(error?.message || '操作失败')
        }
      }
      break
    case 'delete':
      try {
        await ElMessageBox.confirm(
          `确定要删除用户 "${user.user_name}" 吗？此操作不可恢复！`,
          '删除用户',
          {
            confirmButtonText: '确定删除',
            cancelButtonText: '取消',
            type: 'error'
          }
        )

        await userApi.delete(user.user_id)
        ElMessage.success('用户删除成功')
        fetchUsers()
      } catch (error: any) {
        if (error !== 'cancel') {
          ElMessage.error(error?.message || '删除失败')
        }
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
    user_id: '',
    user_name: '',
    email: '',
    phone: '',
    name: '',
    password: '',
    role_ids: [],
    status: 1,
    avatar: '',
    sex: 0
  })
}

// 获取用户列表
const fetchUsers = async () => {
  loading.value = true
  try {
    const params: UserListParams = {
      page: pagination.page,
      page_size: pagination.size
    }

    if (searchForm.keyword) {
      params.name = searchForm.keyword
    }
    if (searchForm.phone) {
      params.phone = searchForm.phone
    }
    if (searchForm.status !== '') {
      params.status = searchForm.status === 'active' ? 1 : 0
    }

    const response = await userApi.getList(params)
    // 将后端返回的 { user: {...} } 结构转换为前端期望的格式
    tableData.value = response.list.map(item => ({
      user_id: item.user.user_id,
      user_name: item.user.username,
      name: item.user.nickname,
      phone: item.user.phone,
      email: item.user.email,
      avatar: item.user.avatar,
      status: item.user.status,
      created_at: item.user.created_at,
      role_list: [] // 后端暂未返回角色列表
    }))
    pagination.total = response.page.total
  } catch (error: any) {
    ElMessage.error(error?.message || '获取用户列表失败')
  } finally {
    loading.value = false
  }
}

// 获取角色列表
const fetchRoles = async () => {
  try {
    const response = await roleApi.getAllRoles()
    roleOptions.value = response.list
  } catch (error: any) {
    ElMessage.error(error?.message || '获取角色列表失败')
  }
}

// 生命周期
onMounted(() => {
  fetchUsers()
  fetchRoles()
})
</script>

<style scoped lang="scss">
.users-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .header-left {
      display: flex;
      align-items: center;
      gap: 12px;

      .card-title {
        font-size: 16px;
        font-weight: 600;
        color: var(--text-primary);
      }
    }

    .header-actions {
      display: flex;
      gap: 10px;
    }
  }

  .search-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 0;
    border-bottom: 1px solid var(--border-lighter);
    margin-bottom: 16px;

    .search-form {
      .el-form-item {
        margin-bottom: 0;
      }
    }

    .view-toggle {
      flex-shrink: 0;
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

// 危险操作样式
:deep(.danger-item) {
  color: var(--danger-color);
}
</style>