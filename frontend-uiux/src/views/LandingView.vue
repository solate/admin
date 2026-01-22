<script setup>
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const { t } = useI18n()
const { Home, Cube, Shield, ChartBar, ArrowRight, CheckCircle, Cloud, Building } = icons

const features = [
  {
    nameKey: 'multiTenant',
    icon: Building
  },
  {
    nameKey: 'basicServices',
    icon: Cube
  },
  {
    nameKey: 'security',
    icon: Shield
  },
  {
    nameKey: 'monitoring',
    icon: ChartBar
  },
  {
    nameKey: 'highAvailability',
    icon: Cloud
  },
  {
    nameKey: 'quickDeploy',
    icon: Home
  }
]

const plans = [
  {
    key: 'basic',
    popular: false
  },
  {
    key: 'professional',
    popular: true
  },
  {
    key: 'enterprise',
    popular: false
  }
]

const stats = [
  { value: '500+', labelKey: 'customers' },
  { value: '10万+', labelKey: 'users' },
  { value: '99.9%', labelKey: 'availability' },
  { value: '24/7', labelKey: 'support' }
]

const goToRegister = () => {
  router.push({ name: 'register' })
}

const goToLogin = () => {
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-b from-slate-50 to-white">
    <!-- Navigation -->
    <nav class="fixed top-0 left-0 right-0 z-50 glass">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-16">
          <div class="flex items-center gap-2">
            <component :is="Cube" class="w-8 h-8 text-primary-600" />
            <span class="text-xl font-display font-bold text-slate-900">{{ $t('common.appName') }}</span>
          </div>
          <div class="hidden md:flex items-center gap-8">
            <a href="#features" class="text-slate-600 hover:text-slate-900 transition-colors">{{ $t('landing.features') }}</a>
            <a href="#pricing" class="text-slate-600 hover:text-slate-900 transition-colors">{{ $t('landing.pricingTitle') }}</a>
            <a href="#about" class="text-slate-600 hover:text-slate-900 transition-colors">{{ $t('landing.about') }}</a>
          </div>
          <div class="flex items-center gap-4">
            <button
              @click="goToLogin"
              class="text-slate-600 hover:text-slate-900 transition-colors font-medium"
            >
              {{ $t('landing.login') }}
            </button>
            <button
              @click="goToRegister"
              class="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors font-medium"
            >
              {{ $t('landing.freeTrial') }}
            </button>
          </div>
        </div>
      </div>
    </nav>

    <!-- Hero Section -->
    <section class="pt-32 pb-20 px-4 sm:px-6 lg:px-8">
      <div class="max-w-7xl mx-auto">
        <div class="text-center">
          <h1 class="text-4xl sm:text-5xl lg:text-6xl font-display font-bold text-slate-900 mb-6">
            {{ $t('landing.heroTitle') }}
            <span class="text-transparent bg-clip-text bg-gradient-to-r from-primary-600 to-primary-400">
              {{ $t('landing.heroTitleHighlight') }}
            </span>
          </h1>
          <p class="text-lg sm:text-xl text-slate-600 max-w-3xl mx-auto mb-10">
            {{ $t('landing.heroSubtitle') }}
          </p>
          <div class="flex flex-col sm:flex-row items-center justify-center gap-4">
            <button
              @click="goToRegister"
              class="w-full sm:w-auto px-8 py-4 bg-primary-600 text-white rounded-xl hover:bg-primary-700 transition-all shadow-lg hover:shadow-xl font-medium text-lg cursor-pointer flex items-center justify-center gap-2"
            >
              {{ $t('landing.startFreeTrial') }}
              <component :is="ArrowRight" class="w-5 h-5" />
            </button>
            <button
              @click="goToLogin"
              class="w-full sm:w-auto px-8 py-4 bg-white text-slate-700 rounded-xl hover:bg-slate-50 transition-all border border-slate-200 font-medium text-lg cursor-pointer"
            >
              {{ $t('landing.viewDemo') }}
            </button>
          </div>

          <!-- Stats -->
          <div class="mt-16 grid grid-cols-2 md:grid-cols-4 gap-8">
            <div
              v-for="stat in stats"
              :key="stat.labelKey"
              class="text-center"
            >
              <div class="text-3xl sm:text-4xl font-display font-bold text-primary-600 mb-1">
                {{ stat.value }}
              </div>
              <div class="text-slate-600">{{ $t(`landing.stats.${stat.labelKey}`) }}</div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Product Demo Section -->
    <section class="py-20 px-4 sm:px-6 lg:px-8">
      <div class="max-w-7xl mx-auto">
        <div class="glass-card rounded-3xl p-2 sm:p-4">
          <div class="bg-slate-900 rounded-2xl p-4 sm:p-8 aspect-video flex items-center justify-center">
            <div class="text-center">
              <component :is="ChartBar" class="w-16 h-16 text-primary-400 mx-auto mb-4" />
              <p class="text-slate-400 text-lg">{{ $t('landing.productDemo') }}</p>
              <p class="text-slate-500 text-sm mt-2">{{ $t('landing.clickToView') }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Features Section -->
    <section id="features" class="py-20 px-4 sm:px-6 lg:px-8 bg-white">
      <div class="max-w-7xl mx-auto">
        <div class="text-center mb-16">
          <h2 class="text-3xl sm:text-4xl font-display font-bold text-slate-900 mb-4">
            {{ $t('landing.featuresTitle') }}
          </h2>
          <p class="text-lg text-slate-600 max-w-2xl mx-auto">
            {{ $t('landing.featuresSubtitle') }}
          </p>
        </div>

        <div class="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
          <div
            v-for="feature in features"
            :key="feature.nameKey"
            class="glass-card p-8 rounded-2xl hover:shadow-xl transition-all cursor-pointer group"
          >
            <div class="w-14 h-14 bg-primary-100 rounded-xl flex items-center justify-center mb-6 group-hover:bg-primary-600 transition-colors">
              <component :is="feature.icon" class="w-7 h-7 text-primary-600 group-hover:text-white transition-colors" />
            </div>
            <h3 class="text-xl font-display font-semibold text-slate-900 mb-3">
              {{ $t(`landing.featureNames.${feature.nameKey}`) }}
            </h3>
            <p class="text-slate-600 leading-relaxed">
              {{ $t(`landing.featureDescs.${feature.nameKey}`) }}
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- Pricing Section -->
    <section id="pricing" class="py-20 px-4 sm:px-6 lg:px-8 bg-slate-50">
      <div class="max-w-7xl mx-auto">
        <div class="text-center mb-16">
          <h2 class="text-3xl sm:text-4xl font-display font-bold text-slate-900 mb-4">
            {{ $t('landing.pricingTitle') }}
          </h2>
          <p class="text-lg text-slate-600 max-w-2xl mx-auto">
            {{ $t('landing.pricingSubtitle') }}
          </p>
        </div>

        <div class="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto">
          <div
            v-for="plan in plans"
            :key="plan.key"
            class="relative glass-card rounded-2xl p-8"
            :class="{ 'ring-2 ring-primary-600 shadow-xl': plan.popular }"
          >
            <div
              v-if="plan.popular"
              class="absolute -top-4 left-1/2 -translate-x-1/2 px-4 py-1 bg-primary-600 text-white text-sm font-medium rounded-full"
            >
              {{ $t('landing.mostPopular') }}
            </div>
            <div class="text-center mb-8">
              <h3 class="text-xl font-display font-semibold text-slate-900 mb-2">
                {{ $t(`landing.plans.${plan.key}`) }}
              </h3>
              <p class="text-slate-600 text-sm mb-4">{{ $t(`landing.planDescs.${plan.key}`) }}</p>
              <div class="flex items-baseline justify-center gap-1">
                <span class="text-4xl font-bold text-slate-900">{{ $t(`landing.pricing.${plan.key}.price`) }}</span>
                <span class="text-slate-600">{{ $t('landing.planPeriods.perMonth') }}</span>
              </div>
            </div>
            <ul class="space-y-4 mb-8">
              <li
                v-for="feature in $t(`landing.pricing.${plan.key}.features`)"
                :key="feature"
                class="flex items-center gap-3 text-slate-600"
              >
                <component :is="CheckCircle" class="w-5 h-5 text-primary-600 flex-shrink-0" />
                <span>{{ feature }}</span>
              </li>
            </ul>
            <button
              @click="goToRegister"
              class="w-full py-3 rounded-xl font-medium transition-all cursor-pointer"
              :class="plan.popular
                ? 'bg-primary-600 text-white hover:bg-primary-700'
                : 'bg-white text-slate-700 border border-slate-200 hover:bg-slate-50'"
            >
              {{ $t(`landing.planActions.${plan.key}`) }}
            </button>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA Section -->
    <section class="py-20 px-4 sm:px-6 lg:px-8">
      <div class="max-w-4xl mx-auto">
        <div class="gradient-primary rounded-3xl p-12 text-center text-white">
          <h2 class="text-3xl sm:text-4xl font-display font-bold mb-4">
            {{ $t('landing.ctaTitle') }}
          </h2>
          <p class="text-lg text-white/90 mb-8 max-w-2xl mx-auto">
            {{ $t('landing.ctaSubtitle') }}
          </p>
          <button
            @click="goToRegister"
            class="px-8 py-4 bg-white text-primary-600 rounded-xl hover:bg-slate-50 transition-all font-medium text-lg shadow-lg cursor-pointer"
          >
            {{ $t('landing.ctaButton') }}
          </button>
        </div>
      </div>
    </section>

    <!-- Footer -->
    <footer id="about" class="py-12 px-4 sm:px-6 lg:px-8 bg-slate-900 text-white">
      <div class="max-w-7xl mx-auto">
        <div class="grid md:grid-cols-4 gap-8 mb-8">
          <div>
            <div class="flex items-center gap-2 mb-4">
              <component :is="Cube" class="w-8 h-8 text-primary-400" />
              <span class="text-xl font-display font-bold">{{ $t('common.appName') }}</span>
            </div>
            <p class="text-slate-400 text-sm">
              {{ $t('landing.footer.description') }}
            </p>
          </div>
          <div>
            <h4 class="font-semibold mb-4">{{ $t('landing.footer.product') }}</h4>
            <ul class="space-y-2 text-slate-400 text-sm">
              <li><a href="#" class="hover:text-white transition-colors">{{ $t('landing.footer.links.features') }}</a></li>
              <li><a href="#" class="hover:text-white transition-colors">{{ $t('landing.footer.links.pricing') }}</a></li>
              <li><a href="#" class="hover:text-white transition-colors">{{ $t('landing.footer.links.changelog') }}</a></li>
            </ul>
          </div>
          <div>
            <h4 class="font-semibold mb-4">{{ $t('landing.footer.resources') }}</h4>
            <ul class="space-y-2 text-slate-400 text-sm">
              <li><a href="#" class="hover:text-white transition-colors">{{ $t('landing.footer.links.docs') }}</a></li>
              <li><a href="#" class="hover:text-white transition-colors">{{ $t('landing.footer.links.api') }}</a></li>
              <li><a href="#" class="hover:text-white transition-colors">{{ $t('landing.footer.links.help') }}</a></li>
            </ul>
          </div>
          <div>
            <h4 class="font-semibold mb-4">{{ $t('landing.footer.company') }}</h4>
            <ul class="space-y-2 text-slate-400 text-sm">
              <li><a href="#" class="hover:text-white transition-colors">{{ $t('landing.footer.links.about') }}</a></li>
              <li><a href="#" class="hover:text-white transition-colors">{{ $t('landing.footer.links.contact') }}</a></li>
              <li><a href="#" class="hover:text-white transition-colors">{{ $t('landing.footer.links.privacy') }}</a></li>
            </ul>
          </div>
        </div>
        <div class="border-t border-slate-800 pt-8 text-center text-slate-400 text-sm">
          <p>{{ $t('landing.footer.copyright') }}</p>
        </div>
      </div>
    </footer>
  </div>
</template>
