<template>
  <div class="tenant-packages-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>套餐管理</span>
          <el-button type="primary" @click="handleCreate">
            <el-icon><Plus /></el-icon>
            新建套餐
          </el-button>
        </div>
      </template>

      <!-- 搜索筛选 -->
      <div class="search-bar">
        <el-input
          v-model="searchForm.name"
          placeholder="搜索套餐名称"
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
          <el-option label="上架" :value="1" />
          <el-option label="下架" :value="2" />
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
        <el-table-column label="套餐名称" prop="name" width="200" />
        <el-table-column label="套餐代码" prop="code" width="150" />
        <el-table-column label="价格（元/月）" prop="price" width="120">
          <template #default="{ row }">
            ¥{{ row.price }}
          </template>
        </el-table-column>
        <el-table-column label="用户数限制" prop="max_users" width="120">
          <template #default="{ row }">
            {{ row.max_users === -1 ? '不限' : row.max_users }}
          </template>
        </el-table-column>
        <el-table-column label="存储空间(GB)" prop="max_storage" width="120">
          <template #default="{ row }">
            {{ row.max_storage === -1 ? '不限' : row.max_storage }}
          </template>
        </el-table-column>
        <el-table-column label="功能权限" prop="features" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag v-for="feature in row.features" :key="feature" size="small" style="margin-right: 4px;">
              {{ feature }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="Number(row.status) === 1 ? 'success' : 'info'">
              {{ Number(row.status) === 1 ? '上架' : '下架' }}
            </el-tag>
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
                :type="Number(row.status) === 1 ? 'warning' : 'success'"
                plain
                @click="handleToggleStatus(row)"
              >
                <el-icon><component :is="Number(row.status) === 1 ? 'Lock' : 'Unlock'" /></el-icon>
                {{ Number(row.status) === 1 ? '下架' : '上架' }}
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

    <!-- 套餐表单对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'create' ? '创建套餐' : '编辑套餐'"
      width="700px"
      :close-on-click-modal="false"
      @close="handleDialogClose"
    >
      <el-form ref="formRef" :model="formData" :rules="formRules" label-width="120px">
        <el-form-item label="套餐名称" prop="name">
          <el-input
            v-model="formData.name"
            placeholder="请输入套餐名称"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="套餐代码" prop="code">
          <el-input
            v-model="formData.code"
            placeholder="请输入套餐代码"
            maxlength="50"
            show-word-limit
            :disabled="dialogType === 'edit'"
          />
        </el-form-item>
        <el-form-item label="价格" prop="price">
          <el-input-number v-model="formData.price" :min="0" :precision="2" :step="100" />
          <span style="margin-left: 8px; color: var(--text-secondary);">元/月</span>
        </el-form-item>
        <el-form-item label="用户数限制" prop="max_users">
          <el-input-number v-model="formData.max_users" :min="-1" :step="10" />
          <span style="margin-left: 8px; color: var(--text-secondary);">-1 表示不限</span>
        </el-form-item>
        <el-form-item label="存储空间(GB)" prop="max_storage">
          <el-input-number v-model="formData.max_storage" :min="-1" :step="10" />
          <span style="margin-left: 8px; color: var(--text-secondary);">-1 表示不限</span>
        </el-form-item>
        <el-form-item label="功能权限" prop="features">
          <el-checkbox-group v-model="formData.features">
            <el-checkbox label="数据报表" />
            <el-checkbox label="API访问" />
            <el-checkbox label="自定义域名" />
            <el-checkbox label="SSO单点登录" />
            <el-checkbox label="高级分析" />
            <el-checkbox label="优先支持" />
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入套餐描述"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item v-if="dialogType === 'edit'" label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :label="1">上架</el-radio>
            <el-radio :label="2">下架</el-radio>
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

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const dialogType = ref<'create' | 'edit'>('create')

const formRef = ref<FormInstance>()

// 搜索表单
const searchForm = reactive({
  name: '',
  status: undefined as number | undefined
})

// 分页
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 套餐数据
interface PackageInfo {
  id: string
  name: string
  code: string
  price: number
  max_users: number
  max_storage: number
  features: string[]
  description: string
  status: number
  created_at: string
}

const tableData = ref<PackageInfo[]>([
  {
    id: '1',
    name: '基础版',
    code: 'basic',
    price: 99,
    max_users: 10,
    max_storage: 100,
    features: ['数据报表', 'API访问'],
    description: '适合小型团队使用',
    status: 1,
    created_at: new Date().toISOString()
  },
  {
    id: '2',
    name: '专业版',
    code: 'professional',
    price: 299,
    max_users: 50,
    max_storage: 500,
    features: ['数据报表', 'API访问', '自定义域名', 'SSO单点登录'],
    description: '适合中型企业使用',
    status: 1,
    created_at: new Date().toISOString()
  },
  {
    id: '3',
    name: '企业版',
    code: 'enterprise',
    price: 999,
    max_users: -1,
    max_storage: -1,
    features: ['数据报表', 'API访问', '自定义域名', 'SSO单点登录', '高级分析', '优先支持'],
    description: '适合大型企业使用',
    status: 1,
    created_at: new Date().toISOString()
  }
])

// 套餐表单
const formData = reactive({
  id: '',
  name: '',
  code: '',
  price: 0,
  max_users: 10,
  max_storage: 100,
  features: [] as string[],
  description: '',
  status: 1
})

// 表单验证规则
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入套餐名称', trigger: 'blur' },
    { min: 2, max: 200, message: '套餐名称长度在 2 到 200 个字符', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入套餐代码', trigger: 'blur' },
    { min: 2, max: 50, message: '套餐代码长度在 2 到 50 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_-]+$/, message: '套餐代码只能包含字母、数字、下划线和连字符', trigger: 'blur' }
  ],
  price: [
    { required: true, message: '请输入价格', trigger: 'blur' }
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
    status: undefined
  })
  handleSearch()
}

