<script setup>
import { computed } from 'vue'

const props = defineProps({
  label: {
    type: String,
    required: true
  },
  value: {
    type: [String, Number],
    required: true
  },
  unit: {
    type: String,
    default: ''
  },
  trend: {
    type: Number,
    default: null
  },
  period: {
    type: String,
    default: 'vs 上月'
  },
  color: {
    type: String,
    default: 'primary',
    validator: (value) => ['primary', 'success', 'warning', 'error', 'info'].includes(value)
  }
})

const colorClasses = computed(() => {
  const colors = {
    primary: {
      bg: 'bg-primary-100 dark:bg-primary-900/30',
      text: 'text-primary-600 dark:text-primary-400',
      border: 'border-primary-200 dark:border-primary-800'
    },
    success: {
      bg: 'bg-success-100 dark:bg-success-900/30',
      text: 'text-success-600 dark:text-success-400',
      border: 'border-success-200 dark:border-success-800'
    },
    warning: {
      bg: 'bg-warning-100 dark:bg-warning-900/30',
      text: 'text-warning-600 dark:text-warning-400',
      border: 'border-warning-200 dark:border-warning-800'
    },
    error: {
      bg: 'bg-error-100 dark:bg-error-900/30',
      text: 'text-error-600 dark:text-error-400',
      border: 'border-error-200 dark:border-error-800'
    },
    info: {
      bg: 'bg-info-100 dark:bg-info-900/30',
      text: 'text-info-600 dark:text-info-400',
      border: 'border-info-200 dark:border-info-800'
    }
  }
  return colors[props.color]
})

const trendIcon = computed(() => {
  if (props.trend === null) return null
  if (props.trend > 0) return '↑'
  if (props.trend < 0) return '↓'
  return '→'
})

const trendColor = computed(() => {
  if (props.trend === null) return 'text-slate-500 dark:text-slate-400'
  if (props.trend > 0) return 'text-success-600 dark:text-success-400'
  if (props.trend < 0) return 'text-error-600 dark:text-error-400'
  return 'text-slate-500 dark:text-slate-400'
})
</script>

<template>
  <div class="card p-5 hover:shadow-card-hover transition-all duration-200">
    <!-- Label -->
    <p class="text-sm font-medium text-slate-500 dark:text-slate-400 mb-1">
      {{ label }}
    </p>

    <!-- Value with unit -->
    <div class="flex items-baseline gap-1">
      <span class="text-2xl font-bold text-slate-900 dark:text-slate-100">
        {{ value }}
      </span>
      <span v-if="unit" class="text-sm text-slate-500 dark:text-slate-400">
        {{ unit }}
      </span>
    </div>

    <!-- Trend -->
    <div v-if="trend !== null" class="flex items-center gap-1 mt-2">
      <span :class="['text-sm font-medium', trendColor]">
        {{ trendIcon }} {{ Math.abs(trend) }}%
      </span>
      <span class="text-xs text-slate-400 dark:text-slate-500">
        {{ period }}
      </span>
    </div>

    <!-- Progress bar (optional slot) -->
    <slot name="progress">
      <div
        v-if="trend !== null"
        class="mt-3 h-1.5 bg-slate-100 dark:bg-slate-700 rounded-full overflow-hidden"
      >
        <div
          :class="['h-full transition-all duration-500', colorClasses.bg.replace('/30', '')]"
          :style="{ width: `${Math.min(Math.abs(trend), 100)}%` }"
        />
      </div>
    </slot>
  </div>
</template>
