<!--
基础徽章组件
提供标签、状态指示等功能
-->
<script setup lang="ts">
import { computed } from 'vue'

/** 徽章变体类型 */
export type BadgeVariant =
  | 'default'
  | 'primary'
  | 'success'
  | 'warning'
  | 'error'
  | 'info'
  | 'outline'
  | 'outline-primary'
  | 'outline-success'
  | 'outline-warning'
  | 'outline-error'

/** 徽章尺寸类型 */
export type BadgeSize = 'xs' | 'sm' | 'md' | 'lg'

/** 徽章组件属性 */
export interface BadgeProps {
  variant?: BadgeVariant
  size?: BadgeSize
  dot?: boolean
}

const props = withDefaults(defineProps<BadgeProps>(), {
  variant: 'default',
  size: 'md',
  dot: false
})

const VARIANT_CLASSES: Record<BadgeVariant, string> = {
  default: 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300',
  primary: 'bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300',
  success: 'bg-success-100 dark:bg-success-900/30 text-success-700 dark:text-success-300',
  warning: 'bg-warning-100 dark:bg-warning-900/30 text-warning-700 dark:text-warning-300',
  error: 'bg-error-100 dark:bg-error-900/30 text-error-700 dark:text-error-300',
  info: 'bg-info-100 dark:bg-info-900/30 text-info-700 dark:text-info-300',
  outline: 'border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-300',
  'outline-primary': 'border border-primary-300 dark:border-primary-600 text-primary-700 dark:text-primary-300',
  'outline-success': 'border border-success-300 dark:border-success-600 text-success-700 dark:text-success-300',
  'outline-warning': 'border border-warning-300 dark:border-warning-600 text-warning-700 dark:text-warning-300',
  'outline-error': 'border border-error-300 dark:border-error-600 text-error-700 dark:text-error-300'
}

const SIZE_CLASSES: Record<BadgeSize, string> = {
  xs: 'px-2 py-0.5 text-xs font-medium',
  sm: 'px-2 py-1 text-xs font-medium',
  md: 'px-2.5 py-1 text-sm font-medium',
  lg: 'px-3 py-1.5 text-sm font-medium'
}

const DOT_SIZE_CLASSES: Record<BadgeSize, string> = {
  xs: 'w-1.5 h-1.5',
  sm: 'w-2 h-2',
  md: 'w-2 h-2',
  lg: 'w-2.5 h-2.5'
}

const badgeClass = computed(() => [
  'inline-flex items-center gap-1.5 rounded-full font-medium',
  VARIANT_CLASSES[props.variant],
  SIZE_CLASSES[props.size]
].join(' '))

const dotClass = computed(() => [
  'rounded-full flex-shrink-0',
  DOT_SIZE_CLASSES[props.size],
  props.variant.includes('outline') ? '' : 'bg-current opacity-75'
].join(' '))
</script>

<template>
  <span :class="badgeClass">
    <span v-if="dot" :class="dotClass" />
    <slot />
  </span>
</template>
