<script setup lang="ts">
/**
 * 顶部导航布局模式
 * 所有导航菜单在顶部水平排列，适合菜单项较少的系统
 */
import { computed, ref, onMounted, onUnmounted, markRaw } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUiStore } from '@/stores/modules/ui'
import { useAuthStore } from '@/stores/modules/auth'
import { usePreferencesStore } from '@/stores/modules/preferences'
import { useTagsViewStore } from '@/stores/modules/tagsView'
import { useI18n } from '@/locales/composables'
import { useLayout } from '@/composables/useLayout'
import AppBreadcrumbs from '@/layouts/components/AppBreadcrumbs.vue'
import TagsView from '@/components/layout/TagsView/index.vue'
import AppFooter from '@/layouts/components/AppFooter.vue'
import LanguageSwitcher from '@/components/language/LanguageSwitcher.vue'
import {
  Box,
  Building,
  Home,
  User,
  BarChart3,
  Settings,
  Sun,
  Moon,
  Search,
  Bell,
  ChevronDown,
  LogOut,
  Shield,
  Mail,
  Minimize2,
  Menu,
  X
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const uiStore = useUiStore()
const authStore = useAuthStore()
const preferencesStore = usePreferencesStore()
const tagsViewStore = useTagsViewStore()

// 使用布局 composable
const {
  headerHeightPx,
  headerMode,
  showBreadcrumbs,
  showTabs,
  showFooter,
  footerFixed,
  showCopyright,
  contentStyle,
  isDesktop,
  showSettingsDrawer,
  showSearchDialog
} = useLayout()

// ==================== 导航数据 ====================

const navigation = computed(() => {
  const baseNav = [
    { name: t('nav.overview'), path: '/dashboard/overview', icon: markRaw(Home), key: 'overview' },
    { name: t('nav.tenants'), path: '/dashboard/tenants', icon: markRaw(Building), key: 'tenants' },
    { name: t('nav.services'), path: '/dashboard/services', icon: markRaw(Box), key: 'services' },
    { name: t('nav.users'), path: '/dashboard/users', icon: markRaw(User), key: 'users' },
    { name: t('nav.analytics'), path: '/dashboard/analytics', icon: markRaw(BarChart3), key: 'analytics' },
    { name: t('nav.settings'), path: '/dashboard/settings', icon: markRaw(Settings), key: 'settings' }
  ]

  const userRole = authStore.userRole
  return baseNav.filter(item => {
    if (item.key === 'tenants' && !['super_admin', 'auditor'].includes(userRole)) {
      return false
    }
    return true
  })
})

const isActive = (path: string) => {
  return route.path === path || route.path.startsWith(path + '/')
}

// ==================== 计算属性 ====================

const generalPrefs = computed(() => preferencesStore.general)
const layoutPrefs = computed(() => preferencesStore.layout)

// 动画控制
const animationClasses = computed(() => {
  if (!generalPrefs.value.enableAnimations) {
    return 'no-animation'
  }
  return ''
})

// 主内容区样式
const mainContentStyle = computed(() => {
  const styles: Record<string, string> = {
    paddingTop: headerHeightPx.value
  }
  return styles
})

// 全屏状态
const isFullscreen = computed(() => tagsViewStore.isMaximized)

// 退出全屏
const exitFullscreen = () => {
  tagsViewStore.setMaximized(false)
}

// ==================== 用户菜单 ====================

const showUserMenu = ref(false)

const userInitials = computed(() => {
  const name = authStore.user?.name || 'Admin'
  return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2)
})

const userRoleLabel = computed(() => t(`user.roles.${authStore.userRole}`))

// ==================== 移动端菜单 ====================

const isMobileMenuOpen = ref(false)

const toggleMobileMenu = () => {
  isMobileMenuOpen.value = !isMobileMenuOpen.value
}

// ==================== 键盘事件 ====================

function handleKeyDown(e: KeyboardEvent) {
  if (e.key === 'Escape' && isFullscreen.value) {
    exitFullscreen()
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
})
</script>

