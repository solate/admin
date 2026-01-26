<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/modules/auth'
import { User, Settings, LogOut, Shield, ChevronDown, Mail, Building } from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()

const isOpen = ref(false)

const userInitials = computed(() => {
  const name = authStore.user?.name || 'User'
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
    super_admin: 'Super Admin',
    admin: 'Administrator',
    auditor: 'Auditor',
    user: 'User'
  }
  return roleLabels[role] || 'User'
})

const menuItems = computed(() => [
  {
    icon: User,
    label: 'Profile',
    description: 'Manage your account',
    action: () => goToProfile()
  },
  {
    icon: Shield,
    label: 'Security',
    description: 'Password & 2FA settings',
    action: () => goToProfile('security')
  },
  {
    icon: Settings,
    label: 'Settings',
    description: 'App preferences',
    action: () => goToSettings()
  }
])

function goToProfile(tab = 'profile') {
  router.push({ name: 'profile', query: { tab } })
  isOpen.value = false
}

function goToSettings() {
  router.push({ name: 'settings' })
  isOpen.value = false
}

function handleLogout() {
  authStore.logout()
  router.push({ name: 'login' })
  isOpen.value = false
}

const handleClickOutside = (event) => {
  const menu = document.querySelector('[data-user-menu]')
  if (menu && !menu.contains(event.target)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div class="relative" data-user-menu>
    <!-- Trigger Button -->
    <button
      @click.stop="isOpen = !isOpen"
      class="flex items-center gap-3 px-3 py-2 bg-white/80 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-700 transition-all cursor-pointer"
      aria-label="User menu"
      aria-haspopup="menu"
      :aria-expanded="isOpen"
    >
      <div class="w-8 h-8 bg-gradient-to-br from-primary-500 to-primary-600 rounded-lg flex items-center justify-center">
        <span class="text-sm font-semibold text-white">{{ userInitials }}</span>
      </div>
      <div class="text-left hidden md:block">
        <p class="text-sm font-medium text-slate-900 dark:text-slate-100">{{ authStore.user?.name || 'User' }}</p>
        <p class="text-xs text-slate-500 dark:text-slate-400">{{ userRoleLabel }}</p>
      </div>
      <el-icon :size="16" class="text-slate-500 transition-transform hidden md:block" :class="{ 'rotate-180': isOpen }">
        <ChevronDown />
      </el-icon>
    </button>

    <!-- Dropdown Menu -->
    <Transition
      enter-active-class="transition ease-out duration-200"
      enter-from-class="opacity-0 scale-95 translate-y-2"
      enter-to-class="opacity-100 scale-100 translate-y-0"
      leave-active-class="transition ease-in duration-150"
      leave-from-class="opacity-100 scale-100 translate-y-0"
      leave-to-class="opacity-0 scale-95 translate-y-2"
    >
      <div
        v-if="isOpen"
        @click.stop
        class="absolute right-0 mt-2 w-72 bg-white dark:bg-slate-800 rounded-2xl shadow-xl border border-slate-200 dark:border-slate-700 z-50 overflow-hidden"
      >
        <!-- User Info Header -->
        <div class="p-4 bg-gradient-to-r from-primary-50 to-primary-100/50 dark:from-primary-900/30 dark:to-primary-800/20 border-b border-slate-200 dark:border-slate-700">
          <div class="flex items-center gap-3">
            <div class="w-12 h-12 bg-gradient-to-br from-primary-500 to-primary-600 rounded-xl flex items-center justify-center">
              <span class="text-lg font-semibold text-white">{{ userInitials }}</span>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-semibold text-slate-900 dark:text-slate-100 truncate">
                {{ authStore.user?.name || 'User' }}
              </p>
              <p class="text-xs text-slate-600 dark:text-slate-400 truncate flex items-center gap-1">
                <Mail  :size="12"  />
                {{ authStore.user?.email || '' }}
              </p>
            </div>
          </div>

          <!-- Role Badge -->
          <div class="mt-3 flex items-center gap-2">
            <span class="px-2 py-1 bg-white dark:bg-slate-700 rounded-lg text-xs font-medium text-primary-700 dark:text-primary-300">
              {{ userRoleLabel }}
            </span>
            <span
              v-if="authStore.tenantId"
              class="px-2 py-1 bg-white dark:bg-slate-700 rounded-lg text-xs font-medium text-slate-600 dark:text-slate-400 flex items-center gap-1"
            >
              <Building  :size="12"  />
              {{ authStore.tenantId }}
            </span>
          </div>
        </div>

        <!-- Menu Items -->
        <div class="p-2">
          <button
            v-for="item in menuItems"
            :key="item.label"
            @click="item.action"
            class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors cursor-pointer text-left group"
          >
            <div class="w-9 h-9 rounded-lg flex items-center justify-center bg-slate-100 dark:bg-slate-700 group-hover:bg-primary-100 dark:group-hover:bg-primary-900/30 transition-colors">
              <el-icon :size="20" class="text-slate-600 dark:text-slate-400 group-hover:text-primary-600 dark:group-hover:text-primary-400 transition-colors">
                <component :is="item.icon" />
              </el-icon>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-slate-900 dark:text-slate-100">{{ item.label }}</p>
              <p class="text-xs text-slate-500 dark:text-slate-400">{{ item.description }}</p>
            </div>
          </button>
        </div>

        <!-- Logout Button -->
        <div class="p-2 border-t border-slate-200 dark:border-slate-700">
          <button
            @click="handleLogout"
            class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors cursor-pointer text-left group"
          >
            <div class="w-9 h-9 rounded-lg flex items-center justify-center bg-red-100 dark:bg-red-900/30 group-hover:bg-red-200 dark:group-hover:bg-red-900/50 transition-colors">
              <el-icon :size="20" class="text-red-600 dark:text-red-400">
                <LogOut />
              </el-icon>
            </div>
            <div class="flex-1">
              <p class="text-sm font-medium text-red-600 dark:text-red-400">Logout</p>
              <p class="text-xs text-red-500/70 dark:text-red-400/70">Sign out of your account</p>
            </div>
          </button>
        </div>
      </div>
    </Transition>

    <!-- Overlay -->
    <div
      v-if="isOpen"
      @click="isOpen = false"
      class="fixed inset-0 z-40"
    ></div>
  </div>
</template>
