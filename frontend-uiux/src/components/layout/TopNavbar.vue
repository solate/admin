<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useUiStore } from '@/stores/modules/ui'
import { useAuthStore } from '@/stores/modules/auth'
import {
  Search,
  Bell,
  Sun,
  Moon,
  Settings,
  User,
  ChevronDown,
  LogOut,
  PanelLeftClose,
  PanelLeftOpen,
  RotateCw
} from 'lucide-vue-next'
import LanguageSwitcher from '@/components/language/LanguageSwitcher.vue'
import SearchDialog from '@/components/search/SearchDialog.vue'

const route = useRoute()
const uiStore = useUiStore()
const authStore = useAuthStore()

const searchQuery = ref('')
const showUserMenu = ref(false)
const showSearchDialog = ref(false)
const isRefreshing = ref(false)

// 刷新页面
const refreshPage = () => {
  isRefreshing.value = true
  // 使用 window.location.reload() 刷新页面
  setTimeout(() => {
    window.location.reload()
  }, 150)
}

// Fullscreen functionality
const isFullscreen = ref(false)

const toggleFullscreen = () => {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen().catch(err => {
      console.error('进入全屏失败:', err)
    })
  } else {
    document.exitFullscreen()
  }
}

const updateFullscreenState = () => {
  isFullscreen.value = !!document.fullscreenElement
}

onMounted(() => {
  document.addEventListener('fullscreenchange', updateFullscreenState)
  updateFullscreenState()
})

onUnmounted(() => {
  document.removeEventListener('fullscreenchange', updateFullscreenState)
})

const pageTitle = computed(() => {
  return route.meta?.title || '概览'
})

const breadcrumbs = computed(() => {
  const pathSegments = route.path.split('/').filter(Boolean)
  const crumbs = []

  if (pathSegments[0] === 'dashboard') {
    const labels = {
      dashboard: '仪表盘',
      overview: '概览',
      tenants: '租户管理',
      services: '服务管理',
      users: '用户管理',
      analytics: '数据分析',
      settings: '设置',
      profile: '个人资料',
      notifications: '通知'
    }

    let currentPath = ''
    for (let i = 0; i < pathSegments.length; i++) {
      currentPath += `/${pathSegments[i]}`
      crumbs.push({
        label: labels[pathSegments[i]] || pathSegments[i],
        path: currentPath
      })
    }
  }

  return crumbs
})

const notificationCount = computed(() => 5)
</script>

<style scoped>
/* 铃铛摇摆动画 */
@keyframes bell-ring {
  0% { transform: rotate(0deg); }
  15% { transform: rotate(15deg); }
  30% { transform: rotate(-12deg); }
  45% { transform: rotate(10deg); }
  60% { transform: rotate(-8deg); }
  75% { transform: rotate(5deg); }
  90% { transform: rotate(-2deg); }
  100% { transform: rotate(0deg); }
}

.bell-button:hover svg {
  animation: bell-ring 0.6s ease-in-out;
  transform-origin: top center;
}
</style>

