<template>
  <div class="login-container">
    <!-- 背景装饰 -->
    <div class="login-bg">
      <div class="bg-shape bg-shape-1"></div>
      <div class="bg-shape bg-shape-2"></div>
      <div class="bg-shape bg-shape-3"></div>
    </div>

    <!-- 主题切换按钮 -->
    <div class="theme-toggle">
      <el-button text @click="themeStore.toggleTheme()" class="theme-btn">
        <el-icon :size="20">
          <Moon v-if="themeStore.theme === 'light'" />
          <Sunny v-else />
        </el-icon>
      </el-button>
    </div>

    <!-- 登录表单 -->
    <div class="login-content">
      <div class="login-form">
        <!-- Logo和标题 -->
        <div class="login-header">
          <div class="logo-container">
            <el-icon class="logo-icon" size="48"><Promotion /></el-icon>
          </div>
          <h1 class="system-title">多租户管理系统</h1>
          <p class="system-subtitle">Multi-Tenant Management System</p>
        </div>

        <!-- 表单区域 -->
        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          class="login-form-inner"
          @keyup.enter="onSubmit"
          v-if="loginStep === 'credentials'"
        >
          <el-form-item prop="username">
            <el-input
              v-model="form.username"
              placeholder="请输入用户名"
              size="large"
              clearable
              autocomplete="username"
              name="login_username"
            >
              <template #prefix>
                <el-icon><User /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="请输入密码"
              size="large"
              autocomplete="current-password"
              name="login_password"
            >
              <template #prefix>
                <el-icon><Lock /></el-icon>
              </template>
              <template #suffix>
                <el-icon
                  class="password-toggle"
                  @click="showPassword = !showPassword"
                >
                  <View v-if="!showPassword" />
                  <Hide v-else />
                </el-icon>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item prop="captcha">
            <div class="captcha-container">
              <el-input
                v-model="form.captcha"
                placeholder="请输入验证码"
                size="large"
                clearable
                autocomplete="off"
                name="login_captcha"
              >
                <template #prefix>
                  <el-icon><Picture /></el-icon>
                </template>
              </el-input>
              <div class="captcha-image" @click="loadCaptcha">
                <img v-if="captchaUrl" :src="captchaUrl" alt="验证码" />
                <el-button v-else size="large" :loading="loadingCaptcha">
                  获取验证码
                </el-button>
              </div>
            </div>
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              size="large"
              class="login-button"
              :loading="loading"
              @click="onSubmit"
            >
              <span v-if="!loading">登录系统</span>
              <span v-else>正在登录...</span>
            </el-button>
          </el-form-item>

          <div class="login-options">
            <el-checkbox v-model="rememberMe">记住用户名</el-checkbox>
            <el-button type="text">忘记密码？</el-button>
          </div>
        </el-form>

        <!-- 租户选择区域 -->
        <div v-else-if="loginStep === 'select_tenant'" class="tenant-selector">
          <div class="tenant-selector-header">
            <el-icon class="selector-icon" size="40"><OfficeBuilding /></el-icon>
            <h2>选择租户</h2>
            <p>检测到您属于多个租户，请选择要登录的租户</p>
          </div>

          <div class="tenant-list">
            <div
              v-for="tenant in availableTenants"
              :key="tenant.tenant_id"
              class="tenant-card"
              :class="{ 'is-loading': selectingTenant === tenant.tenant_id }"
              @click="selectTenant(tenant)"
            >
              <div class="tenant-avatar">
                <el-icon :size="28"><OfficeBuilding /></el-icon>
              </div>
              <div class="tenant-info">
                <div class="tenant-name">{{ tenant.tenant_name }}</div>
                <div class="tenant-code">{{ tenant.tenant_code }}</div>
              </div>
              <el-icon v-if="selectingTenant === tenant.tenant_id" class="is-loading-icon">
                <Loading />
              </el-icon>
              <el-icon v-else class="arrow-icon">
                <ArrowRight />
              </el-icon>
            </div>
          </div>

          <el-button text @click="backToLogin" class="back-button">
            <el-icon><ArrowLeft /></el-icon>
            返回登录
          </el-button>
        </div>
      </div>
    </div>

    <!-- 版权信息 -->
    <div class="copyright">
      <p>&copy; 2025 Multi-Tenant Management System. All rights reserved.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance } from 'element-plus'
