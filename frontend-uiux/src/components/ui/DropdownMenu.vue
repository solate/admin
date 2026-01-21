<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { ClickOutside as vClickOutside } from 'directives'

const props = defineProps({
  trigger: {
    type: String,
    default: 'click',
    validator: (value) => ['click', 'hover'].includes(value)
  },
  placement: {
    type: String,
    default: 'bottom',
    validator: (value) => ['top', 'bottom', 'left', 'right'].includes(value)
  },
  disabled: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['open', 'close'])

const isOpen = ref(false)
const triggerRef = ref(null)
const menuRef = ref(null)

const placementClasses = computed(() => {
  const placements = {
    top: 'bottom-full mb-1',
    bottom: 'top-full mt-1',
    left: 'right-full mr-1',
    right: 'left-full ml-1'
  }
  return placements[props.placement]
})

const toggle = () => {
  if (props.disabled) return
  isOpen.value = !isOpen.value
}

const open = () => {
  if (props.disabled) return
  isOpen.value = true
}

const close = () => {
  isOpen.value = false
}

const handleClickOutside = () => {
  close()
}

watch(isOpen, (value) => {
  emit(value ? 'open' : 'close')
})

defineExpose({
  toggle,
  open,
  close
})
</script>

<template>
  <div ref="triggerRef" class="relative inline-block">
    <slot
      :is-open="isOpen"
      :toggle="toggle"
      :open="open"
      :close="close"
    />

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
        ref="menuRef"
        :class="[
          'absolute z-50 min-w-[160px] max-w-xs',
          placementClasses
        ]"
        v-click-outside="handleClickOutside"
      >
        <div class="card shadow-panel py-1">
          <slot name="menu" :close="close" />
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
/* Vue 3 doesn't have built-in clickOutside, so we need to handle it differently */
</style>
