<script setup>
import { computed } from 'vue'

const props = defineProps({
  columns: {
    type: Array,
    required: true
  },
  data: {
    type: Array,
    required: true
  },
  loading: {
    type: Boolean,
    default: false
  },
  striped: {
    type: Boolean,
    default: true
  },
  hoverable: {
    type: Boolean,
    default: true
  },
  compact: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['row-click'])

const tableClasses = computed(() => {
  return [
    'w-full',
    props.compact ? 'text-sm' : 'text-base'
  ].join(' ')
})

const rowClasses = computed(() => {
  return (index) => [
    'border-b border-slate-200 dark:border-slate-700 transition-colors duration-150',
    props.striped && index % 2 === 0 ? 'bg-slate-50/50 dark:bg-slate-800/50' : 'bg-white dark:bg-slate-800',
    props.hoverable ? 'hover:bg-slate-100 dark:hover:bg-slate-700 cursor-pointer' : ''
  ].filter(Boolean).join(' ')
})

const headerClasses = 'px-4 py-3 text-left text-xs font-semibold tracking-wide text-slate-500 dark:text-slate-400 uppercase bg-slate-50 dark:bg-slate-900/50'
const cellClasses = 'px-4 py-3 text-slate-700 dark:text-slate-300'

const handleRowClick = (row, index) => {
  emit('row-click', { row, index })
}
</script>

<template>
  <div class="overflow-x-auto rounded-xl border border-slate-200 dark:border-slate-700">
    <table :class="tableClasses">
      <thead>
        <tr>
          <th
            v-for="column in columns"
            :key="column.key"
            :class="headerClasses"
            :style="column.width ? `width: ${column.width}` : ''"
          >
            {{ column.label }}
          </th>
        </tr>
      </thead>
      <tbody>
        <template v-if="loading">
          <tr v-for="i in 3" :key="`skeleton-${i}`">
            <td
              v-for="column in columns"
              :key="column.key"
              :class="cellClasses"
            >
              <div class="h-4 bg-slate-200 dark:bg-slate-700 rounded animate-pulse" />
            </td>
          </tr>
        </template>
        <template v-else-if="data.length === 0">
          <tr>
            <td
              :colspan="columns.length"
              class="px-4 py-12 text-center text-slate-500 dark:text-slate-400"
            >
              <slot name="empty">
                <p class="text-sm">暂无数据</p>
              </slot>
            </td>
          </tr>
        </template>
        <template v-else>
          <tr
            v-for="(row, index) in data"
            :key="index"
            :class="rowClasses(index)"
            @click="handleRowClick(row, index)"
          >
            <td
              v-for="column in columns"
              :key="column.key"
              :class="cellClasses"
            >
              <slot
                :name="`cell-${column.key}`"
                :row="row"
                :value="row[column.key]"
              >
                {{ row[column.key] }}
              </slot>
            </td>
          </tr>
        </template>
      </tbody>
    </table>
  </div>
</template>
