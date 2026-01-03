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
      <el-table v-if="viewMode === 'list'" :data="menuList" v-loading="loading" style="width: 100%;" row-key="menu_id">
        <el-table-column prop="name" label="菜单名称" width="200" />
        <el-table-column prop="path" label="路由路径" min-width="150" show-overflow-tooltip />
        <el-table-column prop="component" label="组件路径" min-width="150" show-overflow-tooltip />
        <el-table-column prop="icon" label="图标" width="100" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="StatusUtils.getTagType(row.status)" size="small">
              {{ Number(row.status) === 1 ? '显示' : '隐藏' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
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
        row-key="menu_id"
        :tree-props="{ children: 'children' }"
        default-expand-all
      >
        <el-table-column prop="name" label="菜单名称" width="250">
          <template #default="{ row }">
            <span style="display: flex; align-items: center; gap: 8px;">
              <el-icon v-if="row.icon"><component :is="row.icon" /></el-icon>
              {{ row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路由路径" min-width="200" show-overflow-tooltip />
        <el-table-column prop="component" label="组件路径" min-width="200" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="StatusUtils.getTagType(row.status)" size="small">
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
      width="700px"
      :close-on-click-modal="false"
      @close="handleDialogClose"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px">
        <el-form-item label="上级菜单" prop="parent_id">
          <el-tree-select
            v-model="form.parent_id"
            :data="menuTreeForSelect"
            :props="{ label: 'name', value: 'menu_id' }"
            placeholder="选择上级菜单（不选则为顶级菜单）"
            clearable
            check-strictly
            style="width: 100%;"
          />
        </el-form-item>
        <el-form-item label="菜单名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入菜单名称" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="路由路径" prop="path">
          <el-input v-model="form.path" placeholder="/system/users（前端路由路径）" />
        </el-form-item>
        <el-form-item label="组件路径">
          <el-input v-model="form.component" placeholder="views/system/Users.vue" />
        </el-form-item>
        <el-form-item label="重定向路径">
          <el-input v-model="form.redirect" placeholder="/system/users" />
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="form.icon" placeholder="User（Element Plus 图标名称）">
            <template #append>
              <el-button @click="showIconPicker = true">选择图标</el-button>
            </template>
          </el-input>
          <div v-if="form.icon" style="margin-top: 8px;">
            <el-icon><component :is="form.icon" /></el-icon>
            {{ form.icon }}
          </div>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="9999" />
          <span style="margin-left: 12px; color: #999;">数值越小越靠前</span>
        </el-form-item>
        <el-form-item label="关联 API">
          <div style="width: 100%;">
            <div v-for="(apiPath, index) in apiPathsList" :key="index" style="display: flex; gap: 8px; margin-bottom: 8px;">
              <el-input v-model="apiPath.path" placeholder="/api/v1/users" style="flex: 2;" />
              <el-select v-model="apiPath.methods" multiple placeholder="请求方法" style="flex: 1;">
                <el-option label="GET" value="GET" />
                <el-option label="POST" value="POST" />
                <el-option label="PUT" value="PUT" />
                <el-option label="DELETE" value="DELETE" />
                <el-option label="PATCH" value="PATCH" />
              </el-select>
              <el-button type="danger" plain @click="removeApiPath(index)" :disabled="apiPathsList.length === 1">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button type="primary" plain @click="addApiPath" style="width: 100%;">
              <el-icon><Plus /></el-icon>
              添加 API
            </el-button>
            <div style="margin-top: 8px; color: #999; font-size: 12px;">
              关联的 API 会在分配菜单权限时自动添加到角色的 API 权限中
            </div>
          </div>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :label="1">显示</el-radio>
            <el-radio :label="2">隐藏</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="菜单描述" maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 图标选择器 -->
    <el-dialog v-model="showIconPicker" title="选择图标" width="800px">
      <div style="max-height: 400px; overflow-y: auto;">
        <div style="display: grid; grid-template-columns: repeat(8, 1fr); gap: 8px;">
          <div
            v-for="icon in commonIcons"
            :key="icon"
            @click="selectIcon(icon)"
            style="padding: 12px; text-align: center; cursor: pointer; border: 1px solid #eee; border-radius: 4px;"
            :style="{ background: form.icon === icon ? '#f0f9ff' : '', borderColor: form.icon === icon ? '#1890ff' : '#eee' }"
          >
            <el-icon><component :is="icon" /></el-icon>
            <div style="font-size: 12px; margin-top: 4px;">{{ icon }}</div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showIconPicker = false">取消</el-button>
        <el-button type="primary" @click="showIconPicker = false">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Edit, Delete, Plus, Search, Lock, Unlock, List, Grid } from '@element-plus/icons-vue'
import { StatusUtils } from '@/utils/status'
import { formatTime } from '@/utils/date'
import { menuApi, type MenuInfo, type MenuTreeNode, type CreateMenuRequest, type UpdateMenuRequest, type ApiPath } from '@/api/menu'

const viewMode = ref<'list' | 'tree'>('list')
const loading = ref(false)
const searchKeyword = ref('')
const statusFilter = ref<number>(0)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

const menuList = ref<MenuInfo[]>([])
const menuTree = ref<MenuTreeNode[]>([])
const menuTreeForSelect = ref<MenuTreeNode[]>([])

const dialogVisible = ref(false)
const showIconPicker = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()

const currentMenu = ref<MenuInfo | null>(null)

// API 路径列表
const apiPathsList = ref<ApiPath[]>([{ path: '', methods: [] }])

const form = reactive<CreateMenuRequest & UpdateMenuRequest>({
  name: '',
  parent_id: '',
  path: '',
  component: '',
  redirect: '',
  icon: '',
  sort: 0,
  status: 1,
  description: '',
  api_paths: ''
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }]
}

// 常用图标列表
const commonIcons = [
  'Home', 'Dashboard', 'Document', 'Folder', 'FolderOpened', 'User', 'UserFilled',
  'Setting', 'Tools', 'Management', 'DataLine', 'DataAnalysis', 'TrendCharts',
  'List', 'Grid', 'Menu', 'Operation', 'MoreFilled', 'Plus', 'Minus',
  'Edit', 'Delete', 'Search', 'Refresh', 'Filter', 'Sort', 'Download', 'Upload',
  'Share', 'Link', 'Message', 'ChatLineSquare', 'Bell', 'Warning', 'InfoFilled',
  'SuccessFilled', 'CircleCheck', 'CircleClose', 'Loading', 'View', 'Hide'
]

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
    parent_id: '',
    path: '',
    component: '',
    redirect: '',
    icon: '',
    sort: 0,
    status: 1,
    description: '',
    api_paths: ''
  })
  apiPathsList.value = [{ path: '', methods: [] }]
  dialogVisible.value = true
}

