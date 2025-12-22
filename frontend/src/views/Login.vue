<template>
  <div class="login-page">
    <el-card class="login-card">
      <h2 class="title">后台管理系统</h2>
      <el-form :model="form" :rules="rules" ref="formRef" label-position="top" autocomplete="off">
        <el-form-item label="账号" prop="username">
          <el-input 
            v-model="form.username" 
            placeholder="请输入账号"
            autocomplete="username"
            name="login_username"
          />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input 
            v-model="form.password" 
            :type="showPwd ? 'text' : 'password'" 
            placeholder="请输入密码"
            autocomplete="current-password"
            name="login_password"
          >
            <template #suffix>
              <el-icon @click="showPwd = !showPwd" class="clickable">
                <component :is="showPwd ? 'View' : 'Hide'" />
              </el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="验证码" prop="captcha">
          <div class="captcha-wrapper">
            <el-input 
              v-model="form.captcha" 
              placeholder="请输入验证码" 
              style="flex: 1;"
              autocomplete="off"
              name="login_captcha"
            />
            <img 
              v-if="captchaUrl" 
              :src="captchaUrl" 
              @click="loadCaptcha" 
              class="captcha-img"
              alt="验证码"
            />
            <el-button v-else @click="loadCaptcha" :loading="loadingCaptcha">获取验证码</el-button>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="onSubmit" style="width: 100%;">登录</el-button>
        </el-form-item>
        
        <el-form-item>
          <el-button text @click="goToRegister" style="width: 100%;">还没有账号？去注册</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { authApi } from '../api'
import { saveTokens } from '../utils/token'

interface LoginForm { 
  username: string
  password: string
  captcha: string
}

const router = useRouter()
const formRef = ref()
const form = ref<LoginForm>({ username: '', password: '', captcha: '' })
const showPwd = ref(false)
const loading = ref(false)
const loadingCaptcha = ref(false)
const captchaId = ref('')
const captchaUrl = ref('')

const rules = {
  username: [{ required: true, message: '请输入账号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  captcha: [{ required: true, message: '请输入验证码', trigger: 'blur' }]
}

onMounted(() => {
  loadCaptcha()
  
  // 如果 URL 中有 username 参数，自动填充到表单
  const usernameParam = router.currentRoute.value.query.username as string
  if (usernameParam) {
    form.value.username = usernameParam
    // 清除 URL 中的 username 参数
    router.replace({ path: '/login', query: {} })
  }
})

async function loadCaptcha() {
  loadingCaptcha.value = true
  try {
    const res = await authApi.getCaptcha()
    captchaId.value = res.captcha_id
    captchaUrl.value = res.captcha_url
  } catch (error) {
    ElMessage.error('获取验证码失败')
  } finally {
    loadingCaptcha.value = false
  }
}

async function onSubmit() {
  await formRef.value?.validate()
  loading.value = true
  try {
    const res = await authApi.login({
      username: form.value.username,
      password: form.value.password,
      captcha_id: captchaId.value,
      captcha: form.value.captcha
    })
    
    // 使用 token 管理模块保存 token
    saveTokens({
      access_token: res.access_token,
      refresh_token: res.refresh_token,
      user_id: res.user_id,
      user_name: res.user_name
    })
    
    ElMessage.success('登录成功')
    
    // 检查是否有重定向参数
    const redirect = (router.currentRoute.value.query.redirect as string) || '/'
    router.push(redirect)
  } catch (error) {
    // 登录失败，刷新验证码
    loadCaptcha()
    form.value.captcha = ''
  } finally {
    loading.value = false
  }
}

function goToRegister() {
  router.push('/register')
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
}
.login-card { width: 400px; }
.title { text-align: center; margin-bottom: 16px; }
.clickable { cursor: pointer; }
.captcha-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
}
.captcha-img {
  height: 40px;
  cursor: pointer;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}
</style>