const handleCreate = () => {
  dialogType.value = 'create'
  resetFormData()
  dialogVisible.value = true
}

const handleEdit = (pkg: PackageInfo) => {
  dialogType.value = 'edit'
  Object.assign(formData, {
    id: pkg.id,
    name: pkg.name,
    code: pkg.code,
    price: pkg.price,
    max_users: pkg.max_users,
    max_storage: pkg.max_storage,
    features: [...pkg.features],
    description: pkg.description,
    status: pkg.status
  })
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
    // TODO: 调用 API 保存数据
    if (dialogType.value === 'create') {
      ElMessage.success('套餐创建成功')
    } else {
      ElMessage.success('套餐更新成功')
    }

    dialogVisible.value = false
    loadData()
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

const handleToggleStatus = async (pkg: PackageInfo) => {
  const newStatus = pkg.status === 1 ? 2 : 1
  const action = newStatus === 1 ? '上架' : '下架'

  try {
    await ElMessageBox.confirm(
      `确定要${action}套餐 "${pkg.name}" 吗？`,
      `${action}套餐`,
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // TODO: 调用 API 更新状态
    pkg.status = newStatus
    ElMessage.success(`套餐${action}成功`)
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || `${action}失败`)
    }
  }
}

const handleDelete = async (pkg: PackageInfo) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除套餐 "${pkg.name}" 吗？此操作不可恢复！`,
      '删除套餐',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'error'
      }
    )

    // TODO: 调用 API 删除
    ElMessage.success('套餐删除成功')
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
    id: '',
    name: '',
    code: '',
    price: 0,
    max_users: 10,
    max_storage: 100,
    features: [],
    description: '',
    status: 1
  })
}

const loadData = async () => {
  loading.value = true
  try {
    // TODO: 调用 API 获取数据
    // 模拟分页
    pagination.total = tableData.value.length
  } catch (error: any) {
    ElMessage.error(error.message || '获取套餐列表失败')
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
.tenant-packages-page {
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
