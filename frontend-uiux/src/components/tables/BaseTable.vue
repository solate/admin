<!--
基础表格组件
提供统一的表格样式和交互
-->
<script setup lang="ts">
import { ref, computed } from 'vue'

/**
 * 列定义类型
 */
interface Column {
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
  columns: Column[]
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
  hoverable?: boolean
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
  stripe: false,
  border: false,
  size: 'default',
  fit: true,
  showHeader: true,
  highlightCurrentRow: false,
  hoverable: false,
  emptyText: '暂无数据',
})

const emit = defineEmits<Emits>()

const tableRef = ref()

/**
 * 处理选择变化
 */
function handleSelectionChange(selection: any[]) {
  emit('selection-change', selection)
}

/**
 * 处理行点击
 */
function handleRowClick(row: any, column: any, event: Event) {
  emit('row-click', { row, column, event })
}

/**
 * 处理行双击
 */
function handleRowDblclick(row: any, column: any, event: Event) {
  emit('row-dblclick', { row, column, event })
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

/**
 * 获取列的插槽名称
 */
function getColumnSlot(column: Column): string {
  return column.prop ? `cell-${column.prop}` : column.type || ''
}
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
      :hoverable="hoverable"
      :empty-text="emptyText"
      @selection-change="handleSelectionChange"
      @row-click="handleRowClick"
      @row-dblclick="handleRowDblclick"
    >
      <!-- 动态生成列 -->
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
          <!-- 优先使用插槽，如果没有插槽则使用格式化函数或默认值 -->
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

    <!-- 分页插槽 -->
    <slot name="pagination" />
  </div>
</template>

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
