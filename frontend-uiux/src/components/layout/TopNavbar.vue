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
  RotateCw,
  Shield,
  Mail,
  Building,
  CreditCard,
  HelpCircle
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

// 用户菜单相关
const userInitials = computed(() => {
  const name = authStore.user?.name || 'Admin'
  return name
    .split(' ')
    .map(n => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)
})

const userRoleLabel = computed(() => {
  const role = authStore.userRole
  const roleLabels = {
    super_admin: '超级管理员',
    admin: '管理员',
    auditor: '审计员',
    user: '用户'
  }
  return roleLabels[role] || '用户'
})

// 获取当前租户名称
const currentTenantName = computed(() => {
  // 从租户 store 获取当前租户名称
  return authStore.user?.tenantName || '默认租户'
})
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
          <!-- Trigger Button -->
          <button
            class="group flex items-center gap-2.5 px-2.5 py-1.5 rounded-xl transition-all duration-200 cursor-pointer"
            :class="showUserMenu
              ? 'bg-primary-50 dark:bg-primary-900/20 shadow-sm'
              : 'hover:bg-slate-100 dark:hover:bg-slate-700/50'"
            @click="showUserMenu = !showUserMenu"
          >
            <!-- Avatar with Gradient -->
            <div class="relative">
              <div class="w-9 h-9 rounded-xl bg-gradient-to-br from-primary-500 to-primary-600 flex items-center justify-center shadow-lg shadow-primary-500/25 ring-2 ring-white dark:ring-slate-700 ring-offset-2 dark:ring-offset-slate-800 transition-all duration-200"
                   :class="showUserMenu ? 'ring-primary-500 dark:ring-primary-400 scale-105' : 'group-hover:scale-105'">
                <span class="text-sm font-bold text-white">{{ userInitials }}</span>
              </div>
              <!-- Online Status Indicator -->
              <span class="absolute -bottom-0.5 -right-0.5 w-3 h-3 bg-green-500 rounded-full border-2 border-white dark:border-slate-800"></span>
            </div>

            <!-- User Info -->
            <div class="hidden lg:block text-left">
              <p class="text-sm font-semibold text-slate-900 dark:text-slate-100 leading-tight">
                {{ authStore.user?.name || 'Admin' }}
              </p>
              <p class="text-xs text-slate-500 dark:text-slate-400 flex items-center gap-1">
                {{ userRoleLabel }}
              </p>
            </div>

            <!-- Chevron -->
            <ChevronDown
              class="hidden lg:block text-slate-400 transition-transform duration-200"
              :size="16"
              :class="{ 'rotate-180 text-primary-600 dark:text-primary-400': showUserMenu }"
            />
          </button>

          <!-- Dropdown -->
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
              v-click-outside="() => showUserMenu = false"
              class="absolute right-0 mt-3 w-80 bg-white dark:bg-slate-800 rounded-2xl shadow-2xl border border-slate-200/80 dark:border-slate-700/50 overflow-hidden z-50"
            >
              <!-- User Info Header -->
              <div class="relative p-5 bg-gradient-to-br from-primary-50 via-blue-50 to-indigo-50 dark:from-primary-900/30 dark:via-blue-900/20 dark:to-indigo-900/20">
                <!-- Decorative Pattern -->
                <div class="absolute inset-0 opacity-[0.03] dark:opacity-[0.05]" style="background-image: radial-gradient(circle, currentColor 1px, transparent 1px); background-size: 20px 20px;"></div>

                <div class="relative flex items-center gap-4">
                  <!-- Large Avatar -->
                  <div class="w-14 h-14 rounded-2xl bg-gradient-to-br from-primary-500 to-primary-600 flex items-center justify-center shadow-xl shadow-primary-500/30">
                    <span class="text-xl font-bold text-white">{{ userInitials }}</span>
                  </div>

                  <!-- User Details -->
                  <div class="flex-1 min-w-0">
                    <h3 class="text-base font-bold text-slate-900 dark:text-slate-100 truncate">
                      {{ authStore.user?.name || 'Admin' }}
                    </h3>
                    <p class="text-sm text-slate-600 dark:text-slate-400 truncate flex items-center gap-1.5 mt-0.5">
                      <Mail :size="13" class="flex-shrink-0" />
                      {{ authStore.user?.email || 'admin@example.com' }}
                    </p>
                  </div>
                </div>

                <!-- Role & Tenant Badges -->
                <div class="relative flex items-center gap-2 mt-4">
                  <span class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-white dark:bg-slate-700 rounded-lg text-xs font-semibold text-primary-700 dark:text-primary-300 shadow-sm border border-primary-100 dark:border-primary-800/50">
                    <Shield :size="12" class="flex-shrink-0" />
                    {{ userRoleLabel }}
                  </span>
                  <span class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-white dark:bg-slate-700 rounded-lg text-xs font-medium text-slate-600 dark:text-slate-400 shadow-sm border border-slate-200 dark:border-slate-600">
                    <Building :size="12" class="flex-shrink-0" />
                    {{ currentTenantName }}
                  </span>
                </div>
              </div>

              <!-- Menu Items -->
              <div class="p-2">
                <!-- Account Section -->
                <div class="px-3 py-2">
                  <p class="text-xs font-semibold text-slate-400 dark:text-slate-500 uppercase tracking-wider mb-1.5">账户</p>
                </div>

                <router-link
                  to="/dashboard/profile"
                  class="flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-all duration-150 cursor-pointer group/item"
                  @click="showUserMenu = false"
                >
                  <div class="w-9 h-9 rounded-lg flex items-center justify-center bg-slate-100 dark:bg-slate-700/50 group-hover/item:bg-primary-100 dark:group-hover/item:bg-primary-900/30 transition-colors">
                    <User :size="18" class="text-slate-600 dark:text-slate-400 group-hover/item:text-primary-600 dark:group-hover/item:text-primary-400 transition-colors" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <p class="text-sm font-medium text-slate-900 dark:text-slate-100">个人资料</p>
                    <p class="text-xs text-slate-500 dark:text-slate-400">管理您的个人信息</p>
                  </div>
                </router-link>

                <router-link
                  to="/dashboard/profile?tab=security"
                  class="flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-all duration-150 cursor-pointer group/item"
                  @click="showUserMenu = false"
                >
                  <div class="w-9 h-9 rounded-lg flex items-center justify-center bg-slate-100 dark:bg-slate-700/50 group-hover/item:bg-primary-100 dark:group-hover/item:bg-primary-900/30 transition-colors">
                    <Shield :size="18" class="text-slate-600 dark:text-slate-400 group-hover/item:text-primary-600 dark:group-hover/item:text-primary-400 transition-colors" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <p class="text-sm font-medium text-slate-900 dark:text-slate-100">安全设置</p>
                    <p class="text-xs text-slate-500 dark:text-slate-400">密码与两步验证</p>
                  </div>
                </router-link>

                <!-- Workspace Section -->
                <div class="px-3 py-2 mt-1">
                  <p class="text-xs font-semibold text-slate-400 dark:text-slate-500 uppercase tracking-wider mb-1.5">工作区</p>
                </div>

                <router-link
                  to="/dashboard/settings"
                  class="flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-all duration-150 cursor-pointer group/item"
                  @click="showUserMenu = false"
                >
                  <div class="w-9 h-9 rounded-lg flex items-center justify-center bg-slate-100 dark:bg-slate-700/50 group-hover/item:bg-primary-100 dark:group-hover/item:bg-primary-900/30 transition-colors">
                    <Settings :size="18" class="text-slate-600 dark:text-slate-400 group-hover/item:text-primary-600 dark:group-hover/item:text-primary-400 transition-colors" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <p class="text-sm font-medium text-slate-900 dark:text-slate-100">系统设置</p>
                    <p class="text-xs text-slate-500 dark:text-slate-400">应用偏好配置</p>
                  </div>
                </router-link>

                <button
                  class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-all duration-150 cursor-pointer group/item text-left"
                >
                  <div class="w-9 h-9 rounded-lg flex items-center justify-center bg-slate-100 dark:bg-slate-700/50 group-hover/item:bg-primary-100 dark:group-hover/item:bg-primary-900/30 transition-colors">
                    <HelpCircle :size="18" class="text-slate-600 dark:text-slate-400 group-hover/item:text-primary-600 dark:group-hover/item:text-primary-400 transition-colors" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <p class="text-sm font-medium text-slate-900 dark:text-slate-100">帮助中心</p>
                    <p class="text-xs text-slate-500 dark:text-slate-400">获取支持与文档</p>
                  </div>
                </button>
              </div>

              <!-- Logout Section -->
              <div class="p-2 border-t border-slate-200 dark:border-slate-700/50">
                <button
                  class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-red-50 dark:hover:bg-red-900/20 transition-all duration-150 cursor-pointer group/logout text-left"
                  @click="authStore.logout(); showUserMenu = false"
                >
                  <div class="w-9 h-9 rounded-lg flex items-center justify-center bg-red-100 dark:bg-red-900/30 group-hover/logout:bg-red-200 dark:group-hover/logout:bg-red-900/50 transition-colors">
                    <LogOut :size="18" class="text-red-600 dark:text-red-400" />
                  </div>
                  <div class="flex-1">
                    <p class="text-sm font-semibold text-red-600 dark:text-red-400">退出登录</p>
                    <p class="text-xs text-red-500/70 dark:text-red-400/70">安全退出您的账户</p>
                  </div>
                </button>
              </div>
            </div>
          </Transition>

          <!-- Overlay -->
          <div
            v-if="showUserMenu"
            @click="showUserMenu = false"
            class="fixed inset-0 z-40"
          ></div>
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
