<script setup>
import { useRouter } from 'vue-router'
import { Box } from 'lucide-vue-next'
import { useI18n } from '@/locales/composables'

const { t } = useI18n()
const router = useRouter()

const goHome = () => {
  router.push({ name: 'landing' })
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 flex items-center justify-center p-4">
    <div class="text-center">
      <div class="w-24 h-24 bg-primary-100 rounded-3xl flex items-center justify-center mx-auto mb-6">
        <component :is="Box" class="w-12 h-12 text-primary-600" />
      </div>
      <h1 class="text-6xl font-display font-bold text-slate-900 mb-4">{{ t('common.404') }}</h1>
      <h2 class="text-2xl font-display font-semibold text-slate-900 mb-2">{{ t('common.pageNotFound') }}</h2>
      <p class="text-slate-600 mb-8 max-w-md mx-auto">
        {{ t('common.pageNotFoundDesc') }}
      </p>
      <button
        @click="goHome"
        class="px-8 py-3 bg-primary-600 text-white rounded-xl hover:bg-primary-700 transition-colors font-medium cursor-pointer"
      >
        {{ t('common.backToHome') }}
      </button>
    </div>
  </div>
</template>
