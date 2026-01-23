<script setup>
import { useRouter } from 'vue-router'
import { Building, Box, Lock, BarChart3, ArrowRight, Cloud, Home } from 'lucide-vue-next'

const router = useRouter()

const features = [
  {
    name: '多租户管理',
    description: '灵活的多租户架构，支持数据隔离和个性化配置',
    icon: Building
  },
  {
    name: '基础服务集成',
    description: '云存储、消息队列、身份认证等开箱即用的基础服务',
    icon: Box
  },
  {
    name: '数据安全',
    description: '企业级安全防护，包括数据加密、访问控制、审计日志',
    icon: Lock
  },
  {
    name: '实时监控',
    description: '全方位系统监控和数据分析，实时掌握服务状态',
    icon: BarChart3
  },
  {
    name: '高可用架构',
    description: '分布式集群部署，支持弹性扩展，确保99.9%可用性',
    icon: Cloud
  },
  {
    name: '快速部署',
    description: '一键部署和自动化运维，大幅降低运维成本',
    icon: Home
  }
]

const plans = [
  {
    name: '基础版',
    price: '299',
    period: '/月',
    popular: false,
    description: '适合小型团队和初创企业',
    features: ['最多 5 个租户', '基础数据隔离', '核心服务访问', '邮件支持', '99% SLA 保障']
  },
  {
    name: '专业版',
    price: '899',
    period: '/月',
    popular: true,
    description: '适合成长型企业和中等规模团队',
    features: ['最多 20 个租户', '高级数据隔离', '全部服务访问', '优先技术支持', '99.9% SLA 保障', '自定义域名', 'API 访问权限']
  },
  {
    name: '企业版',
    price: '定制',
    period: '',
    popular: false,
    description: '适合大型企业和特殊需求',
    features: ['无限租户数', '私有化部署', '专属客户经理', '7x24 技术支持', '99.95% SLA 保障', '定制功能开发', '源码授权', '培训服务']
  }
]

