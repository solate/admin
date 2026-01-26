<!--
基础表格组件
提供统一的表格样式和交互
兼容原有 API 并正确处理百分比宽度
-->
<script setup lang="ts">
import { ref, computed } from 'vue'

/**
 * 列定义类型（兼容旧格式）
 */
interface LegacyColumn {
  key: string       // 列的键名（兼容）
  label?: string     // 列标题
  width?: string | number
  align?: 'left' | 'center' | 'right'
  formatter?: (row: any, column: any, value: any, index: number) => any
}

/**
 * 新列定义类型
 */
interface StandardColumn {
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

type Column = LegacyColumn | StandardColumn

interface Props {
  data: any[]
  columns: Column[]
  loading?: boolean
  selectable?: boolean
  showIndex?: boolean
  striped?: boolean    // 兼容旧属性
  stripe?: boolean      // 新属性名
  border?: boolean
  size?: 'large' | 'default' | 'small'
  height?: string | number
  maxHeight?: string | number
  fit?: boolean
  showHeader?: boolean
  highlightCurrentRow?: boolean
  hoverable?: boolean   // 兼容旧属性
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
  striped: false,     // 兼容旧属性
  stripe: false,      // 新属性名
  border: false,
  size: 'default',
  fit: true,
  showHeader: true,
  highlightCurrentRow: false,
  hoverable: false,   // 兼容旧属性
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
 * 转换列定义格式（兼容旧格式）
 */
const normalizedColumns = computed(() => {
  return props.columns.map(col => {
    // 如果是旧格式（有 key 但没有 prop），转换为标准格式
    if ('key' in col && !('prop' in col)) {
      const legacyCol = col as LegacyColumn
      const width = legacyCol.width
      const isPercentage = typeof width === 'string' && width.includes('%')

      return {
        prop: legacyCol.key,
        label: legacyCol.label,
        // 百分比宽度使用 minWidth，固定宽度使用 width
        ...(isPercentage ? {} : { width }),
        minWidth: isPercentage ? width : undefined,
        align: legacyCol.align,
        formatter: legacyCol.formatter,
      }
    }
    return col as StandardColumn
  })
})

/**
 * 计算是否使用斑马纹（兼容旧属性）
 */
const isStriped = computed(() => {
  return props.striped || props.stripe
})

/**
 * 获取列的插槽名称
 */
function getColumnSlot(column: StandardColumn): string {
  return column.prop ? `cell-${column.prop}` : column.type || ''
}
</script>

<template>
  <div class="base-table-wrapper w-full">
    <el-table
      ref="tableRef"
      :data="data"
      :loading="loading"
      :stripe="isStriped"
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
        v-for="col in normalizedColumns"
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
