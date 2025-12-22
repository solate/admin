<template>
  <div class="register-page">
    <el-card class="register-card">
      <h2 class="title">用户注册</h2>
      <el-form :model="form" :rules="rules" ref="formRef" label-position="top" autocomplete="off">
        <el-form-item label="用户名" prop="user_name">
          <el-input 
            v-model="form.user_name" 
            placeholder="请输入用户名（用于登录）"
            autocomplete="off"
            name="register_username"
          />
        </el-form-item>
        
        <el-form-item label="昵称" prop="nick_name">
          <el-input 
            v-model="form.nick_name" 
            placeholder="请输入昵称（用于显示）"
            autocomplete="off"
            name="register_nickname"
          />
        </el-form-item>
        
        <el-form-item label="手机号" prop="phone">
          <el-input 
            v-model="form.phone" 
            placeholder="请输入手机号" 
            maxlength="11"
            autocomplete="off"
            name="register_phone"
          />
        </el-form-item>
        
        <el-form-item label="邮箱" prop="email">
          <el-input 
            v-model="form.email" 
            placeholder="请输入邮箱（可选）"
            autocomplete="off"
            name="register_email"
          />
        </el-form-item>
        
        <el-form-item label="密码" prop="password">
          <el-input 
            v-model="form.password" 
            :type="showPwd ? 'text' : 'password'" 
            placeholder="请输入密码"
            autocomplete="new-password"
            name="register_password"
          >
            <template #suffix>
              <el-icon @click="showPwd = !showPwd" class="clickable">
                <component :is="showPwd ? 'View' : 'Hide'" />
              </el-icon>
            </template>
          </el-input>
        </el-form-item>
        
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input 
            v-model="form.confirmPassword" 
            :type="showConfirmPwd ? 'text' : 'password'" 
            placeholder="请再次输入密码"
            autocomplete="new-password"
            name="register_confirm_password"
          >
            <template #suffix>
              <el-icon @click="showConfirmPwd = !showConfirmPwd" class="clickable">
                <component :is="showConfirmPwd ? 'View' : 'Hide'" />
              </el-icon>
            </template>
          </el-input>
        </el-form-item>
        
        <el-form-item label="性别" prop="sex">
          <el-radio-group v-model="form.sex">
            <el-radio :label="1">男</el-radio>
            <el-radio :label="2">女</el-radio>
            <el-radio :label="0">保密</el-radio>
          </el-radio-group>
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="onSubmit" style="width: 100%;">
            注册
          </el-button>
        </el-form-item>
        
        <el-form-item>
          <el-button text @click="goToLogin" style="width: 100%;">
            已有账号？去登录
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { authApi } from '../api'

interface RegisterForm {
  user_name: string
  nick_name: string
  phone: string
  email: string
  password: string
  confirmPassword: string
  sex: number
}

const router = useRouter()
const formRef = ref()
const loading = ref(false)
const showPwd = ref(false)
const showConfirmPwd = ref(false)

const form = ref<RegisterForm>({
  user_name: '',
  nick_name: '',
  phone: '',
  email: '',
  password: '',
  confirmPassword: '',
  sex: 0
})

// 手机号验证
const validatePhone = (_rule: any, value: string, callback: any) => {
  if (!value) {
    callback(new Error('请输入手机号'))
  } else if (!/^1[3-9]\d{9}$/.test(value)) {
    callback(new Error('请输入正确的手机号'))
  } else {
    callback()
  }
}

// 邮箱验证
const validateEmail = (_rule: any, value: string, callback: any) => {
  if (value && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
    callback(new Error('请输入正确的邮箱'))
  } else {
    callback()
  }
}

// 密码验证
const validatePassword = (_rule: any, value: string, callback: any) => {
  if (!value) {
    callback(new Error('请输入密码'))
  } else if (value.length < 6) {
    callback(new Error('密码长度至少6位'))
  } else {
    callback()
  }
}

// 确认密码验证
const validateConfirmPassword = (_rule: any, value: string, callback: any) => {
  if (!value) {
    callback(new Error('请再次输入密码'))
  } else if (value !== form.value.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules = {
  user_name: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在3到20个字符', trigger: 'blur' }
  ],
  nick_name: [
    { required: true, message: '请输入昵称', trigger: 'blur' },
    { min: 2, max: 20, message: '昵称长度在2到20个字符', trigger: 'blur' }
  ],
  phone: [
    { required: true, validator: validatePhone, trigger: 'blur' }
  ],
  email: [
    { validator: validateEmail, trigger: 'blur' }
  ],
  password: [
    { required: true, validator: validatePassword, trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

async function onSubmit() {
  await formRef.value?.validate()
  loading.value = true
  try {
    const username = form.value.user_name
    
    await authApi.register({
      user_name: form.value.user_name,
      nick_name: form.value.nick_name,
      phone: form.value.phone,
      email: form.value.email || undefined,
      password: form.value.password,
      sex: form.value.sex
    })
    
    ElMessage.success('注册成功，请使用用户名登录')
    
    // 跳转到登录页面，并通过query参数传递用户名提示
    router.push({
      path: '/login',
      query: { username: username }
    })
  } catch (error: any) {
    ElMessage.error(error?.msg || error?.message || '注册失败')
  } finally {
    loading.value = false
  }
}

function goToLogin() {
  router.push('/login')
}
</script>

<style scoped>
.register-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
}

.register-card {
  width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.title {
  text-align: center;
  margin-bottom: 16px;
}

.clickable {
  cursor: pointer;
}
</style>

