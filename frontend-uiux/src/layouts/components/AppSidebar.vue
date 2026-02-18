<script setup lang="ts">
/**
 * 应用侧边栏组件
 * 支持折叠/展开、手风琴模式
 */
import { computed, ref, watch, markRaw } from 'vue'
import { useRoute } from 'vue-router'
import { useUiStore } from '@/stores/modules/ui'
import { useAuthStore } from '@/stores/modules/auth'
import { useTenantsStore } from '@/stores/modules/tenants'
import { usePreferencesStore } from '@/stores/modules/preferences'
import { useI18n } from '@/locales/composables'
import {
  Box,
  Building,
  ChevronLeft
} from 'lucide-vue-next'
import type { NavStyle } from '@/types/preferences'

const props = withDefaults(defineProps<{
  /** 侧边栏宽度（像素） */
  width?: number
  /** 折叠宽度（像素） */
  collapsedWidth?: number
  /** 是否折叠 */
  collapsed?: boolean
  /** 导航样式 */
  navStyle?: NavStyle
  /** 手风琴模式 */
  navAccordion?: boolean
}>(), {
  width: 256,
  collapsedWidth: 64,
  collapsed: false,
  navStyle: 'icon-text',
  navAccordion: true
})

const route = useRoute()
const { t } = useI18n()
const uiStore = useUiStore()
const authStore = useAuthStore()
const tenantsStore = useTenantsStore()
const preferencesStore = usePreferencesStore()

// ==================== 导航数据 ====================

interface NavigationItem {
  name: string
  path: string
  icon: unknown
  key: string
  children?: NavigationItem[]
}

const navigation = computed((): NavigationItem[] => {
  const baseNav: NavigationItem[] = [
    {
      name: t('nav.overview'),
      path: '/dashboard/overview',
      icon: markRaw(Box),
      key: 'overview'
    },
    {
      name: t('nav.tenants'),
      path: '/dashboard/tenants',
      icon: markRaw(Building),
      key: 'tenants'
    }
  ]

  const userRole = authStore.userRole
  return baseNav.filter(item => {
    if (item.key === 'tenants' && !['super_admin', 'auditor'].includes(userRole)) {
      return false
    }
    return true
  })
})

const bottomNavigation = computed((): NavigationItem[] => [])

// ==================== 手风琴模式 ====================

/** 展开的菜单项 */
const expandedMenus = ref<string[]>([])

const toggleMenu = (key: string) => {
  if (props.navAccordion) {
    // 手风琴模式：关闭其他，切换当前
    expandedMenus.value = expandedMenus.value.includes(key) ? [] : [key]
  } else {
    // 非手风琴模式：独立切换
    const index = expandedMenus.value.indexOf(key)
    if (index > -1) {
      expandedMenus.value.splice(index, 1)
    } else {
      expandedMenus.value.push(key)
    }
  }
}

const isMenuExpanded = (key: string) => expandedMenus.value.includes(key)

// ==================== 激活状态 ====================

const isActive = (path: string) => {
  return route.path === path || route.path.startsWith(path + '/')
}

// ==================== 计算属性 ====================

const isIconOnly = computed(() => props.navStyle === 'icon-only' || props.collapsed)

const sidebarWidthPx = computed(() =>
  props.collapsed ? props.collapsedWidth : props.width
)

const sidebarStyle = computed(() => ({
  width: `${sidebarWidthPx.value}px`
}))

// 外观设置
const appearancePrefs = computed(() => preferencesStore.appearance)

// 深色侧边栏样式（仅在浅色模式下生效）
const darkSidebarClass = computed(() => {
  if (appearancePrefs.value.darkSidebar && !uiStore.darkMode) {
    return 'dark-sidebar'
  }
  return ''
})

// ==================== 监听路由变化关闭展开的菜单 ====================

watch(
  () => route.path,
  () => {
    // 路由变化时，如果启用手风琴模式，关闭所有展开的菜单
    if (props.navAccordion) {
      expandedMenus.value = []
    }
  }
)
</script>

