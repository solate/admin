# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 语言规范
- 所有对话和文档都使用中文
- 文档使用 markdown 格式

## 项目概述

这是一个基于 **Vue 3 + Vite + Tailwind CSS + Element Plus** 的多租户 SaaS 管理平台前端项目，采用现代化的 Glassmorphism 设计风格。

### 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Vue | 3.4.21 | 渐进式 JavaScript 框架 |
| Vite | 7.3.1 | 快速的开发构建工具 |
| Vue Router | 4.3.0 | 路由管理 |
| Pinia | 2.1.7 | 状态管理 |
| Vue I18n | 9.14.5 | 国际化 (中文/英文) |
| Axios | 1.13.2 | HTTP 客户端 |
| Element Plus | 2.13.1 | UI 组件库 |
| Tailwind CSS | 3.4.1 | 原子化 CSS 框架 |
| ECharts | 6.0.0 | 图表可视化 |

---

## 快速开始

```bash
# 安装依赖
npm install

# 开发模式 (端口 3000)
npm run dev

# 生产构建
npm run build

# 预览构建
npm run preview

# 代码检查
npm run lint
```

### 环境变量

创建 `.env` 文件配置后端 API 地址：

```bash
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

默认值为 `http://localhost:8080/api/v1`

---

## 项目结构

```
frontend-uiux/
├── src/
│   ├── assets/              # 静态资源
│   ├── components/          # 可复用组件
│   │   ├── icons/          # 图标组件
│   │   ├── layout/         # 布局组件 (TopNavbar, UserMenu)
│   │   ├── notification/   # 通知组件
│   │   ├── tenant/         # 租户选择器
│   │   ├── ui/             # 基础 UI 组件
│   │   └── user/           # 用户相关组件
│   ├── directives/         # 自定义指令 (clickOutside)
│   ├── layouts/            # 页面布局 (DashboardLayout)
│   ├── locales/            # 国际化文件 (zh-CN, en-US)
│   ├── plugins/            # 插件配置 (Element Plus)
│   ├── router/             # 路由配置
│   ├── services/           # API 服务层
│   ├── stores/             # Pinia 状态管理
│   ├── views/              # 页面组件
│   │   ├── auth/          # 登录/注册页面
│   │   ├── dashboard/     # 仪表板
│   │   ├── tenants/       # 租户管理
│   │   ├── services/      # 服务管理
│   │   ├── users/         # 用户管理
│   │   ├── business/      # 业务管理
│   │   ├── analytics/     # 数据分析
│   │   ├── settings/      # 系统设置
│   │   ├── profile/       # 个人中心
│   │   └── notifications/ # 通知中心
│   ├── App.vue             # 根组件
│   └── main.js             # 应用入口
├── index.html
├── package.json
├── vite.config.js
├── tailwind.config.js
└── postcss.config.js
```

---

## 核心架构

### Element Plus 集成 (`src/plugins/element.js`)

项目集成了 Element Plus 组件库，并通过自定义 CSS 变量与现有设计系统保持一致：

```javascript
// 所有 Element Plus 图标已在 main.js 中全局注册
// 可直接在任何组件中使用，如 <el-icon><Plus /></el-icon>

// 自定义主题颜色与 Tailwind 保持一致
--el-color-primary: #2563eb  // primary-600
--el-color-success: #22c55e  // success-600
--el-color-warning: #f59e0b  // warning-600
--el-color-danger: #ef4444   // error-600
```

### API 服务层 (`src/services/api.js`)

集中式 API 管理，使用 Axios 拦截器：

```javascript
// 请求拦截器 - 自动添加认证和租户信息
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`

  const tenantId = localStorage.getItem('currentTenantId')
  if (tenantId) config.headers['X-Tenant-ID'] = tenantId

  return config
})

// 响应拦截器 - 统一错误处理 (401/403/404/500)
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)
```

**API 服务模块**：
- `auth` - 认证相关 (login, logout, register, refreshToken, me)
- `tenants` - 租户管理 (list, getAll, getById, create, update, delete, updateStatus)
- `users` - 用户管理 (list, getById, create, update, delete, assignRoles, changePassword, profile)
- `roles` - 角色管理
- `permissions` - 权限管理
- `auditLogs` - 审计日志
- `services` - 服务管理 (list, getById, create, update, delete, toggle)
- `notifications` - 通知管理 (list, markAsRead, markAllAsRead, unreadCount)
- `dashboard` - 仪表板数据 (stats, charts, activity)
- `settings` - 系统设置

### 状态管理 (Pinia)

| Store | 文件 | 职责 |
|-------|------|------|
| `useAuthStore` | `stores/auth.js` | 用户认证状态 (当前为 Mock 实现) |
| `useTenantsStore` | `stores/tenants.js` | 租户数据管理 |
| `useServicesStore` | `stores/services.js` | 服务状态管理 |
| `useUIStore` | `stores/ui.js` | UI 状态 (侧边栏、主题等) |

**注意**: `useAuthStore` 目前使用 Mock 认证，需要对接真实后端 API。

### 路由结构

**公共路由**:
- `/` - 落地页 (LandingView)
- `/login` - 登录页 (LoginView)
- `/register` - 注册页 (RegisterView)

**Dashboard 路由** (需认证):
- `/dashboard/overview` - 概览
- `/dashboard/tenants` - 租户列表
- `/dashboard/tenants/create` - 创建租户
- `/dashboard/tenants/:id` - 租户详情
- `/dashboard/services` - 服务列表
- `/dashboard/services/:id` - 服务详情
- `/dashboard/users` - 用户列表
- `/dashboard/users/create` - 创建用户
- `/dashboard/users/:id` - 用户详情
- `/dashboard/business` - 业务管理
- `/dashboard/analytics` - 数据分析
- `/dashboard/settings` - 系统设置
- `/dashboard/profile` - 个人中心
- `/dashboard/notifications` - 通知中心

---

## 设计系统

### 颜色规范 (Tailwind + Element Plus)

```javascript
// 主色 - 专业与信任
primary: {
  DEFAULT: '#2563eb',  // blue-600
  // 50-900 色阶...
}

