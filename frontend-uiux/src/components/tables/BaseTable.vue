<!--
基础表格组件
提供统一的表格样式和交互
-->
<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { TableColumnCtx } from 'element-plus'

export interface TableColumn {
  prop?: string
  label?: string
  width?: string | number
  minWidth?: string | number
  align?: 'left' | 'center' | 'right'
  fixed?: boolean | 'left' | 'right'
  type?: 'selection' | 'index' | 'expand'
  sortable?: boolean
  formatter?: (row: any, column: any, cellValue: any) => any
}

interface Props {
  data: any[]
  columns: TableColumn[]
  loading?: boolean
  selectable?: boolean
  showIndex?: boolean
  stripe?: boolean
  border?: boolean
  size?: 'large' | 'default' | 'small'
  height?: string | number
  maxHeight?: string | number
  fit?: boolean
  showHeader?: boolean
  highlightCurrentRow?: boolean
  emptyText?: string
}

interface Emits {
  (e: 'selection-change', selection: any[]): void
  (e: 'row-click', row: any, column: any, event: Event): void
  (e: 'row-dblclick', row: any, column: any, event: Event): void
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  selectable: false,
  showIndex: false,
  stripe: true,
  border: false,
  size: 'default',
  fit: true,
  showHeader: true,
  highlightCurrentRow: false,
  emptyText: '暂无数据',
})

const emit = defineEmits<Emits>()

const tableRef = ref()
const currentRow = ref()
const selectedRows = ref<any[]>([])

/**
 * 处理选择变化
 */
function handleSelectionChange(selection: any[]) {
  selectedRows.value = selection
  emit('selection-change', selection)
}

/**
 * 处理行点击
 */
function handleRowClick(row: any, column: any, event: Event) {
  currentRow.value = row
  emit('row-click', row, column, event)
}

/**
 * 处理行双击
 */
function handleRowDblclick(row: any, column: any, event: Event) {
  emit('row-dblclick', row, column, event)
}

/**
 * 清除选择
 */
function clearSelection() {
  tableRef.value?.clearSelection()
}

/**
 * 切换某一行的选中状态
 */
function toggleRowSelection(row: any, selected?: boolean) {
  tableRef.value?.toggleRowSelection(row, selected)
}

/**
 * 清除排序
 */
function clearSort() {
  tableRef.value?.clearSort()
}

/**
 * 切换所有行的选中状态
 */
function toggleAllSelection() {
  tableRef.value?.toggleAllSelection()
}

// 暴露方法给父组件
defineExpose({
  clearSelection,
  toggleRowSelection,
  clearSort,
  toggleAllSelection,
  tableRef,
})

// 计算完整的列（包含选择框和索引列）
const fullColumns = computed(() => {
  const cols: TableColumn[] = [...props.columns]

  // 如果需要索引列，添加到最前面
  if (props.showIndex) {
    cols.unshift({
      type: 'index',
      label: '序号',
      width: 60,
      align: 'center',
    })
  }

  // 如果需要选择框，添加到最前面
  if (props.selectable) {
    cols.unshift({
      type: 'selection',
      width: 50,
      align: 'center',
    })
  }

  return cols
})
</script>

<template>
  <div class="base-table-wrapper">
    <el-table
      ref="tableRef"
      :data="data"
      :columns="fullColumns"
      :loading="loading"
      :stripe="stripe"
      :border="border"
      :size="size"
      :height="height"
      :max-height="maxHeight"
      :fit="fit"
      :show-header="showHeader"
      :highlight-current-row="highlightCurrentRow"
      :empty-text="emptyText"
      @selection-change="handleSelectionChange"
      @row-click="handleRowClick"
      @row-dblclick="handleRowDblclick"
    >
      <template v-for="column in columns" :key="column.prop || column.type" #[getColumnSlot(column)]="scope">
        <slot :name="column.prop || column.type" :row="scope.row" :column="scope.column" :$index="scope.$index">
          {{ formatCellValue(scope.row, column) }}
        </slot>
      </template>
    </el-table>

    <!-- 分页插槽 -->
    <slot name="pagination" />
  </div>
</template>

<script lang="ts">
/**
 * 获取列的插槽名称
 */
function getColumnSlot(column: TableColumn): string {
  return column.prop ? `cell-${column.prop}` : column.type || ''
}

/**
 * 格式化单元格值
 */
function formatCellValue(row: any, column: TableColumn): any {
  if (column.formatter) {
    return column.formatter(row, column, row[column.prop!])
  }
  return row[column.prop!]
}
</script>

<style scoped>
.base-table-wrapper {
  width: 100%;
}

.el-table {
  width: 100%;
}

/* 添加表格过渡动画 */
.el-table :deep(.el-table__body tr) {
  transition: background-color 0.2s ease;
}

.el-table :deep(.el-table__body tr:hover > td) {
  background-color: var(--el-table-row-hover-bg-color);
}
</style>
