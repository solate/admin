<script setup>
import { computed, watch, onMounted, onBeforeUnmount } from 'vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: ''
  },
  size: {
    type: String,
    default: 'md',
    validator: (value) => ['xs', 'sm', 'md', 'lg', 'xl', 'full'].includes(value)
  },
  closable: {
    type: Boolean,
    default: true
  },
  maskClosable: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['close', 'open'])

const modalSizes = {
  xs: 'max-w-xs',
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
  xl: 'max-w-xl',
  full: 'max-w-6xl'
}

const modalClass = computed(() => {
  return [
    'relative bg-white dark:bg-slate-800 rounded-2xl shadow-panel w-full',
    modalSizes[props.size],
    'transform transition-all duration-200'
  ].join(' ')
})

const handleMaskClick = () => {
  if (props.maskClosable) {
    emit('close')
  }
}

const handleEscape = (e) => {
  if (e.key === 'Escape' && props.closable) {
    emit('close')
  }
}

watch(() => props.open, (isOpen) => {
  emit('open', isOpen)
  if (isOpen) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})

onMounted(() => {
  document.addEventListener('keydown', handleEscape)
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleEscape)
  document.body.style.overflow = ''
})
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-all duration-200"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-all duration-200"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="open"
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
      >
        <!-- Backdrop -->
        <div
          class="absolute inset-0 bg-black/50 backdrop-blur-sm"
          @click="handleMaskClick"
        />

        <!-- Modal -->
        <Transition
          enter-active-class="transition-all duration-200"
          enter-from-class="scale-95 opacity-0"
          enter-to-class="scale-100 opacity-100"
          leave-active-class="transition-all duration-150"
          leave-from-class="scale-100 opacity-100"
          leave-to-class="scale-95 opacity-0"
        >
          <div
            v-if="open"
            :class="modalClass"
            role="dialog"
            aria-modal="true"
          >
            <!-- Header -->
            <div
              v-if="title || closable || $slots.header"
              class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700"
            >
              <slot name="header">
                <h3 class="text-lg font-semibold text-slate-900 dark:text-slate-100">
                  {{ title }}
                </h3>
              </slot>
              <button
                v-if="closable"
                class="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors cursor-pointer"
                @click="emit('close')"
              >
                <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <!-- Body -->
            <div class="px-6 py-4">
              <slot />
            </div>

            <!-- Footer -->
            <div
              v-if="$slots.footer"
              class="flex items-center justify-end gap-3 px-6 py-4 border-t border-slate-200 dark:border-slate-700"
            >
              <slot name="footer" />
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>
