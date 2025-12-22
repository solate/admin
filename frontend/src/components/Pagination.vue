<template>
  <div class="pagination-wrapper">
    <el-pagination
      v-model:current-page="currentPageValue"
      v-model:page-size="pageSizeValue"
      :page-sizes="pageSizes"
      :total="total"
      :layout="layout"
      :background="true"
      :hide-on-single-page="hideOnSinglePage"
      :small="small"
      @size-change="handleSizeChange"
      @current-change="handleCurrentChange"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  total: number
  currentPage?: number
  pageSize?: number
  pageSizes?: number[]
  layout?: string
  hideOnSinglePage?: boolean
  small?: boolean
}

interface Emits {
  (e: 'update:currentPage', value: number): void
  (e: 'update:pageSize', value: number): void
  (e: 'change', data: { page: number; pageSize: number }): void
}

const props = withDefaults(defineProps<Props>(), {
  currentPage: 1,
  pageSize: 10,
  pageSizes: () => [10, 20, 50, 100],
  layout: 'total, sizes, prev, pager, next, jumper',
  hideOnSinglePage: false,
  small: false
})

const emit = defineEmits<Emits>()

const currentPageValue = computed({
  get: () => props.currentPage,
  set: (val) => emit('update:currentPage', val)
})

const pageSizeValue = computed({
  get: () => props.pageSize,
  set: (val) => emit('update:pageSize', val)
})

function handleSizeChange(val: number) {
  emit('update:pageSize', val)
  emit('change', { page: props.currentPage, pageSize: val })
}

function handleCurrentChange(val: number) {
  emit('update:currentPage', val)
  emit('change', { page: val, pageSize: props.pageSize })
}
</script>

<style scoped>
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
  padding: 10px 0;
}
</style>
