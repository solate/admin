# 多租户架构

> 本文档详细说明项目的多租户 SaaS 架构实现

---

## 概述

本项目采用多租户 SaaS 架构，支持数据隔离、租户上下文管理和权限控制。

---

## 核心概念

### 租户隔离策略

| 层级 | 隔离方式 | 实现位置 |
|------|----------|----------|
| **API 层** | 请求头 `X-Tenant-ID` | `src/utils/request.ts` |
| **状态层** | localStorage + Pinia Store | `src/stores/modules/tenants.ts` |
| **UI 层** | 租户选择器 | `src/components/business/tenant/` |

---

## 实现细节

### 1. 租户上下文管理

**文件**: `src/lib/tenant/context.ts`

```typescript
import { ref, computed } from 'vue'

export class TenantContext {
  private currentTenantId = ref<string | null>(null)

  get currentTenant() {
    return computed(() => {
      if (!this.currentTenantId.value) return null
      return tenants.value.find(t => t.id === this.currentTenantId.value)
    })
  }

  switchTenant(tenantId: string) {
    this.currentTenantId.value = tenantId
    localStorage.setItem('currentTenantId', tenantId)

    // 更新 API 请求头
    api.defaults.headers['X-Tenant-ID'] = tenantId

    // 重新加载数据
    this.reloadTenantData()
  }

  initialize() {
    const saved = localStorage.getItem('currentTenantId')
    if (saved) {
      this.switchTenant(saved)
    }
  }
}

export const useTenantContext = () => new TenantContext()
```

### 2. API 请求拦截

**文件**: `src/utils/request.ts`

```typescript
// 请求拦截器 - 自动添加租户 ID
request.interceptors.request.use((config) => {
  const tenantId = localStorage.getItem('currentTenantId')
  if (tenantId) {
    config.headers['X-Tenant-ID'] = tenantId
  }
  return config})
```

### 3. 租户状态管理

**文件**: `src/stores/modules/tenants.ts`

```typescript
export const useTenantsStore = defineStore('tenants', {
  state: () => ({
    currentTenant: null as Tenant | null,
    tenants: [] as Tenant[]
  }),

  actions: {
    async fetchTenants() {
      const { data } = await api.get('/tenants')
      this.tenants = data
    },

    setCurrentTenant(tenant: Tenant) {
      this.currentTenant = tenant
      localStorage.setItem('currentTenantId', tenant.id)
    }
  }
})
```

### 4. 路由守卫

**文件**: `src/router/guards/tenant.ts`

```typescript
export const tenantGuard: RouterGuard = async (to, from, next) => {
  const tenantsStore = useTenantsStore()

  // 初始化租户数据
  if (!tenantsStore.tenants.length) {
    await tenantsStore.fetchTenants()
  }

  // 检查是否需要选择租户
  if (to.meta.requiresTenant && !tenantsStore.currentTenant) {
    // 重定向到租户选择页面或使用默认租户
    const defaultTenant = tenantsStore.tenants[0]
    tenantsStore.setCurrentTenant(defaultTenant)
  }

  next()
}
```

---

## 使用方式

### 在组件中使用

```vue
<script setup>
import { useTenantContext } from '@/lib/tenant/context'

const { currentTenant, switchTenant } = useTenantContext()
</script>

<template>
  <div>
    <p>当前租户: {{ currentTenant?.name }}</p>
    <button @click="switchTenant('tenant-123')">
      切换租户
    </button>
  </div>
</template>
```

### 租户选择器组件

**文件**: `src/components/business/tenant/TenantSelector.vue`

```vue
<template>
  <el-dropdown @command="handleTenantChange">
    <span class="tenant-selector">
      {{ currentTenant?.name }}
      <el-icon><Switch /></el-icon>
    </span>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item
          v-for="tenant in tenants"
          :key="tenant.id"
          :command="tenant.id"
        >
          {{ tenant.name }}
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>
```

---

## 数据模型

### Tenant 类型定义

**文件**: `src/types/models.ts`

```typescript
export interface Tenant {
  id: string
  name: string
  domain?: string
  status: 'active' | 'suspended' | 'pending'
  plan: 'basic' | 'pro' | 'enterprise'
  quota: {
    users: number
    storage: number
    apiCalls: number
  }
  createdAt: string
  updatedAt: string
}
```

---

## 最佳实践

### 1. 租户切换时清理数据

```typescript
function switchTenant(tenantId: string) {
  // 清理缓存数据
  queryClient.clear()

  // 切换租户
  setCurrentTenant(tenantId)

  // 重新加载数据
  router.push('/dashboard/overview')
}
```

### 2. 租户级缓存

```typescript
// 使用租户 ID 作为缓存键的一部分
const cacheKey = computed(() => {
  return `users:${currentTenant.value?.id}`
})
```

### 3. 错误处理

```typescript
// API 返回 403 时检查是否是租户问题
api.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === 403) {
      // 可能是租户权限问题
      const tenantsStore = useTenantsStore()
      tenantsStore.checkTenantAccess()
    }
    return Promise.reject(error)
  }
)
```

---

## 相关文件

| 文件 | 说明 |
|------|------|
| `src/lib/tenant/context.ts` | 租户上下文管理 |
| `src/stores/modules/tenants.ts` | 租户状态管理 |
| `src/router/guards/tenant.ts` | 租户路由守卫 |
| `src/components/business/tenant/` | 租户相关组件 |
| `src/api/modules/tenants.ts` | 租户 API 接口 |

---

| 状态 | ✅ 已实现 |
|------|----------|
| 日期 | 2026-01-26 |
