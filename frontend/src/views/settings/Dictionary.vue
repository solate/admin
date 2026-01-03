<template>
  <div class="dictionary-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>字典管理</span>
          <el-button type="primary" @click="handleCreateDict">
            <el-icon><Plus /></el-icon>
            新建字典
          </el-button>
        </div>
      </template>

      <!-- 搜索栏 -->
      <div class="search-bar">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索字典名称或编码"
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

      <!-- 字典类型列表 -->
      <el-table :data="dictTypeList" v-loading="loading" style="width: 100%;">
        <el-table-column prop="type_name" label="字典名称" width="200" />
        <el-table-column prop="type_code" label="字典编码" width="200" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" plain @click="handleViewItems(row)">
                <el-icon><View /></el-icon>
                查看字典项
              </el-button>
              <el-button
                v-if="isSuperAdmin"
                size="small"
                type="danger"
                plain
                @click="handleDeleteDict(row)"
              >
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

    <!-- 字典项管理对话框 -->
    <el-dialog
      v-model="itemsDialogVisible"
      :title="`${currentDict?.type_name} - 字典项管理`"
      width="800px"
      :close-on-click-modal="false"
      @close="handleItemsDialogClose"
      class="dict-items-dialog"
    >
      <div class="dict-dialog-header">
        <div class="dict-info">
          <span class="dict-code">{{ currentDict?.type_code }}</span>
          <span class="dict-desc">{{ currentDict?.description || '暂无描述' }}</span>
        </div>
        <el-button type="primary" size="small" @click="handleAddItem">
          添加字典项
        </el-button>
      </div>

      <el-table :data="dictItems" v-loading="itemsLoading" size="small">
        <el-table-column prop="label" label="显示文本" min-width="120" />
        <el-table-column prop="value" label="实际值" width="120" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="来源" width="80">
          <template #default="{ row }">
            <el-tag :type="row.source === 'custom' ? 'warning' : 'info'" size="small">
              {{ row.source === 'custom' ? '自定义' : '系统' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleEditItem(row)">编辑</el-button>
            <el-button v-if="isSuperAdmin" link type="danger" size="small" @click="handleDeleteItem(row)">删除</el-button>
            <el-button v-else-if="row.source === 'custom'" link type="warning" size="small" @click="handleResetItem(row)">恢复默认</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 编辑字典项对话框 -->
    <el-dialog
      v-model="itemEditDialogVisible"
      :title="isEditingNewItem ? '添加字典项' : '编辑字典项'"
      width="500px"
      :close-on-click-modal="false"
      @close="handleItemEditDialogClose"
    >
      <el-form :model="itemForm" :rules="itemRules" ref="itemFormRef" label-width="80px">
        <el-form-item label="显示文本" prop="label">
          <el-input v-model="itemForm.label" placeholder="请输入显示文本" />
        </el-form-item>
        <el-form-item v-if="isEditingNewItem" label="实际值" prop="value">
          <el-input v-model="itemForm.value" placeholder="请输入实际值" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="itemForm.sort" :min="0" :max="9999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="itemEditDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="itemSaveLoading" @click="handleSaveItemForm">确定</el-button>
      </template>
    </el-dialog>

    <!-- 新建字典对话框 -->
    <el-dialog
      v-model="dictDialogVisible"
      :title="isEditDict ? '编辑字典' : '新建字典'"
      width="600px"
      :close-on-click-modal="false"
      @close="handleDictDialogClose"
    >
      <el-form :model="dictForm" :rules="dictRules" ref="dictFormRef" label-width="100px">
        <el-form-item label="字典编码" prop="type_code">
          <el-input
            v-model="dictForm.type_code"
            placeholder="请输入字典编码（如：order_status）"
            maxlength="50"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="字典名称" prop="type_name">
          <el-input
            v-model="dictForm.type_name"
            placeholder="请输入字典名称（如：订单状态）"
            maxlength="100"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="dictForm.description"
            type="textarea"
            :rows="2"
            placeholder="请输入描述"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>

        <!-- 字典项列表 -->
        <el-form-item label="字典项">
          <div class="dict-items-section">
            <div v-if="dictForm.items.length === 0" class="empty-hint">
              暂无字典项，请点击下方按钮添加
            </div>
            <div class="dict-items-table">
              <div
                v-for="(item, index) in dictForm.items"
                :key="index"
                class="dict-item-row"
              >
                <span class="row-index">{{ index + 1 }}</span>
                <el-input
                  v-model="item.label"
                  placeholder="显示文本"
                  size="small"
                />
                <el-input
                  v-model="item.value"
                  placeholder="实际值"
                  size="small"
                />
                <el-input-number
                  v-model="item.sort"
                  :min="0"
                  :max="9999"
                  placeholder="排序"
                  size="small"
                  :controls="false"
                />
                <el-button
                  type="danger"
                  size="small"
                  link
                  @click="handleRemoveFormItem(index)"
                >
                  <el-icon><Delete /></el-icon>
                  删除
                </el-button>
              </div>
            </div>
            <el-button
              type="primary"
              size="default"
              @click="handleAddFormItem"
              class="add-item-btn"
            >
              <el-icon><Plus /></el-icon>
              添加字典项
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dictDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmitDict">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  Plus,
  Search,
  Edit,
  Delete,
  View,
  Check,
  Close,
  RefreshLeft
} from '@element-plus/icons-vue'
import {
  dictApi,
  type DictTypeInfo,
  type DictInfo,
  type CreateSystemDictRequest,
  type CreateDictItemRequest,
  type UpdateDictItemRequest
} from '../../api/dict'
import { getRolesInfo } from '../../utils/token'
import { formatTime } from '../../utils/date'

const loading = ref(false)
const itemsLoading = ref(false)
const submitLoading = ref(false)
const itemSaveLoading = ref(false)
const dictDialogVisible = ref(false)
const itemsDialogVisible = ref(false)
const itemEditDialogVisible = ref(false)
const isEditDict = ref(false)
const isEditingNewItem = ref(false)
const dictFormRef = ref<FormInstance>()
const itemFormRef = ref<FormInstance>()
const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const dictTypeList = ref<DictTypeInfo[]>([])
const currentDict = ref<DictInfo | null>(null)
const currentEditingItem = ref<DictInfo['items'][0] | null>(null)
const dictItems = ref<DictInfo['items'][0]>([])

// 判断是否是超管
const isSuperAdmin = computed(() => {
  const roles = getRolesInfo()
  return roles.some(role => role.role_code === 'super_admin')
})

const dictForm = reactive<CreateSystemDictRequest>({
  type_code: '',
  type_name: '',
  description: '',
  items: []
})

const itemForm = reactive({
  label: '',
  value: '',
  sort: 0
})

const dictRules: FormRules = {
  type_code: [
    { required: true, message: '请输入字典编码', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_]+$/, message: '只能包含字母、数字和下划线', trigger: 'blur' }
  ],
  type_name: [{ required: true, message: '请输入字典名称', trigger: 'blur' }]
}

