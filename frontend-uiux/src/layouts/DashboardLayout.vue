<script setup lang="ts">
/**
 * 仪表板布局
 * 根据用户偏好设置动态切换布局模式
 * 支持四种布局模式：sidebar、topbar、mixed、horizontal
 */
import { computed, defineAsyncComponent, watch } from 'vue'
import { usePreferencesStore } from '@/stores/modules/preferences'
import { useUiStore } from '@/stores/modules/ui'
import { useLayout } from '@/composables/useLayout'
import type { LayoutMode } from '@/types/preferences'

const preferencesStore = usePreferencesStore()
const uiStore = useUiStore()

// 使用布局 composable
const { layoutMode, generalPrefs } = useLayout()

// 布局组件映射 - 使用异步组件懒加载
const layoutComponents: Record<LayoutMode, ReturnType<typeof defineAsyncComponent>> = {
  sidebar: defineAsyncComponent(() => import('./modes/SidebarLayout.vue')),
  topbar: defineAsyncComponent(() => import('./modes/TopbarLayout.vue')),
  mixed: defineAsyncComponent(() => import('./modes/MixedLayout.vue')),
  horizontal: defineAsyncComponent(() => import('./modes/HorizontalLayout.vue')),
  'double-sidebar': defineAsyncComponent(() => import('./modes/DoubleSidebarLayout.vue'))
}

// 当前布局组件
const currentLayoutComponent = computed(() => layoutComponents[layoutMode.value])

// 动画控制
const animationClasses = computed(() => {
  if (!generalPrefs.value.enableAnimations) {
    return 'no-animation'
  }
  return ''
})

// 初始化时同步侧边栏默认折叠状态
watch(
  () => preferencesStore.layout.sidebarCollapsed,
  (collapsed) => {
    uiStore.setSidebarOpen(!collapsed)
  },
  { immediate: true }
)
</script>

<template>
  <div :class="animationClasses">
    <!-- 布局切换过渡 -->
    <Transition
      mode="out-in"
      enter-active-class="transition-opacity duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <component :is="currentLayoutComponent" :key="layoutMode" />
    </Transition>
  </div>
</template>
