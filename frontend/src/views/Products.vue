<template>
  <div class="products">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>商品管理</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新建商品
          </el-button>
        </div>
      </template>
      
      <div class="search-bar">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索商品名称"
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
        <el-table-column prop="product_id" label="商品ID" width="200" />
        <el-table-column prop="product_name" label="商品名称" />
        <el-table-column prop="unit" label="单位" width="80" />
        <el-table-column label="工厂" width="100">
          <template #default="{ row }">
            {{ getFactoryName(row.factory_id) }}
          </template>
        </el-table-column>
        <el-table-column label="采购价格" width="100">
          <template #default="{ row }">
            ¥{{ row.purchase_price }}
          </template>
        </el-table-column>
        <el-table-column label="销售价格" width="100">
          <template #default="{ row }">
            ¥{{ row.sale_price }}
          </template>
        </el-table-column>
        <el-table-column label="当前库存" width="100">
          <template #default="{ row }">
            <el-tag :type="getStockType(row)">
              {{ row.current_stock }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="350" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" plain @click="handleEdit(row)">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button size="small" type="success" plain @click="handleStockIn(row)">
                <el-icon><Plus /></el-icon>
                入库
              </el-button>
              <el-button size="small" type="warning" plain @click="handleStockOut(row)">
                <el-icon><Minus /></el-icon>
                出库
              </el-button>
              <el-button size="small" type="info" plain @click="handleViewHistory(row)">
                <el-icon><Tickets /></el-icon>
                记录
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
    
    <!-- 商品编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑商品' : '新建商品'"
      width="600px"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="商品名称" prop="product_name">
          <el-input v-model="form.product_name" placeholder="请输入商品名称" />
        </el-form-item>
        <el-form-item label="单位" prop="unit">
          <el-input v-model="form.unit" placeholder="请输入单位，如：个、台、箱" />
        </el-form-item>
        <el-form-item label="所属工厂" prop="factory_id">
          <el-select v-model="form.factory_id" placeholder="请选择所属工厂" clearable>
            <el-option
              v-for="factory in factories"
              :key="factory.factory_id"
              :label="factory.factory_name"
              :value="factory.factory_id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="采购价格" prop="purchase_price">
          <el-input v-model="form.purchase_price" placeholder="请输入采购价格">
            <template #prefix>¥</template>
          </el-input>
        </el-form-item>
        <el-form-item label="销售价格" prop="sale_price">
          <el-input v-model="form.sale_price" placeholder="请输入销售价格">
            <template #prefix>¥</template>
          </el-input>
        </el-form-item>
        <el-form-item label="初始库存" prop="current_stock" v-if="!isEdit">
          <el-input-number v-model="form.current_stock" :min="0" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="最小库存" prop="min_stock">
          <el-input-number v-model="form.min_stock" :min="0" style="width: 100%;" />
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
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
    
    <!-- 库存操作对话框 -->
    <el-dialog
      v-model="stockDialogVisible"
      :title="stockType === 'in' ? '商品入库' : '商品出库'"
      width="500px"
    >
      <el-form :model="stockForm" :rules="stockRules" ref="stockFormRef" label-width="100px">
        <el-form-item label="商品名称">
          <el-input v-model="currentProduct.product_name" disabled />
        </el-form-item>
        <el-form-item label="当前库存">
          <el-input v-model="currentProduct.current_stock" disabled />
        </el-form-item>
        <el-form-item :label="stockType === 'in' ? '入库数量' : '出库数量'" prop="quantity">
          <el-input-number v-model="stockForm.quantity" :min="1" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="单价" prop="unit_price">
          <el-input v-model="stockForm.unit_price" placeholder="请输入单价">
            <template #prefix>¥</template>
          </el-input>
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="stockForm.remark" type="textarea" placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="stockDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleStockSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 操作记录对话框 -->
    <el-dialog
      v-model="historyDialogVisible"
      title="商品操作记录"
      width="900px"
    >
      <div class="history-header">
        <div class="product-info">
          <el-descriptions :column="3" border>
            <el-descriptions-item label="商品名称">{{ currentProduct.product_name }}</el-descriptions-item>
            <el-descriptions-item label="当前库存">
              <el-tag :type="getStockType(currentProduct)">{{ currentProduct.current_stock }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="单位">{{ currentProduct.unit }}</el-descriptions-item>
          </el-descriptions>
        </div>
      </div>

      <el-table 
        :data="historyData" 
        v-loading="historyLoading" 
        style="width: 100%; margin-top: 20px;"
        max-height="400"
      >
        <el-table-column prop="operation_time" label="操作时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.operation_time) }}
          </template>
        </el-table-column>
        <el-table-column label="操作类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.operation_type === 'in' ? 'success' : 'warning'">
              {{ row.operation_type === 'in' ? '入库' : '出库' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="operator_name" label="操作人" width="100" />
        <el-table-column prop="quantity" label="数量" width="80" />
        <el-table-column label="单价" width="100">
          <template #default="{ row }">
            ¥{{ row.unit_price }}
          </template>
        </el-table-column>
        <el-table-column label="总金额" width="100">
          <template #default="{ row }">
            ¥{{ row.total_amount }}
          </template>
        </el-table-column>
        <el-table-column label="操作前库存" width="100">
          <template #default="{ row }">
            {{ row.before_stock }}
          </template>
        </el-table-column>
        <el-table-column label="操作后库存" width="100">
          <template #default="{ row }">
            {{ row.after_stock }}
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" show-overflow-tooltip />
      </el-table>

      <Pagination
        v-model:current-page="historyPage"
        v-model:page-size="historyPageSize"
        :total="historyTotal"
        :page-sizes="[10, 20, 50]"
        @change="loadHistory"
      />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Edit, Plus, Minus, Delete, Search, Tickets } from '@element-plus/icons-vue'
import { productApi, inventoryApi, factoryApi, type ProductInfo, type InventoryInfo, type FactoryInfo } from '../api'
import Pagination from '../components/Pagination.vue'

const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const stockDialogVisible = ref(false)
const historyDialogVisible = ref(false)
const isEdit = ref(false)
const stockType = ref<'in' | 'out'>('in')
const formRef = ref()
const stockFormRef = ref()
const loading = ref(false)
const historyLoading = ref(false)

const tableData = ref<ProductInfo[]>([])
const currentProduct = ref<ProductInfo>({} as ProductInfo)
const historyData = ref<InventoryInfo[]>([])
const historyPage = ref(1)
const historyPageSize = ref(10)
const historyTotal = ref(0)
const factories = ref<FactoryInfo[]>([])

const form = reactive({
  product_id: '',
  product_name: '',
  unit: '',
  factory_id: '',
  purchase_price: '',
  sale_price: '',
  current_stock: 0,
  min_stock: 0,
  status: 1
})

const stockForm = reactive({
  quantity: 1,
  unit_price: '',
  remark: ''
})

const rules = {
  product_name: [{ required: true, message: '请输入商品名称', trigger: 'blur' }],
  unit: [{ required: true, message: '请输入单位', trigger: 'blur' }],
  factory_id: [{ required: true, message: '请选择所属工厂', trigger: 'change' }],
  purchase_price: [{ required: true, message: '请输入采购价格', trigger: 'blur' }],
  sale_price: [{ required: true, message: '请输入销售价格', trigger: 'blur' }],
  current_stock: [{ required: true, message: '请输入初始库存', trigger: 'blur' }],
  min_stock: [{ required: true, message: '请输入最小库存', trigger: 'blur' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }]
}

const stockRules = {
  quantity: [{ required: true, message: '请输入数量', trigger: 'blur' }],
  unit_price: [{ required: true, message: '请输入单价', trigger: 'blur' }],
  remark: [{ required: true, message: '请输入备注', trigger: 'blur' }]
}

onMounted(() => {
  loadFactories()
  loadData()
})

async function loadData() {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      page_size: pageSize.value,
      product_name: searchKeyword.value || undefined
    }
    const res = await productApi.getList(params)
    tableData.value = res.list
    total.value = res.page.total
  } catch (error) {
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

async function loadFactories() {
  try {
    const res = await factoryApi.getList({ page: 1, page_size: 100 })
    factories.value = res.list || []
  } catch (error) {
    ElMessage.error('加载工厂列表失败')
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

async function handleEdit(row: ProductInfo) {
  isEdit.value = true
  dialogVisible.value = true
  try {
    const detail = await productApi.getDetail(row.product_id)
    Object.assign(form, detail)
  } catch (error) {
    ElMessage.error('获取详情失败')
  }
}

function handleStockIn(row: ProductInfo) {
  stockType.value = 'in'
  currentProduct.value = row
  stockDialogVisible.value = true
  resetStockForm()
  stockForm.unit_price = row.purchase_price
}

function handleStockOut(row: ProductInfo) {
  stockType.value = 'out'
  currentProduct.value = row
  stockDialogVisible.value = true
  resetStockForm()
  stockForm.unit_price = row.sale_price
}

function handleDelete(row: ProductInfo) {
  ElMessageBox.confirm('确定要删除这个商品吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await productApi.delete(row.product_id)
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
      await productApi.update(form.product_id, {
        product_name: form.product_name,
        unit: form.unit,
        factory_id: form.factory_id,
        purchase_price: form.purchase_price,
        sale_price: form.sale_price,
        min_stock: form.min_stock,
        status: form.status
      })
      ElMessage.success('更新成功')
    } else {
      await productApi.create({
        product_name: form.product_name,
        unit: form.unit,
        factory_id: form.factory_id,
        purchase_price: form.purchase_price,
        sale_price: form.sale_price,
        current_stock: form.current_stock,
        min_stock: form.min_stock,
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

async function handleStockSubmit() {
  await stockFormRef.value?.validate()
  try {
    const userId = localStorage.getItem('user_id') || ''
    if (stockType.value === 'in') {
      await inventoryApi.productIn({
        product_id: currentProduct.value.product_id,
        quantity: stockForm.quantity,
        unit_price: stockForm.unit_price,
        operator_id: userId,
        remark: stockForm.remark
      })
      ElMessage.success('入库成功')
    } else {
      await inventoryApi.productOut({
        product_id: currentProduct.value.product_id,
        quantity: stockForm.quantity,
        unit_price: stockForm.unit_price,
        operator_id: userId,
        remark: stockForm.remark
      })
      ElMessage.success('出库成功')
    }
    stockDialogVisible.value = false
    loadData()
  } catch (error) {
    ElMessage.error(stockType.value === 'in' ? '入库失败' : '出库失败')
  }
}

function resetForm() {
  Object.assign(form, {
    product_id: '',
    product_name: '',
    unit: '',
    factory_id: '',
    purchase_price: '',
    sale_price: '',
    current_stock: 0,
    min_stock: 0,
    status: 1
  })
}

function resetStockForm() {
  Object.assign(stockForm, {
    quantity: 1,
    unit_price: '',
    remark: ''
  })
}

function getStockType(row: ProductInfo) {
  if (row.current_stock <= row.min_stock) return 'danger'
  if (row.current_stock <= row.min_stock * 2) return 'warning'
  return 'success'
}

function getFactoryName(factoryId: string) {
  const factory = factories.value.find((f) => f.factory_id === factoryId)
  return factory ? factory.factory_name : '-'
}

// 查看操作记录
function handleViewHistory(row: ProductInfo) {
  currentProduct.value = row
  historyDialogVisible.value = true
  historyPage.value = 1
  loadHistory()
}

// 加载操作记录
async function loadHistory() {
  historyLoading.value = true
  try {
    const res = await inventoryApi.getHistory({
      product_id: currentProduct.value.product_id,
      page: historyPage.value,
      page_size: historyPageSize.value
    })
    historyData.value = res.list || []
    historyTotal.value = res.page?.total || 0
  } catch (error) {
    ElMessage.error('加载操作记录失败')
  } finally {
    historyLoading.value = false
  }
}

// 格式化时间
function formatTime(timestamp: number) {
  const date = new Date(timestamp)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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

/* 操作记录对话框样式 */
.history-header {
  margin-bottom: 16px;
}

.product-info {
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
}
</style>
