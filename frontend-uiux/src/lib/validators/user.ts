/**
 * 用户相关验证器
 */

import { REGEX_PATTERNS } from '@/config/constants'

/**
 * 验证邮箱格式
 */
export function validateEmail(email: string): boolean {
  return REGEX_PATTERNS.EMAIL.test(email)
}

/**
 * 验证手机号（中国大陆）
 */
export function validatePhone(phone: string): boolean {
  return REGEX_PATTERNS.PHONE_CN.test(phone)
}

/**
 * 验证用户名格式
 * 规则：字母开头，允许字母数字下划线，4-16位
 */
export function validateUsername(username: string): boolean {
  return REGEX_PATTERNS.USERNAME.test(username)
}

/**
 * 验证密码强度
 * 规则：至少8位，包含字母和数字
 */
export function validatePassword(password: string): {
  isValid: boolean
  strength: 'weak' | 'medium' | 'strong'
  message?: string
} {
  if (!REGEX_PATTERNS.PASSWORD.test(password)) {
    return {
      isValid: false,
      strength: 'weak',
      message: '密码至少8位，包含字母和数字',
    }
  }

  // 计算密码强度
  let strength: 'weak' | 'medium' | 'strong' = 'weak'
  const hasUpperCase = /[A-Z]/.test(password)
  const hasLowerCase = /[a-z]/.test(password)
  const hasNumber = /\d/.test(password)
  const hasSpecial = /[@$!%*#?&]/.test(password)

  const varietyCount = [hasUpperCase, hasLowerCase, hasNumber, hasSpecial].filter(
    Boolean
  ).length

  if (varietyCount >= 3 && password.length >= 12) {
    strength = 'strong'
  } else if (varietyCount >= 2 && password.length >= 8) {
    strength = 'medium'
  }

  return {
    isValid: true,
    strength,
  }
}

/**
 * 验证用户数据
 */
export function validateUserData(data: {
  username?: string
  email?: string
  phone?: string
  password?: string
}): {
  isValid: boolean
  errors: Record<string, string>
} {
  const errors: Record<string, string> = {}

  if (data.username && !validateUsername(data.username)) {
    errors.username = '用户名格式不正确（字母开头，4-16位）'
  }

  if (data.email && !validateEmail(data.email)) {
    errors.email = '邮箱格式不正确'
  }

  if (data.phone && !validatePhone(data.phone)) {
    errors.phone = '手机号格式不正确'
  }

  if (data.password) {
    const passwordValidation = validatePassword(data.password)
    if (!passwordValidation.isValid) {
      errors.password = passwordValidation.message || '密码格式不正确'
    }
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  }
}

/**
 * 生成随机密码
 */
export function generatePassword(length = 12): string {
  const charset = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@$!%*#?&'
  let password = ''

  // 确保至少包含一个大写字母、小写字母、数字和特殊字符
  password += 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'[Math.floor(Math.random() * 26)]
  password += 'abcdefghijklmnopqrstuvwxyz'[Math.floor(Math.random() * 26)]
  password += '0123456789'[Math.floor(Math.random() * 10)]
  password += '@$!%*#?&'[Math.floor(Math.random() * 8)]

  // 填充剩余长度
  for (let i = password.length; i < length; i++) {
    password += charset[Math.floor(Math.random() * charset.length)]
  }

  // 打乱顺序
  return password
    .split('')
    .sort(() => Math.random() - 0.5)
    .join('')
}
