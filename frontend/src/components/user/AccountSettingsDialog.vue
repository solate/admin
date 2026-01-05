<template>
  <el-dialog
    v-model="visible"
    title="账号设置"
    width="500px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-tabs v-model="activeTab">
      <!-- 基本信息 -->
      <el-tab-pane label="基本信息" name="info">
        <div class="info-section">
          <div class="info-item">
            <span class="info-label">用户名</span>
            <span class="info-value">{{ userInfo?.user_name || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">昵称</span>
            <span class="info-value">{{ userInfo?.nickname || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">邮箱</span>
            <span class="info-value">{{ userInfo?.email || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">手机号</span>
            <span class="info-value">{{ userInfo?.phone || '-' }}</span>
          </div>
        </div>
      </el-tab-pane>

      <!-- 修改密码 -->
      <el-tab-pane label="修改密码" name="password">
        <el-form
          ref="passwordFormRef"
          :model="passwordForm"
          :rules="passwordRules"
          label-width="100px"
          class="password-form"
        >
          <el-form-item label="当前密码" prop="old_password">
            <el-input
              v-model="passwordForm.old_password"
              type="password"
              placeholder="请输入当前密码"
              show-password
            />
          </el-form-item>
          <el-form-item label="新密码" prop="new_password">
            <el-input
              v-model="passwordForm.new_password"
              type="password"
              placeholder="请输入新密码（至少6位）"
              show-password
            />
          </el-form-item>
          <el-form-item label="确认密码" prop="confirm_password">
            <el-input
              v-model="passwordForm.confirm_password"
              type="password"
              placeholder="请再次输入新密码"
              show-password
            />
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              :loading="submitting"
              @click="handleChangePassword"
              style="width: 100%"
            >
              修改密码
            </el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { ElMessage, type FormInstance } from 'element-plus'
import { authApi } from '@/api'

// Props
interface Props {
  modelValue: boolean
  userInfo?: {
    user_name?: string
    nickname?: string
    email?: string
    phone?: string
  } | null
}

const props = withDefaults(defineProps<Props>(), {
  userInfo: null
})

// Emits
const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const activeTab = ref('info')
const passwordFormRef = ref<FormInstance>()
const submitting = ref(false)

// 修改密码表单
const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

// 密码验证规则
const validateConfirmPassword = (_rule: any, value: any, callback: any) => {
  if (value === '') {
    callback(new Error('请再次输入新密码'))
  } else if (value !== passwordForm.new_password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const passwordRules = {
  old_password: [
    { required: true, message: '请输入当前密码', trigger: 'blur' }
  ],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度至少为6位', trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

// 修改密码
const handleChangePassword = async () => {
  if (!passwordFormRef.value) return

  await passwordFormRef.value.validate()
  submitting.value = true

  try {
    await authApi.changePassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password
    })

    ElMessage.success('密码修改成功，请重新登录')

    // 清空表单并关闭对话框
    resetPasswordForm()
    visible.value = false

    // TODO: 可以在这里触发登出，让用户重新登录
    // setTimeout(() => {
    //   logout()
    // }, 1000)
  } catch (error: any) {
    ElMessage.error(error?.message || '修改密码失败')
  } finally {
    submitting.value = false
  }
}

const handleClose = () => {
  resetPasswordForm()
  activeTab.value = 'info'
}

const resetPasswordForm = () => {
  passwordForm.old_password = ''
  passwordForm.new_password = ''
  passwordForm.confirm_password = ''
  passwordFormRef.value?.clearValidate()
}
</script>

<style scoped lang="scss">
.info-section {
  .info-item {
    display: flex;
    padding: 16px 0;
    border-bottom: 1px solid var(--border-lighter);

    &:last-child {
      border-bottom: none;
    }

    .info-label {
      width: 100px;
      font-weight: 500;
      color: var(--text-secondary);
      flex-shrink: 0;
    }

    .info-value {
      flex: 1;
      color: var(--text-primary);
      font-weight: 400;
    }
  }
}

.password-form {
  padding: 20px 0;
}
</style>