<template>
  <header class="sticky top-0 z-20 bg-white/80 dark:bg-slate-800/80 backdrop-blur-lg border-b border-slate-200 dark:border-slate-700">
    <div class="flex items-center justify-between h-16 px-4 lg:px-6">
      <!-- Left: Sidebar Toggle, Refresh & Breadcrumbs -->
      <div class="hidden sm:flex items-center gap-2 flex-1">
        <!-- Sidebar Toggle -->
        <button
          class="hidden lg:flex items-center justify-center p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
          :aria-label="uiStore.sidebarOpen ? '收起侧边栏' : '展开侧边栏'"
          @click="uiStore.toggleSidebar()"
        >
          <PanelLeftClose v-if="uiStore.sidebarOpen" :size="18" class="text-slate-600 dark:text-slate-400" />
          <PanelLeftOpen v-else :size="18" class="text-slate-600 dark:text-slate-400" />
        </button>

        <!-- Refresh -->
        <button
          class="flex items-center justify-center p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
          :aria-label="'刷新'"
          :disabled="isRefreshing"
          @click="refreshPage"
        >
          <RotateCw :size="18" class="text-slate-600 dark:text-slate-400" :class="{ 'animate-spin': isRefreshing }" />
        </button>

        <!-- Breadcrumbs -->
        <nav class="flex items-center gap-2">
          <template v-for="(crumb, index) in breadcrumbs" :key="crumb.path">
            <router-link
              :to="crumb.path"
              class="text-sm font-medium transition-colors cursor-pointer"
              :class="index === breadcrumbs.length - 1
                ? 'text-slate-900 dark:text-slate-100'
                : 'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300'"
            >
              {{ crumb.label }}
            </router-link>
            <span
              v-if="index < breadcrumbs.length - 1"
              class="text-slate-400"
            >/</span>
          </template>
        </nav>
      </div>

      <!-- Page Title (Mobile) -->
      <h1 class="sm:hidden text-base font-semibold text-slate-900 dark:text-slate-100">
        {{ pageTitle }}
      </h1>

      <!-- Center: Search -->
      <div class="hidden md:flex flex-1 justify-end px-8">
        <button
          class="group flex h-9 cursor-pointer items-center gap-2 rounded-full bg-slate-100 dark:bg-slate-800 px-3 py-1.5 outline-none transition-all duration-300 hover:bg-slate-200 dark:hover:bg-slate-700"
          @click="showSearchDialog = true"
        >
          <Search :size="16" class="text-slate-500 dark:text-slate-400 group-hover:text-slate-700 dark:group-hover:text-slate-200 transition-colors" />
          <span class="text-slate-500 dark:text-slate-400 group-hover:text-slate-700 dark:group-hover:text-slate-200 text-xs transition-colors duration-300">搜索</span>
          <span class="ml-0.5 flex items-center gap-1 rounded-md bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-600 px-1.5 py-0.5 text-xs text-slate-500 dark:text-slate-400 group-hover:text-slate-700 dark:group-hover:text-slate-200 transition-colors duration-300">
            <kbd class="font-sans">⌘</kbd>
            <kbd class="font-sans">K</kbd>
          </span>
        </button>
      </div>

      <!-- Right: Actions -->
      <div class="flex items-center gap-1 lg:gap-2">
        <!-- Settings -->
        <router-link
          to="/dashboard/settings"
          class="hidden sm:flex items-center justify-center p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
        >
          <Settings :size="20" class="text-slate-600 dark:text-slate-400" />
        </router-link>

        <!-- Dark Mode Toggle -->
        <button
          class="flex items-center justify-center p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
          :aria-label="uiStore.darkMode ? '切换到浅色模式' : '切换到深色模式'"
          @click="uiStore.toggleDarkMode()"
        >
          <Sun v-if="uiStore.darkMode" :size="20" class="text-slate-600 dark:text-slate-400" />
          <Moon v-else :size="20" class="text-slate-600 dark:text-slate-400" />
        </button>

        <!-- Language Switcher -->
        <LanguageSwitcher />

        <!-- Fullscreen Toggle -->
        <button
          class="flex items-center justify-center p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
          :aria-label="isFullscreen ? '退出全屏' : '全屏'"
          @click="toggleFullscreen"
        >
          <!-- Maximize Icon (进入全屏) -->
          <svg
            v-if="!isFullscreen"
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="text-slate-600 dark:text-slate-400"
          >
            <path d="M8 3H5a2 2 0 0 0-2 2v3"></path>
            <path d="M21 8V5a2 2 0 0 0-2-2h-3"></path>
            <path d="M3 16v3a2 2 0 0 0 2 2h3"></path>
            <path d="M16 21h3a2 2 0 0 0 2-2v-3"></path>
          </svg>
          <!-- Minimize Icon (退出全屏) -->
          <svg
            v-else
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="text-slate-600 dark:text-slate-400"
          >
            <path d="M8 3v3a2 2 0 0 1-2 2H3"></path>
            <path d="M21 8h-3a2 2 0 0 1-2-2V3"></path>
            <path d="M3 16h3a2 2 0 0 1 2 2v3"></path>
            <path d="M16 21v-3a2 2 0 0 1 2-2h3"></path>
          </svg>
        </button>

        <!-- Notifications -->
        <button class="bell-button relative flex items-center justify-center p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer">
          <Bell :size="20" class="text-slate-600 dark:text-slate-400" />
          <span
            v-if="notificationCount > 0"
            class="absolute top-1 right-1 w-2.5 h-2.5 bg-error-500 rounded-full"
          />
        </button>

        <!-- User Menu -->
        <div class="relative">
          <button
            class="flex items-center gap-2 px-2 py-1.5 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
            @click="showUserMenu = !showUserMenu"
          >
            <div class="w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center">
              <User :size="16" class="text-primary-600 dark:text-primary-400" />
            </div>
            <div class="hidden lg:block text-left">
              <p class="text-sm font-medium text-slate-900 dark:text-slate-100">
                {{ authStore.user?.name || 'Admin' }}
              </p>
              <p class="text-xs text-slate-500 dark:text-slate-400">
                {{ authStore.userRole || 'admin' }}
              </p>
            </div>
            <ChevronDown class="hidden lg:block text-slate-400" :size="16" />
          </button>

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
              v-if="showUserMenu"
              class="absolute right-0 mt-2 w-48 bg-white dark:bg-slate-800 rounded-lg shadow-panel border border-slate-200 dark:border-slate-700 py-1"
              v-click-outside="() => showUserMenu = false"
            >
              <router-link
                to="/dashboard/profile"
                class="flex items-center gap-2 px-3 py-2 text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-700 cursor-pointer"
                @click="showUserMenu = false"
              >
                <User :size="16" />
                个人资料
              </router-link>
              <router-link
                to="/dashboard/settings"
                class="flex items-center gap-2 px-3 py-2 text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-700 cursor-pointer"
                @click="showUserMenu = false"
              >
                <Settings :size="16" />
                设置
              </router-link>
              <hr class="my-1 border-slate-200 dark:border-slate-700">
              <button
                class="w-full flex items-center gap-2 px-3 py-2 text-sm text-error-600 dark:text-error-400 hover:bg-error-50 dark:hover:bg-error-900/30 cursor-pointer text-left"
                @click="authStore.logout(); showUserMenu = false"
              >
                <LogOut :size="16" />
                退出登录
              </button>
            </div>
          </Transition>
        </div>
      </div>
    </div>

    <!-- Mobile Search Bar -->
    <div class="sm:hidden px-4 pb-3">
      <div class="relative">
        <Search :size="20" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
        <input
          v-model="searchQuery"
          type="search"
          placeholder="搜索..."
          class="w-full pl-10 pr-4 py-2 bg-slate-100 dark:bg-slate-700 border-0 rounded-lg text-sm text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:ring-2 focus:ring-primary-500 outline-none"
        />
      </div>
    </div>

    <!-- Search Dialog -->
    <SearchDialog v-model:visible="showSearchDialog" />
  </header>
</template>