const itemRules: FormRules = {
  label: [{ required: true, message: '请输入显示文本', trigger: 'blur' }],
  value: [{ required: true, message: '请输入实际值', trigger: 'blur' }],
  sort: [{ required: true, message: '请输入排序', trigger: 'blur' }]
}

onMounted(() => {
  loadData()
})

async function loadData() {
  loading.value = true
  try {
    const apiFunc = isSuperAdmin.value ? dictApi.listSystemTypes : dictApi.listTypes
    const params = {
      page: currentPage.value,
      page_size: pageSize.value,
      keyword: searchKeyword.value || undefined
    }
    const response = await apiFunc(params)
    dictTypeList.value = response.list || []
    total.value = response.total
  } catch (error: any) {
    ElMessage.error(error.message || '加载字典列表失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPage.value = 1
  loadData()
}

async function handleViewItems(row: DictTypeInfo) {
  itemsDialogVisible.value = true
  itemsLoading.value = true
  try {
    const response = await dictApi.getDict(row.type_code)
    currentDict.value = response
    dictItems.value = response.items
  } catch (error: any) {
    ElMessage.error(error.message || '加载字典项失败')
  } finally {
    itemsLoading.value = false
  }
}

function handleAddItem() {
  isEditingNewItem.value = true
  currentEditingItem.value = null
  Object.assign(itemForm, {
    label: '',
    value: '',
    sort: dictItems.value.length
  })
  itemEditDialogVisible.value = true
}

function handleEditItem(row: DictInfo['items'][0]) {
  isEditingNewItem.value = false
  currentEditingItem.value = row
  Object.assign(itemForm, {
    label: row.label,
    value: row.value,
    sort: row.sort
  })
  itemEditDialogVisible.value = true
}

function handleItemEditDialogClose() {
  itemFormRef.value?.resetFields()
  currentEditingItem.value = null
}

async function handleSaveItemForm() {
  if (!itemFormRef.value) return

  try {
    await itemFormRef.value.validate()
  } catch {
    return
  }

  itemSaveLoading.value = true
  try {
    if (isEditingNewItem.value) {
      // 添加新项
      await dictApi.batchUpdateItems({
        type_code: currentDict.value!.type_code,
        items: dictItems.value.map(item => ({ label: item.label, value: item.value, sort: item.sort })).concat({
          label: itemForm.label,
          value: itemForm.value,
          sort: itemForm.sort
        })
      })
      ElMessage.success('添加成功')
    } else {
      // 编辑现有项
      await dictApi.batchUpdateItems({
        type_code: currentDict.value!.type_code,
        items: dictItems.value.map(item =>
          item.item_id === currentEditingItem.value?.item_id
            ? { label: itemForm.label, value: itemForm.value, sort: itemForm.sort }
            : { label: item.label, value: item.value, sort: item.sort }
        )
      })
      ElMessage.success('修改成功')
    }

    itemEditDialogVisible.value = false
    await handleViewItems(currentDict.value as unknown as DictTypeInfo)
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
  } finally {
    itemSaveLoading.value = false
  }
}

async function handleResetItem(row: DictInfo['items'][0]) {
  try {
    await ElMessageBox.confirm(
      `确定要恢复 "${row.label}" 的系统默认值吗？`,
      '恢复默认值',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await dictApi.resetItem(currentDict.value!.type_code, row.value)
    ElMessage.success('恢复成功')
    await handleViewItems(currentDict.value as unknown as DictTypeInfo)
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '恢复失败')
    }
  }
}

async function handleDeleteItem(row: DictInfo['items'][0]) {
  try {
    await ElMessageBox.confirm(
      `确定要删除字典项 "${row.label}" 吗？删除后不可恢复！`,
      '删除字典项',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'error'
      }
    )

    await dictApi.deleteSystemDictItem(currentDict.value!.type_code, row.value)
    ElMessage.success('删除成功')
    await handleViewItems(currentDict.value as unknown as DictTypeInfo)
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

function handleItemsDialogClose() {
  currentDict.value = null
  dictItems.value = []
}

function handleCreateDict() {
  isEditDict.value = false
  resetDictForm()
  dictDialogVisible.value = true
}

function handleAddFormItem() {
  dictForm.items.push({
    label: '',
    value: '',
    sort: dictForm.items.length
  })
}

function handleRemoveFormItem(index: number) {
  dictForm.items.splice(index, 1)
  // 重新排序
  dictForm.items.forEach((item, i) => {
    item.sort = i
  })
}

async function handleDeleteDict(row: DictTypeInfo) {
  try {
    await ElMessageBox.confirm(
      `确定要删除字典 "${row.type_name}" 吗？`,
      '删除字典',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await dictApi.deleteSystemDict(row.type_code)
    ElMessage.success('删除成功')
    loadData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

async function handleSubmitDict() {
  if (!dictFormRef.value) return

  try {
    await dictFormRef.value.validate()
  } catch {
    return
  }

  // 验证字典项
  if (dictForm.items.length === 0) {
    ElMessage.warning('请至少添加一个字典项')
    return
  }

  for (const item of dictForm.items) {
    if (!item.label || !item.value) {
      ElMessage.warning('请填写完整的字典项信息')
      return
    }
  }

  submitLoading.value = true

  try {
    await dictApi.createSystemDict(dictForm)
    ElMessage.success('创建成功')
    dictDialogVisible.value = false
    loadData()
  } catch (error: any) {
    ElMessage.error(error.message || '创建失败')
  } finally {
    submitLoading.value = false
  }
}

function handleDictDialogClose() {
  dictFormRef.value?.resetFields()
  resetDictForm()
}

function resetDictForm() {
  Object.assign(dictForm, {
    type_code: '',
    type_name: '',
    description: '',
    items: []
  })
}
</script>

<style scoped lang="scss">
.dictionary-page {
  :deep(.el-card) {
    border-radius: 8px;
    box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  }

  :deep(.el-card__header) {
    padding: 16px 20px;
    border-bottom: 1px solid #f0f0f0;
    background: linear-gradient(to bottom, #fafafa, #fff);
  }

  :deep(.el-card__body) {
    padding: 20px;
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;

  span {
    font-size: 16px;
    font-weight: 600;
    color: #303133;
  }

  .el-button {
    border-radius: 6px;
  }
}

.search-bar {
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 10px;

  :deep(.el-input) {
    border-radius: 6px;

    .el-input__wrapper {
      border-radius: 6px;
      transition: all 0.3s ease;

      &:hover {
        box-shadow: 0 0 0 1px var(--el-color-primary) inset;
      }
    }
  }
}

// 主列表表格样式
:deep(.el-table) {
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #ebeef5;

  .el-table__header-wrapper {
    th {
      background: linear-gradient(to bottom, #f8f9fa, #f5f7fa);
      color: #606266;
      font-weight: 600;
      border-bottom: 2px solid #e4e7ed;
    }
  }

  .el-table__body-wrapper {
    .el-table__row {
      transition: background-color 0.25s ease;

      &:hover {
        background-color: #f5f7fa !important;
      }

      td {
        border-bottom: 1px solid #f0f0f0;
      }
    }
  }
}

.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;

  :deep(.el-pagination) {
    .btn-prev,
    .btn-next,
    .el-pager li {
      border-radius: 6px;
      transition: all 0.3s ease;

      &:hover {
        transform: translateY(-1px);
      }
    }

    .el-pager li.is-active {
      background: var(--el-color-primary);
    }
  }
}

.action-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;

  .el-button {
    border-radius: 6px;
    font-weight: 500;

    .el-icon {
      margin-right: 4px;
    }
  }
}

// 对话框样式
:deep(.el-dialog) {
  border-radius: 12px;
  overflow: hidden;

  .el-dialog__header {
    padding: 20px 24px;
    border-bottom: 1px solid #f0f0f0;
    background: linear-gradient(to bottom, #fafafa, #fff);

    .el-dialog__title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
    }
  }

  .el-dialog__body {
    padding: 24px;
  }

  .el-dialog__footer {
    padding: 16px 24px;
    border-top: 1px solid #f0f0f0;
    background: #fafafa;
  }
}

// 字典项表单区域（新建字典对话框中使用）
.dict-items-section {
  width: 100%;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  padding: 16px;
  background: #fafafa;

  .empty-hint {
    text-align: center;
    padding: 30px 20px;
    color: #909399;
    font-size: 14px;
    background: #fff;
    border-radius: 6px;
    border: 1px dashed #dcdfe6;
  }

  .dict-items-table {
    max-height: 300px;
    overflow-y: auto;
    margin-bottom: 12px;
    background: #fff;
    border-radius: 6px;
    border: 1px solid #ebeef5;

    .dict-item-row {
      display: grid;
      grid-template-columns: 40px 1fr 1fr 100px auto;
      gap: 10px;
      align-items: center;
      padding: 12px 16px;
      border-bottom: 1px solid #f0f0f0;

      &:last-child {
        border-bottom: none;
      }

      .row-index {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        width: 28px;
        height: 28px;
        background: linear-gradient(135deg, var(--el-color-primary), var(--el-color-primary-light-3));
        color: #fff;
        border-radius: 50%;
        font-size: 12px;
        font-weight: 600;
      }
    }
  }

  .add-item-btn {
    width: 100%;
    border-radius: 6px;
    padding: 12px;
    font-weight: 500;

    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
    }

    .el-icon {
      margin-right: 6px;
    }
  }
}

// 字典项管理对话框样式
.dict-dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 16px;
  margin-bottom: 16px;
  border-bottom: 1px solid #ebeef5;

  .dict-info {
    display: flex;
    align-items: center;
    gap: 12px;

    .dict-code {
      padding: 4px 12px;
      background: #f4f4f5;
      border-radius: 4px;
      color: #606266;
      font-size: 13px;
      font-weight: 500;
    }

    .dict-desc {
      color: #909399;
      font-size: 14px;
    }
  }
}

.dict-items-dialog {
  :deep(.el-dialog__body) {
    padding-top: 16px;
  }
}

// 表单样式优化
:deep(.el-form-item__label) {
  font-weight: 500;
  color: #606266;
}

:deep(.el-input) {
  .el-input__wrapper {
    border-radius: 6px;
    transition: all 0.3s ease;

    &:hover {
      box-shadow: 0 0 0 1px var(--el-color-primary) inset;
    }

    &.is-focus {
      box-shadow: 0 0 0 1px var(--el-color-primary) inset;
    }
  }
}

:deep(.el-textarea) {
  .el-textarea__inner {
    border-radius: 6px;

    &:hover {
      border-color: var(--el-color-primary);
    }

    &:focus {
      border-color: var(--el-color-primary);
    }
  }
}

// 标签样式
:deep(.el-tag) {
  border-radius: 6px;
  padding: 4px 10px;
  font-weight: 500;

  &.el-tag--info {
    background: #f4f4f5;
    border-color: #e9e9eb;
    color: #909399;
  }

  &.el-tag--warning {
    background: #fdf6ec;
    border-color: #f5dab1;
    color: #e6a23c;
  }

  &.el-tag--success {
    background: #f0f9ff;
    border-color: #b3d8ff;
    color: #409eff;
  }

  &.el-tag--primary {
    background: #ecf5ff;
    border-color: #d9ecff;
    color: #409eff;
  }
}
</style>
