<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  modelValue: {
    type: [String, Number, Object],
    default: ''
  },
  options: {
    type: Array,
    required: true
  },
  labelKey: {
    type: String,
    default: 'label'
  },
  valueKey: {
    type: String,
    default: 'value'
  },
  placeholder: {
    type: String,
    default: '请选择'
  },
  label: {
    type: String,
    default: ''
  },
  disabled: {
    type: Boolean,
    default: false
  },
  required: {
    type: Boolean,
    default: false
  },
  error: {
    type: String,
    default: ''
  },
  size: {
    type: String,
    default: 'md',
    validator: (value) => ['sm', 'md', 'lg'].includes(value)
  }
})

const emit = defineEmits(['update:modelValue'])

const isOpen = ref(false)

const selectedOption = computed(() => {
  if (!props.modelValue) return null
  return props.options.find(option => {
    const value = typeof option === 'object' ? option[props.valueKey] : option
    return value === props.modelValue
  })
})

const displayValue = computed(() => {
  if (!selectedOption.value) return props.placeholder
  return typeof selectedOption.value === 'object'
    ? selectedOption.value[props.labelKey]
    : selectedOption.value
})

const selectOption = (option) => {
  const value = typeof option === 'object' ? option[props.valueKey] : option
  emit('update:modelValue', value)
  isOpen.value = false
}

const sizes = {
  sm: 'px-3 py-1.5 text-sm',
  md: 'px-3 py-2 text-sm',
  lg: 'px-4 py-2.5 text-base'
}
</script>

<template>
  <div class="relative w-full">
    <label
      v-if="label"
      class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1"
    >
      {{ label }}
      <span v-if="required" class="text-error-500 ml-1">*</span>
    </label>

    <button
      type="button"
      :class="[
        'relative w-full flex items-center justify-between',
        'rounded-lg border transition-colors duration-200',
        'bg-white dark:bg-slate-800',
        'text-slate-900 dark:text-slate-100',
        'focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent',
        'disabled:bg-slate-100 dark:disabled:bg-slate-700 disabled:cursor-not-allowed',
        error ? 'border-error-500' : 'border-slate-300 dark:border-slate-600',
        sizes[size]
      ]"
      :disabled="disabled"
      @click="isOpen = !isOpen"
    >
      <span :class="displayValue === placeholder ? 'text-slate-400 dark:text-slate-500' : ''">
        {{ displayValue }}
      </span>
      <svg
        :class="['w-5 h-5 text-slate-400 transition-transform duration-200', isOpen && 'rotate-180']"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>

    <p
      v-if="error"
      class="mt-1 text-sm text-error-600 dark:text-error-400"
    >
      {{ error }}
    </p>

    <!-- Dropdown -->
    <Transition
      enter-active-class="transition-all duration-150"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition-all duration-150"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-if="isOpen"
        class="absolute z-50 w-full mt-1 max-h-60 overflow-auto rounded-lg shadow-panel bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700"
        v-click-outside="() => isOpen = false"
      >
        <div class="py-1">
          <button
            v-for="(option, index) in options"
            :key="index"
            type="button"
            :class="[
              'w-full px-3 py-2 text-left text-sm transition-colors duration-150',
              'hover:bg-slate-100 dark:hover:bg-slate-700',
              'focus:outline-none focus:bg-slate-100 dark:focus:bg-slate-700',
              selectedOption === option ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300' : 'text-slate-700 dark:text-slate-200'
            ]"
            @click="selectOption(option)"
          >
            <slot name="option" :option="option">
              {{ typeof option === 'object' ? option[labelKey] : option }}
            </slot>
          </button>

          <div v-if="options.length === 0" class="px-3 py-4 text-center text-sm text-slate-500 dark:text-slate-400">
            暂无选项
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>
