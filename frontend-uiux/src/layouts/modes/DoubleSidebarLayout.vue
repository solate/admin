<script setup lang="ts">
/**
 * 双列菜单布局模式
 * 左侧第一列为图标主菜单，第二列为子菜单列表
 * 类似于 VS Code 的侧边栏布局
 */
import { computed, ref, watch, onMounted, onUnmounted, markRaw } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUiStore } from '@/stores/modules/ui'
import { useAuthStore } from '@/stores/modules/auth'
import { useTenantsStore } from '@/stores/modules/tenants'
import { usePreferencesStore } from '@/stores/modules/preferences'
import { useTagsViewStore } from '@/stores/modules/tagsView'
import { useI18n } from '@/locales/composables'
import { useLayout } from '@/composables/useLayout'
import TopNavbar from '@/components/layout/TopNavbar.vue'
import TagsView from '@/components/layout/TagsView/index.vue'
import AppFooter from '@/layouts/components/AppFooter.vue'
import {
  Box,
  Building,
  Home,
  User,
  BarChart3,
  Settings,
  Menu,
  ChevronLeft,
  X,
  Minimize2,
  ChevronRight
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const uiStore = useUiStore()
const authStore = useAuthStore()
const tenantsStore = useTenantsStore()
const preferencesStore = usePreferencesStore()
const tagsViewStore = useTagsViewStore()

// 使用布局 composable
const {
  showTabs,
  showFooter,
  footerFixed,
  showCopyright,
  contentStyle,
  isDesktop,
  isMobile
} = useLayout()

// 初始化租户
tenantsStore.initialize()

// ==================== 主题相关 ====================

// 是否深色模式
const isDarkMode = computed(() => uiStore.darkMode)

// 是否使用深色侧边栏（仅在浅色模式下生效）
const useDarkSidebar = computed(() => {
  return preferencesStore.appearance.darkSidebar && !isDarkMode.value
})

// 主侧边栏样式类
const primarySidebarClass = computed(() => {
  if (isDarkMode.value || useDarkSidebar.value) {
    return 'bg-slate-900 dark:bg-slate-950 border-slate-800'
  }
  return 'bg-white dark:bg-slate-900 border-slate-200/60 dark:border-slate-800'
})

// 主侧边栏图标样式
const primaryIconClass = computed(() => {
  if (isDarkMode.value || useDarkSidebar.value) {
    return 'text-slate-400 hover:text-white hover:bg-slate-800'
  }
  return 'text-slate-500 hover:text-slate-900 hover:bg-slate-100 dark:text-slate-400 dark:hover:text-white dark:hover:bg-slate-800'
})

// 主侧边栏边框样式
const primaryBorderClass = computed(() => {
  if (isDarkMode.value || useDarkSidebar.value) {
    return 'border-slate-800'
  }
  return 'border-slate-200/60 dark:border-slate-800'
})

// 双列菜单样式
const doubleSidebarStyle = computed(() => preferencesStore.layout.doubleSidebarStyle || 'icon-left')

// 是否为图标在左边的样式
const isIconLeftStyle = computed(() => doubleSidebarStyle.value === 'icon-left')

// ==================== 移动端菜单 ====================

const isMobileMenuOpen = ref(false)

const toggleMobileMenu = () => {
  isMobileMenuOpen.value = !isMobileMenuOpen.value
}

const closeMobileMenu = () => {
  isMobileMenuOpen.value = false
}

// ==================== 导航数据 ====================

// 主菜单（第一列图标）
const primaryNavigation = computed(() => [
  {
    name: t('nav.overview'),
    path: '/dashboard/overview',
    icon: markRaw(Home),
    key: 'overview'
  },
  {
    name: t('nav.management'),
    path: '/dashboard/tenants',
    icon: markRaw(Building),
    key: 'management'
  },
  {
    name: t('nav.analytics'),
    path: '/dashboard/analytics',
    icon: markRaw(BarChart3),
    key: 'analytics'
  },
  {
    name: t('nav.settings'),
    path: '/dashboard/settings',
    icon: markRaw(Settings),
    key: 'settings'
  }
])

// 二级菜单映射
const secondaryNavigationMap: Record<string, Array<{ name: string; path: string; icon: any; key: string }>> = {
  management: [
    { name: t('nav.tenants'), path: '/dashboard/tenants', icon: markRaw(Building), key: 'tenants' },
    { name: t('nav.services'), path: '/dashboard/services', icon: markRaw(Box), key: 'services' },
    { name: t('nav.users'), path: '/dashboard/users', icon: markRaw(User), key: 'users' }
  ]
}

// 当前选中的主菜单
const activePrimaryMenu = ref('overview')

// 当前二级菜单
const secondaryNavigation = computed(() => {
  return secondaryNavigationMap[activePrimaryMenu.value] || []
})

// 根据当前路由设置主菜单
const updateActiveMenuFromRoute = () => {
  const currentPath = route.path

  // 检查是否属于某个有二级菜单的主菜单
  for (const [key, items] of Object.entries(secondaryNavigationMap)) {
    if (items.some(item => currentPath.startsWith(item.path) || currentPath === item.path)) {
      activePrimaryMenu.value = key
      return
    }
  }

  // 否则根据路径匹配主菜单
  if (currentPath.includes('/overview')) {
    activePrimaryMenu.value = 'overview'
  } else if (currentPath.includes('/analytics')) {
    activePrimaryMenu.value = 'analytics'
  } else if (currentPath.includes('/settings')) {
    activePrimaryMenu.value = 'settings'
  }
}

// 监听路由变化
watch(() => route.path, updateActiveMenuFromRoute, { immediate: true })

// 选择主菜单
const selectPrimaryMenu = (key: string, path: string) => {
  activePrimaryMenu.value = key

  // 如果有二级菜单，导航到第一个二级菜单或保持当前位置
  const secondaryItems = secondaryNavigationMap[key]
  if (secondaryItems && secondaryItems.length > 0) {
    // 检查当前路径是否已经在该菜单的子项中
    const isInSubmenu = secondaryItems.some(item => route.path.startsWith(item.path))
    if (!isInSubmenu) {
      router.push(secondaryItems[0].path)
    }
  } else {
    // 没有二级菜单，直接导航
    router.push(path)
  }
}

const isActive = (path: string) => {
  return route.path === path || route.path.startsWith(path + '/')
}

// ==================== 全屏状态 ====================

const isFullscreen = computed(() => tagsViewStore.isMaximized)

const exitFullscreen = () => {
  tagsViewStore.setMaximized(false)
}

// ESC 键退出全屏
function handleKeyDown(e: KeyboardEvent) {
  if (e.key === 'Escape' && isFullscreen.value) {
    exitFullscreen()
  }
}

// ==================== 窗口尺寸 ====================

const windowWidth = ref(typeof window !== 'undefined' ? window.innerWidth : 1024)

const handleResize = () => {
  windowWidth.value = window.innerWidth
  if (window.innerWidth >= 1024) {
    isMobileMenuOpen.value = false
  }
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
  window.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  window.removeEventListener('keydown', handleKeyDown)
})