const handleEdit = (row: MenuInfo | MenuTreeNode) => {
  isEdit.value = true
  currentMenu.value = row as MenuInfo

  // 解析 API 路径
  let apiPaths: ApiPath[] = []
  if (row.api_paths) {
    try {
      apiPaths = JSON.parse(row.api_paths)
    } catch (e) {
      apiPaths = []
    }
  }
  if (apiPaths.length === 0) {
    apiPaths = [{ path: '', methods: [] }]
  }
  apiPathsList.value = apiPaths

  Object.assign(form, {
    name: row.name,
    parent_id: row.parent_id || '',
    path: row.path || '',
    component: row.component || '',
    redirect: row.redirect || '',
    icon: row.icon || '',
    sort: row.sort || 0,
    status: row.status,
    description: row.description || '',
    api_paths: row.api_paths || ''
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitLoading.value = true
    try {
      // 将 apiPathsList 转换为 JSON 字符串
      const validApiPaths = apiPathsList.value.filter(p => p.path.trim())
      const submitData = {
        ...form,
        api_paths: validApiPaths.length > 0 ? JSON.stringify(validApiPaths) : ''
      }

      if (isEdit.value && currentMenu.value) {
        await menuApi.update(currentMenu.value.menu_id, submitData)
        ElMessage.success('更新成功')
      } else {
        await menuApi.create(submitData as CreateMenuRequest)
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
    await menuApi.updateStatus(row.menu_id, newStatus)
    ElMessage.success(newStatus === 1 ? '已显示' : '已隐藏')
    loadData()
  } catch (error: any) {
    ElMessage.error('操作失败')
  }
}

const handleDelete = (row: MenuInfo | MenuTreeNode) => {
  ElMessageBox.confirm('确定要删除此菜单吗？删除后相关的权限配置也会被清理。', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await menuApi.delete(row.menu_id)
      ElMessage.success('删除成功')
      loadData()
    } catch (error: any) {
      ElMessage.error(error.message || '删除失败')
    }
  })
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
  apiPathsList.value = [{ path: '', methods: [] }]
}

const addApiPath = () => {
  apiPathsList.value.push({ path: '', methods: [] })
}

const removeApiPath = (index: number) => {
  if (apiPathsList.value.length > 1) {
    apiPathsList.value.splice(index, 1)
  }
}

const selectIcon = (icon: string) => {
  form.icon = icon
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
