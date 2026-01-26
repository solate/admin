<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { notificationsApi } from '@/api'
import { Bell, X, Check, Trash2, User, Shield, AlertTriangle, CircleCheck, Info } from 'lucide-vue-next'

const router = useRouter()

const isOpen = ref(false)
const notifications = ref<any[]>([])
const unreadCount = ref(0)
const loading = ref(false)

// Notification icons mapping
const notificationIcons = {
  user: User,
  security: Shield,
  warning: AlertTriangle,
  success: CircleCheck,
  info: Info
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
  const groups: Record<string, any[]> = {
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

const groupLabels = {
  today: '今天',
  yesterday: '昨天',
  older: '更早'
}

async function fetchNotifications() {
  loading.value = true
  try {
    const response = await notificationsApi.list({ limit: 20 })
    notifications.value = response.data.data || []
    unreadCount.value = response.data.unreadCount || 0
  } catch (error) {
    console.error('Failed to fetch notifications:', error)
  } finally {
    loading.value = false
  }
}

async function markAsRead(notificationId: string) {
  try {
    await notificationsApi.markAsRead(notificationId)
    const notification = notifications.value.find(n => n.id === notificationId)
    if (notification) {
      notification.isRead = true
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    }
  } catch (error) {
    console.error('Failed to mark notification as read:', error)
  }
}

async function markAllAsRead() {
  try {
    await notificationsApi.markAllAsRead()
    notifications.value.forEach(n => n.isRead = true)
    unreadCount.value = 0
  } catch (error) {
    console.error('Failed to mark all notifications as read:', error)
  }
}

async function deleteNotification(notificationId: string) {
  try {
    await notificationsApi.delete(notificationId)
    notifications.value = notifications.value.filter(n => n.id !== notificationId)
  } catch (error) {
    console.error('Failed to delete notification:', error)
  }
}

function getNotificationIcon(type: string) {
  return notificationIcons[type as keyof typeof notificationIcons] || Info
}

function getNotificationColor(type: string) {
  return notificationColors[type as keyof typeof notificationColors] || notificationColors.info
}

onMounted(() => {
  fetchNotifications()
})
</script>

<template>
  <div class="relative">
    <!-- Trigger Button -->
    <button
      @click="isOpen = !isOpen"
      class="relative p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
      aria-label="Notifications"
    >
      <el-icon :size="20" class="text-slate-600 dark:text-slate-400">
        <Bell />
      </el-icon>
      <span
        v-if="unreadCount > 0"
        class="absolute -top-1 -right-1 w-5 h-5 bg-red-500 text-white text-xs font-medium rounded-full flex items-center justify-center"
      >
        {{ unreadCount > 9 ? '9+' : unreadCount }}
      </span>
    </button>

    <!-- Dropdown -->
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
        @click.outside="isOpen = false"
        class="absolute right-0 mt-2 w-96 bg-white dark:bg-slate-800 rounded-xl shadow-xl border border-slate-200 dark:border-slate-700 z-50 overflow-hidden"
      >
        <!-- Header -->
        <div class="flex items-center justify-between px-4 py-3 border-b border-slate-200 dark:border-slate-700">
          <h3 class="font-semibold text-slate-900 dark:text-slate-100">通知中心</h3>
          <div class="flex items-center gap-2">
            <button
              v-if="unreadCount > 0"
              @click="markAllAsRead"
              class="text-xs text-primary-600 hover:text-primary-700 dark:text-primary-400 transition-colors cursor-pointer"
            >
              全部已读
            </button>
          </div>
        </div>

        <!-- Notifications List -->
        <div class="max-h-96 overflow-y-auto">
          <div
            v-if="loading"
            class="px-4 py-8 text-center text-sm text-slate-500 dark:text-slate-400"
          >
            加载中...
          </div>

          <div
            v-else-if="notifications.length === 0"
            class="px-4 py-8 text-center text-sm text-slate-500 dark:text-slate-400"
          >
            暂无通知
          </div>

          <div v-else>
            <div
              v-for="(group, groupKey) in groupedNotifications"
              :key="groupKey"
            >
              <div
                v-if="group.length > 0"
                class="px-4 py-2 text-xs font-medium text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-slate-900/50"
              >
                {{ groupLabels[groupKey as keyof typeof groupLabels] }}
              </div>
              <div
                v-for="notification in group"
                :key="notification.id"
                @click="markAsRead(notification.id)"
                class="px-4 py-3 hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-colors cursor-pointer border-b border-slate-100 dark:border-slate-700/50 last:border-0"
                :class="{ 'bg-primary-50/50 dark:bg-primary-900/10': !notification.isRead }"
              >
                <div class="flex gap-3">
                  <!-- Icon -->
                  <div class="flex-shrink-0 mt-0.5">
                    <div :class="['w-8 h-8 rounded-lg flex items-center justify-center', getNotificationColor(notification.type)]">
                      <el-icon :size="16">
                        <component :is="getNotificationIcon(notification.type)" />
                      </el-icon>
                    </div>
                  </div>

                  <!-- Content -->
                  <div class="flex-1 min-w-0">
                    <div class="flex items-start justify-between gap-2">
                      <p class="text-sm font-medium text-slate-900 dark:text-slate-100">
                        {{ notification.title }}
                      </p>
                      <button
                        @click.stop="deleteNotification(notification.id)"
                        class="flex-shrink-0 text-slate-400 hover:text-red-500 transition-colors cursor-pointer"
                      >
                        <el-icon :size="14">
                          <Trash2 />
                        </el-icon>
                      </button>
                    </div>
                    <p class="text-sm text-slate-600 dark:text-slate-400 mt-0.5">
                      {{ notification.message }}
                    </p>
                    <p class="text-xs text-slate-500 dark:text-slate-500 mt-1">
                      {{ new Date(notification.createdAt).toLocaleString() }}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-4 py-3 border-t border-slate-200 dark:border-slate-700">
          <router-link
            to="/dashboard/notifications"
            @click="isOpen = false"
            class="block text-center text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400 transition-colors cursor-pointer"
          >
            查看全部通知
          </router-link>
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