const stats = [
  { value: '500+', label: '企业客户' },
  { value: '10万+', label: '活跃用户' },
  { value: '99.9%', label: '服务可用性' },
  { value: '24/7', label: '技术支持' }
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
            <Box :size="32" class="text-primary-600" />
            <span class="text-xl font-display font-bold text-slate-900">MultiTenant</span>
          </div>
          <div class="hidden md:flex items-center gap-8">
            <a href="#features" class="text-slate-600 hover:text-slate-900 transition-colors">功能特性</a>
            <a href="#pricing" class="text-slate-600 hover:text-slate-900 transition-colors">价格方案</a>
            <a href="#about" class="text-slate-600 hover:text-slate-900 transition-colors">关于我们</a>
          </div>
          <div class="flex items-center gap-4">
            <button
              @click="goToLogin"
              class="text-slate-600 hover:text-slate-900 transition-colors font-medium"
            >
              登录
            </button>
            <button
              @click="goToRegister"
              class="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors font-medium"
            >
              免费试用
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
            企业级多租户
            <span class="text-transparent bg-clip-text bg-gradient-to-r from-primary-600 to-primary-400">
              SaaS 管理平台
            </span>
          </h1>
          <p class="text-lg sm:text-xl text-slate-600 max-w-3xl mx-auto mb-10">
            一站式多租户管理解决方案，帮助企业快速构建和部署 SaaS 应用。包含基础服务管理和业务模块，让您专注于核心业务创新。
          </p>
          <div class="flex flex-col sm:flex-row items-center justify-center gap-4">
            <button
              @click="goToRegister"
              class="w-full sm:w-auto px-8 py-4 bg-primary-600 text-white rounded-xl hover:bg-primary-700 transition-all shadow-lg hover:shadow-xl font-medium text-lg cursor-pointer flex items-center justify-center gap-2"
            >
              免费开始 14 天试用
              <ArrowRight :size="20" />
            </button>
            <button
              @click="goToLogin"
              class="w-full sm:w-auto px-8 py-4 bg-white text-slate-700 rounded-xl hover:bg-slate-50 transition-all border border-slate-200 font-medium text-lg cursor-pointer"
            >
              查看演示
            </button>
          </div>

          <!-- Stats -->
          <div class="mt-16 grid grid-cols-2 md:grid-cols-4 gap-8">
            <div
              v-for="stat in stats"
              :key="stat.label"
              class="text-center"
            >
              <div class="text-3xl sm:text-4xl font-display font-bold text-primary-600 mb-1">
                {{ stat.value }}
              </div>
              <div class="text-slate-600">{{ stat.label }}</div>
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
              <BarChart3 :size="64" class="text-primary-400 mx-auto mb-4" />
              <p class="text-slate-400 text-lg">产品演示视频</p>
              <p class="text-slate-500 text-sm mt-2">点击查看完整功能演示</p>
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
            强大的功能特性
          </h2>
          <p class="text-lg text-slate-600 max-w-2xl mx-auto">
            完整的多租户管理解决方案，从基础服务到业务模块，全方位满足您的需求
          </p>
        </div>

        <div class="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
          <div
            v-for="feature in features"
            :key="feature.name"
            class="glass-card p-8 rounded-2xl hover:shadow-xl transition-all cursor-pointer group"
          >
            <div class="w-14 h-14 bg-primary-100 rounded-xl flex items-center justify-center mb-6 group-hover:bg-primary-600 transition-colors">
              <component :is="feature.icon" :size="28" class="text-primary-600 group-hover:text-white transition-colors" />
            </div>
            <h3 class="text-xl font-display font-semibold text-slate-900 mb-3">
              {{ feature.name }}
            </h3>
            <p class="text-slate-600 leading-relaxed">
              {{ feature.description }}
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
            灵活的价格方案
          </h2>
          <p class="text-lg text-slate-600 max-w-2xl mx-auto">
            选择适合您企业的方案，随时可以升级
          </p>
        </div>

        <div class="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto">
          <div
            v-for="plan in plans"
            :key="plan.name"
            class="relative glass-card rounded-2xl p-8"
            :class="{ 'ring-2 ring-primary-600 shadow-xl': plan.popular }"
          >
            <div
              v-if="plan.popular"
              class="absolute -top-4 left-1/2 -translate-x-1/2 px-4 py-1 bg-primary-600 text-white text-sm font-medium rounded-full"
            >
              最受欢迎
            </div>
            <div class="text-center mb-8">
              <h3 class="text-xl font-display font-semibold text-slate-900 mb-2">{{ plan.name }}</h3>
              <div class="flex items-baseline justify-center gap-1 mb-2">
                <span class="text-4xl font-display font-bold text-slate-900">{{ plan.price }}</span>
                <span class="text-slate-600">{{ plan.period }}</span>
              </div>
              <p class="text-slate-600 text-sm">{{ plan.description }}</p>
            </div>
            <ul class="space-y-3 mb-8">
              <li
                v-for="feature in plan.features"
                :key="feature"
                class="flex items-start gap-3"
              >
                <ArrowRight :size="16" class="text-primary-600 flex-shrink-0 mt-0.5" />
                <span class="text-slate-700">{{ feature }}</span>
              </li>
            </ul>
            <button
              @click="goToRegister"
              class="w-full px-6 py-3 bg-slate-900 text-white rounded-xl hover:bg-slate-800 transition-colors font-medium"
            >
              {{ plan.popular ? '立即购买' : '开始试用' }}
            </button>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA Section -->
    <section class="py-20 px-4 sm:px-6 lg:px-8 bg-gradient-to-br from-primary-50 to-blue-50">
      <div class="max-w-4xl mx-auto text-center">
        <h2 class="text-3xl sm:text-4xl font-display font-bold text-slate-900 mb-4">
          准备好开始了吗？
        </h2>
        <p class="text-lg text-slate-600 mb-8">
          立即注册，获得 14 天免费试用，无需信用卡
        </p>
        <button
          @click="goToRegister"
          class="px-8 py-4 bg-primary-600 text-white rounded-xl hover:bg-primary-700 transition-all shadow-lg hover:shadow-xl font-medium text-lg"
        >
          免费开始使用
        </button>
      </div>
    </section>

    <!-- Footer -->
    <footer class="py-12 px-4 sm:px-6 lg:px-8 bg-slate-900 text-slate-400">
      <div class="max-w-7xl mx-auto">
        <div class="grid grid-cols-2 md:grid-cols-4 gap-8 mb-8">
          <div>
            <div class="flex items-center gap-2 mb-4">
              <Box :size="24" class="text-primary-400" />
              <span class="text-lg font-semibold text-white">MultiTenant</span>
            </div>
            <p class="text-sm text-slate-400">
              企业级多租户 SaaS 管理平台，帮助企业快速构建和部署应用。
            </p>
          </div>
          <div>
            <h3 class="text-white font-semibold mb-4">产品</h3>
            <ul class="space-y-2 text-sm">
              <li><a href="#features" class="hover:text-white transition-colors">功能特性</a></li>
              <li><a href="#pricing" class="hover:text-white transition-colors">价格方案</a></li>
            </ul>
          </div>
          <div>
            <h3 class="text-white font-semibold mb-4">资源</h3>
            <ul class="space-y-2 text-sm">
              <li><a href="#" class="hover:text-white transition-colors">文档中心</a></li>
              <li><a href="#" class="hover:text-white transition-colors">API 参考</a></li>
            </ul>
          </div>
          <div>
            <h3 class="text-white font-semibold mb-4">公司</h3>
            <ul class="space-y-2 text-sm">
              <li><a href="#" class="hover:text-white transition-colors">关于我们</a></li>
              <li><a href="#" class="hover:text-white transition-colors">联系我们</a></li>
            </ul>
          </div>
        </div>
        <div class="border-t border-slate-800 pt-8 text-center text-sm">
          <p>© 2024 MultiTenant. All rights reserved.</p>
        </div>
      </div>
    </footer>
  </div>
</template>

<style scoped>
/* Glass morphism effect */
.glass {
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.glass-card {
  background: rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
}
</style>
