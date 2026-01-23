<script setup>
import { computed } from 'vue'
import { BarChart3, TrendingUp, TrendingDown, Minus } from 'lucide-vue-next'

const props = defineProps({
  title: {
    type: String,
    required: true
  },
  value: {
    type: [String, Number],
    required: true
  },
  change: {
    type: String,
    default: null
  },
  changeType: {
    type: String,
    default: 'neutral',
    validator: (value) => ['positive', 'negative', 'neutral'].includes(value)
  },
  icon: {
    type: [Object, Function],
    default: null
  },
  loading: {
    type: Boolean,
    default: false
  },
  trend: {
    type: String,
    default: null
  }
})

const trendIcon = computed(() => {
  switch (props.changeType) {
    case 'positive':
      return TrendingUp
    case 'negative':
      return TrendingDown
    default:
      return Minus
  }
})

const trendColor = computed(() => {
  switch (props.changeType) {
    case 'positive':
      return 'text-success-600 dark:text-success-400'
    case 'negative':
      return 'text-error-600 dark:text-error-400'
    default:
      return 'text-slate-500 dark:text-slate-400'
  }
})

const iconBgColor = computed(() => {
  if (props.changeType === 'positive') return 'bg-success-100 dark:bg-success-900/30'
  if (props.changeType === 'negative') return 'bg-error-100 dark:bg-error-900/30'
  return 'bg-primary-100 dark:bg-primary-900/30'
})

const iconColor = computed(() => {
  if (props.changeType === 'positive') return 'text-success-600 dark:text-success-400'
  if (props.changeType === 'negative') return 'text-error-600 dark:text-error-400'
  return 'text-primary-600 dark:text-primary-400'
})
</script>

<template>
  <div class="card hover:shadow-card-hover transition-all duration-200 p-6">
    <div v-if="loading" class="animate-pulse">
      <div class="flex items-center gap-4">
        <div class="w-14 h-14 bg-slate-200 dark:bg-slate-700 rounded-xl flex-shrink-0" />
        <div class="flex-1 space-y-2">
          <div class="h-4 bg-slate-200 dark:bg-slate-700 rounded w-24" />
          <div class="h-7 bg-slate-200 dark:bg-slate-700 rounded w-20" />
        </div>
      </div>
    </div>

    <div v-else class="flex items-center gap-4">
      <!-- Icon -->
      <div
        v-if="icon"
        :class="['w-14 h-14 rounded-xl flex items-center justify-center flex-shrink-0', iconBgColor]"
      >
        <component :is="icon" :class="iconColor"  :size="28"  />
      </div>

      <!-- Content -->
      <div class="flex-1 min-w-0">
        <!-- Title -->
        <p class="text-sm font-medium text-slate-500 dark:text-slate-400 mb-0.5">
          {{ title }}
        </p>

        <!-- Value -->
        <p class="text-2xl font-bold text-slate-900 dark:text-slate-100 mb-1.5">
          {{ value }}
        </p>

        <!-- Change indicator -->
        <div v-if="change" class="flex items-center gap-1">
          <component :is="trendIcon" :class="['flex-shrink-0', trendColor]"  :size="16"  />
          <span :class="['text-sm font-semibold', trendColor]">
            {{ change }}
          </span>
          <span v-if="trend" class="text-sm text-slate-500 dark:text-slate-400">
            {{ trend }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
