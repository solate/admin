<script setup>
import { computed } from 'vue'

const props = defineProps({
  variant: {
    type: String,
    default: 'default',
    validator: (value) => ['default', 'elevated', 'borderless', 'glass'].includes(value)
  },
  padding: {
    type: String,
    default: 'md',
    validator: (value) => ['none', 'sm', 'md', 'lg', 'xl'].includes(value)
  },
  hoverable: {
    type: Boolean,
    default: false
  },
  clickable: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['click'])

const classes = computed(() => {
  const variants = {
    default: 'bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 shadow-card',
    elevated: 'bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 shadow-panel',
    borderless: 'bg-white dark:bg-slate-800 shadow-none',
    glass: 'glass-card'
  }

  const paddings = {
    none: '',
    sm: 'p-4',
    md: 'p-5',
    lg: 'p-6',
    xl: 'p-8'
  }

  return [
    'rounded-xl transition-all duration-200',
    variants[props.variant],
    paddings[props.padding],
    props.hoverable ? 'hover:shadow-card-hover hover:border-slate-300 dark:hover:border-slate-600' : '',
    props.clickable ? 'cursor-pointer' : ''
  ].filter(Boolean).join(' ')
})

const handleClick = () => {
  if (props.clickable) {
    emit('click')
  }
}
</script>

<template>
  <div :class="classes" @click="handleClick">
    <slot />
  </div>
</template>
