<script setup>
import { ref } from 'vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import BaseInput from '@/components/ui/BaseInput.vue'
import {
  Settings as SettingsIcon,
  Bell,
  Shield,
  CreditCard,
  User,
  Globe,
  Mail,
  Moon,
  Sun,
  Lock,
  Eye,
  EyeOff
} from 'lucide-vue-next'

const activeTab = ref('general')

const tabs = [
  { id: 'general', name: '通用设置', icon: SettingsIcon },
  { id: 'profile', name: '个人资料', icon: User },
  { id: 'notifications', name: '通知设置', icon: Bell },
  { id: 'security', name: '安全设置', icon: Shield },
  { id: 'billing', name: '账单设置', icon: CreditCard }
]

const settings = ref({
  siteName: 'AdminSystem',
  supportEmail: 'support@example.com',
  maxTenants: 1000,
  defaultPlan: 'basic',
  enableRegistration: true,
  maintenanceMode: false,
  language: 'zh-CN',
  timezone: 'Asia/Shanghai'
})

const profileSettings = ref({
  name: 'Admin User',
  email: 'admin@example.com',
  avatar: '',
  bio: '',
  company: 'Example Company',
  position: 'Administrator'
})

const notificationSettings = ref({
  emailNotifications: true,
  pushNotifications: false,
  weeklyReports: true,
  alertThreshold: 90,
  tenantAlerts: true,
  securityAlerts: true
})

const securitySettings = ref({
  twoFactorAuth: false,
  sessionTimeout: 30,
  passwordMinLength: 8,
  requireStrongPassword: true,
  loginNotifications: true,
  ipWhitelist: ''
})

const billingSettings = ref({
  billingCycle: 'monthly',
  paymentMethods: ['alipay', 'wechat'],
  invoiceEmail: 'finance@example.com',
  taxId: '',
  companyAddress: ''
})