import { useThemeStore } from '../stores/theme'
import { authApi, type TenantInfo } from '../api'
import { saveTokens, getLastTenantId } from '../utils/token'

type LoginStep = 'credentials' | 'select_tenant'

interface LoginForm {
  username: string
  password: string
  captcha: string
}

const router = useRouter()
const themeStore = useThemeStore()
const formRef = ref<FormInstance>()

// 登录步骤
const loginStep = ref<LoginStep>('credentials')

// 表单数据
const form = ref<LoginForm>({
  username: '',
  password: '',
  captcha: ''
})

// 状态变量
const showPassword = ref(false)
const loading = ref(false)
const loadingCaptcha = ref(false)
const rememberMe = ref(false)
const captchaId = ref('')
const captchaUrl = ref('')

// 租户选择相关
const pendingUserId = ref('')
const availableTenants = ref<TenantInfo[]>([])
const selectingTenant = ref('')

// 表单验证规则
const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 2, max: 50, message: '用户名长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 50, message: '密码长度在 6 到 50 个字符', trigger: 'blur' }
  ],
  captcha: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { min: 4, max: 6, message: '验证码长度在 4 到 6 位', trigger: 'blur' }
  ]
}

onMounted(() => {
  loadCaptcha()

  // 如果URL中有username参数，自动填充
  const usernameParam = router.currentRoute.value.query.username as string
  if (usernameParam) {
    form.value.username = usernameParam
    router.replace({ path: '/login', query: {} })
  }

  // 如果有记住的用户名，自动填充
  const savedUsername = localStorage.getItem('remember_username')
  if (savedUsername) {
    form.value.username = savedUsername
    rememberMe.value = true
  }
})

// 加载验证码
async function loadCaptcha() {
  loadingCaptcha.value = true
  try {
    const res = await authApi.getCaptcha()
    captchaId.value = res.captcha_id
    captchaUrl.value = res.captcha_data
  } catch (error) {
    console.error('获取验证码失败:', error)
    ElMessage.error('获取验证码失败，请稍后重试')
  } finally {
    loadingCaptcha.value = false
  }
}

// 返回登录页面
function backToLogin() {
  loginStep.value = 'credentials'
  pendingUserId.value = ''
  availableTenants.value = []
}

// 选择租户
async function selectTenant(tenant: TenantInfo) {
  if (!pendingUserId.value) return

  selectingTenant.value = tenant.tenant_id

  try {
    const res = await authApi.selectTenant(pendingUserId.value, { tenant_id: tenant.tenant_id })

    // 保存token和租户信息
    saveTokens({
      access_token: res.access_token,
      refresh_token: res.refresh_token,
      user_id: pendingUserId.value,
      current_tenant: res.current_tenant
    })

    // 记住用户名
    if (rememberMe.value) {
      localStorage.setItem('remember_username', form.value.username)
    } else {
      localStorage.removeItem('remember_username')
    }

    ElMessage.success(`登录成功！欢迎来到 ${res.current_tenant.tenant_name}`)

    // 跳转到首页或重定向页面
    const redirect = (router.currentRoute.value.query.redirect as string) || '/'
    router.push(redirect)
  } catch (error: any) {
    console.error('选择租户失败:', error)
    ElMessage.error(error.message || '选择租户失败，请重试')
  } finally {
    selectingTenant.value = ''
  }
}

