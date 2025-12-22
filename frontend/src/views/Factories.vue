<template>
  <div class="factories">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>工厂管理</span>
          <div class="header-actions">
            <el-button @click="loadData" :loading="loading">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
            <el-button type="primary" @click="handleAdd">
              <el-icon><Plus /></el-icon>
              新建工厂
            </el-button>
          </div>
        </div>
      </template>
      
      <div class="search-bar">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索工厂名称"
          style="width: 300px;"
          clearable
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button type="primary" @click="handleSearch" style="margin-left: 10px;">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </div>
      
      <el-table :data="tableData" v-loading="loading" style="width: 100%; margin-top: 16px;">
        <el-table-column prop="factory_id" label="工厂ID" width="200" />
        <el-table-column prop="factory_name" label="工厂名称" />
        <el-table-column prop="address" label="地址" />
        <el-table-column prop="contact_phone" label="联系电话" width="120" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" plain @click="handleEdit(row)">
                <el-icon><Edit /></el-icon>
                编辑
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
    
    <!-- 编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑工厂' : '新建工厂'"
      width="500px"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="90px">
        <el-form-item label="工厂名称" prop="factory_name">
          <el-input v-model="form.factory_name" placeholder="请输入工厂名称" autocomplete="off" />
        </el-form-item>
        <el-form-item label="地址" prop="address">
          <el-input v-model="form.address" placeholder="请输入地址" autocomplete="off" />
        </el-form-item>
        <el-form-item label="联系电话" prop="contact_phone">
          <el-input v-model="form.contact_phone" placeholder="请输入联系电话" autocomplete="off" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="2">禁用</el-radio>
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Edit, Delete, Refresh } from '@element-plus/icons-vue'
import { factoryApi, type FactoryInfo } from '../api'
import Pagination from '../components/Pagination.vue'

const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref()
const loading = ref(false)
const statusFilter = ref<number>()

const tableData = ref<FactoryInfo[]>([])

const form = reactive({
  factory_id: '',
  factory_name: '',
  address: '',
  contact_phone: '',
  status: 1
})

const rules = {
  factory_name: [{ required: true, message: '请输入工厂名称', trigger: 'blur' }],
  address: [{ required: true, message: '请输入地址', trigger: 'blur' }],
  contact_phone: [{ required: true, message: '请输入联系电话', trigger: 'blur' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }]
}

onMounted(() => {
  loadData()
})

async function loadData() {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      page_size: pageSize.value,
      factory_name: searchKeyword.value || undefined,
      status: statusFilter.value
    }
    const res = await factoryApi.getList(params)
    tableData.value = res.list
    total.value = res.page.total
  } catch (error) {
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPage.value = 1
  loadData()
}

function handleAdd() {
  isEdit.value = false
  dialogVisible.value = true
  resetForm()
}

async function handleEdit(row: FactoryInfo) {
  isEdit.value = true
  dialogVisible.value = true
  try {
    const detail = await factoryApi.getDetail(row.factory_id)
    Object.assign(form, detail)
  } catch (error) {
    ElMessage.error('获取详情失败')
  }
}

function handleDelete(row: FactoryInfo) {
  ElMessageBox.confirm('确定要删除这个工厂吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await factoryApi.delete(row.factory_id)
      ElMessage.success('删除成功')
      loadData()
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

async function handleSubmit() {
  await formRef.value?.validate()
  try {
    if (isEdit.value) {
      await factoryApi.update(form.factory_id, {
        factory_name: form.factory_name,
        address: form.address,
        contact_phone: form.contact_phone,
        status: form.status
      })
      ElMessage.success('更新成功')
    } else {
      await factoryApi.create({
        factory_name: form.factory_name,
        address: form.address,
        contact_phone: form.contact_phone,
        status: form.status
      })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadData()
  } catch (error) {
    ElMessage.error(isEdit.value ? '更新失败' : '创建失败')
  }
}

function resetForm() {
  Object.assign(form, {
    factory_id: '',
    factory_name: '',
    address: '',
    contact_phone: '',
    status: 1
  })
}

function formatDate(timestamp: number) {
  return new Date(timestamp * 1000).toLocaleString('zh-CN')
}

function getStatusText(status: number) {
  return status === 1 ? '启用' : '禁用'
}

function getStatusType(status: number) {
  return status === 1 ? 'success' : 'info'
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.search-bar {
  margin-bottom: 16px;
  display: flex;
  align-items: center;
}

.action-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  align-items: center;
}

.action-buttons .el-button {
  margin: 0;
  padding: 4px 8px;
  font-size: 12px;
  border-radius: 4px;
  transition: all 0.3s ease;
}

.action-buttons .el-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.action-buttons .el-button .el-icon {
  margin-right: 2px;
  font-size: 12px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .action-buttons {
    flex-direction: column;
    gap: 2px;
  }
  
  .action-buttons .el-button {
    width: 100%;
    justify-content: center;
  }
}
</style>
