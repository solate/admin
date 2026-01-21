<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { apiService } from '@/services/api'
import icons from '@/components/icons/index.js'

const router = useRouter()

const { Bell, XMark, Check, Trash, User, Shield, AlertCircle, CheckCircle, InformationCircle, Sparkles } = icons

const isOpen = ref(false)
const notifications = ref([])
const unreadCount = ref(0)
const loading = ref(false)

// Notification icons mapping
const notificationIcons = {
  user: User,
  security: Shield,
  warning: AlertCircle,
  success: CheckCircle,
  info: InformationCircle
}

// Notification colors mapping
const notificationColors = {
  user: 'text-blue-500 bg-blue-50 dark:bg-blue-900/30',
  security: 'text-amber-500 bg-amber-50 dark:bg-amber-900/30',
  warning: 'text-red-500 bg-red-50 dark:bg-red-900/30',
  success: 'text-emerald-500 bg-emerald-50 dark:bg-emerald-900/30',
  info: 'text-primary-500 bg-primary-50 dark:bg-primary-900/30'
}

const groupedNotifications = computed(() => {
  const groups = {
    today: [],
    yesterday: [],
    older: []
  }

  const today = new Date()
  today.setHours(0, 0, 0, 0)

  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  notifications.value.forEach(notification => {
    const date = new Date(notification.createdAt)
    if (date >= today) {
      groups.today.push(notification)
    } else if (date >= yesterday) {
      groups.yesterday.push(notification)
    } else {
      groups.older.push(notification)
    }
  })

  return groups
})

async function fetchNotifications() {
  loading.value = true
  try {
    const response = await apiService.notifications.list({ limit: 20 })
    notifications.value = response.data?.items || []
  } catch (error) {
    console.error('Failed to fetch notifications:', error)
  } finally {
    loading.value = false
  }
}

async function fetchUnreadCount() {
  try {
    const response = await apiService.notifications.unreadCount()
    unreadCount.value = response.data?.count || 0
  } catch (error) {
    console.error('Failed to fetch unread count:', error)
  }
}

async function markAsRead(notificationId) {
  try {
    await apiService.notifications.markAsRead(notificationId)
    const notification = notifications.value.find(n => n.id === notificationId)
    if (notification) {
      notification.read = true
    }
    unreadCount.value = Math.max(0, unreadCount.value - 1)
  } catch (error) {
    console.error('Failed to mark notification as read:', error)
  }
}

async function markAllAsRead() {
  try {
    await apiService.notifications.markAllAsRead()
    notifications.value.forEach(n => n.read = true)
    unreadCount.value = 0
  } catch (error) {
    console.error('Failed to mark all as read:', error)
  }
}

async function deleteNotification(notificationId) {
  try {
    await apiService.notifications.delete(notificationId)
    notifications.value = notifications.value.filter(n => n.id !== notificationId)
    if (!notifications.value.find(n => n.id === notificationId)?.read) {
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    }
  } catch (error) {
    console.error('Failed to delete notification:', error)
  }
}

function handleNotificationClick(notification) {
  if (!notification.read) {
    markAsRead(notification.id)
  }

  // Navigate to related page if available
  if (notification.link) {
    router.push(notification.link)
    isOpen.value = false
  }
}

function getNotificationIcon(type) {
  return notificationIcons[type] || InformationCircle
}

function getNotificationColor(type) {
  return notificationColors[type] || notificationColors.info
}

