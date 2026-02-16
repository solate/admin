<!--
基础模态框组件
提供弹窗交互功能，支持 v-model:open 双向绑定
-->
<script setup lang="ts">
import { computed, watch, onMounted, onBeforeUnmount, useId } from 'vue'
import { X } from 'lucide-vue-next'

/** 模态框尺寸类型 */
export type ModalSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl' | 'full'

/** 模态框组件属性 */
export interface ModalProps {
  open?: boolean
  title?: string
  size?: ModalSize
  closable?: boolean
  maskClosable?: boolean
}

interface Emits {
  (e: 'update:open', value: boolean): void
  (e: 'close'): void
  (e: 'open'): void
}

const props = withDefaults(defineProps<ModalProps>(), {
  open: false,
  title: '',
  size: 'md',
  closable: true,
  maskClosable: true
})

const emit = defineEmits<Emits>()

// 生成唯一 ID 用于无障碍属性
const titleId = useId()

const SIZE_CLASSES: Record<ModalSize, string> = {
  xs: 'max-w-xs',
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
  xl: 'max-w-xl',
  full: 'max-w-6xl'
}

const modalClass = computed(() => [
  'relative bg-white dark:bg-slate-800 rounded-2xl shadow-panel w-full',
  SIZE_CLASSES[props.size],
  'transform transition-all duration-200'
].join(' '))

function close() {
  emit('update:open', false)
  emit('close')
}

function handleMaskClick() {
  if (props.maskClosable) {
    close()
  }
}

function handleEscape(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.closable && props.open) {
    close()
  }
}

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    emit('open')
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
            :aria-labelledby="title ? titleId : undefined"
          >
            <!-- Header -->
            <div
              v-if="title || closable || $slots.header"
              class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700"
            >
              <slot name="header">
                <h3
                  v-if="title"
                  :id="titleId"
                  class="text-lg font-semibold text-slate-900 dark:text-slate-100"
                >
                  {{ title }}
                </h3>
              </slot>
              <button
                v-if="closable"
                type="button"
                class="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors cursor-pointer"
                aria-label="关闭"
                @click="close"
              >
                <X :size="20" />
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
