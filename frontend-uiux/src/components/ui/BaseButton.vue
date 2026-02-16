<!--
基础按钮组件
提供统一的按钮样式和交互
-->
<script setup lang="ts">
import { computed } from 'vue'

/** 按钮变体类型 */
export type ButtonVariant = 'primary' | 'secondary' | 'cta' | 'ghost' | 'danger'

/** 按钮尺寸类型 */
export type ButtonSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl'

/** 按钮组件属性 */
export interface ButtonProps {
  variant?: ButtonVariant
  size?: ButtonSize
  disabled?: boolean
  loading?: boolean
  fullWidth?: boolean
  type?: 'button' | 'submit' | 'reset'
}

interface Emits {
  (e: 'click', event: MouseEvent): void
}

const props = withDefaults(defineProps<ButtonProps>(), {
  variant: 'primary',
  size: 'md',
  disabled: false,
  loading: false,
  fullWidth: false,
  type: 'button'
})

const emit = defineEmits<Emits>()

const VARIANT_CLASSES: Record<ButtonVariant, string> = {
  primary: 'bg-primary-600 text-white hover:bg-primary-700 focus:ring-primary-500',
  secondary: 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200 hover:bg-slate-200 dark:hover:bg-slate-600 focus:ring-slate-500',
  cta: 'gradient-cta text-white hover:opacity-90 focus:ring-cta-500',
  ghost: 'bg-transparent text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 focus:ring-slate-500',
  danger: 'bg-error-600 text-white hover:bg-error-700 focus:ring-error-500'
}

const SIZE_CLASSES: Record<ButtonSize, string> = {
  xs: 'px-2.5 py-1 text-xs font-medium',
  sm: 'px-3 py-1.5 text-sm font-medium',
  md: 'px-4 py-2 text-sm font-medium',
  lg: 'px-5 py-2.5 text-base font-medium',
  xl: 'px-6 py-3 text-base font-semibold'
}

const buttonClass = computed(() => [
  'inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-all duration-200',
  'focus:outline-none focus:ring-2 focus:ring-offset-2 dark:focus:ring-offset-slate-900',
  'disabled:opacity-50 disabled:cursor-not-allowed',
  props.fullWidth ? 'w-full' : '',
  VARIANT_CLASSES[props.variant],
  SIZE_CLASSES[props.size],
  props.loading ? 'cursor-wait' : 'cursor-pointer'
].filter(Boolean).join(' '))

function handleClick(event: MouseEvent) {
  emit('click', event)
}
</script>

<template>
  <button
    :type="type"
    :class="buttonClass"
    :disabled="disabled || loading"
    @click="handleClick"
  >
    <!-- Loading Spinner -->
    <svg
      v-if="loading"
      class="animate-spin h-4 w-4"
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
    >
      <circle
        class="opacity-25"
        cx="12"
        cy="12"
        r="10"
        stroke="currentColor"
        stroke-width="4"
      />
      <path
        class="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
      />
    </svg>
    <slot v-else />
  </button>
</template>
