<script setup>
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useUiStore } from '@/stores/modules/ui'
import { useAuthStore } from '@/stores/modules/auth'
import {
  Search,
  Bell,
  Sunny,
  Moon,
  Setting,
  User,
  ArrowDown,
  SwitchButton
} from '@element-plus/icons-vue'
import LanguageSwitcher from '@/components/language/LanguageSwitcher.vue'

const route = useRoute()
const uiStore = useUiStore()
const authStore = useAuthStore()

const searchQuery = ref('')
const showUserMenu = ref(false)

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

<template>
  <header class="sticky top-0 z-20 bg-white/80 dark:bg-slate-800/80 backdrop-blur-lg border-b border-slate-200 dark:border-slate-700">
    <div class="flex items-center justify-between h-16 px-4 lg:px-6">
      <!-- Left: Breadcrumbs -->
      <nav class="hidden sm:flex items-center gap-2 flex-1">
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

      <!-- Page Title (Mobile) -->
      <h1 class="sm:hidden text-base font-semibold text-slate-900 dark:text-slate-100">
        {{ pageTitle }}
      </h1>

      <!-- Center: Search -->
      <div class="hidden md:flex flex-1 justify-center px-8">
        <div class="relative w-full max-w-md">
          <el-icon :size="20" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"><Search /></el-icon>
          <input
            v-model="searchQuery"
            type="search"
            placeholder="搜索..."
            class="w-full pl-10 pr-4 py-2 bg-slate-100 dark:bg-slate-700 border-0 rounded-lg text-sm text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:ring-2 focus:ring-primary-500 outline-none transition-all"
          />
        </div>
      </div>

      <!-- Right: Actions -->
      <div class="flex items-center gap-1 lg:gap-2">
        <!-- Language Switcher -->
        <LanguageSwitcher />

        <!-- Dark Mode Toggle -->
        <button
          class="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
          :aria-label="uiStore.darkMode ? '切换到浅色模式' : '切换到深色模式'"
          @click="uiStore.toggleDarkMode()"
        >
          <el-icon v-if="uiStore.darkMode" :size="20" class="text-slate-600 dark:text-slate-400"><Sunny /></el-icon>
          <el-icon v-else :size="20" class="text-slate-600 dark:text-slate-400"><Moon /></el-icon>
        </button>

        <!-- Notifications -->
        <button class="relative p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer">
          <el-icon :size="20" class="text-slate-600 dark:text-slate-400"><Bell /></el-icon>
          <span
            v-if="notificationCount > 0"
            class="absolute top-1 right-1 w-2.5 h-2.5 bg-error-500 rounded-full"
          />
        </button>

        <!-- Settings -->
        <router-link
          to="/dashboard/settings"
          class="hidden sm:flex p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
        >
          <el-icon :size="20" class="text-slate-600 dark:text-slate-400"><Setting /></el-icon>
        </router-link>

        <!-- User Menu -->
        <div class="relative">
          <button
            class="flex items-center gap-2 px-2 py-1.5 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
            @click="showUserMenu = !showUserMenu"
          >
            <div class="w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center">
              <el-icon :size="16" class="text-primary-600 dark:text-primary-400"><User /></el-icon>
            </div>
            <div class="hidden lg:block text-left">
              <p class="text-sm font-medium text-slate-900 dark:text-slate-100">
                {{ authStore.user?.name || 'Admin' }}
              </p>
              <p class="text-xs text-slate-500 dark:text-slate-400">
                {{ authStore.userRole || 'admin' }}
              </p>
            </div>
            <el-icon class="hidden lg:block text-slate-400"><ArrowDown /></el-icon>
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
                <el-icon :size="16"><User /></el-icon>
                个人资料
              </router-link>
              <router-link
                to="/dashboard/settings"
                class="flex items-center gap-2 px-3 py-2 text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-700 cursor-pointer"
                @click="showUserMenu = false"
              >
                <el-icon :size="16"><Setting /></el-icon>
                设置
              </router-link>
              <hr class="my-1 border-slate-200 dark:border-slate-700">
              <button
                class="w-full flex items-center gap-2 px-3 py-2 text-sm text-error-600 dark:text-error-400 hover:bg-error-50 dark:hover:bg-error-900/30 cursor-pointer text-left"
                @click="authStore.logout(); showUserMenu = false"
              >
                <el-icon :size="16"><SwitchButton /></el-icon>
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
        <el-icon :size="20" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"><Search /></el-icon>
        <input
          v-model="searchQuery"
          type="search"
          placeholder="搜索..."
          class="w-full pl-10 pr-4 py-2 bg-slate-100 dark:bg-slate-700 border-0 rounded-lg text-sm text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:ring-2 focus:ring-primary-500 outline-none"
        />
      </div>
    </div>
  </header>
</template>
