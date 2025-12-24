<template>
  <div class="permissions-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ pageTitle }}</span>
          <el-button type="primary" @click="handleSave">
            <el-icon><Check /></el-icon>
            保存配置
          </el-button>
        </div>
      </template>

      <div class="search-bar">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索权限名称"
          clearable
          style="width: 300px;"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button type="primary" @click="handleSearch">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </div>

      <el-table
        :data="filteredPermissions"
        v-loading="loading"
        style="width: 100%"
        row-key="id"
        default-expand-all
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
      >
        <el-table-column prop="name" label="权限名称" min-width="200" />
        <el-table-column prop="code" label="权限代码" width="200" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="getTypeColor(row.type)" size="small">
              {{ getTypeText(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路径" width="200" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              active-value="active"
              inactive-value="inactive"
              @change="handleStatusChange(row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" plain @click="handleEdit(row)">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button size="small" type="success" plain @click="handleAddChild(row)">
                <el-icon><Plus /></el-icon>
                添加子项
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

    <!-- 编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="权限名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入权限名称" />
        </el-form-item>
        <el-form-item label="权限代码" prop="code">
          <el-input v-model="form.code" placeholder="请输入权限代码" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" placeholder="请选择类型" style="width: 100%">
            <el-option label="菜单" value="menu" />
            <el-option label="按钮" value="button" />
            <el-option label="接口" value="api" />
            <el-option label="数据" value="data" />
          </el-select>
        </el-form-item>
        <el-form-item label="路径" prop="path" v-if="form.type !== 'button'">
          <el-input v-model="form.path" placeholder="请输入路径" />
        </el-form-item>
        <el-form-item label="图标" prop="icon">
          <el-input v-model="form.icon" placeholder="请输入图标名称" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio label="active">启用</el-radio>
            <el-radio label="inactive">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Check, Search, Edit, Delete, Plus } from '@element-plus/icons-vue'

interface Permission {
  id: number
  name: string
  code: string
  type: 'menu' | 'button' | 'api' | 'data'
  path?: string
  icon?: string
  sort: number
  status: 'active' | 'inactive'
  parentId?: number
  children?: Permission[]
}

const route = useRoute()
const loading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('新建权限')
const formRef = ref()
const searchKeyword = ref('')
const currentParent = ref<Permission | null>(null)

const pageTitle = computed(() => {
  const titles: Record<string, string> = {
    'permissions-menu': '菜单权限',
    'permissions-api': '接口权限',
    'permissions-data': '数据权限'
  }
  return titles[route.name as string] || '权限管理'
})

const permissions = ref<Permission[]>([
  {
    id: 1,
    name: '系统管理',
    code: 'system',
    type: 'menu',
    path: '/system',
    icon: 'Setting',
    sort: 1,
    status: 'active',
    children: [
      {
        id: 11,
        name: '用户管理',
        code: 'system:user',
        type: 'menu',
        path: '/system/users',
        sort: 1,
        status: 'active',
        parentId: 1
      },
      {
        id: 12,
        name: '角色管理',
        code: 'system:role',
        type: 'menu',
        path: '/system/roles',
        sort: 2,
        status: 'active',
        parentId: 1
      }
    ]
  },
  {
    id: 2,
    name: '业务管理',
    code: 'business',
    type: 'menu',
    path: '/business',
    icon: 'Briefcase',
    sort: 2,
    status: 'active',
    children: [
      {
        id: 21,
        name: '商品管理',
        code: 'business:product',
        type: 'menu',
        path: '/business/products',
        sort: 1,
        status: 'active',
        parentId: 2
      },
      {
        id: 22,
        name: '订单管理',
        code: 'business:order',
        type: 'menu',
        path: '/business/orders',
        sort: 2,
        status: 'active',
        parentId: 2
      }
    ]
  }
])

const form = reactive({
  id: 0,
  name: '',
  code: '',
  type: 'menu' as 'menu' | 'button' | 'api' | 'data',
  path: '',
  icon: '',
  sort: 0,
  status: 'active' as 'active' | 'inactive',
  parentId: undefined as number | undefined
})

const rules = {
  name: [{ required: true, message: '请输入权限名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入权限代码', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  path: [{ required: true, message: '请输入路径', trigger: 'blur' }]
}

const filteredPermissions = computed(() => {
  if (!searchKeyword.value) return permissions.value
  const keyword = searchKeyword.value.toLowerCase()
  return filterPermissions(permissions.value, keyword)
})

function filterPermissions(list: Permission[], keyword: string): Permission[] {
  return list.reduce((acc: Permission[], item) => {
    const matchName = item.name.toLowerCase().includes(keyword)
    const matchCode = item.code.toLowerCase().includes(keyword)
    const filteredChildren = item.children ? filterPermissions(item.children, keyword) : []

    if (matchName || matchCode || filteredChildren.length > 0) {
      acc.push({
        ...item,
        children: filteredChildren.length > 0 ? filteredChildren : item.children
      })
    }
    return acc
  }, [])
}

function getTypeColor(type: string) {
  const colors: Record<string, string> = {
    menu: 'primary',
    button: 'success',
    api: 'warning',
    data: 'info'
  }
  return colors[type] || ''
}

function getTypeText(type: string) {
  const texts: Record<string, string> = {
    menu: '菜单',
    button: '按钮',
    api: '接口',
    data: '数据'
  }
  return texts[type] || type
}

function handleSearch() {
  // 搜索由 computed 自动处理
}

function handleEdit(row: Permission) {
  dialogTitle.value = '编辑权限'
  Object.assign(form, {
    id: row.id,
    name: row.name,
    code: row.code,
    type: row.type,
    path: row.path || '',
    icon: row.icon || '',
    sort: row.sort,
    status: row.status,
    parentId: row.parentId
  })
  dialogVisible.value = true
}

function handleAddChild(row: Permission) {
  dialogTitle.value = '添加子权限'
  Object.assign(form, {
    id: 0,
    name: '',
    code: '',
    type: 'menu',
    path: '',
    icon: '',
    sort: 0,
    status: 'active',
    parentId: row.id
  })
  dialogVisible.value = true
}

function handleDelete(row: Permission) {
  ElMessage.warning('删除功能开发中...')
}

function handleStatusChange(row: Permission) {
  ElMessage.success(`权限 "${row.name}" 状态已更新`)
}

function handleSubmit() {
  formRef.value?.validate((valid: boolean) => {
    if (valid) {
      ElMessage.success('保存成功')
      dialogVisible.value = false
    }
  })
}

function handleSave() {
  ElMessage.success('权限配置已保存')
}

onMounted(() => {
  // 初始化数据
})
</script>

<style scoped lang="scss">
.permissions-page {
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
}
</style>