// 主内容区样式
const mainContentStyle = computed(() => {
  const styles: Record<string, string> = {}

  if (isDesktop.value) {
    if (isIconLeftStyle.value) {
      // 图标在左：第一列 64px + 第二列 200px (如果有二级菜单)
      const secondaryWidth = secondaryNavigation.value.length > 0 ? 200 : 0
      styles.marginLeft = `${64 + secondaryWidth}px`
    } else {
      // 文字在左：第一列 160px + 第二列 64px (如果有二级菜单)
      const primaryWidth = 160
      const secondaryWidth = secondaryNavigation.value.length > 0 ? 64 : 0
      styles.marginLeft = `${primaryWidth + secondaryWidth}px`
    }
  }

  return styles
})

// 第一列宽度
const primarySidebarWidth = computed(() => isIconLeftStyle.value ? '64px' : '160px')

// 二级菜单宽度样式
const secondarySidebarStyle = computed(() => {
  if (isIconLeftStyle.value) {
    return {
      width: '200px',
      left: '64px'
    }
  } else {
    return {
      width: '64px',
      left: '160px'
    }
  }
})
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-50 via-slate-50/80 to-blue-50/40 dark:from-slate-950 dark:via-slate-950 dark:to-slate-900/60">
    <!-- ==================== 全屏内容区 ==================== -->
    <Transition
      enter-active-class="transition-all duration-300 ease-out"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition-all duration-200 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-if="isFullscreen"
        class="fixed inset-0 z-[100] bg-gradient-to-br from-slate-50 via-slate-50/80 to-blue-50/40 dark:from-slate-950 dark:via-slate-950 dark:to-slate-900/60 overflow-auto"
      >
        <button
          @click="exitFullscreen"
          class="fixed top-4 right-4 z-[101] flex items-center gap-2 px-3 py-2 bg-white dark:bg-slate-800 rounded-lg shadow-lg border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors cursor-pointer"
        >
          <Minimize2 :size="16" class="text-slate-600 dark:text-slate-400" />
          <span class="text-sm text-slate-600 dark:text-slate-400">{{ t('tagsView.restore') }}</span>
          <kbd class="px-1.5 py-0.5 text-xs bg-slate-100 dark:bg-slate-700 rounded text-slate-500 dark:text-slate-400">ESC</kbd>
        </button>

        <div class="h-full p-4 lg:p-6">
          <router-view v-slot="{ Component }">
            <keep-alive :include="tagsViewStore.cachedTags">
              <component :is="Component" :key="route.path" />
            </keep-alive>
          </router-view>
        </div>
      </div>
    </Transition>

    <!-- ==================== 正常布局 ==================== -->
    <!-- Mobile Header -->
    <header class="lg:hidden fixed top-0 left-0 right-0 z-40 bg-white/80 dark:bg-slate-900/80 backdrop-blur-md shadow-sm">
      <div class="flex items-center justify-between px-4 h-16">
        <div class="flex items-center gap-3">
          <button
            class="p-2 -ml-2 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors cursor-pointer"
            @click="toggleMobileMenu"
          >
            <Menu :size="24" />
          </button>
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-lg bg-primary-600 flex items-center justify-center">
              <Box :size="20" class="text-white" />
            </div>
            <span class="text-lg font-semibold text-slate-900 dark:text-slate-100">AdminSystem</span>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <div class="w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center">
            <User :size="20" class="text-primary-600 dark:text-primary-400" />
          </div>
        </div>
      </div>
    </header>

    <!-- Mobile Overlay -->
    <Transition
      enter-active-class="transition-opacity duration-200"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-200"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="isMobileMenuOpen"
        class="lg:hidden fixed inset-0 bg-black/50 z-40"
        @click="closeMobileMenu"
      />
    </Transition>

    <!-- Mobile Sidebar -->
    <Transition
      enter-active-class="transition-transform duration-300"
      enter-from-class="-translate-x-full"
      enter-to-class="translate-x-0"
      leave-active-class="transition-transform duration-300"
      leave-from-class="translate-x-0"
      leave-to-class="-translate-x-full"
    >
      <aside
        v-if="isMobileMenuOpen"
        class="lg:hidden fixed inset-y-0 left-0 z-50 w-72 bg-white dark:bg-slate-900 border-r border-slate-200/80 dark:border-slate-800"
      >
        <div class="flex flex-col h-full">
          <div class="flex items-center justify-between p-4 border-b border-slate-200/60 dark:border-slate-800/40">
            <div class="flex items-center gap-2">
              <div class="w-8 h-8 rounded-lg bg-primary-600 flex items-center justify-center">
                <Box :size="20" class="text-white" />
              </div>
              <span class="text-xl font-semibold text-slate-900 dark:text-slate-100">AdminSystem</span>
            </div>
            <button
              class="p-2 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors cursor-pointer"
              @click="closeMobileMenu"
            >
              <ChevronLeft :size="20" />
            </button>
          </div>

          <nav class="flex-1 overflow-y-auto p-4 space-y-1">
            <router-link
              v-for="item in primaryNavigation"
              :key="item.key"
              :to="item.path"
              @click="closeMobileMenu"
              class="flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all cursor-pointer group"
              :class="isActive(item.path)
                ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400'
                : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700'"
            >
              <component :is="item.icon" class="w-5 h-5" />
              <span class="font-medium">{{ item.name }}</span>
            </router-link>
          </nav>

          <div class="p-4 border-t border-slate-200/60 dark:border-slate-700/30 space-y-2">
            <div class="flex items-center gap-3 p-3 bg-slate-50 dark:bg-slate-700 rounded-lg">
              <div class="w-10 h-10 rounded-full bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center">
                <User :size="20" class="text-primary-600 dark:text-primary-400" />
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-slate-900 dark:text-slate-100 truncate">
                  {{ authStore.user?.name || 'Admin User' }}
                </p>
                <p class="text-xs text-slate-500 dark:text-slate-400 truncate">
                  {{ authStore.user?.email || 'admin@example.com' }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </aside>
    </Transition>

    <!-- Desktop Primary Sidebar (Icons or Text based on style) -->
    <aside
      class="hidden lg:flex flex-col fixed inset-y-0 left-0 z-30 border-r transition-all duration-300 overflow-hidden"
      :class="primarySidebarClass"
      :style="{ width: primarySidebarWidth }"
    >
      <!-- Logo -->
      <div class="flex items-center h-16 border-b" :class="[primaryBorderClass, isIconLeftStyle ? 'justify-center' : 'px-4']">
        <div class="w-8 h-8 rounded-lg bg-primary-600 flex items-center justify-center flex-shrink-0">
          <Box :size="20" class="text-white" />
        </div>
        <span v-if="!isIconLeftStyle" class="ml-3 text-sm font-semibold text-slate-900 dark:text-slate-100 truncate">AdminSystem</span>
      </div>

      <!-- Primary Navigation -->
      <nav class="flex-1 overflow-y-auto overflow-x-hidden py-4">
        <div :class="isIconLeftStyle ? 'flex flex-col items-center space-y-2' : 'flex flex-col space-y-1 px-2'">
          <button
            v-for="item in primaryNavigation"
            :key="item.key"
            @click="selectPrimaryMenu(item.key, item.path)"
            class="relative flex items-center rounded-xl transition-all cursor-pointer group"
            :class="[
              activePrimaryMenu === item.key
                ? 'bg-primary-600 text-white'
                : primaryIconClass,
              isIconLeftStyle ? 'w-12 h-12 justify-center' : 'w-full h-10 px-3 gap-3'
            ]"
            :title="isIconLeftStyle ? item.name : ''"
          >
            <component :is="item.icon" :class="isIconLeftStyle ? 'w-5 h-5' : 'w-5 h-5 flex-shrink-0'" />

            <!-- Text for text-left style -->
            <span v-if="!isIconLeftStyle" class="text-sm font-medium truncate">{{ item.name }}</span>

            <!-- Active indicator -->
            <div
              v-if="activePrimaryMenu === item.key"
              class="absolute top-1/2 -translate-y-1/2 w-1 h-6 bg-primary-500"
              :class="isIconLeftStyle ? 'left-0 rounded-r-full' : 'right-0 rounded-l-full'"
            />

            <!-- Tooltip for icon-left style -->
            <div
              v-if="isIconLeftStyle"
              class="absolute left-full ml-2 px-2 py-1 bg-slate-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none z-50"
            >
              {{ item.name }}
            </div>
          </button>
        </div>
      </nav>

      <!-- Bottom Actions -->
      <div class="py-4 border-t" :class="primaryBorderClass">
        <div :class="isIconLeftStyle ? 'flex flex-col items-center space-y-2' : 'flex flex-col space-y-1 px-2'">
          <button
            class="flex items-center rounded-xl transition-colors cursor-pointer"
            :class="[
              activePrimaryMenu === 'settings' ? 'bg-primary-600 text-white' : primaryIconClass,
              isIconLeftStyle ? 'w-12 h-12 justify-center' : 'w-full h-10 px-3 gap-3'
            ]"
            :title="isIconLeftStyle ? '设置' : ''"
            @click="selectPrimaryMenu('settings', '/dashboard/settings')"
          >
            <Settings :size="isIconLeftStyle ? 20 : 18" />
            <span v-if="!isIconLeftStyle" class="text-sm font-medium">设置</span>
          </button>
        </div>
      </div>
    </aside>

    <!-- Desktop Secondary Sidebar (Sub-menu) -->
    <Transition
      enter-active-class="transition-all duration-200 ease-out"
      enter-from-class="opacity-0 -translate-x-4"
      enter-to-class="opacity-100 translate-x-0"
      leave-active-class="transition-all duration-150 ease-in"
      leave-from-class="opacity-100 translate-x-0"
      leave-to-class="opacity-0 -translate-x-4"
    >
      <aside
        v-if="isDesktop && secondaryNavigation.length > 0"
        class="hidden lg:flex flex-col fixed inset-y-0 z-20 bg-white dark:bg-slate-900 border-r border-slate-200/60 dark:border-slate-800/60 overflow-hidden"
        :style="secondarySidebarStyle"
      >
        <!-- Secondary Header -->
        <div class="flex items-center h-16 border-b border-slate-200/60 dark:border-slate-800/40" :class="isIconLeftStyle ? 'px-4' : 'justify-center'">
          <span v-if="isIconLeftStyle" class="text-sm font-semibold text-slate-700 dark:text-slate-300 truncate">
            {{ primaryNavigation.find(p => p.key === activePrimaryMenu)?.name }}
          </span>
          <component
            v-else
            :is="primaryNavigation.find(p => p.key === activePrimaryMenu)?.icon"
            class="w-5 h-5 text-slate-600 dark:text-slate-400"
          />
        </div>

        <!-- Secondary Navigation -->
        <nav class="flex-1 overflow-y-auto overflow-x-hidden" :class="isIconLeftStyle ? 'p-3 space-y-1' : 'py-4'">
          <div :class="isIconLeftStyle ? '' : 'flex flex-col items-center space-y-2'">
            <router-link
              v-for="item in secondaryNavigation"
              :key="item.key"
              :to="item.path"
              class="flex items-center rounded-lg transition-all cursor-pointer group relative"
              :class="[
                isActive(item.path)
                  ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400'
                  : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700',
                isIconLeftStyle ? 'gap-3 px-3 py-2.5 w-full' : 'w-12 h-12 justify-center'
              ]"
              :title="!isIconLeftStyle ? item.name : ''"
            >
              <component :is="item.icon" class="w-5 h-5 flex-shrink-0" />
              <span v-if="isIconLeftStyle" class="font-medium text-sm">{{ item.name }}</span>

              <!-- Tooltip for icon-right style -->
              <div
                v-if="!isIconLeftStyle"
                class="absolute left-full ml-2 px-2 py-1 bg-slate-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none z-50"
              >
                {{ item.name }}
              </div>
            </router-link>
          </div>
        </nav>

        <!-- Tenant Info -->
        <div v-if="tenantsStore.currentTenant && isIconLeftStyle" class="p-3 border-t border-slate-200/60 dark:border-slate-800/40">
          <div class="px-3 py-2 bg-primary-50 dark:bg-primary-900/30 rounded-lg">
            <div class="flex items-center gap-2 text-primary-700 dark:text-primary-300">
              <Building :size="14" class="flex-shrink-0" />
              <span class="text-xs font-medium truncate">
                {{ tenantsStore.currentTenant?.name || 'Default Tenant' }}
              </span>
            </div>
          </div>
        </div>
      </aside>
    </Transition>

    <!-- Main Content -->
    <main
      class="transition-all duration-300 min-h-screen pt-16 lg:pt-0"
      :style="mainContentStyle"
    >
      <!-- Top Navbar -->
      <TopNavbar />

      <!-- TagsView -->
      <TagsView v-if="showTabs" />

      <!-- Page Content -->
      <div class="p-4 lg:p-6">
        <div
          class="bg-white/70 dark:bg-slate-900/60 backdrop-blur-sm rounded-xl p-4 lg:p-6 min-h-[calc(100vh-8rem)] shadow-sm border border-slate-200/50 dark:border-slate-800/50"
          :style="contentStyle"
        >
          <router-view v-slot="{ Component }">
            <keep-alive :include="tagsViewStore.cachedTags">
              <component :is="Component" :key="route.path" />
            </keep-alive>
          </router-view>
        </div>
      </div>

      <!-- Footer -->
      <AppFooter
        v-if="showFooter"
        :fixed="footerFixed"
        :show-copyright="showCopyright"
      />
    </main>
  </div>
</template>
