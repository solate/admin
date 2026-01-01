<template>
  <div class="tenants-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>租户管理</span>
          <el-button type="primary" @click="handleCreate">
            <el-icon><Plus /></el-icon>
            新建租户
          </el-button>
        </div>
      </template>

      <!-- 搜索筛选 -->
      <div class="search-bar">
        <el-input
          v-model="searchForm.name"
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
        <el-input
          v-model="searchForm.code"
          placeholder="搜索租户编码"
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
          <el-option label="正常" :value="1" />
          <el-option label="禁用" :value="2" />
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
        <el-table-column label="租户名称" prop="name" width="200" />
        <el-table-column label="租户编码" prop="code" width="200" />
        <el-table-column label="描述" prop="description" min-width="200" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="StatusUtils.getTagType(row.status)">
              {{ StatusUtils.getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" plain @click="handleEdit(row)">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button
                size="small"
                :type="StatusUtils.getButtonType(row.status, 'warning', 'success')"
                plain
                @click="handleToggleStatus(row)"
              >
                <el-icon><component :is="StatusUtils.isActive(row.status) ? 'Lock' : 'Unlock'" /></el-icon>
                {{ StatusUtils.getToggleActionText(row.status) }}
              </el-button>
              <el-button size="small" type="danger" plain @click="handleDelete(row)">
                <el-icon><Delete /></el-icon>
                删除
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

    <!-- 租户表单对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'create' ? '创建租户' : '编辑租户'"
      width="600px"
      :close-on-click-modal="false"
      @close="handleDialogClose"
    >
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="100px">
        <el-form-item label="租户名称" prop="name">
          <el-input
            v-model="formData.name"
            placeholder="请输入租户名称"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="租户编码" prop="code">
          <el-input
            v-model="formData.code"
            placeholder="请输入租户编码（全局唯一）"
            maxlength="50"
            show-word-limit
            :disabled="dialogType === 'edit'"
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入租户描述"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item v-if="dialogType === 'edit'" label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :label="1">正常</el-radio>
            <el-radio :label="2">禁用</el-radio>
          </el-radio-group>
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Plus, Search, Refresh, Edit, Delete } from '@element-plus/icons-vue'
import {
  tenantApi,
  type TenantInfo,
  type TenantListParams,
  type CreateTenantRequest,
  type UpdateTenantRequest
} from '../../api/tenant'
import { formatTime } from '../../utils/date'
import { StatusUtils } from '../../utils/status'

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const dialogType = ref<'create' | 'edit'>('create')

const formRef = ref<FormInstance>()

// 搜索表单
const searchForm = reactive({
  name: '',
  code: '',
  status: undefined as number | undefined
})

// 分页
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 租户数据
const tableData = ref<TenantInfo[]>([])

// 租户表单
const formData = reactive<CreateTenantRequest & { status?: number; tenant_id?: string }>({
  name: '',
  code: '',
  description: '',
  status: 1,
  tenant_id: undefined
})

// 表单验证规则
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入租户名称', trigger: 'blur' },
    { min: 2, max: 200, message: '租户名称长度在 2 到 200 个字符', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入租户编码', trigger: 'blur' },
    { min: 2, max: 50, message: '租户编码长度在 2 到 50 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_-]+$/, message: '租户编码只能包含字母、数字、下划线和连字符', trigger: 'blur' }
  ]
}

// 事件处理
const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  Object.assign(searchForm, {
    name: '',
    code: '',
    status: undefined
  })
  handleSearch()
}

const handleCreate = () => {
  dialogType.value = 'create'
  resetFormData()
  dialogVisible.value = true
}

const handleEdit = (tenant: TenantInfo) => {
  dialogType.value = 'edit'
  Object.assign(formData, {
    name: tenant.name,
    code: tenant.code,
    description: tenant.description || '',
    status: tenant.status
  })
  // 存储当前编辑的租户ID
  formData.tenant_id = tenant.tenant_id
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
  } catch {
    return
  }

  submitLoading.value = true

  try {
    if (dialogType.value === 'create') {
      await tenantApi.create({
        name: formData.name,
        code: formData.code,
        description: formData.description
      })
      ElMessage.success('租户创建成功')
    } else {
      const updateData: UpdateTenantRequest = {
        name: formData.name,
        description: formData.description,
        status: formData.status
      }
      await tenantApi.update(formData.tenant_id!, updateData)
      ElMessage.success('租户更新成功')
    }

    dialogVisible.value = false
    loadData()
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

const handleToggleStatus = async (tenant: TenantInfo) => {
  const newStatus = StatusUtils.toggleStatus(tenant.status)
  const action = StatusUtils.getToggleActionText(newStatus, '启用', '禁用')

  try {
    await ElMessageBox.confirm(
      `确定要${action}租户 "${tenant.name}" 吗？`,
      `${action}租户`,
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await tenantApi.updateStatus(tenant.tenant_id, newStatus)
    ElMessage.success(`租户${action}成功`)
    loadData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || `${action}失败`)
    }
  }
}

const handleDelete = async (tenant: TenantInfo) => {
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

    await tenantApi.delete(tenant.tenant_id)
    ElMessage.success('租户删除成功')
    loadData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
  resetFormData()
}

const resetFormData = () => {
  Object.assign(formData, {
    tenant_id: undefined,
    name: '',
    code: '',
    description: '',
    status: 1
  })
}

const loadData = async () => {
  loading.value = true
  try {
    const params: TenantListParams = {
      page: pagination.page,
      page_size: pagination.size,
      name: searchForm.name || undefined,
      code: searchForm.code || undefined,
      status: searchForm.status
    }
    const response = await tenantApi.getList(params)
    tableData.value = response.list || []
    pagination.total = response.total
  } catch (error: any) {
    ElMessage.error(error.message || '获取租户列表失败')
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
.tenants-page {
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
