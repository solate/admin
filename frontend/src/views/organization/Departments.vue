<template>
  <div class="departments-page">
    <!-- 左侧部门树 -->
    <div class="dept-tree-panel">
      <el-card class="tree-card" shadow="never">
        <template #header>
          <div class="tree-header">
            <span class="tree-title">组织架构</span>
            <div class="tree-actions">
              <el-button type="primary" size="small" @click="handleCreate">
                <el-icon><Plus /></el-icon>
                新建部门
              </el-button>
            </div>
          </div>
        </template>

        <div class="tree-search">
          <el-input
            v-model="treeSearchKeyword"
            placeholder="搜索部门"
            clearable
            size="small"
            :prefix-icon="Search"
          />
        </div>

        <el-tree
          ref="treeRef"
          :data="filteredDepartmentTree"
          :props="{ label: 'department_name', children: 'children' }"
          node-key="department_id"
          :highlight-current="true"
          :default-expand-all="false"
          :expand-on-click-node="false"
          :filter-node-method="filterNode"
          @node-click="handleNodeClick"
          class="dept-tree"
        >
          <template #default="{ node, data }">
            <div class="tree-node">
              <span class="node-label">{{ node.label }}</span>
              <span class="node-actions" @click.stop>
                <el-dropdown trigger="click">
                  <el-icon class="more-icon"><MoreFilled /></el-icon>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item @click="handleCreateChild(data)">
                        <el-icon><Plus /></el-icon>
                        添加子部门
                      </el-dropdown-item>
                      <el-dropdown-item @click="handleEdit(data)">
                        <el-icon><Edit /></el-icon>
                        编辑
                      </el-dropdown-item>
                      <el-dropdown-item @click="handleToggleStatus(data)">
                        <el-icon><component :is="Number(data.status) === 1 ? 'Lock' : 'Unlock'" /></el-icon>
                        {{ Number(data.status) === 1 ? '禁用' : '启用' }}
                      </el-dropdown-item>
                      <el-dropdown-item divided @click="handleDelete(data)">
                        <el-icon><Delete /></el-icon>
                        删除
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </span>
            </div>
          </template>
        </el-tree>
      </el-card>
    </div>

    <!-- 右侧部门详情 -->
    <div class="dept-detail-panel">
      <el-card class="detail-card" shadow="never" v-if="selectedDepartment">
        <template #header>
          <div class="detail-header">
            <span class="detail-title">部门详情</span>
            <div class="detail-actions">
              <el-button size="small" @click="handleEdit(selectedDepartment)">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button
                size="small"
                :type="Number(selectedDepartment.status) === 1 ? 'warning' : 'success'"
                @click="handleToggleStatus(selectedDepartment)"
              >
                <el-icon><component :is="Number(selectedDepartment.status) === 1 ? 'Lock' : 'Unlock'" /></el-icon>
                {{ Number(selectedDepartment.status) === 1 ? '禁用' : '启用' }}
              </el-button>
              <el-button size="small" type="danger" @click="handleDelete(selectedDepartment)">
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
          </div>
        </template>

        <div class="detail-content">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="部门名称">
              {{ selectedDepartment.department_name }}
            </el-descriptions-item>
            <el-descriptions-item label="上级部门">
              {{ getParentDepartmentName(selectedDepartment.parent_id) }}
            </el-descriptions-item>
            <el-descriptions-item label="部门描述" :span="2">
              {{ selectedDepartment.description || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="排序">
              {{ selectedDepartment.sort }}
            </el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="StatusUtils.getTagType(selectedDepartment.status)">
                {{ Number(selectedDepartment.status) === 1 ? '启用' : '禁用' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ formatTime(selectedDepartment.created_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="更新时间">
              {{ formatTime(selectedDepartment.updated_at) }}
            </el-descriptions-item>
          </el-descriptions>

          <!-- 子部门列表 -->
          <div class="children-section" v-if="selectedDepartment.children && selectedDepartment.children.length > 0">
            <div class="section-title">下级部门 ({{ selectedDepartment.children.length }})</div>
            <div class="children-list">
              <el-tag
                v-for="child in selectedDepartment.children"
                :key="child.department_id"
                class="child-tag"
                @click="handleNodeClick(child)"
              >
                {{ child.department_name }}
              </el-tag>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 空状态 -->
      <el-card class="detail-card empty-state" shadow="never" v-else>
        <el-empty description="请从左侧选择一个部门查看详情" />
      </el-card>
    </div>

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
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Edit, Delete, Plus, Search, Refresh, Lock, Unlock, MoreFilled } from '@element-plus/icons-vue'
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
const treeRef = ref()
const treeSearchKeyword = ref('')
const departmentTree = ref<DepartmentTreeNode[]>([])
const selectedDepartment = ref<DepartmentTreeNode | null>(null)
const parentDepartmentId = ref('')
const parentDepartmentName = ref('')

// 过滤后的部门树
const filteredDepartmentTree = computed(() => {
  if (!treeSearchKeyword.value) {
    return departmentTree.value
  }
  const filterTree = (tree: DepartmentTreeNode[]): DepartmentTreeNode[] => {
    return tree
      .filter(node => node.department_name.includes(treeSearchKeyword.value))
      .map(node => ({
        ...node,
        children: node.children ? filterTree(node.children) : undefined
      }))
  }
  return filterTree(departmentTree.value)
})

// 获取父部门名称
function getParentDepartmentName(parentId: string | undefined): string {
  if (!parentId) return '根部门'
  const findParent = (tree: DepartmentTreeNode[], id: string): string | null => {
    for (const node of tree) {
      if (node.department_id === id) return node.department_name
      if (node.children) {
        const found = findParent(node.children, id)
        if (found) return found
      }
    }
    return null
  }
  return findParent(departmentTree.value, parentId) || '-'
}

// 树节点过滤
function filterNode(value: string, data: DepartmentTreeNode) {
  if (!value) return true
  return data.department_name.includes(value)
}

const dialogTitle = computed(() => {
  if (isAddingChild.value) return '添加子部门'
  return isEdit.value ? '编辑部门' : '新建部门'
})

const parentDepartmentOptions = computed(() => {
  const filterTree = (tree: DepartmentTreeNode[], excludeId?: string): DepartmentTreeNode[] => {
    return tree
      .filter(node => node.department_id !== excludeId)
      .map(node => ({
        ...node,
        children: node.children ? filterTree(node.children, excludeId) : undefined
      }))
  }

  if (isEdit.value && form.department_id) {
    return filterTree(departmentTree.value, form.department_id)
  }
  return departmentTree.value
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

// 监听树搜索关键词
watch(treeSearchKeyword, (val) => {
  treeRef.value?.filter(val)
})

onMounted(() => {
  loadTree()
})

async function loadTree() {
  loading.value = true
  try {
    const response = await departmentApi.getTree()
    departmentTree.value = response.tree || []
  } catch (error: any) {
    ElMessage.error(error.message || '加载部门树失败')
  } finally {
    loading.value = false
  }
}

// 处理节点点击
function handleNodeClick(data: DepartmentTreeNode) {
  selectedDepartment.value = data
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
  resetForm()
  parentDepartmentId.value = row.department_id
  parentDepartmentName.value = row.department_name
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
    await loadTree()
    // 更新选中的部门
    if (selectedDepartment.value?.department_id === row.department_id) {
      const findUpdated = (tree: DepartmentTreeNode[], id: string): DepartmentTreeNode | null => {
        for (const node of tree) {
          if (node.department_id === id) return node
          if (node.children) {
            const found = findUpdated(node.children, id)
            if (found) return found
          }
        }
        return null
      }
      const updated = findUpdated(departmentTree.value, row.department_id)
      if (updated) selectedDepartment.value = updated
    }
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
    await loadTree()
    selectedDepartment.value = null
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
    await loadTree()
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
.departments-page {
  display: flex;
  gap: 16px;
  height: calc(100vh - 120px);
}

// 左侧树形面板
.dept-tree-panel {
  width: 320px;
  flex-shrink: 0;

  .tree-card {
    height: 100%;
    display: flex;
    flex-direction: column;

    :deep(.el-card__body) {
      flex: 1;
      overflow: hidden;
      display: flex;
      flex-direction: column;
      padding: 16px;
    }
  }

  .tree-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .tree-title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
    }
  }

  .tree-search {
    margin-bottom: 12px;
  }

  .dept-tree {
    flex: 1;
    overflow: auto;
    border: 1px solid #ebeef5;
    border-radius: 4px;
    padding: 8px;

    :deep(.el-tree-node__content) {
      height: 36px;
      padding-right: 8px;

      &:hover {
        background-color: #f5f7fa;
      }
    }
  }

  .tree-node {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 14px;

    .node-label {
      flex: 1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .node-actions {
      margin-left: 8px;
      opacity: 0;
      transition: opacity 0.2s;

      .more-icon {
        cursor: pointer;
        font-size: 16px;
        color: #909399;

        &:hover {
          color: #409eff;
        }
      }
    }

    &:hover .node-actions {
      opacity: 1;
    }
  }
}

// 右侧详情面板
.dept-detail-panel {
  flex: 1;
  min-width: 0;

  .detail-card {
    height: 100%;
    overflow: auto;

    &.empty-state {
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }

  .detail-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .detail-title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
    }

    .detail-actions {
      display: flex;
      gap: 8px;
    }
  }

  .detail-content {
    .children-section {
      margin-top: 24px;
      padding: 16px;
      background-color: #f5f7fa;
      border-radius: 4px;

      .section-title {
        font-size: 14px;
        font-weight: 600;
        color: #606266;
        margin-bottom: 12px;
      }

      .children-list {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;

        .child-tag {
          cursor: pointer;
          transition: all 0.2s;

          &:hover {
            transform: translateY(-2px);
            box-shadow: 0 2px 8px rgba(64, 158, 255, 0.3);
          }
        }
      }
    }
  }
}
</style>
