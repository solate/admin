<script setup>
import { computed, ref, onMounted, onUnmounted, markRaw } from 'vue'
import { useRoute } from 'vue-router'
import { useUiStore } from '@/stores/modules/ui'
import { useAuthStore } from '@/stores/modules/auth'
import { useTenantsStore } from '@/stores/modules/tenants'
import TopNavbar from '@/components/layout/TopNavbar.vue'
import {
  Home,
  Building,
  Box,
  User,
  BarChart3,
  Settings,
  Bell,
  Search,
  Moon,
  Sun,
  Menu,
  ChevronLeft,
  X
} from 'lucide-vue-next'

const route = useRoute()
const uiStore = useUiStore()
const authStore = useAuthStore()
const tenantsStore = useTenantsStore()

const isMobileMenuOpen = ref(false)

// Initialize tenant store
tenantsStore.initialize()

const navigation = computed(() => {
  const baseNav = [
    {
      name: '概览',
      path: '/dashboard/overview',
      icon: markRaw(Home),
      key: 'overview'
    },
    {
      name: '租户管理',
      path: '/dashboard/tenants',
      icon: markRaw(Building),
      key: 'tenants'
    },
    {
      name: '服务管理',
      path: '/dashboard/services',
      icon: markRaw(Box),
      key: 'services'
    },
    {
      name: '用户管理',
      path: '/dashboard/users',
      icon: markRaw(User),
      key: 'users'
    },
    {
      name: '数据分析',
      path: '/dashboard/analytics',
      icon: markRaw(BarChart3),
      key: 'analytics'
    }
  ]

  const userRole = authStore.userRole
  // Filter navigation based on user role if needed
  return baseNav.filter(item => {
    // Example: super_admin and auditor only access
    if (item.key === 'tenants' && !['super_admin', 'auditor'].includes(userRole)) {
      return false
    }
    return true
  })
})

const bottomNavigation = [
  {
    name: '设置',
    path: '/dashboard/settings',
    icon: markRaw(Settings),
    key: 'settings'
  }
]

const isActive = (path) => {
  return route.path === path || route.path.startsWith(path + '/')
}

const toggleMobileMenu = () => {
  isMobileMenuOpen.value = !isMobileMenuOpen.value
}

const closeMobileMenu = () => {
  isMobileMenuOpen.value = false
}

// Handle window resize
const handleResize = () => {
  if (window.innerWidth >= 1024) {
    isMobileMenuOpen.value = false
  }
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-50 via-slate-50/80 to-blue-50/40 dark:from-slate-950 dark:via-slate-950 dark:to-slate-900/60">
    <!-- Mobile Header - 使用毛玻璃+阴影代替边框，更现代 -->
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

        <!-- User Menu -->
        <div class="flex items-center gap-2">
          <button class="p-2 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors cursor-pointer">
            <Bell :size="20" />
          </button>
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
          <!-- Logo -->
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

          <!-- Navigation -->
          <nav class="flex-1 overflow-y-auto p-4 space-y-1">
            <router-link
              v-for="item in navigation"
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

            <div class="pt-4 mt-4 border-t border-slate-200/60 dark:border-slate-700/30">
              <router-link
                v-for="item in bottomNavigation"
                :key="item.key"
                :to="item.path"
                @click="closeMobileMenu"
                class="flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all cursor-pointer"
                :class="isActive(item.path)
                  ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400'
                  : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700'"
              >
                <component :is="item.icon" class="w-5 h-5" />
                <span class="font-medium">{{ item.name }}</span>
              </router-link>
            </div>
          </nav>

          <!-- User Section -->
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

    <!-- Desktop Sidebar -->
    <aside
      :class="[
        'hidden lg:flex flex-col fixed inset-y-0 left-0 transition-all duration-300 z-30',
        'bg-white/90 dark:bg-slate-900/90',
        'backdrop-blur-xl',
        'border-r border-slate-200/60 dark:border-slate-800/60',
        'shadow-[4px_0_24px_-8px_rgba(0,0,0,0.06)] dark:shadow-[4px_0_24px_-8px_rgba(0,0,0,0.3)]',
        uiStore.sidebarOpen ? 'w-64' : 'w-20'
      ]"
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
              v-if="uiStore.sidebarOpen"
              class="text-lg font-semibold text-slate-900 dark:text-slate-100 whitespace-nowrap overflow-hidden"
            >
              AdminSystem
            </span>
          </Transition>
        </div>
      </div>

      <!-- Tenant Info (when expanded) -->
      <Transition
        enter-active-class="transition-all duration-200"
        enter-from-class="opacity-0 h-0 mt-0 mx-0"
        enter-to-class="opacity-100 h-auto mt-4 mx-4"
        leave-active-class="transition-all duration-200"
        leave-from-class="opacity-100 h-auto mt-4 mx-4"
        leave-to-class="opacity-0 h-0 mt-0 mx-0"
      >
        <div
          v-if="uiStore.sidebarOpen && tenantsStore.currentTenant"
          class="px-4 py-3 bg-primary-50 dark:bg-primary-900/30 rounded-lg overflow-hidden"
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

      <!-- Navigation -->
      <nav class="flex-1 overflow-y-auto p-3 space-y-1">
        <router-link
          v-for="item in navigation"
          :key="item.key"
          :to="item.path"
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all cursor-pointer group"
          :class="isActive(item.path)
            ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400'
            : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700'"
          :title="!uiStore.sidebarOpen ? item.name : ''"
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
              v-if="uiStore.sidebarOpen"
              class="font-medium whitespace-nowrap overflow-hidden"
            >
              {{ item.name }}
            </span>
          </Transition>
        </router-link>

        <!-- Divider - 柔和的分割线 -->
        <div class="border-t border-slate-200/60 dark:border-slate-700/30" :class="uiStore.sidebarOpen ? 'pt-4 mt-4' : 'pt-3 mt-3'">
          <router-link
            v-for="item in bottomNavigation"
            :key="item.key"
            :to="item.path"
            class="flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all cursor-pointer"
            :class="isActive(item.path)
              ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400'
              : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700'"
            :title="!uiStore.sidebarOpen ? item.name : ''"
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
                v-if="uiStore.sidebarOpen"
                class="font-medium whitespace-nowrap overflow-hidden"
              >
                {{ item.name }}
              </span>
            </Transition>
          </router-link>
        </div>
      </nav>
    </aside>

    <!-- Main Content -->
    <main
      :class="[
        'transition-all duration-300 min-h-screen',
        uiStore.sidebarOpen ? 'lg:ml-64' : 'lg:ml-20',
        'pt-16 lg:pt-0'
      ]"
    >
      <!-- Top Navbar -->
      <TopNavbar />

      <!-- Page Content -->
      <div class="p-4 lg:p-6">
        <div class="bg-white/70 dark:bg-slate-900/60 backdrop-blur-sm rounded-xl p-4 lg:p-6 min-h-[calc(100vh-8rem)] shadow-sm border border-slate-200/50 dark:border-slate-800/50">
          <router-view />
        </div>
      </div>
    </main>
  </div>
</template>
