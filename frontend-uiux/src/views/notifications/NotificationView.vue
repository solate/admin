<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { apiService } from '@/api'
import { Bell, Check, Trash22, ChevronLeft, User, Shield, CircleCheck, Info, AlertTriangle, TriangleAlert, Clock, Filter } from 'lucide-vue-next'

const router = useRouter()

const notifications = ref([])
const loading = ref(false)
const error = ref(null)
const selectedFilter = ref('all')
const searchQuery = ref('')

const filterOptions = [
  { label: 'All', value: 'all' },
  { label: 'Unread', value: 'unread' },
  { label: 'User', value: 'user' },
  { label: 'Security', value: 'security' },
  { label: 'System', value: 'system' }
]

const notificationIcons = {
  user: User,
  security: Shield,
  success: CircleCheck,
  info: Info,
  warning: TriangleAlert,
  error: AlertTriangle
}

const notificationColors = {
  user: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
  security: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
  success: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
  info: 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400',
  warning: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
  error: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
}

const unreadCount = computed(() => {
  return notifications.value.filter(n => !n.read).length
})

const filteredNotifications = computed(() => {
  let filtered = notifications.value

  // Apply status filter
  if (selectedFilter.value === 'unread') {
    filtered = filtered.filter(n => !n.read)
  } else if (selectedFilter.value !== 'all') {
    filtered = filtered.filter(n => n.type === selectedFilter.value)
  }

  // Apply search
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(n =>
      n.title?.toLowerCase().includes(query) ||
      n.message?.toLowerCase().includes(query)
    )
  }

  return filtered
})

async function fetchNotifications() {
  loading.value = true
  error.value = null
  try {
    const response = await apiService.notifications.list({ limit: 100 })
    notifications.value = response.data?.items || []
  } catch (err) {
    error.value = err.message || 'Failed to fetch notifications'
    console.error('Error fetching notifications:', err)
  } finally {
    loading.value = false
  }
}

async function markAsRead(notificationId) {
  try {
    await apiService.notifications.markAsRead(notificationId)
    const notification = notifications.value.find(n => n.id === notificationId)
    if (notification) {
      notification.read = true
    }
  } catch (err) {
    console.error('Failed to mark notification as read:', err)
  }
}

async function markAllAsRead() {
  try {
    await apiService.notifications.markAllAsRead()
    notifications.value.forEach(n => n.read = true)
  } catch (err) {
    console.error('Failed to mark all as read:', err)
  }
}

async function deleteNotification(notificationId) {
  try {
    await apiService.notifications.delete(notificationId)
    notifications.value = notifications.value.filter(n => n.id !== notificationId)
  } catch (err) {
    console.error('Failed to delete notification:', err)
  }
}

function handleNotificationClick(notification) {
  if (!notification.read) {
    markAsRead(notification.id)
  }
  if (notification.link) {
    router.push(notification.link)
  }
}

function getNotificationIcon(type) {
  return notificationIcons[type] || Info
}

function getNotificationColor(type) {
  return notificationColors[type] || notificationColors.info
}

