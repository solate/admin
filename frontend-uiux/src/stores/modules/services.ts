// Services store

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Service } from '@/types/models'

export interface ServiceCategory {
  value: string
  label: string
  icon: string
}

export const useServicesStore = defineStore('services', () => {
  // State
  const services = ref<Service[]>([
    {
      id: 'svc-1',
      name: '云存储服务',
      description: '提供高可用、可扩展的云存储解决方案',
      status: 'running',
      endpoint: 'https://storage.example.com',
      config: {},
      createdAt: '2024-01-01',
      updatedAt: '2024-01-01'
    },
    {
      id: 'svc-2',
      name: '消息队列',
      description: '高性能消息队列服务，支持多种消息模式',
      status: 'running',
      endpoint: 'https://mq.example.com',
      config: {},
      createdAt: '2024-01-01',
      updatedAt: '2024-01-01'
    },
    {
      id: 'svc-3',
      name: '数据分析',
      description: '实时数据分析和可视化平台',
      status: 'running',
      endpoint: 'https://analytics.example.com',
      config: {},
      createdAt: '2024-01-01',
      updatedAt: '2024-01-01'
    },
    {
      id: 'svc-4',
      name: '身份认证',
      description: '企业级身份认证和授权服务',
      status: 'running',
      endpoint: 'https://auth.example.com',
      config: {},
      createdAt: '2024-01-01',
      updatedAt: '2024-01-01'
    },
    {
      id: 'svc-5',
      name: '邮件服务',
      description: '可靠的邮件发送和接收服务',
      status: 'stopped',
      endpoint: 'https://mail.example.com',
      config: {},
      createdAt: '2024-01-01',
      updatedAt: '2024-01-01'
    },
    {
      id: 'svc-6',
      name: '支付网关',
      description: '集成多种支付方式的支付处理服务',
      status: 'running',
      endpoint: 'https://payment.example.com',
      config: {},
      createdAt: '2024-01-01',
      updatedAt: '2024-01-01'
    }
  ])

  const selectedService = ref<Service | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Computed
  const activeServices = computed(() =>
    services.value.filter((s) => s.status === 'running')
  )

  const stoppedServices = computed(() =>
    services.value.filter((s) => s.status === 'stopped')
  )

  const serviceCountByStatus = computed(() => {
    return services.value.reduce((acc, service) => {
      acc[service.status] = (acc[service.status] || 0) + 1
      return acc
    }, {} as Record<string, number>)
  })

  // Categories
  const categories: ServiceCategory[] = [
    { value: 'storage', label: '存储服务', icon: 'Cloud' },
    { value: 'messaging', label: '消息服务', icon: 'MessageQueue' },
    { value: 'analytics', label: '数据分析', icon: 'ChartBar' },
    { value: 'security', label: '安全服务', icon: 'Shield' },
    { value: 'communication', label: '通信服务', icon: 'Mail' },
    { value: 'payment', label: '支付服务', icon: 'CreditCard' }
  ]

  // Actions
  function getServiceById(id: string): Service | undefined {
    return services.value.find((s) => s.id === id)
  }

  function getServicesByStatus(status: Service['status']): Service[] {
    return services.value.filter((s) => s.status === status)
  }

  function setSelectedService(service: Service | null) {
    selectedService.value = service
  }

  async function createService(
    serviceData: Partial<Service>
  ): Promise<Service> {
    isLoading.value = true
    error.value = null

    return new Promise((resolve) => {
      setTimeout(() => {
        const newService: Service = {
          id: 'svc-' + Date.now(),
          name: serviceData.name || '',
          description: serviceData.description || '',
          status: 'running',
          endpoint: serviceData.endpoint || '',
          config: serviceData.config || {},
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString()
        }
        services.value.push(newService)
        isLoading.value = false
        resolve(newService)
      }, 300)
    })
  }

  async function updateService(
    id: string,
    updates: Partial<Service>
  ): Promise<Service> {
    isLoading.value = true
    error.value = null

    return new Promise((resolve, reject) => {
      setTimeout(() => {
        const index = services.value.findIndex((s) => s.id === id)
        if (index !== -1) {
          services.value[index] = {
            ...services.value[index],
            ...updates,
            updatedAt: new Date().toISOString()
          }
          isLoading.value = false
          resolve(services.value[index])
        } else {
          isLoading.value = false
          reject(new Error('Service not found'))
        }
      }, 300)
    })
  }

  async function deleteService(id: string): Promise<void> {
    isLoading.value = true
    error.value = null

    return new Promise((resolve) => {
      setTimeout(() => {
        const index = services.value.findIndex((s) => s.id === id)
        if (index !== -1) {
          services.value.splice(index, 1)
        }
        isLoading.value = false
        resolve()
      }, 300)
    })
  }

  async function toggleService(
    id: string,
    enabled: boolean
  ): Promise<Service> {
    const index = services.value.findIndex((s) => s.id === id)
    if (index !== -1) {
      services.value[index].status = enabled ? 'running' : 'stopped'
      services.value[index].updatedAt = new Date().toISOString()
      return services.value[index]
    }
    throw new Error('Service not found')
  }

  return {
    // State
    services,
    selectedService,
    isLoading,
    error,
    categories,

    // Computed
    activeServices,
    stoppedServices,
    serviceCountByStatus,

    // Actions
    getServiceById,
    getServicesByStatus,
    setSelectedService,
    createService,
    updateService,
    deleteService,
    toggleService
  }
})