// CTA 色 - 对比与行动
cta: {
  DEFAULT: '#f97316',  // orange-500
}

// 语义色
success: '#22c55e'  // green-500
warning: '#f59e0b'  // amber-500
error: '#ef4444'    // red-500
info: '#0ea5e9'     // sky-500
```

### 字体系统

- **Sans**: Fira Sans (现代化技术字体)
- **Mono**: Fira Code (数据/技术内容)

### 组件选择优先级

1. **优先使用 Element Plus 组件**: `el-button`, `el-input`, `el-table`, `el-dialog`, `el-form` 等
2. **使用 Tailwind CSS**: 用于布局和自定义样式
3. **自定义组件**: `src/components/ui/` 中的基础组件

---

## 多租户架构

### 租户上下文管理

1. **租户选择**: 通过 `TenantSelector` 组件切换
2. **状态存储**: `localStorage.getItem('currentTenantId')`
3. **API 请求**: 自动添加 `X-Tenant-ID` 请求头
4. **状态同步**: `useTenantsStore` 管理当前租户状态

```javascript
// 切换租户
function setCurrentTenant(tenantId) {
  const tenant = tenants.value.find(t => t.id === tenantId)
  currentTenant.value = tenant
  localStorage.setItem('currentTenantId', tenantId)
  api.defaults.headers['X-Tenant-ID'] = tenantId
}
```

---

## 国际化 (i18n)

### 支持语言

- `zh-CN` - 简体中文 (默认)
- `en-US` - 英文

### Element Plus 多语言切换

```javascript
// 在 main.js 中已配置全局属性
app.config.globalProperties.$elementLocales = { zhCn, en }

// 切换语言时需要同时更新 Vue I18n 和 Element Plus locale
```

### 使用方式

```vue
<script setup>
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <h1>{{ t('welcome') }}</h1>
</template>
```

### 语言文件位置

- `src/locales/zh-CN.js`
- `src/locales/en-US.js`

---

## 开发规范

### 通用约定

1. **语言规范**
   - **回复语言**: 与用户对话时使用**中文**回复
   - **代码注释**: 所有代码注释使用**中文**
   - **提交信息**: Git 提交信息使用**中文**
   - **文档编写**: 项目文档使用**中文**

2. **命名约定**
   - 组件文件: PascalCase (例: `UserProfile.vue`)
   - 工具文件: camelCase (例: `formatUtils.js`)
   - 常量: UPPER_SNAKE_CASE (例: `API_BASE_URL`)
   - CSS 类名: kebab-case (配合 Tailwind)

3. **代码风格**
   - 使用 2 空格缩进
   - 使用单引号 (JS/JSX)
   - 语句末尾添加分号
   - 组件名使用多单词 (避免与 HTML 元素冲突)

### 组件开发

**使用 Composition API + `<script setup>`**:

```vue
<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  variant: {
    type: String,
    default: 'primary',
    validator: (value) => ['primary', 'secondary'].includes(value)
  }
})

const emit = defineEmits(['click'])
</script>
```

### 自定义指令

项目包含 `v-click-outside` 指令，用于点击外部关闭下拉菜单等场景。

### 路由守卫

```javascript
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  const tenantsStore = useTenantsStore()

  tenantsStore.initialize()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'login', query: { redirect: to.fullPath } })
  } else {
    next()
  }
})
```

---

## 常见任务

### 添加新页面

1. 在 `src/views/` 创建 `.vue` 文件
2. 在 `src/router/index.js` 添加路由
3. 在 `src/locales/` 添加翻译

### 添加 UI 组件

1. 优先使用 Element Plus 组件
2. 需要自定义时，在 `src/components/ui/` 创建组件
3. 使用 Tailwind 类名进行样式设计

### 添加 API 端点

在 `src/services/api.js` 的 `apiService` 对象中添加：

```javascript
export const apiService = {
  // 新增模块
  newModule: {
    list: (params) => api.get('/new-module', { params }),
    getById: (id) => api.get(`/new-module/${id}`),
    create: (data) => api.post('/new-module', data),
    update: (id, data) => api.put(`/new-module/${id}`, data),
    delete: (id) => api.delete(`/new-module/${id}`)
  }
}
```

---

## 配置文件

### Vite 配置 (`vite.config.js`)

- 开发服务器端口: 3000
- 自动打开浏览器
- 路径别名: `@` → `./src`

### Tailwind 配置 (`tailwind.config.js`)

- 暗色模式: class 策略
- 自定义颜色系统
- 自定义字体 (Fira Sans, Fira Code)
- 自定义动画 (fade-in, slide-up, slide-down, scale-in)

---

## 注意事项

1. **认证**: 目前使用 Mock 认证 (`stores/auth.js`)，需要对接真实后端 API
2. **API 响应**: 期望后端返回 `{ code: 200, data: {...}, message: "success" }` 格式
3. **租户隔离**: 所有 API 请求自动携带 `X-Tenant-ID` 请求头
4. **Token 管理**: 存储在 localStorage，401 时自动清除并跳转登录
5. **国际化**: 新增文本必须同时添加中英文翻译
6. **Element Plus**: 图标已全局注册，可直接使用 `<el-icon><IconName /></el-icon>`