// 登录提交
async function onSubmit() {
  if (!formRef.value) return

  await formRef.value.validate()
  loading.value = true

  try {
    // 获取上次选择的租户ID
    const lastTenantId = getLastTenantId() || undefined

    const res = await authApi.login({
      username: form.value.username,
      password: form.value.password,
      captcha_id: captchaId.value,
      captcha: form.value.captcha,
      last_tenant_id: lastTenantId
    })

    // 检查是否需要选择租户
    if (res.need_select_tenant) {
      // 显示租户选择界面
      loginStep.value = 'select_tenant'
      pendingUserId.value = res.user_id
      availableTenants.value = res.tenants || []
      ElMessage.info('请选择要登录的租户')
      return
    }

    // 直接登录成功
    if (res.access_token && res.current_tenant) {
      saveTokens({
        access_token: res.access_token,
        refresh_token: res.refresh_token!,
        user_id: res.user_id,
        email: res.email,
        phone: res.phone,
        current_tenant: res.current_tenant
      })

      // 记住用户名
      if (rememberMe.value) {
        localStorage.setItem('remember_username', form.value.username)
      } else {
        localStorage.removeItem('remember_username')
      }

      ElMessage.success('登录成功！欢迎回来')

      // 跳转到首页或重定向页面
      const redirect = (router.currentRoute.value.query.redirect as string) || '/'
      router.push(redirect)
    }
  } catch (error: any) {
    console.error('登录失败:', error)
    // 登录失败，刷新验证码
    loadCaptcha()
    form.value.captcha = ''
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
.login-container {
  position: relative;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: var(--gradient-primary);
  overflow: hidden;

  // 背景装饰
  .login-bg {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    overflow: hidden;
    z-index: 0;

    .bg-shape {
      position: absolute;
      border-radius: 50%;
      background: rgba(255, 255, 255, 0.1);
      animation: float 8s ease-in-out infinite;

      &.bg-shape-1 {
        width: 300px;
        height: 300px;
        top: 10%;
        left: 10%;
        animation-delay: 0s;
      }

      &.bg-shape-2 {
        width: 200px;
        height: 200px;
        top: 60%;
        right: 10%;
        animation-delay: 2s;
      }

      &.bg-shape-3 {
        width: 150px;
        height: 150px;
        bottom: 15%;
        left: 20%;
        animation-delay: 4s;
      }
    }
  }

  // 主题切换按钮
  .theme-toggle {
    position: absolute;
    top: 24px;
    right: 24px;
    z-index: 10;

    .theme-btn {
      color: rgba(255, 255, 255, 0.9);
      background: rgba(255, 255, 255, 0.15);
      backdrop-filter: blur(10px);
      border: 1px solid rgba(255, 255, 255, 0.2);
      border-radius: 50%;
      width: 44px;
      height: 44px;
      padding: 0;

      &:hover {
        background: rgba(255, 255, 255, 0.25);
        color: white;
      }
    }
  }

  // 登录内容区域
  .login-content {
    position: relative;
    z-index: 1;
    width: 100%;
    max-width: 480px;
    padding: 20px;
  }

  .login-form {
    background: var(--bg-white);
    backdrop-filter: blur(20px);
    border-radius: var(--border-radius-xl);
    padding: 40px;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
    border: 1px solid rgba(255, 255, 255, 0.2);

    .login-header {
      text-align: center;
      margin-bottom: 32px;

      .logo-container {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        width: 64px;
        height: 64px;
        background: var(--gradient-primary);
        border-radius: var(--border-radius-large);
        margin-bottom: 16px;

        .logo-icon {
          color: white;
        }
      }

      .system-title {
        font-size: 24px;
        font-weight: 600;
        color: var(--text-primary);
        margin: 0 0 8px 0;
      }

      .system-subtitle {
        font-size: 14px;
        color: var(--text-secondary);
        margin: 0;
      }
    }

    .login-form-inner {
      .el-form-item {
        margin-bottom: 20px;
      }

      .captcha-container {
        display: flex;
        gap: 12px;
        align-items: center;

        .el-input {
          flex: 1;
        }

        .captcha-image {
          width: 120px;
          height: 40px;
          cursor: pointer;
          border-radius: var(--border-radius);
          overflow: hidden;
          border: 1px solid var(--border-base);
          background: var(--bg-light);
          display: flex;
          align-items: center;
          justify-content: center;
          transition: var(--transition-base);

          &:hover {
            border-color: var(--primary-color);
            box-shadow: var(--glow-primary);
          }

          img {
            width: 100%;
            height: 100%;
            object-fit: cover;
          }
        }
      }

      .login-button {
        width: 100%;
        height: 44px;
        font-size: 15px;
        font-weight: 500;
        border-radius: var(--border-radius);
        background: var(--gradient-primary);
        border: none;
        box-shadow: var(--glow-primary);
        transition: var(--transition-base);

        &:hover {
          background: var(--gradient-primary-hover);
          box-shadow: var(--glow-hover);
          transform: translateY(-1px);
        }

        &:active {
          transform: translateY(0);
        }
      }

      .login-options {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 20px;

        .el-button--text {
          color: var(--primary-color);

          &:hover {
            opacity: 0.8;
          }
        }
      }
    }

    // 租户选择器
    .tenant-selector {
      .tenant-selector-header {
        text-align: center;
        margin-bottom: 24px;

        .selector-icon {
          color: var(--primary-color);
          margin-bottom: 12px;
        }

        h2 {
          font-size: 20px;
          font-weight: 600;
          color: var(--text-primary);
          margin: 0 0 8px 0;
        }

        p {
          font-size: 14px;
          color: var(--text-secondary);
          margin: 0;
        }
      }

      .tenant-list {
        display: flex;
        flex-direction: column;
        gap: 12px;
        margin-bottom: 20px;

        .tenant-card {
          display: flex;
          align-items: center;
          padding: 16px;
          border: 1px solid var(--border-base);
          border-radius: var(--border-radius);
          cursor: pointer;
          transition: var(--transition-base);
          background: var(--bg-light);

          &:hover {
            border-color: var(--primary-color);
            background: var(--bg-white);
            box-shadow: var(--shadow-sm);
            transform: translateX(4px);
          }

          &.is-loading {
            opacity: 0.7;
            pointer-events: none;
          }

          .tenant-avatar {
            width: 48px;
            height: 48px;
            border-radius: var(--border-radius);
            background: var(--gradient-primary);
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
            margin-right: 16px;
          }

          .tenant-info {
            flex: 1;

            .tenant-name {
              font-size: 16px;
              font-weight: 500;
              color: var(--text-primary);
              margin-bottom: 4px;
            }

            .tenant-code {
              font-size: 13px;
              color: var(--text-secondary);
            }
          }

          .is-loading-icon {
            animation: rotate 1s linear infinite;
            color: var(--primary-color);
          }

          .arrow-icon {
            color: var(--text-placeholder);
            transition: var(--transition-base);
          }

          &:hover .arrow-icon {
            color: var(--primary-color);
          }
        }
      }

      .back-button {
        width: 100%;
        color: var(--text-secondary);

        &:hover {
          color: var(--primary-color);
        }
      }
    }
  }

  // 版权信息
  .copyright {
    position: absolute;
    bottom: 24px;
    text-align: center;
    color: rgba(255, 255, 255, 0.9);
    font-size: 13px;
    z-index: 1;

    p {
      margin: 0;
    }
  }
}

// 浮动动画
@keyframes float {
  0%, 100% {
    transform: translateY(0) rotate(0deg);
  }
  50% {
    transform: translateY(-30px) rotate(5deg);
  }
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

// 暗色主题适配
[data-theme='dark'] {
  .login-container {
    .login-form {
      background: rgba(30, 41, 59, 0.9);
      border: 1px solid rgba(255, 255, 255, 0.1);
      box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);

      .system-title {
        color: var(--text-primary);
      }

      .system-subtitle {
        color: var(--text-secondary);
      }
    }
  }
}

// 响应式设计
@media (max-width: 480px) {
  .login-container {
    padding: 20px;

    .login-form {
      padding: 30px 20px;
    }

    .theme-toggle {
      top: 16px;
      right: 16px;

      .theme-btn {
        width: 40px;
        height: 40px;
      }
    }
  }
}
</style>
