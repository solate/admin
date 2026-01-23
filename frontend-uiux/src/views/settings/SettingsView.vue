<script setup>
import { ref, computed } from 'vue'
import { Settings as SettingsIcon, Bell, Lock, Wallet, User, Eye, EyeOff } from 'lucide-vue-next'
import { useI18n } from '@/locales/composables'

const { t } = useI18n()
const activeTab = ref('general')

const tabs = computed(() => [
  { id: 'general', name: t('settings.tabs.general'), icon: SettingsIcon },
  { id: 'profile', name: t('settings.tabs.profile'), icon: User },
  { id: 'notifications', name: t('settings.tabs.notifications'), icon: Bell },
  { id: 'security', name: t('settings.tabs.security'), icon: Lock },
  { id: 'billing', name: t('settings.tabs.billing'), icon: Wallet }
])

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

const billingOptions = computed(() => [
  { label: t('settings.billing.methods.monthly'), value: 'monthly' },
  { label: t('settings.billing.methods.quarterly'), value: 'quarterly' },
  { label: t('settings.billing.methods.annually'), value: 'annually' }
])

const paymentMethodLabels = {
  alipay: t('common.alipay'),
  wechat: t('common.wechat'),
  bank: t('common.bankTransfer'),
  card: t('common.creditCard')
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">{{ t('settings.title') }}</h1>
      <p class="text-slate-600 dark:text-slate-400 mt-1">
        {{ t('settings.description') }}
      </p>
    </div>

    <el-row :gutter="24">
      <!-- Sidebar -->
      <el-col :lg="6">
        <el-card shadow="never">
          <el-menu
            :default-active="activeTab"
            mode="vertical"
            @select="(key) => activeTab = key"
          >
            <el-menu-item
              v-for="tab in tabs"
              :key="tab.id"
              :index="tab.id"
            >
              <component :is="tab.icon" />
              <span>{{ tab.name }}</span>
            </el-menu-item>
          </el-menu>
        </el-card>
      </el-col>

      <!-- Content -->
      <el-col :lg="18">
        <el-card shadow="never">
          <!-- General Settings -->
          <div v-if="activeTab === 'general'">
            <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">{{ t('settings.general.title') }}</h2>
            <el-form label-width="120px">
              <el-form-item :label="t('settings.general.siteName')">
                <el-input v-model="settings.siteName" />
              </el-form-item>

              <el-form-item :label="t('settings.general.supportEmail')">
                <el-input v-model="settings.supportEmail" type="email" />
              </el-form-item>

              <el-form-item :label="t('settings.general.language')">
                <el-select v-model="settings.language" style="width: 100%">
                  <el-option :label="t('common.language.zhCN')" value="zh-CN" />
                  <el-option :label="t('common.language.enUS')" value="en-US" />
                  <el-option label="日本語" value="ja-JP" />
                </el-select>
              </el-form-item>

              <el-form-item :label="t('settings.general.timezone')">
                <el-select v-model="settings.timezone" style="width: 100%">
                  <el-option label="Asia/Shanghai (UTC+8)" value="Asia/Shanghai" />
                  <el-option label="Asia/Tokyo (UTC+9)" value="Asia/Tokyo" />
                  <el-option label="America/New_York (UTC-5)" value="America/New_York" />
                  <el-option label="Europe/London (UTC+0)" value="Europe/London" />
                </el-select>
              </el-form-item>

              <el-form-item :label="t('settings.general.maxTenants')">
                <el-input-number v-model="settings.maxTenants" :min="1" :max="10000" />
              </el-form-item>

              <el-divider />

              <el-form-item :label="t('settings.general.enableRegistration')">
                <el-switch
                  v-model="settings.enableRegistration"
                  :active-text="t('settings.general.enableRegistrationOn')"
                />
              </el-form-item>

              <el-form-item :label="t('settings.general.maintenanceMode')">
                <el-switch
                  v-model="settings.maintenanceMode"
                  :active-text="t('settings.general.maintenanceModeOn')"
                />
              </el-form-item>
            </el-form>
          </div>

          <!-- Profile Settings -->
          <div v-if="activeTab === 'profile'">
            <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">{{ t('settings.profile.title') }}</h2>
            <el-form label-width="120px">
              <el-form-item :label="t('settings.profile.avatar')">
                <div class="flex items-center gap-4">
                  <el-avatar :size="80" icon="User" />
                  <el-button size="small">{{ t('settings.profile.changeAvatar') }}</el-button>
                </div>
                <div class="text-xs text-slate-500 dark:text-slate-400 mt-1">
                  {{ t('settings.profile.avatarHint') }}
                </div>
              </el-form-item>

              <el-form-item :label="t('settings.profile.name')">
                <el-input v-model="profileSettings.name" />
              </el-form-item>

              <el-form-item :label="t('settings.profile.email')">
                <el-input v-model="profileSettings.email" type="email" />
              </el-form-item>

              <el-form-item :label="t('settings.profile.company')">
                <el-input v-model="profileSettings.company" />
              </el-form-item>

              <el-form-item :label="t('settings.profile.position')">
                <el-input v-model="profileSettings.position" />
              </el-form-item>

              <el-form-item :label="t('settings.profile.bio')">
                <el-input
                  v-model="profileSettings.bio"
                  type="textarea"
                  :rows="3"
                />
              </el-form-item>
            </el-form>
          </div>

          <!-- Notification Settings -->
          <div v-if="activeTab === 'notifications'">
            <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">{{ t('settings.notifications.title') }}</h2>
            <el-space direction="vertical" :size="16" style="width: 100%">
              <div class="flex items-center justify-between py-2">
                <div>
                  <div class="font-medium text-slate-900 dark:text-slate-100">{{ t('settings.notifications.emailNotifications') }}</div>
                  <div class="text-sm text-slate-500 dark:text-slate-400">{{ t('settings.notifications.emailNotificationsDesc') }}</div>
                </div>
                <el-switch v-model="notificationSettings.emailNotifications" />
              </div>

              <el-divider />

              <div class="flex items-center justify-between py-2">
                <div>
                  <div class="font-medium text-slate-900 dark:text-slate-100">{{ t('settings.notifications.pushNotifications') }}</div>
                  <div class="text-sm text-slate-500 dark:text-slate-400">{{ t('settings.notifications.pushNotificationsDesc') }}</div>
                </div>
                <el-switch v-model="notificationSettings.pushNotifications" />
              </div>

              <el-divider />

              <div class="flex items-center justify-between py-2">
                <div>
                  <div class="font-medium text-slate-900 dark:text-slate-100">{{ t('settings.notifications.weeklyReports') }}</div>
                  <div class="text-sm text-slate-500 dark:text-slate-400">{{ t('settings.notifications.weeklyReportsDesc') }}</div>
                </div>
                <el-switch v-model="notificationSettings.weeklyReports" />
              </div>

              <el-divider />

              <div class="flex items-center justify-between py-2">
                <div>
                  <div class="font-medium text-slate-900 dark:text-slate-100">{{ t('settings.notifications.tenantAlerts') }}</div>
                  <div class="text-sm text-slate-500 dark:text-slate-400">{{ t('settings.notifications.tenantAlertsDesc') }}</div>
                </div>
                <el-switch v-model="notificationSettings.tenantAlerts" />
              </div>

              <el-divider />

              <div class="flex items-center justify-between py-2">
                <div>
                  <div class="font-medium text-slate-900 dark:text-slate-100">{{ t('settings.notifications.securityAlerts') }}</div>
                  <div class="text-sm text-slate-500 dark:text-slate-400">{{ t('settings.notifications.securityAlertsDesc') }}</div>
                </div>
                <el-switch v-model="notificationSettings.securityAlerts" />
              </div>

              <el-divider />

              <el-form-item :label="t('settings.notifications.alertThreshold')" label-width="140px">
                <el-input-number
                  v-model="notificationSettings.alertThreshold"
                  :min="0"
                  :max="100"
                />
              </el-form-item>
            </el-space>
          </div>

          <!-- Security Settings -->
          <div v-if="activeTab === 'security'">
            <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">{{ t('settings.security.title') }}</h2>
            <el-form label-width="140px">
              <el-form-item :label="t('settings.security.twoFactorAuth')">
                <el-switch
                  v-model="securitySettings.twoFactorAuth"
                  :active-text="t('settings.security.twoFactorAuthOn')"
                />
              </el-form-item>

              <el-form-item :label="t('settings.security.sessionTimeout')">
                <el-input-number v-model="securitySettings.sessionTimeout" :min="5" :max="120" />
              </el-form-item>

              <el-form-item :label="t('settings.security.passwordMinLength')">
                <el-input-number v-model="securitySettings.passwordMinLength" :min="6" :max="32" />
              </el-form-item>

              <el-form-item :label="t('settings.security.requireStrongPassword')">
                <el-switch
                  v-model="securitySettings.requireStrongPassword"
                  :active-text="t('settings.security.requireStrongPasswordOn')"
                />
              </el-form-item>

              <el-form-item :label="t('settings.security.loginNotifications')">
                <el-switch
                  v-model="securitySettings.loginNotifications"
                  :active-text="t('settings.security.loginNotificationsOn')"
                />
              </el-form-item>

              <el-divider />

              <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100 mb-4">{{ t('settings.security.changePassword') }}</h3>
              <el-form-item :label="t('settings.security.currentPassword')">
                <el-input
                  v-model="securitySettings.currentPassword"
                  :type="showCurrentPassword ? 'text' : 'password'"
                >
                  <template #append>
                    <el-button
                      :icon="showCurrentPassword ? Eye : EyeOff"
                      @click="showCurrentPassword = !showCurrentPassword"
                    />
                  </template>
                </el-input>
              </el-form-item>

              <el-form-item :label="t('settings.security.newPassword')">
                <el-input
                  v-model="securitySettings.newPassword"
                  :type="showNewPassword ? 'text' : 'password'"
                >
                  <template #append>
                    <el-button
                      :icon="showNewPassword ? Eye : EyeOff"
                      @click="showNewPassword = !showNewPassword"
                    />
                  </template>
                </el-input>
              </el-form-item>

              <el-form-item>
                <el-button type="primary">{{ t('settings.security.changePassword') }}</el-button>
              </el-form-item>
            </el-form>
          </div>

          <!-- Billing Settings -->
          <div v-if="activeTab === 'billing'">
            <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">{{ t('settings.billing.title') }}</h2>
            <el-form label-width="120px">
              <el-form-item :label="t('settings.billing.billingCycle')">
                <el-select v-model="billingSettings.billingCycle" style="width: 100%">
                  <el-option
                    v-for="option in billingOptions"
                    :key="option.value"
                    :label="option.label"
                    :value="option.value"
                  />
                </el-select>
              </el-form-item>

              <el-form-item :label="t('settings.billing.invoiceEmail')">
                <el-input v-model="billingSettings.invoiceEmail" type="email" />
              </el-form-item>

              <el-form-item :label="t('settings.billing.taxId')">
                <el-input v-model="billingSettings.taxId" />
              </el-form-item>

              <el-divider />

              <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100 mb-4">{{ t('settings.billing.paymentMethods') }}</h3>
              <el-checkbox-group v-model="billingSettings.paymentMethods">
                <el-checkbox value="alipay">{{ t('common.alipay') }}</el-checkbox>
                <el-checkbox value="wechat">{{ t('common.wechat') }}</el-checkbox>
                <el-checkbox value="bank">{{ t('common.bankTransfer') }}</el-checkbox>
                <el-checkbox value="card">{{ t('common.creditCard') }}</el-checkbox>
              </el-checkbox-group>
            </el-form>
          </div>

          <!-- Save Button -->
          <template #footer>
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-500 dark:text-slate-400">{{ t('settings.lastSaved', { time: '2小时前' }) }}</span>
              <el-button type="primary">{{ t('settings.saveSettings') }}</el-button>
            </div>
          </template>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>
