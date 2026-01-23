<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useServicesStore } from '@/stores/modules/services'
import { useI18n } from '@/locales/composables'
import { Box, ChevronLeft, CircleCheck, Cloud, Shield, BarChart3, Mail, CreditCard } from 'lucide-vue-next'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const servicesStore = useServicesStore()

const service = computed(() => {
  return servicesStore.getServiceById(route.params.id)
})

const categoryIcons = {
  storage: Cloud,
  messaging: Box,
  analytics: BarChart3,
  security: Shield,
  communication: Mail,
  payment: CreditCard
}

const statusColors = {
  active: 'bg-green-100 text-green-700',
  maintenance: 'bg-yellow-100 text-yellow-700',
  deprecated: 'bg-red-100 text-red-700'
}

const statusLabels = computed(() => ({
  active: t('service.running'),
  maintenance: t('service.maintenance'),
  deprecated: t('service.deprecated')
}))

const CategoryIcon = computed(() => {
  return service.value ? categoryIcons[service.value.category] || Box : Box
})

const goBack = () => {
  router.push({ name: 'services' })
}
</script>

<template>
  <div v-if="service" class="space-y-6">
    <!-- Header -->
    <div class="flex items-center gap-4">
      <button
        @click="goBack"
        class="p-2 hover:bg-slate-100 rounded-lg transition-colors cursor-pointer"
      >
        <component :is="ChevronLeft" class="w-5 h-5 text-slate-600" />
      </button>
      <div class="flex-1">
        <div class="flex items-center gap-4">
          <div class="w-14 h-14 bg-primary-100 rounded-xl flex items-center justify-center">
            <component :is="CategoryIcon" class="w-7 h-7 text-primary-600" />
          </div>
          <div>
            <h1 class="text-2xl font-display font-bold text-slate-900">{{ service.name }}</h1>
            <p class="text-slate-600 text-sm">{{ service.code }}</p>
          </div>
          <span
            class="px-4 py-1 rounded-full text-sm font-medium"
            :class="statusColors[service.status]"
          >
            {{ statusLabels[service.status] }}
          </span>
        </div>
      </div>
    </div>

    <!-- Description -->
    <div class="glass-card rounded-2xl p-6">
      <h2 class="text-lg font-display font-semibold text-slate-900 mb-2">{{ t('service.detail.description') }}</h2>
      <p class="text-slate-600">{{ service.description }}</p>
    </div>

    <!-- Pricing -->
    <div class="glass-card rounded-2xl p-6">
      <h2 class="text-lg font-display font-semibold text-slate-900 mb-4">{{ t('service.detail.pricing') }}</h2>
      <div class="grid md:grid-cols-3 gap-6">
        <div
          v-for="(price, plan) in service.pricing"
          :key="plan"
          class="p-6 bg-slate-50 rounded-xl"
        >
          <p class="text-sm text-slate-600 mb-1 capitalize">{{ plan }}</p>
          <p class="text-3xl font-bold text-slate-900">¥{{ price }}</p>
          <p class="text-sm text-slate-500">{{ t('service.detail.perMonth') }}</p>
        </div>
      </div>
    </div>

    <!-- Features -->
    <div class="glass-card rounded-2xl p-6">
      <h2 class="text-lg font-display font-semibold text-slate-900 mb-4">{{ t('service.detail.features') }}</h2>
      <div class="grid md:grid-cols-2 gap-4">
        <div
          v-for="(feature, index) in service.features"
          :key="index"
          class="flex items-center gap-3 p-4 bg-slate-50 rounded-xl"
        >
          <component :is="CircleCheck" class="w-5 h-5 text-green-600 flex-shrink-0" />
          <span class="text-slate-700">{{ feature }}</span>
        </div>
      </div>
    </div>

    <!-- Service Information -->
    <div class="grid lg:grid-cols-2 gap-6">
      <div class="glass-card rounded-2xl p-6">
        <h2 class="text-lg font-display font-semibold text-slate-900 mb-4">{{ t('service.detail.serviceInfo') }}</h2>
        <div class="space-y-4">
          <div class="flex justify-between py-3 border-b border-slate-100">
            <span class="text-slate-600">{{ t('service.detail.serviceId') }}</span>
            <span class="font-medium text-slate-900 font-mono text-sm">{{ service.id }}</span>
          </div>
          <div class="flex justify-between py-3 border-b border-slate-100">
            <span class="text-slate-600">{{ t('service.detail.serviceCode') }}</span>
            <span class="font-medium text-slate-900 font-mono text-sm">{{ service.code }}</span>
          </div>
          <div class="flex justify-between py-3 border-b border-slate-100">
            <span class="text-slate-600">{{ t('service.detail.serviceCategory') }}</span>
            <span class="font-medium text-slate-900">{{ service.category }}</span>
          </div>
          <div class="flex justify-between py-3">
            <span class="text-slate-600">{{ t('service.detail.serviceStatus') }}</span>
            <span
              class="px-3 py-1 rounded-full text-xs font-medium"
              :class="statusColors[service.status]"
            >
              {{ statusLabels[service.status] }}
            </span>
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div class="glass-card rounded-2xl p-6">
        <h2 class="text-lg font-display font-semibold text-slate-900 mb-4">{{ t('service.detail.manageActions') }}</h2>
        <div class="space-y-3">
          <button class="w-full py-3 bg-primary-600 text-white rounded-xl hover:bg-primary-700 transition-colors font-medium cursor-pointer">
            {{ t('service.detail.editService') }}
          </button>
          <button class="w-full py-3 bg-white text-slate-700 rounded-xl hover:bg-slate-50 transition-colors border border-slate-200 font-medium cursor-pointer">
            {{ t('service.detail.viewLogs') }}
          </button>
          <button class="w-full py-3 bg-white text-slate-700 rounded-xl hover:bg-slate-50 transition-colors border border-slate-200 font-medium cursor-pointer">
            {{ t('service.detail.viewSubscriptions') }}
          </button>
          <button class="w-full py-3 bg-red-50 text-red-600 rounded-xl hover:bg-red-100 transition-colors font-medium cursor-pointer">
            {{ t('service.detail.disableService') }}
          </button>
        </div>
      </div>
    </div>
  </div>

  <!-- Not Found -->
  <div v-else class="text-center py-12">
    <div class="w-16 h-16 bg-slate-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
      <component :is="Box" class="w-8 h-8 text-slate-400" />
    </div>
    <h2 class="text-xl font-display font-semibold text-slate-900 mb-2">{{ t('service.detail.notFound') }}</h2>
    <p class="text-slate-600 mb-4">{{ t('service.detail.checkId') }}</p>
    <button
      @click="goBack"
      class="px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors font-medium cursor-pointer"
    >
      {{ t('service.detail.backToList') }}
    </button>
  </div>
</template>
