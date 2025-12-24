<template>
  <div class="roles-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>角色管理</span>
          <el-button type="primary" @click="handleCreate">
            <el-icon><Plus /></el-icon>
            新建角色
          </el-button>
        </div>
      </template>

      <div class="search-bar">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索角色名称或代码"
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

      <el-table :data="filteredRoles" v-loading="loading" style="width: 100%;">
        <el-table-column prop="name" label="角色名称" width="200" />
        <el-table-column prop="code" label="角色代码" width="200" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="userCount" label="用户数" width="100" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'">
              {{ row.status === 'active' ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" plain @click="handleEdit(row)">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button size="small" type="warning" plain @click="handlePermissions(row)">
                <el-icon><Setting /></el-icon>
                权限配置
              </el-button>
              <el-button size="small" type="danger" plain @click="handleDelete(row)">
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <Pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        @change="loadData"
      />
    </el-card>

    <!-- 角色编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑角色' : '新建角色'"
      width="500px"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="角色代码" prop="code">
          <el-input v-model="form.code" placeholder="请输入角色代码" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" placeholder="请输入描述" />
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { Edit, Delete, Plus, Search, Setting } from '@element-plus/icons-vue'
import Pagination from '../../components/Pagination.vue'

interface Role {
  id: number
  name: string
  code: string
  description: string
  userCount: number
  status: string
}

const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref()
const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const roles = ref<Role[]>([
  { id: 1, name: '超级管理员', code: 'super_admin', description: '系统超级管理员', userCount: 1, status: 'active' },
  { id: 2, name: '管理员', code: 'admin', description: '系统管理员', userCount: 5, status: 'active' },
  { id: 3, name: '运营', code: 'operator', description: '运营人员', userCount: 10, status: 'active' },
  { id: 4, name: '普通用户', code: 'user', description: '普通用户', userCount: 50, status: 'active' }
])

const form = reactive({
  id: 0,
  name: '',
  code: '',
  description: '',
  status: 'active'
})

const rules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入角色代码', trigger: 'blur' }],
  description: [{ required: true, message: '请输入描述', trigger: 'blur' }]
}

const filteredRoles = computed(() => {
  if (!searchKeyword.value) return roles.value
  const keyword = searchKeyword.value.toLowerCase()
  return roles.value.filter(role =>
    role.name.toLowerCase().includes(keyword) ||
    role.code.toLowerCase().includes(keyword)
  )
})

onMounted(() => {
  loadData()
})

function loadData() {
  loading.value = true
  setTimeout(() => {
    total.value = filteredRoles.value.length
    loading.value = false
  }, 300)
}

function handleSearch() {
  currentPage.value = 1
  loadData()
}

function handleCreate() {
  isEdit.value = false
  dialogVisible.value = true
  resetForm()
}

function handleEdit(row: Role) {
  isEdit.value = true
  dialogVisible.value = true
  Object.assign(form, row)
}

function handlePermissions(row: Role) {
  ElMessage.info(`配置权限：${row.name}`)
}

function handleDelete(row: Role) {
  ElMessageBox.confirm(`确定要删除角色 "${row.name}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    const index = roles.value.findIndex(r => r.id === row.id)
    if (index !== -1) {
      roles.value.splice(index, 1)
      ElMessage.success('删除成功')
      loadData()
    }
  }).catch(() => {})
}

function handleSubmit() {
  formRef.value?.validate((valid: boolean) => {
    if (valid) {
      if (isEdit.value) {
        const index = roles.value.findIndex(r => r.id === form.id)
        if (index !== -1) {
          Object.assign(roles.value[index], form)
        }
        ElMessage.success('更新成功')
      } else {
        roles.value.unshift({
          id: Date.now(),
          ...form,
          userCount: 0
        } as Role)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      loadData()
    }
  })
}

function resetForm() {
  Object.assign(form, {
    id: 0,
    name: '',
    code: '',
    description: '',
    status: 'active'
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
</style>