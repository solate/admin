<script setup>
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/modules/auth'
import { useI18n } from '@/locales/composables'
import {
  Building,
  Mail,
  Lock,
  Eye,
  EyeOff,
  ArrowRight
} from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const { t } = useI18n()

const showPassword = ref(false)
const isLoading = ref(false)
const error = ref('')

const form = reactive({
  email: '',
  password: '',
  remember: false
})

const handleLogin = async () => {
  error.value = ''

  // Basic validation
  if (!form.email) {
    error.value = t('auth.enterEmail')
    return
  }
  if (!form.password) {
    error.value = t('auth.enterPassword')
    return
  }

  isLoading.value = true

  try {
    await authStore.login({
      email: form.email,
      password: form.password
    })

    // Redirect to the page they were trying to access, or dashboard
    const redirect = route.query.redirect || '/dashboard/overview'
    router.push(redirect)
  } catch (err) {
    error.value = err.message || t('auth.loginError')
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 flex flex-col">
    <!-- Header -->
    <div class="flex items-center justify-between px-6 py-4">
      <div class="flex items-center gap-2">
        <div class="w-8 h-8 rounded-lg bg-primary-600 flex items-center justify-center">
          <Building :size="20" class="text-white" />
        </div>
        <span class="text-lg font-semibold text-slate-900 dark:text-slate-100">{{ t('auth.appName') }}</span>
      </div>
      <router-link
        to="/"
        class="text-sm text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 transition-colors cursor-pointer"
      >
        {{ t('auth.backToHome') }}
      </router-link>
    </div>

    <!-- Main Content -->
    <main class="flex-1 flex items-center justify-center px-4 py-12">
      <div class="w-full max-w-md">
        <!-- Logo and Title (Mobile) -->
        <div class="text-center mb-8 sm:hidden">
          <div class="w-16 h-16 rounded-2xl bg-primary-600 flex items-center justify-center mx-auto mb-4">
            <Building :size="40" class="text-white" />
          </div>
          <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">{{ t('auth.appName') }}</h1>
        </div>

        <!-- Login Card -->
        <div class="card p-8 sm:p-10">
          <div class="text-center mb-8">
            <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100 mb-2">
              {{ t('auth.welcomeBack') }}
            </h1>
            <p class="text-slate-600 dark:text-slate-400">
              {{ t('auth.loginToPlatform') }}
            </p>
          </div>

          <!-- Error Message -->
          <Transition
            enter-active-class="transition-all duration-200"
            enter-from-class="opacity-0 -translate-y-2"
            enter-to-class="opacity-100 translate-y-0"
            leave-active-class="transition-all duration-200"
            leave-from-class="opacity-100 translate-y-0"
            leave-to-class="opacity-0 -translate-y-2"
          >
            <div
              v-if="error"
              class="mb-6 p-4 bg-error-50 dark:bg-error-900/20 border border-error-200 dark:border-error-800 rounded-lg flex items-start gap-3"
            >
              <div class="w-5 h-5 rounded-full bg-error-500 flex items-center justify-center flex-shrink-0 mt-0.5">
                <span class="text-white text-xs font-bold">!</span>
              </div>
              <p class="text-sm text-error-700 dark:text-error-300">
                {{ error }}
              </p>
            </div>
          </Transition>

          <!-- Login Form -->
          <form @submit.prevent="handleLogin" class="space-y-5">
            <!-- Email Field -->
            <div>
              <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                {{ t('auth.emailAddress') }}
              </label>
              <div class="relative">
                <Mail :size="20" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                <input
                  v-model="form.email"
                  type="email"
                  placeholder="your@email.com"
                  autocomplete="email"
                  :disabled="isLoading"
                  class="w-full pl-10 pr-4 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 placeholder:text-slate-400 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                />
              </div>
            </div>

            <!-- Password Field -->
            <div>
              <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                {{ t('auth.password') }}
              </label>
              <div class="relative">
                <Lock :size="20" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                <input
                  v-model="form.password"
                  :type="showPassword ? 'text' : 'password'"
                  placeholder="••••••••"
                  autocomplete="current-password"
                  :disabled="isLoading"
                  class="w-full pl-10 pr-12 py-2.5 bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded-lg text-slate-900 dark:text-slate-100 placeholder:text-slate-400 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                />
                <button
                  type="button"
                  :disabled="isLoading"
                  class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 transition-colors cursor-pointer disabled:cursor-not-allowed disabled:opacity-50"
                  @click="showPassword = !showPassword"
                >
                  <Eye v-if="showPassword" :size="20" /><EyeOff v-else :size="20" />
                </button>
              </div>
            </div>

            <!-- Remember & Forgot Password -->
            <div class="flex items-center justify-between">
              <label class="flex items-center gap-2 cursor-pointer">
                <input
                  v-model="form.remember"
                  type="checkbox"
                  :disabled="isLoading"
                  class="w-4 h-4 rounded border-slate-300 text-primary-600 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
                />
                <span class="text-sm text-slate-600 dark:text-slate-400">{{ t('auth.rememberMe') }}</span>
              </label>
              <router-link
                to="/forgot-password"
                class="text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300 transition-colors cursor-pointer"
              >
                {{ t('auth.forgotPassword') }}
              </router-link>
            </div>

            <!-- Submit Button -->
            <button
              type="submit"
              :disabled="isLoading"
              class="w-full flex items-center justify-center gap-2 px-4 py-2.5 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-all focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
            >
              <svg
                v-if="isLoading"
                class="animate-spin h-5 w-5"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                />
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                />
              </svg>
              <span v-else>{{ t('auth.login') }}</span>
              <ArrowRight v-if="!isLoading" :size="20" />
            </button>
          </form>

          <!-- Divider -->
          <div class="relative my-8">
            <div class="absolute inset-0 flex items-center">
              <div class="w-full border-t border-slate-200 dark:border-slate-700" />
            </div>
            <div class="relative flex justify-center text-sm">
              <span class="px-4 bg-white dark:bg-slate-800 text-slate-500 dark:text-slate-400">
                {{ t('auth.or') }}
              </span>
            </div>
          </div>

          <!-- Register Link -->
          <p class="text-center text-sm text-slate-600 dark:text-slate-400">
            {{ t('auth.noAccount') }}
            <router-link
              to="/register"
              class="font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300 transition-colors cursor-pointer"
            >
              {{ t('auth.createAccount') }}
            </router-link>
          </p>
        </div>

        <!-- Footer Info -->
        <p class="mt-8 text-center text-xs text-slate-500 dark:text-slate-400">
          {{ t('auth.agreeToTerms') }}
          <a href="#" class="text-primary-600 hover:text-primary-700 dark:text-primary-400 transition-colors cursor-pointer">{{ t('auth.termsOfService') }}</a>
          {{ t('auth.and') }}
          <a href="#" class="text-primary-600 hover:text-primary-700 dark:text-primary-400 transition-colors cursor-pointer">{{ t('auth.privacyPolicy') }}</a>
        </p>
      </div>
    </main>
  </div>
</template>