<template>
  <div :class="['min-h-screen bg-gradient-to-br from-slate-50 via-slate-50/80 to-blue-50/40 dark:from-slate-950 dark:via-slate-950 dark:to-slate-900/60', animationClasses]">
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
    <!-- 顶部导航栏 -->
    <header
      :class="[
        'fixed top-0 left-0 right-0 z-30 backdrop-blur-xl transition-colors',
        'bg-white/90 dark:bg-slate-900/90',
        'border-b border-slate-200/60 dark:border-slate-700/60'
      ]"
      :style="{ height: headerHeightPx }"
    >
      <div class="h-full flex items-center justify-between px-4 lg:px-6">
        <!-- 左侧：Logo 和导航 -->
        <div class="flex items-center gap-6 h-full">
          <!-- Logo -->
          <div class="flex items-center gap-2 flex-shrink-0">
            <div class="w-8 h-8 rounded-lg bg-primary-600 flex items-center justify-center">
              <Box :size="20" class="text-white" />
            </div>
            <span class="hidden lg:block text-lg font-semibold text-slate-900 dark:text-slate-100">AdminSystem</span>
          </div>

          <!-- 水平导航 - 桌面端 -->
          <nav class="hidden lg:flex items-center gap-1 h-full">
            <router-link
              v-for="item in navigation"
              :key="item.key"
              :to="item.path"
              class="flex items-center gap-2 px-4 h-10 rounded-lg text-sm font-medium transition-all cursor-pointer"
              :class="isActive(item.path)
                ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400'
                : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800'"
            >
              <component :is="item.icon" :size="18" />
              <span>{{ item.name }}</span>
            </router-link>
          </nav>
        </div>

        <!-- 右侧：工具栏 -->
        <div class="flex items-center gap-2">
          <!-- 搜索 -->
          <button
            class="hidden md:flex items-center gap-2 px-3 py-1.5 bg-slate-100 dark:bg-slate-800 rounded-full text-sm text-slate-500 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors cursor-pointer"
            @click="showSearchDialog = true"
          >
            <Search :size="16" />
            <span class="hidden lg:inline">{{ t('common.search') }}</span>
            <kbd class="hidden lg:inline-flex items-center gap-0.5 px-1.5 py-0.5 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded text-xs">
              ⌘K
            </kbd>
          </button>

          <!-- 设置 -->
          <button
            class="p-2 text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-colors cursor-pointer"
            @click="showSettingsDrawer = true"
          >
            <Settings :size="20" />
          </button>

          <!-- 深色模式切换 -->
          <button
            class="p-2 text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-colors cursor-pointer"
            @click="uiStore.toggleDarkMode()"
          >
            <Sun v-if="uiStore.darkMode" :size="20" />
            <Moon v-else :size="20" />
          </button>

          <!-- 语言切换 -->
          <LanguageSwitcher />

          <!-- 通知 -->
          <button class="relative p-2 text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-colors cursor-pointer">
            <Bell :size="20" />
            <span class="absolute top-1 right-1 w-2.5 h-2.5 bg-red-500 rounded-full" />
          </button>

          <!-- 用户菜单 -->
          <div class="relative">
            <button
              class="flex items-center gap-2 p-1.5 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
              @click="showUserMenu = !showUserMenu"
            >
              <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-primary-500 to-primary-600 flex items-center justify-center">
                <span class="text-sm font-bold text-white">{{ userInitials }}</span>
              </div>
              <ChevronDown :size="16" class="hidden lg:block text-slate-400" :class="{ 'rotate-180': showUserMenu }" />
            </button>

            <!-- 用户下拉菜单 -->
            <Transition
              enter-active-class="transition-all duration-200 ease-out"
              enter-from-class="opacity-0 scale-95 translate-y-[-10px]"
              enter-to-class="opacity-100 scale-100 translate-y-0"
              leave-active-class="transition-all duration-150 ease-in"
              leave-from-class="opacity-100 scale-100 translate-y-0"
              leave-to-class="opacity-0 scale-95 translate-y-[-10px]"
            >
              <div
                v-if="showUserMenu"
                class="absolute right-0 mt-2 w-64 bg-white dark:bg-slate-800 rounded-xl shadow-xl border border-slate-200 dark:border-slate-700 overflow-hidden z-50"
              >
                <div class="p-4 border-b border-slate-200 dark:border-slate-700">
                  <p class="font-semibold text-slate-900 dark:text-slate-100">{{ authStore.user?.name || 'Admin' }}</p>
                  <p class="text-sm text-slate-500 dark:text-slate-400">{{ authStore.user?.email }}</p>
                </div>
                <div class="p-2">
                  <router-link
                    to="/dashboard/profile"
                    class="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 cursor-pointer"
                    @click="showUserMenu = false"
                  >
                    <User :size="18" />
                    <span>{{ t('userMenu.profile') }}</span>
                  </router-link>
                  <button
                    class="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 text-red-600 dark:text-red-400 cursor-pointer"
                    @click="authStore.logout(); showUserMenu = false"
                  >
                    <LogOut :size="18" />
                    <span>{{ t('userMenu.logout') }}</span>
                  </button>
                </div>
              </div>
            </Transition>

            <!-- 遮罩 -->
            <div v-if="showUserMenu" class="fixed inset-0 z-40" @click="showUserMenu = false" />
          </div>

          <!-- 移动端菜单按钮 -->
          <button
            class="lg:hidden p-2 text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-colors cursor-pointer"
            @click="toggleMobileMenu"
          >
            <Menu v-if="!isMobileMenuOpen" :size="20" />
            <X v-else :size="20" />
          </button>
        </div>
      </div>

      <!-- 移动端导航菜单 -->
      <Transition
        enter-active-class="transition-all duration-200 ease-out"
        enter-from-class="opacity-0 max-h-0"
        enter-to-class="opacity-100 max-h-96"
        leave-active-class="transition-all duration-150 ease-in"
        leave-from-class="opacity-100 max-h-96"
        leave-to-class="opacity-0 max-h-0"
      >
        <div
          v-if="isMobileMenuOpen"
          class="lg:hidden border-t border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900"
        >
          <nav class="p-4 space-y-1">
            <router-link
              v-for="item in navigation"
              :key="item.key"
              :to="item.path"
              class="flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-medium transition-all cursor-pointer"
              :class="isActive(item.path)
                ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400'
                : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800'"
              @click="isMobileMenuOpen = false"
            >
              <component :is="item.icon" :size="20" />
              <span>{{ item.name }}</span>
            </router-link>
          </nav>
        </div>
      </Transition>
    </header>

    <!-- 主内容区 -->
    <main class="transition-all duration-300" :style="mainContentStyle">
      <!-- 面包屑 -->
      <div v-if="showBreadcrumbs" class="px-4 lg:px-6 py-3 bg-white/50 dark:bg-slate-900/50 border-b border-slate-200/60 dark:border-slate-700/60">
        <AppBreadcrumbs />
      </div>

      <!-- 标签页 -->
      <TagsView v-if="showTabs" />

      <!-- 页面内容 -->
      <div class="p-4 lg:p-6">
        <div
          class="bg-white/70 dark:bg-slate-900/60 backdrop-blur-sm rounded-xl p-4 lg:p-6 min-h-[calc(100vh-12rem)] shadow-sm border border-slate-200/50 dark:border-slate-800/50"
          :style="contentStyle"
        >
          <router-view v-slot="{ Component }">
            <keep-alive :include="tagsViewStore.cachedTags">
              <component :is="Component" :key="route.path" />
            </keep-alive>
          </router-view>
        </div>
      </div>

      <!-- 页脚 -->
      <AppFooter
        v-if="showFooter"
        :fixed="footerFixed"
        :show-copyright="showCopyright"
      />
    </main>
  </div>
</template>
