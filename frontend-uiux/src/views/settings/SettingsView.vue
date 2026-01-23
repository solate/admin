<script setup>
import { ref } from 'vue'
import { Settings as SettingsIcon, Bell, Lock, Wallet, User, Mail, Eye, EyeOff } from 'lucide-vue-next'

const activeTab = ref('general')

const tabs = [
  { id: 'general', name: '通用设置', icon: SettingsIcon },
  { id: 'profile', name: '个人资料', icon: User },
  { id: 'notifications', name: '通知设置', icon: Bell },
  { id: 'security', name: '安全设置', icon: Lock },
  { id: 'billing', name: '账单设置', icon: Wallet }
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
  <div class="space-y-6 p-6">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">系统设置</h1>
      <p class="text-slate-600 dark:text-slate-400 mt-1">
        管理平台配置和个人偏好
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
              <component :is="tab.icon"  />
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
            <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">通用设置</h2>
            <el-form label-width="120px">
              <el-form-item label="站点名称">
                <el-input v-model="settings.siteName" />
              </el-form-item>

              <el-form-item label="支持邮箱">
                <el-input v-model="settings.supportEmail" type="email" />
              </el-form-item>

              <el-form-item label="语言">
                <el-select v-model="settings.language" style="width: 100%">
                  <el-option label="简体中文" value="zh-CN" />
                  <el-option label="English" value="en-US" />
                  <el-option label="日本語" value="ja-JP" />
                </el-select>
              </el-form-item>

              <el-form-item label="时区">
                <el-select v-model="settings.timezone" style="width: 100%">
                  <el-option label="Asia/Shanghai (UTC+8)" value="Asia/Shanghai" />
                  <el-option label="Asia/Tokyo (UTC+9)" value="Asia/Tokyo" />
                  <el-option label="America/New_York (UTC-5)" value="America/New_York" />
                  <el-option label="Europe/London (UTC+0)" value="Europe/London" />
                </el-select>
              </el-form-item>

              <el-form-item label="最大租户数">
                <el-input-number v-model="settings.maxTenants" :min="1" :max="10000" />
              </el-form-item>

              <el-divider />

              <el-form-item label="允许用户注册">
                <el-switch
                  v-model="settings.enableRegistration"
                  active-text="开启后，新用户可以自行注册账户"
                />
              </el-form-item>

              <el-form-item label="维护模式">
                <el-switch
                  v-model="settings.maintenanceMode"
                  active-text="开启后，只有管理员可以访问平台"
                />
              </el-form-item>
            </el-form>
          </div>

          <!-- Profile Settings -->
          <div v-if="activeTab === 'profile'">
            <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">个人资料</h2>
            <el-form label-width="120px">
              <el-form-item label="头像">
                <div class="flex items-center gap-4">
                  <el-avatar :size="80" icon="User" />
                  <el-button size="small">更换头像</el-button>
                </div>
                <div class="text-xs text-slate-500 dark:text-slate-400 mt-1">
                  支持 JPG, PNG 格式，最大 2MB
                </div>
              </el-form-item>

              <el-form-item label="姓名">
                <el-input v-model="profileSettings.name" />
              </el-form-item>

              <el-form-item label="邮箱">
                <el-input v-model="profileSettings.email" type="email" />
              </el-form-item>

              <el-form-item label="公司">
                <el-input v-model="profileSettings.company" />
              </el-form-item>

              <el-form-item label="职位">
                <el-input v-model="profileSettings.position" />
              </el-form-item>

              <el-form-item label="个人简介">
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
            <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">通知设置</h2>
            <el-space direction="vertical" :size="16" style="width: 100%">
              <div class="flex items-center justify-between py-2">
                <div>
                  <div class="font-medium text-slate-900 dark:text-slate-100">邮件通知</div>
                  <div class="text-sm text-slate-500 dark:text-slate-400">接收重要更新的邮件通知</div>
                </div>
                <el-switch v-model="notificationSettings.emailNotifications" />
              </div>

              <el-divider />

              <div class="flex items-center justify-between py-2">
                <div>
                  <div class="font-medium text-slate-900 dark:text-slate-100">推送通知</div>
                  <div class="text-sm text-slate-500 dark:text-slate-400">接收浏览器推送通知</div>
                </div>
                <el-switch v-model="notificationSettings.pushNotifications" />
              </div>

              <el-divider />

              <div class="flex items-center justify-between py-2">
                <div>
                  <div class="font-medium text-slate-900 dark:text-slate-100">周报</div>
                  <div class="text-sm text-slate-500 dark:text-slate-400">每周发送活动摘要报告</div>
                </div>
                <el-switch v-model="notificationSettings.weeklyReports" />
              </div>

              <el-divider />

              <div class="flex items-center justify-between py-2">
                <div>
                  <div class="font-medium text-slate-900 dark:text-slate-100">租户告警</div>
                  <div class="text-sm text-slate-500 dark:text-slate-400">租户资源使用告警</div>
                </div>
                <el-switch v-model="notificationSettings.tenantAlerts" />
              </div>

              <el-divider />

              <div class="flex items-center justify-between py-2">
                <div>
                  <div class="font-medium text-slate-900 dark:text-slate-100">安全告警</div>
                  <div class="text-sm text-slate-500 dark:text-slate-400">异常登录等安全事件通知</div>
                </div>
                <el-switch v-model="notificationSettings.securityAlerts" />
              </div>

              <el-divider />

              <el-form-item label="告警阈值 (%)" label-width="120px">
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
            <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">安全设置</h2>
            <el-form label-width="140px">
              <el-form-item label="双因素认证">
                <el-switch
                  v-model="securitySettings.twoFactorAuth"
                  active-text="要求用户使用双因素认证"
                />
              </el-form-item>

              <el-form-item label="会话超时 (分钟)">
                <el-input-number v-model="securitySettings.sessionTimeout" :min="5" :max="120" />
              </el-form-item>

              <el-form-item label="最小密码长度">
                <el-input-number v-model="securitySettings.passwordMinLength" :min="6" :max="32" />
              </el-form-item>

              <el-form-item label="强密码要求">
                <el-switch
                  v-model="securitySettings.requireStrongPassword"
                  active-text="密码必须包含大小写字母、数字和特殊字符"
                />
              </el-form-item>

              <el-form-item label="登录通知">
                <el-switch
                  v-model="securitySettings.loginNotifications"
                  active-text="新设备登录时发送通知"
                />
              </el-form-item>

              <el-divider />

              <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100 mb-4">修改密码</h3>
              <el-form-item label="当前密码">
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

              <el-form-item label="新密码">
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
                <el-button type="primary">修改密码</el-button>
              </el-form-item>
            </el-form>
          </div>

          <!-- Billing Settings -->
          <div v-if="activeTab === 'billing'">
            <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">账单设置</h2>
            <el-form label-width="120px">
              <el-form-item label="账单周期">
                <el-select v-model="billingSettings.billingCycle" style="width: 100%">
                  <el-option label="每月" value="monthly" />
                  <el-option label="每季度" value="quarterly" />
                  <el-option label="每年" value="annually" />
                </el-select>
              </el-form-item>

              <el-form-item label="发票邮箱">
                <el-input v-model="billingSettings.invoiceEmail" type="email" />
              </el-form-item>

              <el-form-item label="税号">
                <el-input v-model="billingSettings.taxId" />
              </el-form-item>

              <el-divider />

              <h3 class="text-base font-semibold text-slate-900 dark:text-slate-100 mb-4">支付方式</h3>
              <el-checkbox-group v-model="billingSettings.paymentMethods">
                <el-checkbox value="alipay">支付宝</el-checkbox>
                <el-checkbox value="wechat">微信支付</el-checkbox>
                <el-checkbox value="bank">银行转账</el-checkbox>
                <el-checkbox value="card">信用卡</el-checkbox>
              </el-checkbox-group>
            </el-form>
          </div>

          <!-- Save Button -->
          <template #footer>
            <div class="flex items-center justify-between">
              <span class="text-sm text-slate-500 dark:text-slate-400">上次保存: 2小时前</span>
              <el-button type="primary">保存设置</el-button>
            </div>
          </template>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>
