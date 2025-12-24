<template>
  <div class="tenant-members-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>租户成员管理</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            添加成员
          </el-button>
        </div>
      </template>

      <!-- 搜索栏 -->
      <div class="search-bar">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索用户名或姓名"
          clearable
          style="width: 300px;"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-select
          v-model="statusFilter"
          placeholder="状态筛选"
          clearable
          style="width: 150px;"
          @change="handleSearch"
        >
          <el-option label="全部" :value="0" />
          <el-option label="正常" :value="1" />
          <el-option label="禁用" :value="2" />
        </el-select>
        <el-button type="primary" @click="handleSearch">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </div>

      <!-- 成员列表表格 -->
      <el-table :data="memberList" v-loading="loading" style="width: 100%;">
        <el-table-column prop="name" label="姓名" width="120" />
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column prop="email" label="邮箱" width="200" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="登录状态" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.first_login" type="warning" size="small">首次登录</el-tag>
            <span v-else-if="row.last_login_time === 0" class="text-muted">未登录</span>
            <span v-else class="text-muted">{{ formatTime(row.last_login_time) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="加入时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" plain @click="handleEditRoles(row)">
                <el-icon><Setting /></el-icon>
                角色设置
              </el-button>
              <el-button size="small" type="danger" plain @click="handleDelete(row)">
                <el-icon><Delete /></el-icon>
                移除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadData"
          @current-change="loadData"
        />
      </div>
    </el-card>

    <!-- 添加成员对话框 -->
    <el-dialog
      v-model="addDialogVisible"
      title="添加租户成员"
      width="550px"
      :close-on-click-modal="false"
    >
      <el-form :model="addForm" :rules="addRules" ref="addFormRef" label-width="100px">
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="addForm.username"
            placeholder="请输入用户名（全局唯一）"
            maxlength="50"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="姓名" prop="name">
          <el-input
            v-model="addForm.name"
            placeholder="请输入姓名"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input
            v-model="addForm.phone"
            placeholder="选填"
            maxlength="11"
          />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input
            v-model="addForm.email"
            placeholder="选填"
            maxlength="100"
          />
        </el-form-item>
        <el-form-item label="分配角色" prop="role_ids">
          <el-select
            v-model="addForm.role_ids"
            multiple
            placeholder="请选择角色（至少一个）"
            style="width: 100%;"
            :loading="rolesLoading"
          >
            <el-option
              v-for="role in availableRoles"
              :key="role.role_id"
              :label="role.name"
              :value="role.role_id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleAddSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 初始密码展示对话框 -->
    <el-dialog
      v-model="passwordDialogVisible"
      title="成员添加成功！"
      width="500px"
      :close-on-click-modal="false"
      :show-close="false"
    >
      <div class="password-dialog-content">
        <el-result icon="success" title="成员添加成功" sub-title="请将以下凭据安全地传递给用户" />
        <div class="credential-box">
          <div class="credential-item">
            <span class="label">用户名：</span>
            <span class="value">{{ createdMemberInfo.username }}</span>
          </div>
          <div class="credential-item">
            <span class="label">初始密码：</span>
            <div class="password-display">
              <el-input
                :model-value="createdMemberInfo.initial_password"
                readonly
                size="large"
                style="font-size: 20px; font-weight: bold; letter-spacing: 2px;"
              />
              <el-button type="primary" @click="copyPassword">
                <el-icon><DocumentCopy /></el-icon>
                复制
              </el-button>
            </div>
          </div>
        </div>
        <el-alert
          type="warning"
          :closable="false"
          show-icon
        >
          此密码只显示一次，请立即记录并通过安全方式传递给用户！
        </el-alert>
      </div>
      <template #footer>
        <el-button type="primary" size="large" @click="handlePasswordConfirm">
          我已记下密码
        </el-button>
      </template>
    </el-dialog>

    <!-- 编辑角色对话框 -->
    <el-dialog
      v-model="editRolesDialogVisible"
      title="设置成员角色"
      width="450px"
      :close-on-click-modal="false"
    >
      <div class="member-info">
        <el-icon><User /></el-icon>
        <span>{{ currentMember.name }}（{{ currentMember.username }}）</span>
      </div>
      <el-form :model="editRolesForm" :rules="editRolesRules" ref="editRolesFormRef" label-width="80px">
        <el-form-item label="角色" prop="role_ids">
          <el-select
            v-model="editRolesForm.role_ids"
            multiple
            placeholder="请选择角色（至少一个）"
            style="width: 100%;"
            :loading="rolesLoading"
          >
            <el-option
              v-for="role in availableRoles"
              :key="role.role_id"
              :label="role.name"
              :value="role.role_id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editRolesDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleEditRolesSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Delete, Setting, DocumentCopy, User } from '@element-plus/icons-vue'
import { tenantMemberApi, type TenantMemberInfo, type AddTenantMemberRequest } from '../../api/tenantMember'
import { roleApi, type RoleInfo } from '../../api/role'

// 状态变量
const loading = ref(false)
const rolesLoading = ref(false)
const submitLoading = ref(false)
const searchKeyword = ref('')
const statusFilter = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const memberList = ref<TenantMemberInfo[]>([])

// 对话框状态
const addDialogVisible = ref(false)
const passwordDialogVisible = ref(false)
const editRolesDialogVisible = ref(false)

// 表单引用
const addFormRef = ref()
const editRolesFormRef = ref()

// 可用角色列表
const availableRoles = ref<RoleInfo[]>([])

// 添加成员表单
const addForm = reactive<AddTenantMemberRequest>({
  username: '',
  name: '',
  phone: '',
  email: '',
  role_ids: []
})

// 添加表单验证规则
const addRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 2, max: 50, message: '用户名长度为 2-50 个字符', trigger: 'blur' }
  ],
  name: [
    { required: true, message: '请输入姓名', trigger: 'blur' },
    { min: 2, max: 200, message: '姓名长度为 2-200 个字符', trigger: 'blur' }
  ],
  phone: [
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ],
  email: [
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ],
  role_ids: [
    { required: true, message: '请至少选择一个角色', trigger: 'change' }
  ]
}

