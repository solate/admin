<script setup lang="ts">
/**
 * 应用顶栏组件
 * 支持 static/fixed/auto-hide 三种模式
 */
import { computed, ref, onMounted, onUnmounted } from 'vue'
import type { HeaderMode } from '@/types/preferences'

const props = withDefaults(defineProps<{
  /** 顶栏模式 */
  mode?: HeaderMode
  /** 顶栏高度（像素） */
  height?: number
  /** 自动隐藏阈值 */
  autoHideThreshold?: number
}>(), {
  mode: 'fixed',
  height: 60,
  autoHideThreshold: 60
})

// ==================== 自动隐藏逻辑 ====================

const isVisible = ref(true)
const lastScrollY = ref(0)

const handleScroll = () => {
  if (props.mode !== 'auto-hide') {
    isVisible.value = true
    return
  }

  const currentScrollY = window.scrollY

  if (currentScrollY > lastScrollY.value && currentScrollY > props.autoHideThreshold) {
    // 向下滚动超过阈值，隐藏顶栏
    isVisible.value = false
  } else {
    // 向上滚动或在阈值内，显示顶栏
    isVisible.value = true
  }

  lastScrollY.value = currentScrollY
}

// ==================== 计算属性 ====================

const headerClass = computed(() => [
  'app-header',
  `app-header--${props.mode}`,
  {
    'app-header--hidden': props.mode === 'auto-hide' && !isVisible.value
  }
])

const headerStyle = computed(() => ({
  height: `${props.height}px`
}))

// ==================== 生命周期 ====================

onMounted(() => {
  if (props.mode === 'auto-hide') {
    window.addEventListener('scroll', handleScroll, { passive: true })
  }
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<template>
  <header :class="headerClass" :style="headerStyle">
    <slot />
  </header>
</template>

<style scoped>
.app-header {
  width: 100%;
  box-sizing: border-box;
}

.app-header--static {
  position: relative;
}

.app-header--fixed {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 30;
}

.app-header--auto-hide {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 30;
  transition: transform 0.3s ease;
}

.app-header--auto-hide.app-header--hidden {
  transform: translateY(-100%);
}
</style>
