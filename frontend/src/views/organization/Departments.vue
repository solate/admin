<template>
  <div class="departments-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>部门管理</span>
          <el-button type="primary" @click="handleCreate">
            <el-icon><Plus /></el-icon>
            新建部门
          </el-button>
        </div>
      </template>

      <div class="search-bar">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索部门名称"
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
          style="width: 120px;"
          @change="handleSearch"
        >
          <el-option label="全部" :value="0" />
          <el-option label="启用" :value="1" />
          <el-option label="禁用" :value="2" />
        </el-select>
        <el-button type="primary" @click="handleSearch">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
        <el-button @click="loadTree">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>

      <el-table
        :data="departmentTree"
        v-loading="loading"
        row-key="department_id"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        :default-expand-all="false"
        style="width: 100%;"
      >
        <el-table-column prop="department_name" label="部门名称" min-width="250" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="StatusUtils.getTagType(row.status)">
              {{ Number(row.status) === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" plain @click="handleCreateChild(row)">
                <el-icon><Plus /></el-icon>
                添加子部门
              </el-button>
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
                {{ Number(row.status) === 1 ? '禁用' : '启用' }}
              </el-button>
              <el-button size="small" type="danger" plain @click="handleDelete(row)">
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 部门编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
      :close-on-click-modal="false"
      @close="handleDialogClose"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="上级部门" prop="parent_id" v-if="!isAddingChild">
          <el-tree-select
            v-model="form.parent_id"
            :data="parentDepartmentOptions"
            :props="{ label: 'department_name', value: 'department_id' }"
            placeholder="选择上级部门（不选则为根部门）"
            clearable
            check-strictly
            :render-after-expand="false"
          />
        </el-form-item>
        <el-form-item label="上级部门" v-else>
          <el-tag>{{ parentDepartmentName }}</el-tag>
        </el-form-item>
        <el-form-item label="部门名称" prop="department_name">
          <el-input v-model="form.department_name" placeholder="请输入部门名称" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入描述"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" :max="9999" controls-position="right" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :label="1">启用</el-radio>
            <el-radio :label="2">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Edit, Delete, Plus, Search, Refresh, Lock, Unlock } from '@element-plus/icons-vue'
import { StatusUtils } from '@/utils/status'
import {
  departmentApi,
  type DepartmentTreeNode,
  type CreateDepartmentRequest,
  type UpdateDepartmentRequest
} from '@/api/department'

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const isAddingChild = ref(false)
const formRef = ref<FormInstance>()
const searchKeyword = ref('')
const statusFilter = ref<number>(0)
const departmentTree = ref<DepartmentTreeNode[]>([])
const allDepartmentTree = ref<DepartmentTreeNode[]>([])
const parentDepartmentId = ref('')
const parentDepartmentName = ref('')

const dialogTitle = computed(() => {
  if (isAddingChild.value) return '添加子部门'
  return isEdit.value ? '编辑部门' : '新建部门'
})

const parentDepartmentOptions = computed(() => {
  // 过滤掉当前编辑的部门及其子部门，防止设置自己为父部门
  const filterTree = (tree: DepartmentTreeNode[], excludeId?: string): DepartmentTreeNode[] => {
    return tree
      .filter(node => node.department_id !== excludeId)
      .map(node => ({
        ...node,
        children: node.children ? filterTree(node.children, excludeId) : undefined
      }))
  }

  if (isEdit.value && form.department_id) {
    return filterTree(allDepartmentTree.value, form.department_id)
  }
  return allDepartmentTree.value
})

const form = reactive<{
  department_id?: string
  parent_id?: string
  department_name: string
  description: string
  sort: number
  status: number
}>({
  parent_id: '',
  department_name: '',
  description: '',
  sort: 0,
  status: 1
})

const rules: FormRules = {
  department_name: [{ required: true, message: '请输入部门名称', trigger: 'blur' }]
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

// 根据状态过滤部门树
function filterTreeByStatus(tree: DepartmentTreeNode[]): DepartmentTreeNode[] {
  return tree
    .filter(node => {
      // 如果有状态筛选，应用筛选
      if (statusFilter.value !== 0) {
        return node.status === statusFilter.value
      }
      return true
    })
    .filter(node => {
      // 如果有关键词搜索，搜索部门名称
      if (searchKeyword.value) {
        return node.department_name.includes(searchKeyword.value)
      }
      return true
    })
    .map(node => {
      const children = node.children ? filterTreeByStatus(node.children) : undefined
      // 如果子部门被过滤完，不显示该节点（除非它本身匹配）
      if (children && children.length === 0) {
        return { ...node, children: undefined }
      }
      return { ...node, children }
    })
    .filter(node => {
      // 如果有关键词或状态筛选，确保至少有匹配的子节点或自己匹配
      if (searchKeyword.value || statusFilter.value !== 0) {
        const selfMatches =
          (!searchKeyword.value || node.department_name.includes(searchKeyword.value)) &&
          (statusFilter.value === 0 || node.status === statusFilter.value)
        const hasMatchingChildren = node.children && node.children.length > 0
        return selfMatches || hasMatchingChildren
      }
      return true
    })
}

onMounted(() => {
  loadTree()
})

async function loadTree() {
  loading.value = true
  try {
    const response = await departmentApi.getTree()
    allDepartmentTree.value = response.tree || []
    applyFilters()
  } catch (error: any) {
    ElMessage.error(error.message || '加载部门树失败')
  } finally {
    loading.value = false
  }
}

function applyFilters() {
  if (!searchKeyword.value && statusFilter.value === 0) {
    departmentTree.value = allDepartmentTree.value
  } else {
    departmentTree.value = filterTreeByStatus(allDepartmentTree.value)
  }
}

function handleSearch() {
  applyFilters()
}

function handleCreate() {
  isEdit.value = false
  isAddingChild.value = false
  resetForm()
  dialogVisible.value = true
}

function handleCreateChild(row: DepartmentTreeNode) {
  isEdit.value = false
  isAddingChild.value = true
  parentDepartmentId.value = row.department_id
  parentDepartmentName.value = row.department_name
  resetForm()
  form.parent_id = row.department_id
  dialogVisible.value = true
}

function handleEdit(row: DepartmentTreeNode) {
  isEdit.value = true
  isAddingChild.value = false
  Object.assign(form, {
    department_id: row.department_id,
    parent_id: row.parent_id || '',
    department_name: row.department_name,
    description: row.description || '',
    sort: row.sort,
    status: row.status
  })
  dialogVisible.value = true
}

async function handleToggleStatus(row: DepartmentTreeNode) {
  const newStatus = Number(row.status) === 1 ? 2 : 1
  const action = newStatus === 1 ? '启用' : '禁用'

  try {
    await ElMessageBox.confirm(
      `确定要${action}部门 "${row.department_name}" 吗？`,
      `${action}部门`,
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await departmentApi.updateStatus(row.department_id, newStatus)
    ElMessage.success(`部门${action}成功`)
    loadTree()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || `${action}失败`)
    }
  }
}

async function handleDelete(row: DepartmentTreeNode) {
  // 检查是否有子部门
  if (row.children && row.children.length > 0) {
    ElMessage.warning('该部门下有子部门，无法删除')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除部门 "${row.department_name}" 吗？`,
      '删除部门',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await departmentApi.delete(row.department_id)
    ElMessage.success('删除成功')
    loadTree()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

async function handleSubmit() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
  } catch {
    return
  }

  submitLoading.value = true

  try {
    if (isEdit.value) {
      const updateData: UpdateDepartmentRequest = {
        parent_id: form.parent_id || undefined,
        department_name: form.department_name,
        description: form.description,
        sort: form.sort,
        status: form.status
      }
      await departmentApi.update(form.department_id!, updateData)
      ElMessage.success('更新成功')
    } else {
      const createData: CreateDepartmentRequest = {
        parent_id: isAddingChild.value ? parentDepartmentId.value : form.parent_id,
        department_name: form.department_name,
        description: form.description,
        sort: form.sort,
        status: form.status
      }
      await departmentApi.create(createData)
      ElMessage.success('创建成功')
    }

    dialogVisible.value = false
    loadTree()
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

function handleDialogClose() {
  formRef.value?.resetFields()
  resetForm()
}

function resetForm() {
  Object.assign(form, {
    department_id: undefined,
    parent_id: '',
    department_name: '',
    description: '',
    sort: 0,
    status: 1
  })
  parentDepartmentId.value = ''
  parentDepartmentName.value = ''
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

:deep(.el-table) {
  .el-table__cell {
    padding: 8px 0;
  }
}
</style>
