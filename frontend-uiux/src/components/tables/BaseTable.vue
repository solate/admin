<!--
基础表格组件
封装 Element Plus el-table，提供统一的 API 和样式
-->
<script setup lang="ts">
import { ref } from 'vue'

/**
 * 列定义类型
 * 与 Element Plus el-table-column 属性对齐
 */
export interface TableColumn {
  prop?: string
  label?: string
  width?: string | number
  minWidth?: string | number
  align?: 'left' | 'center' | 'right'
  fixed?: boolean | 'left' | 'right'
  type?: 'selection' | 'index' | 'expand'
  sortable?: boolean | string
  formatter?: (row: any, column: any, cellValue: any, index: number) => any
}

/**
 * 行点击事件参数
 */
export interface RowClickPayload {
  row: any
  column: any
  event: Event
}

interface Props {
  data: any[]
  columns: TableColumn[]
  loading?: boolean
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
  (e: 'row-click', payload: RowClickPayload): void
  (e: 'row-dblclick', payload: RowClickPayload): void
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  stripe: false,
  border: false,
  size: 'default',
  fit: true,
  showHeader: true,
  highlightCurrentRow: false,
  emptyText: '暂无数据',
})

const emit = defineEmits<Emits>()

const tableRef = ref()

function handleSelectionChange(selection: any[]) {
  emit('selection-change', selection)
}

function handleRowClick(row: any, column: any, event: Event) {
  emit('row-click', { row, column, event })
}

function handleRowDblclick(row: any, column: any, event: Event) {
  emit('row-dblclick', { row, column, event })
}

/** 获取列的插槽名称 */
function getColumnSlot(column: TableColumn): string {
  return column.prop ? `cell-${column.prop}` : column.type || ''
}

// 暴露方法给父组件
defineExpose({
  clearSelection: () => tableRef.value?.clearSelection(),
  toggleRowSelection: (row: any, selected?: boolean) => tableRef.value?.toggleRowSelection(row, selected),
  clearSort: () => tableRef.value?.clearSort(),
  toggleAllSelection: () => tableRef.value?.toggleAllSelection(),
  tableRef,
})
</script>

<template>
  <div class="base-table-wrapper w-full">
    <el-table
      ref="tableRef"
      :data="data"
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
      <el-table-column
        v-for="col in columns"
        :key="col.prop || col.type"
        :prop="col.prop"
        :label="col.label"
        :width="col.width"
        :min-width="col.minWidth"
        :align="col.align"
        :fixed="col.fixed"
        :type="col.type"
        :sortable="col.sortable"
      >
        <template #default="scope">
          <slot
            :name="getColumnSlot(col)"
            :row="scope.row"
            :column="scope.column"
            :$index="scope.$index"
          >
            <template v-if="col.formatter">
              {{ col.formatter(scope.row, scope.column, scope.row[col.prop!], scope.$index) }}
            </template>
            <template v-else>
              {{ scope.row[col.prop!] }}
            </template>
          </slot>
        </template>
      </el-table-column>
    </el-table>

    <slot name="pagination" />
  </div>
</template>

<style scoped>
.base-table-wrapper {
  width: 100%;
}

.base-table-wrapper :deep(.el-table) {
  width: 100% !important;
}

.base-table-wrapper :deep(.el-table__header-wrapper),
.base-table-wrapper :deep(.el-table__body-wrapper) {
  width: 100% !important;
}

.base-table-wrapper :deep(.el-table__header),
.base-table-wrapper :deep(.el-table__body) {
  width: 100% !important;
}

.base-table-wrapper :deep(.el-table__body tr) {
  transition: background-color 0.2s ease;
}

.base-table-wrapper :deep(.el-table__body tr:hover > td) {
  background-color: var(--el-table-row-hover-bg-color);
}
</style>
