import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useServicesStore = defineStore('services', () => {
  const services = ref([
    {
      id: 'svc-1',
      name: '云存储服务',
      code: 'cloud-storage',
      description: '提供高可用、可扩展的云存储解决方案',
      category: 'storage',
      status: 'active',
      pricing: { basic: 99, pro: 299, enterprise: 999 },
      features: ['99.9% SLA', '自动备份', 'CDN加速', 'API访问']
    },
    {
      id: 'svc-2',
      name: '消息队列',
      code: 'message-queue',
      description: '高性能消息队列服务，支持多种消息模式',
      category: 'messaging',
      status: 'active',
      pricing: { basic: 149, pro: 399, enterprise: 1299 },
      features: ['发布订阅', '点对点', '死信队列', '消息追踪']
    },
    {
      id: 'svc-3',
      name: '数据分析',
      code: 'data-analytics',
      description: '实时数据分析和可视化平台',
      category: 'analytics',
      status: 'active',
      pricing: { basic: 199, pro: 499, enterprise: 1999 },
      features: ['实时分析', '自定义报表', '数据导出', 'API集成']
    },
    {
      id: 'svc-4',
      name: '身份认证',
      code: 'auth-service',
      description: '企业级身份认证和授权服务',
      category: 'security',
      status: 'active',
      pricing: { basic: 79, pro: 199, enterprise: 799 },
      features: ['SSO单点登录', 'MFA多因素认证', 'LDAP集成', '审计日志']
    },
    {
      id: 'svc-5',
      name: '邮件服务',
      code: 'email-service',
      description: '可靠的邮件发送和接收服务',
      category: 'communication',
      status: 'maintenance',
      pricing: { basic: 49, pro: 149, enterprise: 499 },
      features: ['API发送', 'SMTP支持', '模板管理', '统计分析']
    },
    {
      id: 'svc-6',
      name: '支付网关',
      code: 'payment-gateway',
      description: '集成多种支付方式的支付处理服务',
      category: 'payment',
      status: 'active',
      pricing: { basic: 129, pro: 349, enterprise: 1499 },
      features: ['多渠道支持', '分期付款', '退款管理', '对账报表']
    }
  ])

  const selectedService = ref(null)
  const loading = ref(false)
  const error = ref(null)

  function getServiceById(id) {
    return services.value.find(s => s.id === id)
  }

  function getServiceByCode(code) {
    return services.value.find(s => s.code === code)
  }

  function getServicesByCategory(category) {
    return services.value.filter(s => s.category === category)
  }

  function getActiveServices() {
    return services.value.filter(s => s.status === 'active')
  }

  function createService(serviceData) {
    return new Promise((resolve) => {
      setTimeout(() => {
        const newService = {
          id: 'svc-' + Date.now(),
          ...serviceData,
          status: 'active'
        }
        services.value.push(newService)
        resolve(newService)
      }, 300)
    })
  }

  function updateService(id, updates) {
    const index = services.value.findIndex(s => s.id === id)
    if (index !== -1) {
      services.value[index] = { ...services.value[index], ...updates }
      return Promise.resolve(services.value[index])
    }
    return Promise.reject(new Error('Service not found'))
  }

  function deleteService(id) {
    return new Promise((resolve) => {
      setTimeout(() => {
        const index = services.value.findIndex(s => s.id === id)
        if (index !== -1) {
          services.value.splice(index, 1)
        }
        resolve()
      }, 300)
    })
  }

  const categories = [
    { value: 'storage', label: '存储服务', icon: 'Cloud' },
    { value: 'messaging', label: '消息服务', icon: 'MessageQueue' },
    { value: 'analytics', label: '数据分析', icon: 'ChartBar' },
    { value: 'security', label: '安全服务', icon: 'Shield' },
    { value: 'communication', label: '通信服务', icon: 'Mail' },
    { value: 'payment', label: '支付服务', icon: 'CreditCard' }
  ]

  return {
    services,
    selectedService,
    loading,
    error,
    categories,
    getServiceById,
    getServiceByCode,
    getServicesByCategory,
    getActiveServices,
    createService,
    updateService,
    deleteService
  }
})