<template>
  <aside
    :class="[
      'app-sidebar hidden lg:flex flex-col fixed inset-y-0 left-0 transition-all duration-300 z-30',
      darkSidebarClass,
      'bg-white/90 dark:bg-slate-900/90',
      'backdrop-blur-xl',
      'border-r border-slate-200/60 dark:border-slate-800/60',
      'shadow-[4px_0_24px_-8px_rgba(0,0,0,0.06)] dark:shadow-[4px_0_24px_-8px_rgba(0,0,0,0.3)]',
      { 'app-sidebar--collapsed': collapsed }
    ]"
    :style="sidebarStyle"
  >
    <!-- Logo -->
    <div class="flex items-center h-16 px-4 relative after:content-[''] after:absolute after:inset-x-4 after:-bottom-px after:h-[1px] after:bg-gradient-to-r after:from-transparent after:via-slate-300 dark:after:via-slate-600 after:to-transparent after:opacity-50">
      <div class="flex items-center gap-3 flex-1">
        <div class="w-8 h-8 rounded-lg bg-primary-600 flex items-center justify-center flex-shrink-0">
          <Box :size="20" class="text-white" />
        </div>
        <Transition
          enter-active-class="transition-all duration-200"
          enter-from-class="opacity-0 w-0"
          enter-to-class="opacity-100 w-auto"
          leave-active-class="transition-all duration-200"
          leave-from-class="opacity-100 w-auto"
          leave-to-class="opacity-0 w-0"
        >
          <span
            v-if="!isIconOnly"
            class="text-lg font-semibold text-slate-900 dark:text-slate-100 whitespace-nowrap overflow-hidden"
          >
            AdminSystem
          </span>
        </Transition>
      </div>
    </div>

    <!-- 租户信息 -->
    <Transition
      enter-active-class="transition-all duration-200"
      enter-from-class="opacity-0 h-0 mt-0 mx-0"
      enter-to-class="opacity-100 h-auto mt-4 mx-4"
      leave-active-class="transition-all duration-200"
      leave-from-class="opacity-100 h-auto mt-4 mx-4"
      leave-to-class="opacity-0 h-0 mt-0 mx-0"
    >
      <div
        v-if="!isIconOnly && tenantsStore.currentTenant"
        class="px-4 py-3 bg-primary-50 dark:bg-primary-900/30 rounded-lg overflow-hidden mx-4 mt-4"
      >
        <div class="flex items-center gap-2 text-primary-700 dark:text-primary-300">
          <Building :size="16" class="flex-shrink-0" />
          <span class="text-sm font-medium truncate">
            {{ tenantsStore.currentTenant?.name || 'Default Tenant' }}
          </span>
        </div>
        <p class="text-xs text-primary-600/70 dark:text-primary-400/70 mt-1 capitalize">
          {{ tenantsStore.currentTenant?.status || 'active' }}
        </p>
      </div>
    </Transition>

    <!-- 导航菜单 -->
    <nav class="flex-1 overflow-y-auto p-3 space-y-1">
      <slot
        name="navigation"
        :navigation="navigation"
        :is-active="isActive"
        :is-icon-only="isIconOnly"
        :expanded-menus="expandedMenus"
        :toggle-menu="toggleMenu"
        :is-menu-expanded="isMenuExpanded"
      >
        <!-- 默认导航项渲染 -->
        <router-link
          v-for="item in navigation"
          :key="item.key"
          :to="item.path"
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all cursor-pointer group"
          :class="isActive(item.path)
            ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400'
            : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700'"
          :title="isIconOnly ? item.name : ''"
        >
          <component :is="item.icon" class="w-5 h-5 flex-shrink-0" />
          <Transition
            enter-active-class="transition-all duration-200"
            enter-from-class="opacity-0 w-0"
            enter-to-class="opacity-100 w-auto"
            leave-active-class="transition-all duration-200"
            leave-from-class="opacity-100 w-auto"
            leave-to-class="opacity-0 w-0"
          >
            <span
              v-if="!isIconOnly"
              class="font-medium whitespace-nowrap overflow-hidden"
            >
              {{ item.name }}
            </span>
          </Transition>
        </router-link>
      </slot>

      <!-- 底部导航分隔线 -->
      <div
        v-if="bottomNavigation.length > 0"
        class="border-t border-slate-200/60 dark:border-slate-700/30"
        :class="!isIconOnly ? 'pt-4 mt-4' : 'pt-3 mt-3'"
      >
        <slot
          name="bottom-navigation"
          :navigation="bottomNavigation"
          :is-active="isActive"
          :is-icon-only="isIconOnly"
        >
          <router-link
            v-for="item in bottomNavigation"
            :key="item.key"
            :to="item.path"
            class="flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all cursor-pointer"
            :class="isActive(item.path)
              ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400'
              : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700'"
            :title="isIconOnly ? item.name : ''"
          >
            <component :is="item.icon" class="w-5 h-5 flex-shrink-0" />
            <Transition
              enter-active-class="transition-all duration-200"
              enter-from-class="opacity-0 w-0"
              enter-to-class="opacity-100 w-auto"
              leave-active-class="transition-all duration-200"
              leave-from-class="opacity-100 w-auto"
              leave-to-class="opacity-0 w-0"
            >
              <span
                v-if="!isIconOnly"
                class="font-medium whitespace-nowrap overflow-hidden"
              >
                {{ item.name }}
              </span>
            </Transition>
          </router-link>
        </slot>
      </div>
    </nav>

    <!-- 折叠按钮 -->
    <div
      v-if="$slots.collapse"
      class="p-3 border-t border-slate-200/60 dark:border-slate-700/30"
    >
      <slot name="collapse" :collapsed="collapsed" />
    </div>
  </aside>
</template>

<style scoped>
.app-sidebar {
  box-sizing: border-box;
}

.app-sidebar--collapsed .sidebar-text {
  opacity: 0;
  width: 0;
  overflow: hidden;
}

/* 深色侧边栏（浅色模式下） */
[data-dark-sidebar="true"] .dark-sidebar {
  background-color: rgb(15 23 42 / 0.95) !important;
}

[data-dark-sidebar="true"] .dark-sidebar * {
  color: rgb(203 213 225 / 1);
}
</style>
