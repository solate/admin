<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useTenantsStore } from '@/stores/modules/tenants'
import { apiService } from '@/api'
import { useI18n } from '@/locales/composables'
import { User, Mail, Building, Shield, Key, X, Check, Pencil, ChevronLeft, UserCircle, Clock, Calendar } from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()
const tenantsStore = useTenantsStore()
const { t } = useI18n()

const loading = ref(false)
const saving = ref(false)
const user = ref(null)
const error = ref(null)

const isEditMode = computed(() => !!route.params.id)
const pageTitle = computed(() => isEditMode.value ? t('user.detail.edit') : t('user.detail.create'))

// Form state
const formData = ref({
  name: '',
  email: '',
  phone: '',
  tenantId: '',
  role: 'user',
  status: 'active',
  password: '',
  confirmPassword: ''
})

const roleOptions = computed(() => [
  { label: t('user.roles.user'), value: 'user' },
  { label: t('user.roles.admin'), value: 'admin' },
  { label: t('user.roles.auditor'), value: 'auditor' },
  { label: t('user.roles.super_admin'), value: 'super_admin' }
])

const statusOptions = computed(() => [
  { label: t('user.list.status.active'), value: 'active' },
  { label: t('user.list.status.inactive'), value: 'inactive' },
  { label: t('user.list.status.suspended'), value: 'suspended' }
])

const roleBadgeStyles = {
  super_admin: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400',
  admin: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
  auditor: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
  user: 'bg-slate-100 text-slate-700 dark:bg-slate-700 dark:text-slate-400'
}

const statusStyles = {
  active: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
  inactive: 'bg-slate-100 text-slate-700 dark:bg-slate-700 dark:text-slate-400',
  suspended: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
}

// Form validation
const formErrors = computed(() => {
  const errors = {}

  if (!formData.value.name) {
    errors.name = t('user.detail.validation.nameRequired')
  }

  if (!formData.value.email) {
    errors.email = t('user.detail.validation.emailRequired')
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.value.email)) {
    errors.email = t('user.detail.validation.emailInvalid')
  }

  if (!formData.value.tenantId) {
    errors.tenantId = t('user.detail.validation.tenantRequired')
  }

  if (!isEditMode.value) {
    if (!formData.value.password) {
      errors.password = t('user.detail.validation.passwordRequired')
    } else if (formData.value.password.length < 8) {
      errors.password = t('user.detail.validation.passwordMinLength')
    }

    if (formData.value.password !== formData.value.confirmPassword) {
      errors.confirmPassword = t('user.detail.validation.passwordMismatch')
    }
  }

  return errors
})

const isFormValid = computed(() => Object.keys(formErrors.value).length === 0)

async function fetchUser() {
  if (!isEditMode.value) return

  loading.value = true
  error.value = null

  try {
    const response = await apiService.users.getById(route.params.id)
    user.value = response.data

    // Populate form
    formData.value = {
      name: user.value.name || '',
      email: user.value.email || '',
      phone: user.value.phone || '',
      tenantId: user.value.tenantId || '',
      role: user.value.role || 'user',
      status: user.value.status || 'active',
      password: '',
      confirmPassword: ''
    }
  } catch (err) {
    error.value = err.message || t('user.detail.errors.fetchFailed')
    console.error('Error fetching user:', err)
  } finally {
    loading.value = false
  }
}

async function saveUser() {
  if (!isFormValid.value) return

  saving.value = true
  error.value = null

  try {
    const dataToSave = {
      name: formData.value.name,
      email: formData.value.email,
      phone: formData.value.phone,
      tenantId: formData.value.tenantId,
      role: formData.value.role,
      status: formData.value.status
    }

    if (!isEditMode.value) {
      dataToSave.password = formData.value.password
    }

    if (isEditMode.value) {
      await apiService.users.update(route.params.id, dataToSave)
    } else {
      await apiService.users.create(dataToSave)
    }

    router.push({ name: 'users' })
  } catch (err) {
    error.value = err.response?.data?.message || err.message || t('user.detail.errors.saveFailed')
    console.error('Error saving user:', err)
  } finally {
    saving.value = false
  }
}

