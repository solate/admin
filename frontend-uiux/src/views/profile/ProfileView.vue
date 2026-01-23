<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/modules/auth'
import { apiService } from '@/api'
import { UserCircle, Mail, Shield, Bell, Key, Smartphone, Clock, Check, X, AlertTriangle, Globe, Building, Badge } from 'lucide-vue-next'

const route = useRoute()
const authStore = useAuthStore()

const activeTab = ref(route.query.tab || 'profile')
const saving = ref(false)
const message = ref(null)
const error = ref(null)

const tabs = [
  { id: 'profile', name: 'Profile', icon: UserCircle },
  { id: 'security', name: 'Security', icon: Shield },
  { id: 'sessions', name: 'Sessions', icon: Smartphone },
  { id: 'notifications', name: 'Notifications', icon: Bell }
]

// Profile form
const profile = ref({
  name: '',
  email: '',
  phone: '',
  location: '',
  bio: ''
})

// Password form
const passwordForm = ref({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 2FA state
const twoFactorEnabled = ref(false)
const showTwoFactorSetup = ref(false)

// Sessions
const sessions = ref([])

// Notification preferences
const notifications = ref({
  email: true,
  security: true,
  marketing: false,
  updates: true
})

const hasChanges = computed(() => {
  return JSON.stringify(profile.value) !== JSON.stringify({
    name: authStore.user?.name || '',
    email: authStore.user?.email || '',
    phone: '',
    location: '',
    bio: ''
  })
})

const passwordStrength = computed(() => {
  const password = passwordForm.value.newPassword
  if (!password) return { level: 0, label: '', color: '' }

  let strength = 0
  if (password.length >= 8) strength++
  if (password.length >= 12) strength++
  if (/[a-z]/.test(password) && /[A-Z]/.test(password)) strength++
  if (/\d/.test(password)) strength++
  if (/[^a-zA-Z0-9]/.test(password)) strength++

  const levels = [
    { label: 'Weak', color: 'bg-red-500' },
    { label: 'Fair', color: 'bg-amber-500' },
    { label: 'Good', color: 'bg-yellow-500' },
    { label: 'Strong', color: 'bg-emerald-500' },
    { label: 'Very Strong', color: 'bg-emerald-600' }
  ]

  return { level: strength, ...levels[strength - 1] || levels[0] }
})

async function fetchProfile() {
  try {
    const response = await apiService.users.profile()
    const userData = response.data

    profile.value = {
      name: userData.name || '',
      email: userData.email || '',
      phone: userData.phone || '',
      location: userData.location || '',
      bio: userData.bio || ''
    }

    twoFactorEnabled.value = userData.twoFactorEnabled || false
    notifications.value = userData.notifications || notifications.value
  } catch (err) {
    console.error('Failed to fetch profile:', err)
  }
}

async function fetchSessions() {
  try {
    const response = await apiService.users.sessions()
    sessions.value = response.data?.items || []
  } catch (err) {
    console.error('Failed to fetch sessions:', err)
  }
}

async function saveProfile() {
  saving.value = true
  error.value = null
  message.value = null

  try {
    await apiService.users.update(authStore.user?.id, profile.value)
    message.value = 'Profile updated successfully'
    authStore.user = { ...authStore.user, ...profile.value }
  } catch (err) {
    error.value = err.response?.data?.message || 'Failed to update profile'
  } finally {
    saving.value = false
  }
}

async function changePassword() {
  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    error.value = 'Passwords do not match'
    return
  }

  if (passwordForm.value.newPassword.length < 8) {
    error.value = 'Password must be at least 8 characters'
    return
  }

  saving.value = true
  error.value = null
  message.value = null

  try {
    await apiService.users.changePassword({
      currentPassword: passwordForm.value.currentPassword,
      newPassword: passwordForm.value.newPassword
    })

    message.value = 'Password changed successfully'
    passwordForm.value = {
      currentPassword: '',
      newPassword: '',
      confirmPassword: ''
    }
  } catch (err) {
    error.value = err.response?.data?.message || 'Failed to change password'
  } finally {
    saving.value = false
  }
}

async function toggleTwoFactor() {
  if (twoFactorEnabled.value) {
    // Disable 2FA
    try {
      await apiService.users.disableTwoFactor()
      twoFactorEnabled.value = false
      message.value = 'Two-factor authentication disabled'
    } catch (err) {
      error.value = 'Failed to disable two-factor authentication'
    }
  } else {
    showTwoFactorSetup.value = true
  }
}

async function revokeSession(sessionId) {
  try {
    await apiService.users.revokeSession(sessionId)
    sessions.value = sessions.value.filter(s => s.id !== sessionId)
    message.value = 'Session revoked successfully'
  } catch (err) {
    error.value = 'Failed to revoke session'
  }
}

async function revokeAllOtherSessions() {
  try {
    await apiService.users.revokeAllOtherSessions()
    sessions.value = sessions.value.filter(s => s.current)
    message.value = 'All other sessions revoked'
  } catch (err) {
    error.value = 'Failed to revoke sessions'
  }
}

async function saveNotifications() {
  saving.value = true
  try {
    await apiService.users.updateNotifications(notifications.value)
    message.value = 'Notification preferences saved'
  } catch (err) {
    error.value = 'Failed to save preferences'
  } finally {
    saving.value = false
  }
}

function getInitials(name) {
  if (!name) return '??'
  return name
    .split(' ')
    .map(n => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)
}

function formatDevice(session) {
  const ua = session.userAgent || ''
  if (ua.includes('Mobile')) return 'Mobile'
  if (ua.includes('Tablet')) return 'Tablet'
  return 'Desktop'
}

function formatDate(date) {
  return new Date(date).toLocaleDateString()
}

onMounted(() => {
  fetchProfile()
  fetchSessions()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-display font-bold text-slate-900 dark:text-slate-100">Profile Settings</h1>
      <p class="text-slate-600 dark:text-slate-400">Manage your account and preferences</p>
    </div>

    <!-- Message/Error -->
    <Transition
      enter-active-class="transition ease-out duration-200"
      enter-from-class="opacity-0 translate-y-2"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition ease-in duration-150"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 translate-y-2"
    >
      <div
        v-if="message"
        class="p-4 bg-emerald-50 dark:bg-emerald-900/30 border border-emerald-200 dark:border-emerald-800 rounded-xl flex items-center gap-3"
      >
        <Check class="w-5 h-5 text-emerald-600 dark:text-emerald-400 flex-shrink-0" />
        <p class="text-sm text-emerald-700 dark:text-emerald-400">{{ message }}</p>
      </div>

      <div
        v-if="error"
        class="p-4 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-xl flex items-center gap-3"
      >
        <AlertTriangle class="w-5 h-5 text-red-600 dark:text-red-400 flex-shrink-0" />
        <p class="text-sm text-red-700 dark:text-red-400">{{ error }}</p>
      </div>
    </Transition>

    <div class="flex flex-col lg:flex-row gap-6">
      <!-- Sidebar -->
      <div class="lg:w-64">
        <div class="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-2xl p-4 text-center mb-4">
          <div class="w-20 h-20 bg-gradient-to-br from-primary-500 to-primary-600 rounded-full flex items-center justify-center mx-auto mb-4">
            <span class="text-2xl font-semibold text-white">{{ getInitials(profile.name) }}</span>
          </div>
          <h3 class="font-semibold text-slate-900 dark:text-slate-100">{{ profile.name || 'User' }}</h3>
          <p class="text-sm text-slate-500 dark:text-slate-400">{{ profile.email }}</p>
        </div>

        <div class="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-2xl p-2 space-y-1">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            @click="activeTab = tab.id"
            class="w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all cursor-pointer"
            :class="activeTab === tab.id
              ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400'
              : 'text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-700'"
          >
            <component :is="tab.icon" class="w-5 h-5" />
            <span class="font-medium">{{ tab.name }}</span>
          </button>
        </div>
      </div>

      <!-- Content -->
      <div class="flex-1">
        <div class="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-2xl p-6">
          <!-- Profile Tab -->
          <div v-if="activeTab === 'profile'">
            <h2 class="text-lg font-display font-semibold text-slate-900 dark:text-slate-100 mb-6">Personal Information</h2>

            <div class="space-y-6">
              <div class="grid md:grid-cols-2 gap-6">
                <div>
                  <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">Full Name</label>
                  <div class="relative">
                    <Badge class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                    <input
                      v-model="profile.name"
                      type="text"
                      placeholder="John Doe"
                      class="w-full pl-10 pr-4 py-3 bg-slate-100 dark:bg-slate-700 border-0 rounded-xl focus:ring-2 focus:ring-primary-500 outline-none transition-all"
                    >
                  </div>
                </div>
                <div>
                  <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">Email</label>
                  <div class="relative">
                    <Mail class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                    <input
                      v-model="profile.email"
                      type="email"
                      placeholder="john@example.com"
                      class="w-full pl-10 pr-4 py-3 bg-slate-100 dark:bg-slate-700 border-0 rounded-xl focus:ring-2 focus:ring-primary-500 outline-none transition-all"
                    >
                  </div>
                </div>
              </div>

              <div class="grid md:grid-cols-2 gap-6">
                <div>
                  <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">Phone</label>
                  <div class="relative">
                    <Smartphone class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                    <input
                      v-model="profile.phone"
                      type="tel"
                      placeholder="+1 234 567 890"
                      class="w-full pl-10 pr-4 py-3 bg-slate-100 dark:bg-slate-700 border-0 rounded-xl focus:ring-2 focus:ring-primary-500 outline-none transition-all"
                    >
                  </div>
                </div>
                <div>
                  <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">Location</label>
                  <div class="relative">
                    <Globe class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                    <input
                      v-model="profile.location"
                      type="text"
                      placeholder="San Francisco, CA"
                      class="w-full pl-10 pr-4 py-3 bg-slate-100 dark:bg-slate-700 border-0 rounded-xl focus:ring-2 focus:ring-primary-500 outline-none transition-all"
                    >
                  </div>
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">Bio</label>
                <textarea
                  v-model="profile.bio"
                  rows="3"
                  placeholder="Tell us about yourself..."
                  class="w-full px-4 py-3 bg-slate-100 dark:bg-slate-700 border-0 rounded-xl focus:ring-2 focus:ring-primary-500 outline-none transition-all resize-none"
                ></textarea>
              </div>

              <div class="pt-6 border-t border-slate-200 dark:border-slate-700 flex justify-end gap-3">
                <button
                  @click="fetchProfile"
                  class="px-4 py-2 bg-slate-100 dark:bg-slate-700 hover:bg-slate-200 dark:hover:bg-slate-600 text-slate-700 dark:text-slate-300 rounded-xl transition-colors font-medium cursor-pointer"
                >
                  Cancel
                </button>
                <button
                  @click="saveProfile"
                  :disabled="!hasChanges || saving"
                  class="px-4 py-2 bg-primary-600 hover:bg-primary-700 disabled:bg-slate-300 disabled:cursor-not-allowed text-white rounded-xl transition-colors font-medium cursor-pointer flex items-center gap-2"
                >
                  <div v-if="saving" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                  <span>{{ saving ? 'Saving...' : 'Save Changes' }}</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Security Tab -->
          <div v-if="activeTab === 'security'">
            <h2 class="text-lg font-display font-semibold text-slate-900 dark:text-slate-100 mb-6">Security Settings</h2>

            <div class="space-y-6">
              <!-- Change Password -->
              <div class="p-6 bg-slate-50 dark:bg-slate-900/50 rounded-xl">
                <h3 class="font-medium text-slate-900 dark:text-slate-100 mb-2">Change Password</h3>
                <p class="text-sm text-slate-600 dark:text-slate-400 mb-4">Regular password changes help protect your account</p>

                <div class="space-y-4">
                  <div>
                    <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">Current Password</label>
                    <div class="relative">
                      <Key class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                      <input
                        v-model="passwordForm.currentPassword"
                        type="password"
                        placeholder="••••••••"
                        class="w-full pl-10 pr-4 py-3 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-primary-500 outline-none transition-all"
                      >
                    </div>
                  </div>

                  <div>
                    <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">New Password</label>
                    <div class="relative">
                      <Key class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                      <input
                        v-model="passwordForm.newPassword"
                        type="password"
                        placeholder="••••••••"
                        class="w-full pl-10 pr-4 py-3 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-primary-500 outline-none transition-all"
                      >
                    </div>
                    <!-- Password Strength -->
                    <div v-if="passwordForm.newPassword" class="mt-2">
                      <div class="flex items-center gap-2">
                        <div class="flex-1 h-1.5 bg-slate-200 dark:bg-slate-700 rounded-full overflow-hidden">
                          <div
                            class="h-full transition-all duration-300"
                            :class="passwordStrength.color"
                            :style="{ width: `${passwordStrength.level * 20}%` }"
                          ></div>
                        </div>
                        <span class="text-xs font-medium" :class="`text-${passwordStrength.color.split('-')[1]}-600`">
                          {{ passwordStrength.label }}
                        </span>
                      </div>
                    </div>
                  </div>

                  <div>
                    <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">Confirm New Password</label>
                    <div class="relative">
                      <Key class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                      <input
                        v-model="passwordForm.confirmPassword"
                        type="password"
                        placeholder="••••••••"
                        class="w-full pl-10 pr-4 py-3 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl focus:ring-2 focus:ring-primary-500 outline-none transition-all"
                      >
                    </div>
                  </div>

                  <button
                    @click="changePassword"
                    :disabled="!passwordForm.currentPassword || !passwordForm.newPassword || saving"
                    class="px-4 py-2 bg-primary-600 hover:bg-primary-700 disabled:bg-slate-300 disabled:cursor-not-allowed text-white rounded-xl transition-colors font-medium cursor-pointer"
                  >
                    {{ saving ? 'Updating...' : 'Update Password' }}
                  </button>
                </div>
              </div>

              <!-- Two-Factor Authentication -->
              <div class="p-6 bg-slate-50 dark:bg-slate-900/50 rounded-xl">
                <div class="flex items-start justify-between">
                  <div>
                    <h3 class="font-medium text-slate-900 dark:text-slate-100 mb-2">Two-Factor Authentication</h3>
                    <p class="text-sm text-slate-600 dark:text-slate-400">
                      {{ twoFactorEnabled ? 'Your account is protected with 2FA' : 'Add an extra layer of security to your account' }}
                    </p>
                  </div>
                  <button
                    @click="toggleTwoFactor"
                    class="px-4 py-2 rounded-xl transition-colors font-medium cursor-pointer"
                    :class="twoFactorEnabled
                      ? 'bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 text-slate-700 dark:text-slate-300'
                      : 'bg-primary-600 hover:bg-primary-700 text-white'"
                  >
                    {{ twoFactorEnabled ? 'Disable' : 'Enable' }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Sessions Tab -->
          <div v-if="activeTab === 'sessions'">
            <div class="flex items-center justify-between mb-6">
              <h2 class="text-lg font-display font-semibold text-slate-900 dark:text-slate-100">Active Sessions</h2>
              <button
                v-if="sessions.length > 1"
                @click="revokeAllOtherSessions"
                class="px-3 py-1.5 text-sm bg-red-100 dark:bg-red-900/30 hover:bg-red-200 dark:hover:bg-red-900/50 text-red-700 dark:text-red-400 rounded-lg transition-colors font-medium cursor-pointer"
              >
                Revoke All Others
              </button>
            </div>

            <div class="space-y-3">
              <div
                v-for="session in sessions"
                :key="session.id"
                class="flex items-center gap-4 p-4 bg-slate-50 dark:bg-slate-900/50 rounded-xl"
                :class="{ 'ring-2 ring-primary-500': session.current }"
              >
                <div class="w-10 h-10 bg-slate-200 dark:bg-slate-700 rounded-lg flex items-center justify-center">
                  <Smartphone class="w-5 h-5 text-slate-600 dark:text-slate-400" />
                </div>
                <div class="flex-1 min-w-0">
                  <p class="font-medium text-slate-900 dark:text-slate-100">
                    {{ formatDevice(session) }}
                    <span v-if="session.current" class="ml-2 text-xs bg-primary-100 dark:bg-primary-900/50 text-primary-700 dark:text-primary-300 px-2 py-0.5 rounded-full">Current</span>
                  </p>
                  <p class="text-sm text-slate-500 dark:text-slate-400">
                    {{ session.ipAddress }} • {{ formatDate(session.lastActive) }}
                  </p>
                </div>
                <button
                  v-if="!session.current"
                  @click="revokeSession(session.id)"
                  class="p-2 hover:bg-slate-200 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
                  aria-label="Revoke session"
                >
                  <X class="w-5 h-5 text-slate-400" />
                </button>
              </div>

              <div v-if="sessions.length === 0" class="text-center py-8">
                <Smartphone class="w-12 h-12 text-slate-300 dark:text-slate-600 mx-auto mb-3" />
                <p class="text-slate-500 dark:text-slate-400">No active sessions found</p>
              </div>
            </div>
          </div>

          <!-- Notifications Tab -->
          <div v-if="activeTab === 'notifications'">
            <h2 class="text-lg font-display font-semibold text-slate-900 dark:text-slate-100 mb-6">Notification Preferences</h2>

            <div class="space-y-4">
              <div
                v-for="(setting, key) in {
                  email: 'Email Notifications',
                  security: 'Security Alerts',
                  marketing: 'Marketing Communications',
                  updates: 'Product Updates'
                }"
                :key="key"
                class="flex items-center justify-between p-4 bg-slate-50 dark:bg-slate-900/50 rounded-xl"
              >
                <div>
                  <p class="font-medium text-slate-900 dark:text-slate-100">{{ setting }}</p>
                  <p class="text-sm text-slate-500 dark:text-slate-400">
                    {{
                      key === 'email' ? 'Receive email notifications' :
                      key === 'security' ? 'Important security alerts' :
                      key === 'marketing' ? 'Promotional offers and news' :
                      'New features and improvements'
                    }}
                  </p>
                </div>
                <button
                  @click="notifications[key] = !notifications[key]"
                  class="relative w-12 h-6 rounded-full transition-colors cursor-pointer focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
                  :class="notifications[key] ? 'bg-primary-600' : 'bg-slate-300 dark:bg-slate-600'"
                  role="switch"
                  :aria-checked="notifications[key]"
                >
                  <span
                    class="absolute top-1 left-1 w-4 h-4 bg-white rounded-full shadow transition-transform"
                    :class="{ 'translate-x-6': notifications[key] }"
                  ></span>
                </button>
              </div>
            </div>

            <div class="mt-6 pt-6 border-t border-slate-200 dark:border-slate-700 flex justify-end">
              <button
                @click="saveNotifications"
                :disabled="saving"
                class="px-4 py-2 bg-primary-600 hover:bg-primary-700 disabled:bg-slate-300 disabled:cursor-not-allowed text-white rounded-xl transition-colors font-medium cursor-pointer"
              >
                {{ saving ? 'Saving...' : 'Save Preferences' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
