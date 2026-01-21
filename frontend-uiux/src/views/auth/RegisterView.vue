<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import icons from '@/components/icons/index.js'

const router = useRouter()
const authStore = useAuthStore()

const { Cube, Envelope, User, CheckCircle, ArrowRight } = icons

const form = ref({
  name: '',
  email: '',
  password: '',
  confirmPassword: ''
})

const loading = ref(false)
const error = ref('')

const passwordRequirements = [
  { text: '至少 8 个字符', valid: false },
  { text: '包含大写字母', valid: false },
  { text: '包含小写字母', valid: false },
  { text: '包含数字或特殊字符', valid: false }
]

const handleRegister = async () => {
  error.value = ''

  if (form.value.password !== form.value.confirmPassword) {
    error.value = '两次输入的密码不一致'
    return
  }

  loading.value = true

  try {
    await authStore.register({
      name: form.value.name,
      email: form.value.email,
      password: form.value.password
    })
    router.push({ name: 'dashboard-overview' })
  } catch (e) {
    error.value = '注册失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

const goToLogin = () => {
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 flex items-center justify-center p-4">
    <div class="w-full max-w-md">
      <!-- Logo -->
      <div class="text-center mb-8">
        <div class="flex items-center justify-center gap-2 mb-4">
          <component :is="Cube" class="w-10 h-10 text-primary-600" />
          <span class="text-2xl font-display font-bold text-slate-900">MultiTenant</span>
        </div>
        <p class="text-slate-600">创建您的账户</p>
      </div>

      <!-- Register Form -->
      <div class="glass-card p-8">
        <form @submit.prevent="handleRegister" class="space-y-5">
          <!-- Error Message -->
          <div
            v-if="error"
            class="p-4 bg-red-50 border border-red-200 rounded-xl text-red-600 text-sm"
          >
            {{ error }}
          </div>

          <!-- Name -->
          <div>
            <label for="name" class="block text-sm font-medium text-slate-700 mb-2">
              姓名
            </label>
            <div class="relative">
              <component :is="User" class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
              <input
                id="name"
                v-model="form.name"
                type="text"
                required
                placeholder="张三"
                class="w-full pl-12 pr-4 py-3 bg-white/80 border border-slate-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition-all"
              >
            </div>
          </div>

          <!-- Email -->
          <div>
            <label for="email" class="block text-sm font-medium text-slate-700 mb-2">
              邮箱地址
            </label>
            <div class="relative">
              <component :is="Envelope" class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
              <input
                id="email"
                v-model="form.email"
                type="email"
                required
                placeholder="your@email.com"
                class="w-full pl-12 pr-4 py-3 bg-white/80 border border-slate-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition-all"
              >
            </div>
          </div>

          <!-- Password -->
          <div>
            <label for="password" class="block text-sm font-medium text-slate-700 mb-2">
              密码
            </label>
            <input
              id="password"
              v-model="form.password"
              type="password"
              required
              placeholder="••••••••"
              class="w-full px-4 py-3 bg-white/80 border border-slate-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition-all"
            >
          </div>

          <!-- Confirm Password -->
          <div>
            <label for="confirmPassword" class="block text-sm font-medium text-slate-700 mb-2">
              确认密码
            </label>
            <input
              id="confirmPassword"
              v-model="form.confirmPassword"
              type="password"
              required
              placeholder="••••••••"
              class="w-full px-4 py-3 bg-white/80 border border-slate-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition-all"
            >
          </div>

          <!-- Password Requirements -->
          <div class="bg-slate-50 rounded-xl p-4 space-y-2">
            <p class="text-sm font-medium text-slate-700">密码要求：</p>
            <div
              v-for="(req, index) in passwordRequirements"
              :key="index"
              class="flex items-center gap-2 text-sm"
            >
              <component
                :is="CheckCircle"
                class="w-4 h-4"
                :class="req.valid ? 'text-green-500' : 'text-slate-300'"
              />
              <span :class="req.valid ? 'text-green-600' : 'text-slate-500'">{{ req.text }}</span>
            </div>
          </div>

          <!-- Terms -->
          <div class="flex items-start gap-2">
            <input type="checkbox" required class="w-4 h-4 mt-1 rounded border-slate-300 text-primary-600 focus:ring-primary-500">
            <span class="text-sm text-slate-600">
              我同意
              <a href="#" class="text-primary-600 hover:text-primary-700 font-medium">服务条款</a>
              和
              <a href="#" class="text-primary-600 hover:text-primary-700 font-medium">隐私政策</a>
            </span>
          </div>

          <!-- Submit Button -->
          <button
            type="submit"
            :disabled="loading"
            class="w-full py-3 bg-primary-600 text-white rounded-xl hover:bg-primary-700 transition-all font-medium disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 cursor-pointer"
          >
            <span v-if="!loading">创建账户</span>
            <span v-else>创建中...</span>
            <component v-if="!loading" :is="ArrowRight" class="w-5 h-5" />
          </button>
        </form>

        <!-- Divider -->
        <div class="relative my-6">
          <div class="absolute inset-0 flex items-center">
            <div class="w-full border-t border-slate-200"></div>
          </div>
          <div class="relative flex justify-center text-sm">
            <span class="px-4 bg-white/80 text-slate-500">已有账户？</span>
          </div>
        </div>

        <!-- Login Link -->
        <button
          @click="goToLogin"
          class="w-full py-3 bg-white text-slate-700 rounded-xl hover:bg-slate-50 transition-all border border-slate-200 font-medium cursor-pointer"
        >
          登录
        </button>
      </div>

      <!-- Back to Home -->
      <div class="text-center mt-6">
        <a
          @click="router.push({ name: 'landing' })"
          class="text-slate-600 hover:text-slate-900 text-sm cursor-pointer"
        >
          返回首页
        </a>
      </div>
    </div>
  </div>
</template>
