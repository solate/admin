<!--
基础输入框组件
提供统一的表单输入样式和交互
-->
<script setup lang="ts">
import { computed, type Component } from 'vue'

/** 输入框尺寸类型 */
export type InputSize = 'sm' | 'md' | 'lg'

/** 输入框类型 */
export type InputType = 'text' | 'password' | 'email' | 'number' | 'tel' | 'url' | 'search'

/** 输入框组件属性 */
export interface InputProps {
  modelValue?: string | number
  type?: InputType
  placeholder?: string
  label?: string
  error?: string
  disabled?: boolean
  readonly?: boolean
  required?: boolean
  fullWidth?: boolean
  size?: InputSize
  leftIcon?: Component
  rightIcon?: Component
}

interface Emits {
  (e: 'update:modelValue', value: string | number): void
  (e: 'focus', event: FocusEvent): void
  (e: 'blur', event: FocusEvent): void
}

const props = withDefaults(defineProps<InputProps>(), {
  modelValue: '',
  type: 'text',
  placeholder: '',
  label: '',
  error: '',
  disabled: false,
  readonly: false,
  required: false,
  fullWidth: false,
  size: 'md'
})

const emit = defineEmits<Emits>()

const SIZE_CLASSES: Record<InputSize, string> = {
  sm: 'px-3 py-1.5 text-sm',
  md: 'px-3 py-2 text-sm',
  lg: 'px-4 py-2.5 text-base'
}

const inputClass = computed(() => [
  'block w-full rounded-lg border transition-colors duration-200',
  'bg-white dark:bg-slate-800',
  'text-slate-900 dark:text-slate-100',
  'placeholder:text-slate-400 dark:placeholder:text-slate-500',
  'focus:outline-none focus:ring-2 focus:border-transparent',
  'disabled:bg-slate-100 dark:disabled:bg-slate-700 disabled:cursor-not-allowed',
  'read-only:bg-slate-50 dark:read-only:bg-slate-800/50 read-only:cursor-default',
  props.error
    ? 'border-error-500 focus:ring-error-500'
    : 'border-slate-300 dark:border-slate-600 focus:ring-primary-500 focus:border-primary-500',
  props.leftIcon ? 'pl-10' : '',
  props.rightIcon ? 'pr-10' : '',
  SIZE_CLASSES[props.size]
].join(' '))

function handleInput(event: Event) {
  const target = event.target as HTMLInputElement
  const value = target.value

  // number 类型时，空字符串返回空字符串而非 0
  if (props.type === 'number' && value !== '') {
    emit('update:modelValue', Number(value))
  } else {
    emit('update:modelValue', value)
  }
}

function handleFocus(event: FocusEvent) {
  emit('focus', event)
}

function handleBlur(event: FocusEvent) {
  emit('blur', event)
}
</script>

<template>
  <div :class="fullWidth ? 'w-full' : undefined">
    <label
      v-if="label"
      class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1"
    >
      {{ label }}
      <span v-if="required" class="text-error-500 ml-1">*</span>
    </label>
    <div class="relative">
      <component
        v-if="leftIcon"
        :is="leftIcon"
        class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400"
      />
      <input
        :type="type"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :readonly="readonly"
        :required="required"
        :class="inputClass"
        @input="handleInput"
        @focus="handleFocus"
        @blur="handleBlur"
      />
      <component
        v-if="rightIcon"
        :is="rightIcon"
        class="absolute right-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400"
      />
    </div>
    <p
      v-if="error"
      class="mt-1 text-sm text-error-600 dark:text-error-400"
    >
      {{ error }}
    </p>
  </div>
</template>