function goBack() {
  router.push({ name: 'users' })
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

onMounted(() => {
  tenantsStore.fetchActiveTenants()
  fetchUser()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center gap-4">
      <button
        @click="goBack"
        class="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
        aria-label="Go back"
      >
        <ChevronLeft class="w-5 h-5 text-slate-600 dark:text-slate-400" />
      </button>
      <div class="flex-1">
        <h1 class="text-2xl font-display font-bold text-slate-900 dark:text-slate-100">{{ pageTitle }}</h1>
        <p class="text-slate-600 dark:text-slate-400">
          {{ isEditMode ? t('user.detail.editSubtitle') : t('user.detail.createSubtitle') }}
        </p>
      </div>
    </div>

    <!-- Error Message -->
    <div
      v-if="error"
      class="p-4 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-xl"
    >
      <p class="text-sm text-red-700 dark:text-red-400">{{ error }}</p>
    </div>

    <!-- Loading State -->
    <div
      v-if="loading"
      class="flex items-center justify-center py-12"
    >
      <div class="animate-spin w-8 h-8 border-3 border-primary-200 border-t-primary-600 rounded-full"></div>
    </div>

    <!-- Form -->
    <div v-else class="grid lg:grid-cols-3 gap-6">
      <!-- Main Form -->
      <div class="lg:col-span-2 space-y-6">
        <div class="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-2xl p-6">
          <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">{{ t('user.detail.basicInfo') }}</h2>

          <div class="space-y-6">
            <!-- Name -->
            <div>
              <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                {{ t('user.detail.fullName') }} <span class="text-red-500">*</span>
              </label>
              <div class="relative">
                <User class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                <input
                  v-model="formData.name"
                  type="text"
                  :placeholder="t('user.detail.placeholders.name')"
                  class="w-full pl-10 pr-4 py-3 bg-slate-100 dark:bg-slate-700 border-0 rounded-xl focus:ring-2 focus:ring-primary-500 outline-none transition-all"
                  :class="{ 'ring-2 ring-red-500': formErrors.name }"
                >
              </div>
              <p v-if="formErrors.name" class="mt-1 text-sm text-red-600 dark:text-red-400">{{ formErrors.name }}</p>
            </div>

            <!-- Email -->
            <div>
              <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                {{ t('user.detail.email') }} <span class="text-red-500">*</span>
              </label>
              <div class="relative">
                <Mail class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                <input
                  v-model="formData.email"
                  type="email"
                  :placeholder="t('user.detail.placeholders.email')"
                  class="w-full pl-10 pr-4 py-3 bg-slate-100 dark:bg-slate-700 border-0 rounded-xl focus:ring-2 focus:ring-primary-500 outline-none transition-all"
                  :class="{ 'ring-2 ring-red-500': formErrors.email }"
                >
              </div>
              <p v-if="formErrors.email" class="mt-1 text-sm text-red-600 dark:text-red-400">{{ formErrors.email }}</p>
            </div>

            <!-- Phone -->
            <div>
              <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">{{ t('user.detail.phone') }}</label>
              <div class="relative">
                <div class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 flex items-center justify-center">
                  <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
                  </svg>
                </div>
                <input
                  v-model="formData.phone"
                  type="tel"
                  :placeholder="t('user.detail.placeholders.phone')"
                  class="w-full pl-10 pr-4 py-3 bg-slate-100 dark:bg-slate-700 border-0 rounded-xl focus:ring-2 focus:ring-primary-500 outline-none transition-all"
                >
              </div>
            </div>

            <!-- Password (New user only) -->
            <template v-if="!isEditMode">
              <div class="grid md:grid-cols-2 gap-6">
                <div>
                  <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                    {{ t('user.detail.password') }} <span class="text-red-500">*</span>
                  </label>
                  <div class="relative">
                    <Key class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                    <input
                      v-model="formData.password"
                      type="password"
                      :placeholder="t('user.detail.placeholders.password')"
                      class="w-full pl-10 pr-4 py-3 bg-slate-100 dark:bg-slate-700 border-0 rounded-xl focus:ring-2 focus:ring-primary-500 outline-none transition-all"
                      :class="{ 'ring-2 ring-red-500': formErrors.password }"
                    >
                  </div>
                  <p v-if="formErrors.password" class="mt-1 text-sm text-red-600 dark:text-red-400">{{ formErrors.password }}</p>
                </div>

                <div>
                  <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                    {{ t('user.detail.confirmPassword') }} <span class="text-red-500">*</span>
                  </label>
                  <div class="relative">
                    <Key class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                    <input
                      v-model="formData.confirmPassword"
                      type="password"
                      :placeholder="t('user.detail.placeholders.password')"
                      class="w-full pl-10 pr-4 py-3 bg-slate-100 dark:bg-slate-700 border-0 rounded-xl focus:ring-2 focus:ring-primary-500 outline-none transition-all"
                      :class="{ 'ring-2 ring-red-500': formErrors.confirmPassword }"
                    >
                  </div>
                  <p v-if="formErrors.confirmPassword" class="mt-1 text-sm text-red-600 dark:text-red-400">{{ formErrors.confirmPassword }}</p>
                </div>
              </div>
            </template>
          </div>
        </div>

        <!-- Role & Permissions -->
        <div class="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-2xl p-6">
          <h2 class="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-6">{{ t('user.detail.rolePermissions') }}</h2>

          <div class="space-y-6">
            <!-- Tenant -->
            <div>
              <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                {{ t('user.detail.tenant') }} <span class="text-red-500">*</span>
              </label>
              <div class="relative">
                <Building class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 z-10" />
                <select
                  v-model="formData.tenantId"
                  class="w-full pl-10 pr-4 py-3 bg-slate-100 dark:bg-slate-700 border-0 rounded-xl focus:ring-2 focus:ring-primary-500 outline-none appearance-none cursor-pointer"
                  :class="{ 'ring-2 ring-red-500': formErrors.tenantId }"
                >
                  <option value="">{{ t('user.detail.selectTenant') }}</option>
                  <option
                    v-for="tenant in tenantsStore.activeTenants"
                    :key="tenant.id"
                    :value="tenant.id"
                  >
                    {{ tenant.name }}
                  </option>
                </select>
              </div>
              <p v-if="formErrors.tenantId" class="mt-1 text-sm text-red-600 dark:text-red-400">{{ formErrors.tenantId }}</p>
            </div>

            <!-- Role -->
            <div>
              <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                {{ t('user.detail.role') }} <span class="text-red-500">*</span>
              </label>
              <div class="grid grid-cols-2 gap-3">
                <button
                  v-for="role in roleOptions"
                  :key="role.value"
                  @click="formData.role = role.value"
                  class="p-4 rounded-xl border-2 transition-all cursor-pointer text-left"
                  :class="formData.role === role.value
                    ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/30'
                    : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600'"
                >
                  <p class="font-medium text-slate-900 dark:text-slate-100">{{ role.label }}</p>
                  <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
                    {{ t(`user.detail.roleDescriptions.${role.value}`) }}
                  </p>
                </button>
              </div>
            </div>

            <!-- Status -->
            <div v-if="isEditMode">
              <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">{{ t('user.detail.status') }}</label>
              <div class="flex gap-3">
                <button
                  v-for="status in statusOptions"
                  :key="status.value"
                  @click="formData.status = status.value"
                  class="px-4 py-2 rounded-lg border-2 transition-all cursor-pointer"
                  :class="formData.status === status.value
                    ? 'border-current bg-opacity-10'
                    : 'border-slate-200 dark:border-slate-700 hover:border-slate-300'"
                  >
                  <span
                    class="font-medium"
                    :class="formData.status === status.value
                      ? statusStyles[status.value]
                      : 'text-slate-600 dark:text-slate-400'"
                  >
                    {{ status.label }}
                  </span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Sidebar -->
      <div class="space-y-6">
        <!-- User Preview -->
        <div class="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-2xl p-6">
          <h3 class="text-sm font-medium text-slate-500 dark:text-slate-400 mb-4">{{ t('user.detail.preview') }}</h3>

          <div class="flex flex-col items-center text-center">
            <div class="w-20 h-20 bg-gradient-to-br from-primary-500 to-primary-600 rounded-full flex items-center justify-center mb-4">
              <span class="text-2xl font-semibold text-white">{{ getInitials(formData.name) || '??' }}</span>
            </div>

            <p class="font-semibold text-slate-900 dark:text-slate-100">{{ formData.name || t('user.detail.fullName') }}</p>
            <p class="text-sm text-slate-500 dark:text-slate-400">{{ formData.email || t('user.detail.placeholders.email') }}</p>

            <div class="flex gap-2 mt-3">
              <span
                class="px-2.5 py-1 rounded-lg text-xs font-medium"
                :class="roleBadgeStyles[formData.role]"
              >
                {{ roleOptions.value.find(r => r.value === formData.role)?.label }}
              </span>
              <span
                v-if="isEditMode"
                class="px-2.5 py-1 rounded-lg text-xs font-medium"
                :class="statusStyles[formData.status]"
              >
                {{ statusOptions.value.find(s => s.value === formData.status)?.label }}
              </span>
            </div>
          </div>
        </div>

        <!-- Actions -->
        <div class="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-2xl p-6">
          <button
            @click="saveUser"
            :disabled="!isFormValid || saving"
            class="w-full px-4 py-3 bg-primary-600 hover:bg-primary-700 disabled:bg-slate-300 disabled:cursor-not-allowed text-white rounded-xl transition-colors font-medium cursor-pointer flex items-center justify-center gap-2"
          >
            <Check v-if="!saving" class="w-5 h-5" />
            <div v-else class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
            <span>{{ saving ? t('user.detail.saving') : isEditMode ? t('user.detail.saveChanges') : t('user.detail.createUser') }}</span>
          </button>

          <button
            @click="goBack"
            class="w-full px-4 py-3 mt-3 bg-slate-100 dark:bg-slate-700 hover:bg-slate-200 dark:hover:bg-slate-600 text-slate-700 dark:text-slate-300 rounded-xl transition-colors font-medium cursor-pointer"
          >
            {{ t('common.cancel') }}
          </button>
        </div>

        <!-- Account Info (Edit mode only) -->
        <div
          v-if="isEditMode && user"
          class="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-2xl p-6"
        >
          <h3 class="text-sm font-medium text-slate-500 dark:text-slate-400 mb-4">{{ t('user.detail.accountInfo') }}</h3>

          <div class="space-y-3 text-sm">
            <div class="flex items-center gap-2 text-slate-600 dark:text-slate-400">
              <Calendar class="w-4 h-4" />
              <span>{{ t('user.detail.created') }}: {{ new Date(user.createdAt).toLocaleDateString() }}</span>
            </div>
            <div class="flex items-center gap-2 text-slate-600 dark:text-slate-400">
              <Clock class="w-4 h-4" />
              <span>{{ t('user.detail.lastLogin') }}: {{ user.lastLoginAt ? formatDate(user.lastLoginAt) : t('user.list.timeAgo.never') }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