// 创建成功的成员信息
const createdMemberInfo = reactive({
  username: '',
  initial_password: ''
})

// 当前编辑的成员
const currentMember = reactive<Partial<TenantMemberInfo>>({
  user_id: '',
  name: '',
  username: ''
})

// 编辑角色表单
const editRolesForm = reactive({
  role_ids: [] as string[]
})

const editRolesRules = {
  role_ids: [
    { required: true, message: '请至少选择一个角色', trigger: 'change' }
  ]
}

// 初始化
onMounted(() => {
  loadData()
  loadRoles()
})

// 加载成员列表
async function loadData() {
  loading.value = true
  try {
    const response = await tenantMemberApi.getList({
      page: currentPage.value,
      page_size: pageSize.value,
      keyword: searchKeyword.value || undefined,
      status: statusFilter.value || undefined
    })
    memberList.value = response.list || []
    total.value = response.total
  } catch (error: any) {
    ElMessage.error(error.message || '加载成员列表失败')
  } finally {
    loading.value = false
  }
}

// 加载角色列表
async function loadRoles() {
  rolesLoading.value = true
  try {
    const response = await roleApi.getAllRoles()
    availableRoles.value = response.list || []
  } catch (error: any) {
    ElMessage.error(error.message || '加载角色列表失败')
  } finally {
    rolesLoading.value = false
  }
}

// 搜索
function handleSearch() {
  currentPage.value = 1
  loadData()
}

// 添加成员
function handleAdd() {
  resetAddForm()
  addDialogVisible.value = true
}

// 提交添加
async function handleAddSubmit() {
  const valid = await addFormRef.value?.validate().catch(() => false)
  if (!valid) return

  submitLoading.value = true
  try {
    const response = await tenantMemberApi.add(addForm)
    // 保存创建成功的信息
    createdMemberInfo.username = response.username
    createdMemberInfo.initial_password = response.initial_password

    // 关闭添加对话框，打开密码展示对话框
    addDialogVisible.value = false
    passwordDialogVisible.value = true

    // 刷新列表
    loadData()
  } catch (error: any) {
    ElMessage.error(error.message || '添加成员失败')
  } finally {
    submitLoading.value = false
  }
}

// 复制密码
function copyPassword() {
  navigator.clipboard.writeText(createdMemberInfo.initial_password).then(() => {
    ElMessage.success('密码已复制到剪贴板')
  }).catch(() => {
    ElMessage.error('复制失败，请手动复制')
  })
}

// 确认已记录密码
function handlePasswordConfirm() {
  passwordDialogVisible.value = false
  resetAddForm()
}

// 编辑角色
function handleEditRoles(row: TenantMemberInfo) {
  currentMember.user_id = row.user_id
  currentMember.name = row.name
  currentMember.username = row.username
  editRolesForm.role_ids = [...row.role_ids]
  editRolesDialogVisible.value = true
}

// 提交角色编辑
async function handleEditRolesSubmit() {
  const valid = await editRolesFormRef.value?.validate().catch(() => false)
  if (!valid) return

  if (!currentMember.user_id) {
    ElMessage.error('成员信息异常')
    return
  }

  submitLoading.value = true
  try {
    await tenantMemberApi.updateRoles({
      user_id: currentMember.user_id,
      role_ids: editRolesForm.role_ids
    })
    ElMessage.success('角色设置成功')
    editRolesDialogVisible.value = false
    loadData()
  } catch (error: any) {
    ElMessage.error(error.message || '角色设置失败')
  } finally {
    submitLoading.value = false
  }
}

// 删除成员
function handleDelete(row: TenantMemberInfo) {
  ElMessageBox.confirm(
    `确定要移除成员"${row.name}"（${row.username}）吗？移除后该成员将无法访问当前租户。`,
    '移除成员',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      await tenantMemberApi.removeByPath(row.user_id)
      ElMessage.success('移除成功')
      loadData()
    } catch (error: any) {
      ElMessage.error(error.message || '移除失败')
    }
  }).catch(() => {})
}

// 重置添加表单
function resetAddForm() {
  Object.assign(addForm, {
    username: '',
    name: '',
    phone: '',
    email: '',
    role_ids: []
  })
  addFormRef.value?.resetFields()
}

// 格式化时间
function formatTime(timestamp: number) {
  if (!timestamp) return '-'
  const date = new Date(timestamp)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<style scoped lang="scss">
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

.text-muted {
  color: #909399;
  font-size: 12px;
}

// 密码对话框样式
.password-dialog-content {
  .credential-box {
    margin: 20px 0;
    padding: 20px;
    background: #f5f7fa;
    border-radius: 8px;

    .credential-item {
      margin-bottom: 16px;

      &:last-child {
        margin-bottom: 0;
      }

      .label {
        font-weight: 500;
        color: #606266;
        margin-right: 8px;
      }

      .value {
        font-weight: bold;
        color: #303133;
      }

      .password-display {
        display: flex;
        gap: 10px;
        align-items: center;

        .el-input {
          flex: 1;
        }
      }
    }
  }

  :deep(.el-result) {
    padding: 20px 0;
  }
}

.member-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 0;
  margin-bottom: 16px;
  font-size: 15px;
  font-weight: 500;
  color: #303133;
  border-bottom: 1px solid #ebeef5;

  .el-icon {
    font-size: 18px;
    color: #409eff;
  }
}
</style>
