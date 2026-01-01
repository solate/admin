<template>
  <div class="menus-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>菜单管理</span>
          <div class="header-actions">
            <el-button @click="loadData('tree')">
              <el-icon><List /></el-icon>
              树形视图
            </el-button>
            <el-button @click="loadData('list')">
              <el-icon><Grid /></el-icon>
              列表视图
            </el-button>
            <el-button type="primary" @click="handleCreate">
              <el-icon><Plus /></el-icon>
              新建菜单
            </el-button>
          </div>
        </div>
      </template>

      <!-- 搜索栏 -->
      <div class="search-bar" v-if="viewMode === 'list'">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索菜单名称"
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
          v-model="typeFilter"
          placeholder="类型筛选"
          clearable
          style="width: 120px;"
          @change="handleSearch"
        >
          <el-option label="全部" value="" />
          <el-option label="菜单" value="MENU" />
          <el-option label="按钮" value="BUTTON" />
        </el-select>
        <el-select
          v-model="statusFilter"
          placeholder="状态筛选"
          clearable
          style="width: 120px;"
          @change="handleSearch"
        >
          <el-option label="全部" :value="0" />
          <el-option label="显示" :value="1" />
          <el-option label="隐藏" :value="2" />
        </el-select>
        <el-button type="primary" @click="handleSearch">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </div>

      <!-- 列表视图 -->
      <el-table v-if="viewMode === 'list'" :data="menuList" v-loading="loading" style="width: 100%;" row-key="permission_id">
        <el-table-column prop="name" label="菜单名称" width="200" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.type === 'MENU' ? 'primary' : 'success'" size="small">
              {{ row.type === 'MENU' ? '菜单' : '按钮' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路由路径" min-width="150" show-overflow-tooltip />
        <el-table-column prop="icon" label="图标" width="100" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="Number(row.status) === 1 ? 'success' : 'info'" size="small">
              {{ Number(row.status) === 1 ? '显示' : '隐藏' }}
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
                {{ Number(row.status) === 1 ? '隐藏' : '显示' }}
              </el-button>
              <el-button size="small" type="danger" plain @click="handleDelete(row)">
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 树形视图 -->
      <el-table
        v-if="viewMode === 'tree'"
        :data="menuTree"
        v-loading="loading"
        style="width: 100%;"
        row-key="permission_id"
        :tree-props="{ children: 'children' }"
        default-expand-all
      >
        <el-table-column prop="name" label="菜单名称" width="250">
          <template #default="{ row }">
            <span style="display: flex; align-items: center; gap: 8px;">
              <el-icon v-if="row.icon"><component :is="row.icon" /></el-icon>
              {{ row.name }}
              <el-tag v-if="row.type === 'BUTTON'" size="small" type="success">按钮</el-tag>
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路由路径" min-width="200" show-overflow-tooltip />
        <el-table-column prop="component" label="组件路径" min-width="200" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="Number(row.status) === 1 ? 'success' : 'info'" size="small">
              {{ Number(row.status) === 1 ? '显示' : '隐藏' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
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
                {{ Number(row.status) === 1 ? '隐藏' : '显示' }}
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
      <div class="pagination-wrapper" v-if="viewMode === 'list'">
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

    <!-- 菜单编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑菜单' : '新建菜单'"
      width="600px"
      :close-on-click-modal="false"
      @close="handleDialogClose"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="菜单类型" prop="type">
          <el-radio-group v-model="form.type">
            <el-radio label="MENU">菜单</el-radio>
            <el-radio label="BUTTON">按钮</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="菜单名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入菜单名称" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="父菜单" prop="parent_id">
          <el-tree-select
            v-model="form.parent_id"
            :data="menuTreeForSelect"
            :props="{ label: 'name', value: 'permission_id' }"
            placeholder="选择父菜单（不选则为顶级菜单）"
            clearable
            check-strictly
          />
        </el-form-item>
        <el-form-item label="路由路径" prop="path" v-if="form.type === 'MENU'">
          <el-input v-model="form.path" placeholder="/system/users" />
        </el-form-item>
        <el-form-item label="组件路径" v-if="form.type === 'MENU'">
          <el-input v-model="form.component" placeholder="views/system/Users.vue" />
        </el-form-item>
        <el-form-item label="重定向路径" v-if="form.type === 'MENU'">
          <el-input v-model="form.redirect" placeholder="/system/users" />
        </el-form-item>
        <el-form-item label="图标" v-if="form.type === 'MENU'">
          <el-input v-model="form.icon" placeholder="User" />
        </el-form-item>
        <el-form-item label="资源路径" v-if="form.type === 'BUTTON'">
          <el-input v-model="form.resource" placeholder="/api/v1/users" />
        </el-form-item>
        <el-form-item label="操作方法" v-if="form.type === 'BUTTON'">
          <el-select v-model="form.action" placeholder="选择HTTP方法">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :label="1">显示</el-radio>
            <el-radio :label="2">隐藏</el-radio>
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Edit, Delete, Plus, Search, Lock, Unlock, List, Grid } from '@element-plus/icons-vue'
import { menuApi, type MenuInfo, type MenuTreeNode, type CreateMenuRequest, type UpdateMenuRequest } from '@/api/menu'

const viewMode = ref<'list' | 'tree'>('list')
const loading = ref(false)
const searchKeyword = ref('')
const typeFilter = ref('')
const statusFilter = ref<number>(0)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

const menuList = ref<MenuInfo[]>([])
const menuTree = ref<MenuTreeNode[]>([])
const menuTreeForSelect = ref<MenuTreeNode[]>([])

const dialogVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()

const currentMenu = ref<MenuInfo | null>(null)

const form = reactive<CreateMenuRequest & UpdateMenuRequest>({
  name: '',
  type: 'MENU',
  parent_id: '',
  path: '',
  component: '',
  redirect: '',
  icon: '',
  sort: 0,
  status: 1,
  resource: '',
  action: ''
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择菜单类型', trigger: 'change' }]
}

const formatTime = (timestamp: number) => {
  return new Date(timestamp).toLocaleString('zh-CN')
}

const loadData = async (mode?: 'list' | 'tree') => {
  if (mode) {
    viewMode.value = mode
  }
  loading.value = true

  try {
    if (viewMode.value === 'list') {
      const res = await menuApi.getList({
        page: currentPage.value,
        page_size: pageSize.value,
        name: searchKeyword.value,
        type: typeFilter.value,
        status: statusFilter.value || undefined
      })
      menuList.value = res.list
      total.value = res.total
    } else {
      const res = await menuApi.getMenuTree()
      menuTree.value = res.list
      menuTreeForSelect.value = res.list
    }
  } catch (error) {
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  currentPage.value = 1
  loadData()
}

const handleCreate = () => {
  isEdit.value = false
  Object.assign(form, {
    name: '',
    type: 'MENU',
    parent_id: '',
    path: '',
    component: '',
    redirect: '',
    icon: '',
    sort: 0,
    status: 1,
    resource: '',
    action: ''
  })
  dialogVisible.value = true
}

const handleEdit = (row: MenuInfo | MenuTreeNode) => {
  isEdit.value = true
  currentMenu.value = row as MenuInfo
  Object.assign(form, {
    name: row.name,
    type: row.type as 'MENU' | 'BUTTON',
    parent_id: row.parent_id || '',
    path: row.path || '',
    component: row.component || '',
    redirect: row.redirect || '',
    icon: row.icon || '',
    sort: row.sort || 0,
    status: row.status,
    resource: row.resource || '',
    action: row.action || ''
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitLoading.value = true
    try {
      if (isEdit.value && currentMenu.value) {
        await menuApi.update(currentMenu.value.permission_id, form)
        ElMessage.success('更新成功')
      } else {
        await menuApi.create(form as CreateMenuRequest)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadData()
    } catch (error: any) {
      ElMessage.error(error.message || '操作失败')
    } finally {
      submitLoading.value = false
    }
  })
}

const handleToggleStatus = async (row: MenuInfo | MenuTreeNode) => {
  const newStatus = Number(row.status) === 1 ? 2 : 1
  try {
    await menuApi.updateStatus(row.permission_id, newStatus)
    ElMessage.success(newStatus === 1 ? '已显示' : '已隐藏')
    loadData()
  } catch (error: any) {
    ElMessage.error('操作失败')
  }
}

const handleDelete = (row: MenuInfo | MenuTreeNode) => {
  ElMessageBox.confirm('确定要删除此菜单吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await menuApi.delete(row.permission_id)
      ElMessage.success('删除成功')
      loadData()
    } catch (error: any) {
      ElMessage.error(error.message || '删除失败')
    }
  })
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}

onMounted(() => {
  loadData('list')
  // 同时加载树形数据用于父菜单选择
  menuApi.getMenuTree().then(res => {
    menuTreeForSelect.value = res.list
  })
})
</script>

<style scoped>
.menus-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.search-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
