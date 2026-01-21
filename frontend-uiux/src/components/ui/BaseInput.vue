<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: {
    type: [String, Number],
    default: ''
  },
  type: {
    type: String,
    default: 'text'
  },
  placeholder: {
    type: String,
    default: ''
  },
  label: {
    type: String,
    default: ''
  },
  error: {
    type: String,
    default: ''
  },
  disabled: {
    type: Boolean,
    default: false
  },
  readonly: {
    type: Boolean,
    default: false
  },
  required: {
    type: Boolean,
    default: false
  },
  fullWidth: {
    type: Boolean,
    default: false
  },
  size: {
    type: String,
    default: 'md',
    validator: (value) => ['sm', 'md', 'lg'].includes(value)
  },
  leftIcon: {
    type: Object,
    default: null
  },
  rightIcon: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:modelValue', 'focus', 'blur'])

const inputClasses = computed(() => {
  const sizes = {
    sm: 'px-3 py-1.5 text-sm',
    md: 'px-3 py-2 text-sm',
    lg: 'px-4 py-2.5 text-base'
  }

  return [
    'block w-full rounded-lg border transition-colors duration-200',
    'bg-white dark:bg-slate-800',
    'text-slate-900 dark:text-slate-100',
    'placeholder:text-slate-400 dark:placeholder:text-slate-500',
    'focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent',
    'disabled:bg-slate-100 dark:disabled:bg-slate-700 disabled:cursor-not-allowed',
    'read-only:bg-slate-50 dark:read-only:bg-slate-800/50 read-only:cursor-default',
    props.error
      ? 'border-error-500 focus:ring-error-500'
      : 'border-slate-300 dark:border-slate-600 focus:border-primary-500',
    props.leftIcon ? 'pl-10' : '',
    props.rightIcon ? 'pr-10' : '',
    sizes[props.size]
  ].filter(Boolean).join(' ')
})

const containerClasses = computed(() => {
  return props.fullWidth ? 'w-full' : ''
})

const handleInput = (event) => {
  emit('update:modelValue', event.target.value)
}
</script>

<template>
  <div :class="containerClasses">
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
        :class="inputClasses"
        @input="handleInput"
        @focus="emit('focus')"
        @blur="emit('blur')"
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