const showCurrentPassword = ref(false)
const showNewPassword = ref(false)
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">系统设置</h1>
      <p class="text-slate-600 dark:text-slate-400 mt-1">
        管理平台配置和个人偏好
      </p>
    </div>

    <div class="flex flex-col lg:flex-row gap-6">
      <!-- Sidebar -->
      <div class="lg:w-64">
        <div class="card p-2 space-y-1">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            @click="activeTab = tab.id"
            class="w-full flex items-center gap-3 px-4 py-2.5 rounded-lg transition-all cursor-pointer"
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
        <div class="card p-6">
          <!-- General Settings -->
          <div v-if="activeTab === 'general'">
            <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">通用设置</h2>
            <div class="space-y-5">
              <!-- Site Name -->
              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">站点名称</label>
                <input
                  v-model="settings.siteName"
                  type="text"
                  class="w-full px-4 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
                >
              </div>

              <!-- Support Email -->
              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">支持邮箱</label>
                <input
                  v-model="settings.supportEmail"
                  type="email"
                  class="w-full px-4 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
                >
              </div>

              <!-- Language -->
              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">语言</label>
                <select
                  v-model="settings.language"
                  class="w-full px-4 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
                >
                  <option value="zh-CN">简体中文</option>
                  <option value="en-US">English</option>
                  <option value="ja-JP">日本語</option>
                </select>
              </div>

              <!-- Timezone -->
              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">时区</label>
                <select
                  v-model="settings.timezone"
                  class="w-full px-4 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
                >
                  <option value="Asia/Shanghai">Asia/Shanghai (UTC+8)</option>
                  <option value="Asia/Tokyo">Asia/Tokyo (UTC+9)</option>
                  <option value="America/New_York">America/New_York (UTC-5)</option>
                  <option value="Europe/London">Europe/London (UTC+0)</option>
                </select>
              </div>

              <!-- Max Tenants -->
              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">最大租户数</label>
                <input
                  v-model="settings.maxTenants"
                  type="number"
                  class="w-full px-4 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
                >
              </div>

              <!-- Toggles -->
              <div class="space-y-4 pt-4 border-t border-slate-200 dark:border-slate-700">
                <div class="flex items-center justify-between py-2">
                  <div>
                    <p class="font-medium text-slate-900 dark:text-slate-100">允许用户注册</p>
                    <p class="text-sm text-slate-500 dark:text-slate-400">开启后，新用户可以自行注册账户</p>
                  </div>
                  <button
                    @click="settings.enableRegistration = !settings.enableRegistration"
                    class="relative w-12 h-6 rounded-full transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
                    :class="settings.enableRegistration ? 'bg-primary-600' : 'bg-slate-300 dark:bg-slate-600'"
                  >
                    <span
                      class="absolute top-1 w-4 h-4 bg-white rounded-full transition-transform shadow-sm"
                      :class="settings.enableRegistration ? 'translate-x-7' : 'translate-x-1'"
                    ></span>
                  </button>
                </div>

                <div class="flex items-center justify-between py-2">
                  <div>
                    <p class="font-medium text-slate-900 dark:text-slate-100">维护模式</p>
                    <p class="text-sm text-slate-500 dark:text-slate-400">开启后，只有管理员可以访问平台</p>
                  </div>
                  <button
                    @click="settings.maintenanceMode = !settings.maintenanceMode"
                    class="relative w-12 h-6 rounded-full transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-error-500 focus:ring-offset-2"
                    :class="settings.maintenanceMode ? 'bg-error-600' : 'bg-slate-300 dark:bg-slate-600'"
                  >
                    <span
                      class="absolute top-1 w-4 h-4 bg-white rounded-full transition-transform shadow-sm"
                      :class="settings.maintenanceMode ? 'translate-x-7' : 'translate-x-1'"
                    ></span>
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Profile Settings -->
          <div v-if="activeTab === 'profile'">
            <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">个人资料</h2>
            <div class="space-y-5">
              <!-- Avatar -->
              <div class="flex items-center gap-4">
                <div class="w-20 h-20 rounded-full bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center">
                  <User class="w-10 h-10 text-primary-600 dark:text-primary-400" />
                </div>
                <div>
                  <BaseButton variant="secondary" size="sm">更换头像</BaseButton>
                  <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">支持 JPG, PNG 格式，最大 2MB</p>
                </div>
              </div>

              <!-- Name -->
              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">姓名</label>
                <input
                  v-model="profileSettings.name"
                  type="text"
                  class="w-full px-4 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
                >
              </div>

              <!-- Email -->
              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">邮箱</label>
                <input
                  v-model="profileSettings.email"
                  type="email"
                  class="w-full px-4 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
                >
              </div>

              <!-- Company -->
              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">公司</label>
                <input
                  v-model="profileSettings.company"
                  type="text"
                  class="w-full px-4 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
                >
              </div>

              <!-- Position -->
              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">职位</label>
                <input
                  v-model="profileSettings.position"
                  type="text"
                  class="w-full px-4 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
                >
              </div>

              <!-- Bio -->
              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">个人简介</label>
                <textarea
                  v-model="profileSettings.bio"
                  rows="3"
                  class="w-full px-4 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all resize-none"
                ></textarea>
              </div>
            </div>
          </div>

          <!-- Notification Settings -->
          <div v-if="activeTab === 'notifications'">
            <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">通知设置</h2>
            <div class="space-y-4">
              <div class="flex items-center justify-between py-3">
                <div>
                  <p class="font-medium text-slate-900 dark:text-slate-100">邮件通知</p>
                  <p class="text-sm text-slate-500 dark:text-slate-400">接收重要更新的邮件通知</p>
                </div>
                <button
                  @click="notificationSettings.emailNotifications = !notificationSettings.emailNotifications"
                  class="relative w-12 h-6 rounded-full transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
                  :class="notificationSettings.emailNotifications ? 'bg-primary-600' : 'bg-slate-300 dark:bg-slate-600'"
                >
                  <span
                    class="absolute top-1 w-4 h-4 bg-white rounded-full transition-transform shadow-sm"
                    :class="notificationSettings.emailNotifications ? 'translate-x-7' : 'translate-x-1'"
                  ></span>
                </button>
              </div>

              <div class="flex items-center justify-between py-3">
                <div>
                  <p class="font-medium text-slate-900 dark:text-slate-100">推送通知</p>
                  <p class="text-sm text-slate-500 dark:text-slate-400">接收浏览器推送通知</p>
                </div>
                <button
                  @click="notificationSettings.pushNotifications = !notificationSettings.pushNotifications"
                  class="relative w-12 h-6 rounded-full transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
                  :class="notificationSettings.pushNotifications ? 'bg-primary-600' : 'bg-slate-300 dark:bg-slate-600'"
                >
                  <span
                    class="absolute top-1 w-4 h-4 bg-white rounded-full transition-transform shadow-sm"
                    :class="notificationSettings.pushNotifications ? 'translate-x-7' : 'translate-x-1'"
                  ></span>
                </button>
              </div>

              <div class="flex items-center justify-between py-3">
                <div>
                  <p class="font-medium text-slate-900 dark:text-slate-100">周报</p>
                  <p class="text-sm text-slate-500 dark:text-slate-400">每周发送活动摘要报告</p>
                </div>
                <button
                  @click="notificationSettings.weeklyReports = !notificationSettings.weeklyReports"
                  class="relative w-12 h-6 rounded-full transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
                  :class="notificationSettings.weeklyReports ? 'bg-primary-600' : 'bg-slate-300 dark:bg-slate-600'"
                >
                  <span
                    class="absolute top-1 w-4 h-4 bg-white rounded-full transition-transform shadow-sm"
                    :class="notificationSettings.weeklyReports ? 'translate-x-7' : 'translate-x-1'"
                  ></span>
                </button>
              </div>

              <div class="flex items-center justify-between py-3">
                <div>
                  <p class="font-medium text-slate-900 dark:text-slate-100">租户告警</p>
                  <p class="text-sm text-slate-500 dark:text-slate-400">租户资源使用告警</p>
                </div>
                <button
                  @click="notificationSettings.tenantAlerts = !notificationSettings.tenantAlerts"
                  class="relative w-12 h-6 rounded-full transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
                  :class="notificationSettings.tenantAlerts ? 'bg-primary-600' : 'bg-slate-300 dark:bg-slate-600'"
                >
                  <span
                    class="absolute top-1 w-4 h-4 bg-white rounded-full transition-transform shadow-sm"
                    :class="notificationSettings.tenantAlerts ? 'translate-x-7' : 'translate-x-1'"
                  ></span>
                </button>
              </div>

              <div class="flex items-center justify-between py-3">
                <div>
                  <p class="font-medium text-slate-900 dark:text-slate-100">安全告警</p>
                  <p class="text-sm text-slate-500 dark:text-slate-400">异常登录等安全事件通知</p>
                </div>
                <button
                  @click="notificationSettings.securityAlerts = !notificationSettings.securityAlerts"
                  class="relative w-12 h-6 rounded-full transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
                  :class="notificationSettings.securityAlerts ? 'bg-primary-600' : 'bg-slate-300 dark:bg-slate-600'"
                >
                  <span
                    class="absolute top-1 w-4 h-4 bg-white rounded-full transition-transform shadow-sm"
                    :class="notificationSettings.securityAlerts ? 'translate-x-7' : 'translate-x-1'"
                  ></span>
                </button>
              </div>

              <!-- Alert Threshold -->
              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">告警阈值 (%)</label>
                <input
                  v-model="notificationSettings.alertThreshold"
                  type="number"
                  min="0"
                  max="100"
                  class="w-full px-4 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
                >
              </div>
            </div>
          </div>

          <!-- Security Settings -->
          <div v-if="activeTab === 'security'">
            <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">安全设置</h2>
            <div class="space-y-5">
              <div class="flex items-center justify-between py-3">
                <div>
                  <p class="font-medium text-slate-900 dark:text-slate-100">双因素认证</p>
                  <p class="text-sm text-slate-500 dark:text-slate-400">要求用户使用双因素认证</p>
                </div>
                <button
                  @click="securitySettings.twoFactorAuth = !securitySettings.twoFactorAuth"
                  class="relative w-12 h-6 rounded-full transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
                  :class="securitySettings.twoFactorAuth ? 'bg-primary-600' : 'bg-slate-300 dark:bg-slate-600'"
                >
                  <span
                    class="absolute top-1 w-4 h-4 bg-white rounded-full transition-transform shadow-sm"
                    :class="securitySettings.twoFactorAuth ? 'translate-x-7' : 'translate-x-1'"
                  ></span>
                </button>
              </div>

              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">会话超时 (分钟)</label>
                <input
                  v-model="securitySettings.sessionTimeout"
                  type="number"
                  class="w-full px-4 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
                >
              </div>

              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">最小密码长度</label>
                <input
                  v-model="securitySettings.passwordMinLength"
                  type="number"
                  class="w-full px-4 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
                >
              </div>

              <div class="flex items-center justify-between py-3">
                <div>
                  <p class="font-medium text-slate-900 dark:text-slate-100">强密码要求</p>
                  <p class="text-sm text-slate-500 dark:text-slate-400">密码必须包含大小写字母、数字和特殊字符</p>
                </div>
                <button
                  @click="securitySettings.requireStrongPassword = !securitySettings.requireStrongPassword"
                  class="relative w-12 h-6 rounded-full transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
                  :class="securitySettings.requireStrongPassword ? 'bg-primary-600' : 'bg-slate-300 dark:bg-slate-600'"
                >
                  <span
                    class="absolute top-1 w-4 h-4 bg-white rounded-full transition-transform shadow-sm"
                    :class="securitySettings.requireStrongPassword ? 'translate-x-7' : 'translate-x-1'"
                  ></span>
                </button>
              </div>

              <div class="flex items-center justify-between py-3">
                <div>
                  <p class="font-medium text-slate-900 dark:text-slate-100">登录通知</p>
                  <p class="text-sm text-slate-500 dark:text-slate-400">新设备登录时发送通知</p>
                </div>
                <button
                  @click="securitySettings.loginNotifications = !securitySettings.loginNotifications"
                  class="relative w-12 h-6 rounded-full transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
                  :class="securitySettings.loginNotifications ? 'bg-primary-600' : 'bg-slate-300 dark:bg-slate-600'"
                >
                  <span
                    class="absolute top-1 w-4 h-4 bg-white rounded-full transition-transform shadow-sm"
                    :class="securitySettings.loginNotifications ? 'translate-x-7' : 'translate-x-1'"
                  ></span>
                </button>
              </div>

              <!-- Change Password Section -->
              <div class="pt-4 mt-4 border-t border-slate-200 dark:border-slate-700">
                <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100 mb-4">修改密码</h3>
                <div class="space-y-4">
                  <div class="relative">
                    <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">当前密码</label>
                    <input
                      :type="showCurrentPassword ? 'text' : 'password'"
                      class="w-full px-4 py-2.5 pr-10 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
                    >
                    <button
                      type="button"
                      class="absolute right-3 top-9 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 transition-colors cursor-pointer"
                      @click="showCurrentPassword = !showCurrentPassword"
                    >
                      <Eye v-if="showCurrentPassword" class="w-5 h-5" />
                      <EyeOff v-else class="w-5 h-5" />
                    </button>
                  </div>

                  <div class="relative">
                    <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">新密码</label>
                    <input
                      :type="showNewPassword ? 'text' : 'password'"
                      class="w-full px-4 py-2.5 pr-10 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
                    >
                    <button
                      type="button"
                      class="absolute right-3 top-9 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 transition-colors cursor-pointer"
                      @click="showNewPassword = !showNewPassword"
                    >
                      <Eye v-if="showNewPassword" class="w-5 h-5" />
                      <EyeOff v-else class="w-5 h-5" />
                    </button>
                  </div>

                  <BaseButton variant="secondary" size="sm">修改密码</BaseButton>
                </div>
              </div>
            </div>
          </div>

          <!-- Billing Settings -->
          <div v-if="activeTab === 'billing'">
            <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">账单设置</h2>
            <div class="space-y-5">
              <!-- Billing Cycle -->
              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">账单周期</label>
                <select
                  v-model="billingSettings.billingCycle"
                  class="w-full px-4 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
                >
                  <option value="monthly">每月</option>
                  <option value="quarterly">每季度</option>
                  <option value="annually">每年</option>
                </select>
              </div>

              <!-- Invoice Email -->
              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">发票邮箱</label>
                <input
                  v-model="billingSettings.invoiceEmail"
                  type="email"
                  class="w-full px-4 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
                >
              </div>

              <!-- Tax ID -->
              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">税号</label>
                <input
                  v-model="billingSettings.taxId"
                  type="text"
                  class="w-full px-4 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
                >
              </div>

              <!-- Payment Methods -->
              <div class="pt-4 border-t border-slate-200 dark:border-slate-700">
                <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100 mb-4">支付方式</h3>
                <div class="space-y-3">
                  <label class="flex items-center gap-3 cursor-pointer">
                    <input
                      v-model="billingSettings.paymentMethods"
                      type="checkbox"
                      value="alipay"
                      class="w-5 h-5 rounded border-slate-300 text-primary-600 focus:ring-primary-500"
                    >
                    <span class="text-slate-700 dark:text-slate-300">支付宝</span>
                  </label>
                  <label class="flex items-center gap-3 cursor-pointer">
                    <input
                      v-model="billingSettings.paymentMethods"
                      type="checkbox"
                      value="wechat"
                      class="w-5 h-5 rounded border-slate-300 text-primary-600 focus:ring-primary-500"
                    >
                    <span class="text-slate-700 dark:text-slate-300">微信支付</span>
                  </label>
                  <label class="flex items-center gap-3 cursor-pointer">
                    <input
                      v-model="billingSettings.paymentMethods"
                      type="checkbox"
                      value="bank"
                      class="w-5 h-5 rounded border-slate-300 text-primary-600 focus:ring-primary-500"
                    >
                    <span class="text-slate-700 dark:text-slate-300">银行转账</span>
                  </label>
                  <label class="flex items-center gap-3 cursor-pointer">
                    <input
                      v-model="billingSettings.paymentMethods"
                      type="checkbox"
                      value="card"
                      class="w-5 h-5 rounded border-slate-300 text-primary-600 focus:ring-primary-500"
                    >
                    <span class="text-slate-700 dark:text-slate-300">信用卡</span>
                  </label>
                </div>
              </div>
            </div>
          </div>

          <!-- Save Button -->
          <div class="pt-6 mt-6 border-t border-slate-200 dark:border-slate-700 flex items-center justify-between">
            <span class="text-sm text-slate-500 dark:text-slate-400">上次保存: 2小时前</span>
            <BaseButton variant="primary">
              保存设置
            </BaseButton>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