function formatTime(date) {
  const now = new Date()
  const diff = now - new Date(date)
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes}m ago`
  if (hours < 24) return `${hours}h ago`
  if (days < 7) return `${days}d ago`
  return new Date(date).toLocaleDateString()
}

onMounted(() => {
  fetchNotifications()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center gap-4">
      <button
        @click="router.back()"
        class="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
        aria-label="Go back"
      >
        <ChevronLeft class="w-5 h-5 text-slate-600 dark:text-slate-400" />
      </button>
      <div class="flex-1">
        <div class="flex items-center gap-3">
          <h1 class="text-2xl font-display font-bold text-slate-900 dark:text-slate-100">Notifications</h1>
          <span
            v-if="unreadCount > 0"
            class="px-2.5 py-1 bg-primary-100 dark:bg-primary-900/50 text-primary-700 dark:text-primary-300 text-sm font-medium rounded-full"
          >
            {{ unreadCount }} unread
          </span>
        </div>
        <p class="text-slate-600 dark:text-slate-400">View and manage all your notifications</p>
      </div>
      <button
        v-if="unreadCount > 0"
        @click="markAllAsRead"
        class="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-xl transition-colors font-medium cursor-pointer"
      >
        Mark All Read
      </button>
    </div>

    <!-- Error Message -->
    <div
      v-if="error"
      class="p-4 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-xl"
    >
      <p class="text-sm text-red-700 dark:text-red-400">{{ error }}</p>
    </div>

    <!-- Filters -->
    <div class="flex flex-col sm:flex-row gap-4">
      <!-- Search -->
      <div class="relative flex-1">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search notifications..."
          class="w-full pl-10 pr-4 py-2.5 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-primary-500 outline-none"
        >
        <Bell class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
      </div>

      <!-- Filter -->
      <div class="flex gap-2">
        <button
          v-for="option in filterOptions"
          :key="option.value"
          @click="selectedFilter = option.value"
          class="px-4 py-2.5 rounded-xl transition-colors cursor-pointer font-medium"
          :class="selectedFilter === option.value
            ? 'bg-primary-600 text-white'
            : 'bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700'"
        >
          {{ option.label }}
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div
      v-if="loading"
      class="flex items-center justify-center py-12"
    >
      <div class="animate-spin w-8 h-8 border-3 border-primary-200 border-t-primary-600 rounded-full"></div>
    </div>

    <!-- Empty State -->
    <div
      v-else-if="filteredNotifications.length === 0"
      class="flex flex-col items-center py-16 px-4 text-center"
    >
      <div class="w-20 h-20 bg-slate-100 dark:bg-slate-700 rounded-full flex items-center justify-center mb-4">
        <Bell class="w-10 h-10 text-slate-400" />
      </div>
      <h3 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-1">
        {{ searchQuery || selectedFilter !== 'all' ? 'No notifications found' : 'All caught up!' }}
      </h3>
      <p class="text-slate-500 dark:text-slate-400">
        {{ searchQuery || selectedFilter !== 'all' ? 'Try adjusting your search or filters' : 'You have no notifications at the moment' }}
      </p>
    </div>

    <!-- Notifications List -->
    <div
      v-else
      class="space-y-3"
    >
      <div
        v-for="notification in filteredNotifications"
        :key="notification.id"
        @click="handleNotificationClick(notification)"
        class="flex items-start gap-4 p-4 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-colors cursor-pointer"
        :class="{ 'bg-primary-50/50 dark:bg-primary-900/20': !notification.read }"
      >
        <!-- Icon -->
        <div :class="['w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0', getNotificationColor(notification.type)]">
          <component :is="getNotificationIcon(notification.type)" class="w-5 h-5" />
        </div>

        <!-- Content -->
        <div class="flex-1 min-w-0">
          <div class="flex items-start justify-between gap-2">
            <div class="flex-1">
              <p class="font-medium text-slate-900 dark:text-slate-100" :class="{ 'font-semibold': !notification.read }">
                {{ notification.title }}
              </p>
              <p class="text-sm text-slate-600 dark:text-slate-400 mt-0.5">
                {{ notification.message }}
              </p>
              <p class="text-xs text-slate-500 dark:text-slate-500 mt-2 flex items-center gap-1">
                <Clock class="w-3 h-3" />
                {{ formatTime(notification.createdAt) }}
              </p>
            </div>
          </div>
        </div>

        <!-- Actions -->
        <div class="flex items-center gap-1" @click.stop>
          <button
            v-if="!notification.read"
            @click="markAsRead(notification.id)"
            class="p-2 hover:bg-slate-100 dark:hover:bg-slate-600 rounded-lg transition-colors cursor-pointer"
            aria-label="Mark as read"
          >
            <Check class="w-4 h-4 text-slate-400" />
          </button>
          <button
            @click="deleteNotification(notification.id)"
            class="p-2 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg transition-colors cursor-pointer"
            aria-label="Delete notification"
          >
            <Trash2 class="w-4 h-4 text-slate-400 hover:text-red-600 dark:hover:text-red-400" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