function formatTime(date) {
  const now = new Date()
  const diff = now - new Date(date)
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)

  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes}m ago`
  if (hours < 24) return `${hours}h ago`
  return new Date(date).toLocaleDateString()
}

onMounted(() => {
  fetchNotifications()
  fetchUnreadCount()
})

// Expose refresh function
defineExpose({
  refresh: fetchNotifications
})
</script>

<template>
  <div class="relative">
    <!-- Trigger Button -->
    <button
      @click="isOpen = !isOpen"
      class="relative p-2 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-colors cursor-pointer"
      aria-label="Open notifications"
      :aria-expanded="isOpen"
    >
      <Bell :class="['w-5 h-5 text-slate-600 dark:text-slate-400 transition-colors', unreadCount > 0 ? 'text-primary-600 dark:text-primary-400' : '']" />

      <!-- Unread Badge -->
      <span
        v-if="unreadCount > 0"
        class="absolute -top-1 -right-1 w-5 h-5 bg-red-500 text-white text-xs font-semibold rounded-full flex items-center justify-center"
      >
        {{ unreadCount > 9 ? '9+' : unreadCount }}
      </span>
    </button>

    <!-- Dropdown Panel -->
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
        class="absolute right-0 mt-2 w-96 bg-white dark:bg-slate-800 rounded-2xl shadow-xl border border-slate-200 dark:border-slate-700 z-50 overflow-hidden"
      >
        <!-- Header -->
        <div class="flex items-center justify-between p-4 border-b border-slate-200 dark:border-slate-700">
          <div class="flex items-center gap-2">
            <Sparkles class="w-5 h-5 text-primary-500" />
            <h3 class="font-semibold text-slate-900 dark:text-slate-100">Notifications</h3>
            <span
              v-if="unreadCount > 0"
              class="px-2 py-0.5 bg-primary-100 dark:bg-primary-900/50 text-primary-700 dark:text-primary-300 text-xs font-medium rounded-full"
            >
              {{ unreadCount }} new
            </span>
          </div>
          <button
            v-if="unreadCount > 0"
            @click="markAllAsRead"
            class="text-sm text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 font-medium cursor-pointer"
          >
            Mark all read
          </button>
        </div>

        <!-- Notifications List -->
        <div class="max-h-96 overflow-y-auto">
          <!-- Loading State -->
          <div
            v-if="loading"
            class="flex items-center justify-center py-12"
          >
            <div class="animate-spin w-8 h-8 border-3 border-primary-200 border-t-primary-600 rounded-full"></div>
          </div>

          <!-- Empty State -->
          <div
            v-else-if="notifications.length === 0"
            class="flex flex-col items-center py-12 px-4 text-center"
          >
            <div class="w-16 h-16 bg-slate-100 dark:bg-slate-700 rounded-full flex items-center justify-center mb-4">
              <Bell class="w-8 h-8 text-slate-400" />
            </div>
            <p class="text-sm font-medium text-slate-900 dark:text-slate-100 mb-1">No notifications</p>
            <p class="text-sm text-slate-500 dark:text-slate-400">You're all caught up!</p>
          </div>

          <!-- Notifications Groups -->
          <template v-else>
            <!-- Today -->
            <div v-if="groupedNotifications.today.length > 0">
              <div class="px-4 py-2 bg-slate-50 dark:bg-slate-900/50 text-xs font-medium text-slate-500 dark:text-slate-400">
                Today
              </div>
              <div
                v-for="notification in groupedNotifications.today"
                :key="notification.id"
                @click="handleNotificationClick(notification)"
                class="flex gap-3 p-4 hover:bg-slate-50 dark:hover:bg-slate-700/50 cursor-pointer transition-colors border-b border-slate-100 dark:border-slate-700/50"
                :class="{ 'bg-primary-50/50 dark:bg-primary-900/20': !notification.read }"
              >
                <div :class="['w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0', getNotificationColor(notification.type)]">
                  <component :is="getNotificationIcon(notification.type)" class="w-5 h-5" />
                </div>
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-medium text-slate-900 dark:text-slate-100" :class="{ 'font-semibold': !notification.read }">
                    {{ notification.title }}
                  </p>
                  <p class="text-sm text-slate-600 dark:text-slate-400 line-clamp-2">
                    {{ notification.message }}
                  </p>
                  <p class="text-xs text-slate-500 dark:text-slate-500 mt-1">
                    {{ formatTime(notification.createdAt) }}
                  </p>
                </div>
                <button
                  @click.stop="deleteNotification(notification.id)"
                  class="p-1 hover:bg-slate-200 dark:hover:bg-slate-600 rounded transition-colors cursor-pointer"
                  aria-label="Delete notification"
                >
                  <XMark class="w-4 h-4 text-slate-400" />
                </button>
              </div>
            </div>

            <!-- Yesterday -->
            <div v-if="groupedNotifications.yesterday.length > 0">
              <div class="px-4 py-2 bg-slate-50 dark:bg-slate-900/50 text-xs font-medium text-slate-500 dark:text-slate-400">
                Yesterday
              </div>
              <div
                v-for="notification in groupedNotifications.yesterday"
                :key="notification.id"
                @click="handleNotificationClick(notification)"
                class="flex gap-3 p-4 hover:bg-slate-50 dark:hover:bg-slate-700/50 cursor-pointer transition-colors border-b border-slate-100 dark:border-slate-700/50 opacity-75"
              >
                <div :class="['w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0', getNotificationColor(notification.type)]">
                  <component :is="getNotificationIcon(notification.type)" class="w-5 h-5" />
                </div>
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-medium text-slate-900 dark:text-slate-100">
                    {{ notification.title }}
                  </p>
                  <p class="text-sm text-slate-600 dark:text-slate-400 line-clamp-2">
                    {{ notification.message }}
                  </p>
                  <p class="text-xs text-slate-500 dark:text-slate-500 mt-1">
                    {{ formatTime(notification.createdAt) }}
                  </p>
                </div>
              </div>
            </div>

            <!-- Older -->
            <div v-if="groupedNotifications.older.length > 0">
              <div class="px-4 py-2 bg-slate-50 dark:bg-slate-900/50 text-xs font-medium text-slate-500 dark:text-slate-400">
                Older
              </div>
              <div
                v-for="notification in groupedNotifications.older"
                :key="notification.id"
                @click="handleNotificationClick(notification)"
                class="flex gap-3 p-4 hover:bg-slate-50 dark:hover:bg-slate-700/50 cursor-pointer transition-colors border-b border-slate-100 dark:border-slate-700/50 opacity-60"
              >
                <div :class="['w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0', getNotificationColor(notification.type)]">
                  <component :is="getNotificationIcon(notification.type)" class="w-5 h-5" />
                </div>
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-medium text-slate-900 dark:text-slate-100">
                    {{ notification.title }}
                  </p>
                  <p class="text-sm text-slate-600 dark:text-slate-400 line-clamp-2">
                    {{ notification.message }}
                  </p>
                  <p class="text-xs text-slate-500 dark:text-slate-500 mt-1">
                    {{ formatTime(notification.createdAt) }}
                  </p>
                </div>
              </div>
            </div>
          </template>
        </div>

        <!-- Footer -->
        <div class="p-3 border-t border-slate-200 dark:border-slate-700">
          <button
            @click="() => { router.push('/dashboard/notifications'); isOpen = false }"
            class="w-full py-2 text-sm text-center text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 font-medium cursor-pointer"
          >
            View all notifications
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

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.max-h-96::-webkit-scrollbar {
  width: 4px;
}

.max-h-96::-webkit-scrollbar-track {
  background: transparent;
}

.max-h-96::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 2px;
}

.dark .max-h-96::-webkit-scrollbar-thumb {
  background: #475569;
}
</style>
