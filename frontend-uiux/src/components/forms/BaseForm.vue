<!--
基础表单组件
提供统一的表单样式和交互
-->
<script setup lang="ts">
import { ref, provide } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'

interface Props {
  modelValue: Record<string, any>
  rules?: FormRules
  labelWidth?: string | number
  labelPosition?: 'left' | 'right' | 'top'
  disabled?: boolean
  loading?: boolean
}

interface Emits {
  (e: 'update:modelValue', value: Record<string, any>): void
  (e: 'submit', data: Record<string, any>): void
  (e: 'reset'): void
}

const props = withDefaults(defineProps<Props>(), {
  labelWidth: '120px',
  labelPosition: 'right',
  disabled: false,
  loading: false,
})

const emit = defineEmits<Emits>()

const formData = ref(props.modelValue)
const formRef = ref<FormInstance>()

// 监听外部变化
watch(
  () => props.modelValue,
  (newVal) => {
    formData.value = newVal
  },
  { deep: true }
)

// 监听内部变化
watch(
  formData,
  (newVal) => {
    emit('update:modelValue', newVal)
  },
  { deep: true }
)

// 提供表单引用给子组件
provide('formRef', formRef)

/**
 * 提交表单
 */
async function handleSubmit() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    emit('submit', formData.value)
  } catch (error) {
    console.error('表单验证失败:', error)
  }
}

/**
 * 重置表单
 */
function handleReset() {
  formRef.value?.resetFields()
  emit('reset')
}

/**
 * 验证表单
 */
async function validate() {
  return await formRef.value?.validate()
}

/**
 * 清除验证
 */
function clearValidate() {
  formRef.value?.clearValidate()
}

// 暴露方法给父组件
defineExpose({
  validate,
  clearValidate,
  handleSubmit,
  handleReset,
})
</script>

<template>
  <el-form
    ref="formRef"
    :model="formData"
    :rules="rules"
    :label-width="labelWidth"
    :label-position="labelPosition"
    :disabled="disabled"
    @submit.prevent="handleSubmit"
  >
    <slot />
    <slot name="actions" :submit="handleSubmit" :reset="handleReset" :loading="loading">
      <div class="form-actions">
        <el-button type="primary" :loading="loading" @click="handleSubmit">
          提交
        </el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
    </slot>
  </el-form>
</template>

<style scoped>
.form-actions {
  display: flex;
  gap: 12px;
  padding-top: 16px;
}
</style>
