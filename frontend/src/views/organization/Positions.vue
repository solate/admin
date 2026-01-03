<template>
  <div class="positions-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>岗位管理</span>
          <el-button type="primary" @click="handleCreate">
            <el-icon><Plus /></el-icon>
            新建岗位
          </el-button>
        </div>
      </template>

      <div class="search-bar">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索岗位名称或编码"
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
      </div>

      <el-table :data="positionList" v-loading="loading" style="width: 100%;">
        <el-table-column prop="position_name" label="岗位名称" width="200" />
        <el-table-column prop="position_code" label="岗位编码" width="180" />
        <el-table-column prop="level" label="职级" width="100" align="center" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="StatusUtils.getTagType(row.status)">
              {{ Number(row.status) === 1 ? '启用' : '禁用' }}
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

    <!-- 岗位编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑岗位' : '新建岗位'"
      width="500px"
      :close-on-click-modal="false"
      @close="handleDialogClose"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="岗位编码" prop="position_code">
          <el-input
            v-model="form.position_code"
            placeholder="请输入岗位编码（如：DEPT_LEADER）"
            maxlength="50"
            show-word-limit
            :disabled="isEdit"
          />
        </el-form-item>
        <el-form-item label="岗位名称" prop="position_name">
          <el-input v-model="form.position_name" placeholder="请输入岗位名称" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="职级" prop="level">
          <el-input-number v-model="form.level" :min="0" :max="100" controls-position="right" />
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Edit, Delete, Plus, Search, Lock, Unlock } from '@element-plus/icons-vue'
import { StatusUtils } from '@/utils/status'
import { formatTime } from '@/utils/date'
import {
  positionApi,
  type PositionInfo,
  type PositionListParams,
  type CreatePositionRequest,
  type UpdatePositionRequest
} from '@/api/position'

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const searchKeyword = ref('')
const statusFilter = ref<number>(0)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const positionList = ref<PositionInfo[]>([])

const form = reactive<{
  position_id?: string
  position_code: string
  position_name: string
  level: number
  description: string
  sort: number
  status: number
}>({
  position_code: '',
  position_name: '',
  level: 0,
  description: '',
  sort: 0,
  status: 1
})

const rules: FormRules = {
  position_code: [
    { required: true, message: '请输入岗位编码', trigger: 'blur' },
    {
      pattern: /^[A-Z_][A-Z0-9_]*$/,
      message: '岗位编码只能包含大写字母、数字和下划线，且必须以字母或下划线开头',
      trigger: 'blur'
    }
  ],
  position_name: [{ required: true, message: '请输入岗位名称', trigger: 'blur' }]
}

onMounted(() => {
  loadData()
})

async function loadData() {
  loading.value = true
  try {
    const params: PositionListParams = {
      page: currentPage.value,
      page_size: pageSize.value,
      keyword: searchKeyword.value || undefined,
      status: statusFilter.value || undefined
    }
    const response = await positionApi.getList(params)
    positionList.value = response.list || []
    total.value = response.total
  } catch (error: any) {
    ElMessage.error(error.message || '加载岗位列表失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPage.value = 1
  loadData()
}

function handleCreate() {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

function handleEdit(row: PositionInfo) {
  isEdit.value = true
  Object.assign(form, {
    position_id: row.position_id,
    position_code: row.position_code,
    position_name: row.position_name,
    level: row.level,
    description: row.description || '',
    sort: row.sort,
    status: row.status
  })
  dialogVisible.value = true
}

async function handleToggleStatus(row: PositionInfo) {
  const newStatus = Number(row.status) === 1 ? 2 : 1
  const action = newStatus === 1 ? '启用' : '禁用'

  try {
    await ElMessageBox.confirm(
      `确定要${action}岗位 "${row.position_name}" 吗？`,
      `${action}岗位`,
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await positionApi.updateStatus(row.position_id, newStatus)
    ElMessage.success(`岗位${action}成功`)
    loadData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || `${action}失败`)
    }
  }
}

async function handleDelete(row: PositionInfo) {
  try {
    await ElMessageBox.confirm(
      `确定要删除岗位 "${row.position_name}" 吗？`,
      '删除岗位',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await positionApi.delete(row.position_id)
    ElMessage.success('删除成功')
    loadData()
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
      const updateData: UpdatePositionRequest = {
        position_name: form.position_name,
        level: form.level,
        description: form.description,
        sort: form.sort,
        status: form.status
      }
      await positionApi.update(form.position_id!, updateData)
      ElMessage.success('更新成功')
    } else {
      const createData: CreatePositionRequest = {
        position_code: form.position_code,
        position_name: form.position_name,
        level: form.level,
        description: form.description,
        sort: form.sort,
        status: form.status
      }
      await positionApi.create(createData)
      ElMessage.success('创建成功')
    }

    dialogVisible.value = false
    loadData()
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
    position_id: undefined,
    position_code: '',
    position_name: '',
    level: 0,
    description: '',
    sort: 0,
    status: 1
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
</style>
