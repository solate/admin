<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTenantsStore } from '@/stores/modules/tenants'
import { Building, Users, Box, ChevronLeft, Pencil, Trash22, CircleCheck, Mail } from 'lucide-vue-next'
import { useI18n } from '@/locales/composables'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const tenantsStore = useTenantsStore()

const tenant = computed(() => {
  return tenantsStore.getTenantById(route.params.id)
})

const statusColors = {
  active: 'bg-green-100 text-green-700',
  suspended: 'bg-red-100 text-red-700',
  trial: 'bg-yellow-100 text-yellow-700'
}

const statusLabels = {
  active: computed(() => t('tenant.active')),
  suspended: computed(() => t('tenant.suspended')),
  trial: computed(() => t('tenant.trial'))
}

const planLabels = {
  basic: computed(() => t('tenant.planBasic')),
  pro: computed(() => t('tenant.planPro')),
  enterprise: computed(() => t('tenant.planEnterprise'))
}

const goBack = () => {
  router.push({ name: 'tenants' })
}

const editTenant = () => {
  // TODO: Open edit modal
}

const deleteTenant = () => {
  // TODO: Confirm and delete
}
</script>

<template>
  <div v-if="tenant" class="space-y-6">
    <!-- Header -->
    <div class="flex items-center gap-4">
      <button
        @click="goBack"
        class="p-2 hover:bg-slate-100 rounded-lg transition-colors cursor-pointer"
      >
        <component :is="ChevronLeft" class="w-5 h-5 text-slate-600" />
      </button>
      <div class="flex-1">
        <div class="flex items-center gap-3">
          <div class="w-12 h-12 bg-primary-100 rounded-xl flex items-center justify-center">
            <component :is="Building" class="w-6 h-6 text-primary-600" />
          </div>
          <div>
            <h1 class="text-2xl font-display font-bold text-slate-900">{{ tenant.name }}</h1>
            <p class="text-slate-600 text-sm">{{ tenant.domain }}</p>
          </div>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <button
          @click="editTenant"
          class="px-4 py-2 bg-white text-slate-700 rounded-lg hover:bg-slate-50 transition-colors border border-slate-200 font-medium cursor-pointer flex items-center gap-2"
        >
          <component :is="Pencil" class="w-4 h-4" />
          {{ t('common.edit') }}
        </button>
        <button
          @click="deleteTenant"
          class="px-4 py-2 bg-red-50 text-red-600 rounded-lg hover:bg-red-100 transition-colors font-medium cursor-pointer flex items-center gap-2"
        >
          <component :is="Trash2" class="w-4 h-4" />
          {{ t('common.delete') }}
        </button>
      </div>
    </div>

    <!-- Status Banner -->
    <div
      class="p-4 rounded-xl"
      :class="tenant.status === 'active' ? 'bg-green-50 border border-green-200' : 'bg-yellow-50 border border-yellow-200'"
    >
      <div class="flex items-center gap-3">
        <component :is="CircleCheck" class="w-5 h-5" :class="tenant.status === 'active' ? 'text-green-600' : 'text-yellow-600'" />
        <div>
          <p class="font-medium" :class="tenant.status === 'active' ? 'text-green-900' : 'text-yellow-900'">
            {{ tenant.status === 'active' ? t('tenant.statusNormal') : t('tenant.statusAbnormal') }}
          </p>
          <p class="text-sm" :class="tenant.status === 'active' ? 'text-green-700' : 'text-yellow-700'">
            {{ tenant.status === 'active' ? t('tenant.statusNormalDesc') : t('tenant.statusAbnormalDesc') }}
          </p>
        </div>
      </div>
    </div>

    <!-- Stats Grid -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="glass-card p-6 rounded-2xl">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
            <component :is="Users" class="w-6 h-6 text-blue-600" />
          </div>
          <div>
            <p class="text-sm text-slate-600">{{ t('tenant.users') }}</p>
            <p class="text-2xl font-bold text-slate-900">{{ tenant.users }}</p>
            <p class="text-xs text-slate-500">/ {{ tenant.maxUsers }} {{ t('tenant.maxUsers') }}</p>
          </div>
        </div>
      </div>
      <div class="glass-card p-6 rounded-2xl">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center">
            <component :is="Box" class="w-6 h-6 text-purple-600" />
          </div>
          <div>
            <p class="text-sm text-slate-600">{{ t('tenant.subscriptionPlan') }}</p>
            <p class="text-2xl font-bold text-slate-900">{{ planLabels[tenant.plan] }}</p>
          </div>
        </div>
      </div>
      <div class="glass-card p-6 rounded-2xl">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center">
            <component :is="CircleCheck" class="w-6 h-6 text-green-600" />
          </div>
          <div>
            <p class="text-sm text-slate-600">{{ t('tenant.status') }}</p>
            <p class="text-2xl font-bold text-slate-900">{{ statusLabels[tenant.status] }}</p>
          </div>
        </div>
      </div>
      <div class="glass-card p-6 rounded-2xl">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 bg-orange-100 rounded-xl flex items-center justify-center">
            <component :is="Mail" class="w-6 h-6 text-orange-600" />
          </div>
          <div>
            <p class="text-sm text-slate-600">{{ t('tenant.createdAt') }}</p>
            <p class="text-2xl font-bold text-slate-900">{{ tenant.createdAt }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Details -->
    <div class="grid lg:grid-cols-2 gap-6">
      <!-- Tenant Information -->
      <div class="glass-card rounded-2xl p-6">
        <h2 class="text-lg font-display font-semibold text-slate-900 mb-4">{{ t('tenant.tenantInfo') }}</h2>
        <div class="space-y-4">
          <div class="flex justify-between py-3 border-b border-slate-100">
            <span class="text-slate-600">{{ t('tenant.id') }}</span>
            <span class="font-medium text-slate-900 font-mono text-sm">{{ tenant.id }}</span>
          </div>
          <div class="flex justify-between py-3 border-b border-slate-100">
            <span class="text-slate-600">{{ t('tenant.name') }}</span>
            <span class="font-medium text-slate-900">{{ tenant.name }}</span>
          </div>
          <div class="flex justify-between py-3 border-b border-slate-100">
            <span class="text-slate-600">{{ t('tenant.accessDomain') }}</span>
            <span class="font-medium text-slate-900 font-mono text-sm">{{ tenant.domain }}</span>
          </div>
          <div class="flex justify-between py-3 border-b border-slate-100">
            <span class="text-slate-600">{{ t('tenant.subscriptionPlan') }}</span>
            <span class="font-medium text-slate-900">{{ planLabels[tenant.plan] }}</span>
          </div>
          <div class="flex justify-between py-3">
            <span class="text-slate-600">{{ t('common.createdAt') }}</span>
            <span class="font-medium text-slate-900">{{ tenant.createdAt }}</span>
          </div>
        </div>
      </div>

      <!-- Usage Statistics -->
      <div class="glass-card rounded-2xl p-6">
        <h2 class="text-lg font-display font-semibold text-slate-900 mb-4">{{ t('tenant.usageStats') }}</h2>
        <div class="space-y-6">
          <div>
            <div class="flex justify-between mb-2">
              <span class="text-sm text-slate-600">{{ t('tenant.usersUsage') }}</span>
              <span class="text-sm font-medium text-slate-900">{{ tenant.users }} / {{ tenant.maxUsers }}</span>
            </div>
            <div class="w-full bg-slate-200 rounded-full h-2">
              <div
                class="bg-primary-600 h-2 rounded-full transition-all"
                :style="{ width: `${(tenant.users / tenant.maxUsers) * 100}%` }"
              ></div>
            </div>
          </div>
          <div>
            <div class="flex justify-between mb-2">
              <span class="text-sm text-slate-600">{{ t('tenant.storageUsage') }}</span>
              <span class="text-sm font-medium text-slate-900">45 GB / 100 GB</span>
            </div>
            <div class="w-full bg-slate-200 rounded-full h-2">
              <div class="bg-blue-600 h-2 rounded-full" style="width: 45%"></div>
            </div>
          </div>
          <div>
            <div class="flex justify-between mb-2">
              <span class="text-sm text-slate-600">{{ t('tenant.apiCalls') }}</span>
              <span class="text-sm font-medium text-slate-900">78,432 / 100,000</span>
            </div>
            <div class="w-full bg-slate-200 rounded-full h-2">
              <div class="bg-green-600 h-2 rounded-full" style="width: 78%"></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Not Found -->
  <div v-else class="text-center py-12">
    <div class="w-16 h-16 bg-slate-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
      <component :is="Building" class="w-8 h-8 text-slate-400" />
    </div>
    <h2 class="text-xl font-display font-semibold text-slate-900 mb-2">{{ t('tenant.notFound') }}</h2>
    <p class="text-slate-600 mb-4">{{ t('tenant.notFoundDesc') }}</p>
    <button
      @click="goBack"
      class="px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors font-medium cursor-pointer"
    >
      {{ t('tenant.backToList') }}
    </button>
  </div>
</template>
