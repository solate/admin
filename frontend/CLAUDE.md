# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 语言规范
- 所有对话和文档都使用中文
- 文档使用 markdown 格式


## 命令

### 安装依赖
```bash
npm install          # 安装所有依赖
```

### 开发
```bash
npm run dev          # 启动开发服务器，端口 5173
npm run build        # 构建生产版本 (运行 vue-tsc -b && vite build)
npm run preview      # 预览生产构建
```

### 类型检查
```bash
vue-tsc -b           # 运行 TypeScript 编译器检查，不进行构建
```

## 架构概览

这是一个 Vue 3 + TypeScript 的管理后台前端，与 Go 后端 API 进行通信。应用程序使用现代化的技术栈，使用 Vite 进行快速开发和构建。

### 核心架构模式

**认证与令牌管理：**
- 基于 JWT 的认证，支持自动令牌刷新
- 访问令牌存储在 localStorage 中，支持刷新令牌轮换
- 请求拦截器自动为 API 调用添加 Bearer 令牌
- 响应拦截器处理 401 错误和令牌刷新，支持请求队列
- 令牌刷新使用订阅者模式防止竞态条件

**API 层结构：**
- 在 `src/api/http.ts` 中配置集中式 HTTP 客户端和拦截器
- 基于功能的 API 模块（auth、dict、factory、product、stats、user、tenant、role、permission、menu、inventory 等）
- 所有 API 调用都使用 admin/v1 前缀
- 后端运行在 8080 端口，前端在开发时代理 `/api` 请求

**状态管理：**
- 使用 Pinia 进行全局状态管理
- 通过令牌工具管理认证状态
- 登录后用户信息存储在 localStorage 中

**路由与导航：**
- 使用 Vue Router 和路由守卫进行认证
- 公共路由（login、register）绕过认证
- 主布局组件包含嵌套路由用于已认证内容
- 未认证访问时自动重定向到登录页

**组件结构：**
- 布局组件提供主要应用框架，包含侧边栏和头部
- 基于功能的视图（Dashboard、Factories、Products、Statistics）
- 共享组件如 Pagination 用于通用 UI 模式

### API 配置

**开发环境：** 使用 Vite 代理将 `/api` 请求转发到 `http://localhost:8080`
**生产环境：** 使用 `VITE_API_BASE_URL` 环境变量，默认为 `/api`

后端 API 遵循 RESTful 模式，所有端点都使用 admin/v1 前缀。认证使用 Authorization 头中的 Bearer 令牌。

### 多租户架构

系统支持多租户功能：
- 基于租户的用户管理
- 基于角色的权限控制
- 工厂和产品数据按租户范围隔离

### 关键文件说明

- `src/api/http.ts` - HTTP 客户端，包含拦截器和令牌刷新逻辑
- `src/utils/token.ts` - JWT 令牌管理和刷新机制
- `src/router/index.ts` - 路由配置和认证守卫
- `vite.config.ts` - 开发服务器配置，包含 API 代理
- `src/views/Layout.vue` - 主应用布局组件

### 开发注意事项

- 启用 TypeScript 严格模式，包含额外的代码检查规则（`noUnusedLocals`、`noUnusedParameters`）
- 组件使用 `<script setup lang="ts">` 语法
- 样式使用 SCSS 和 Element Plus 主题
- 所有 Element Plus 图标在 main.ts 中全局注册
- API 响应期望 `code: 200` 或 `code: 0` 表示成功
- 开发服务器将 `/api` 请求代理到 `http://localhost:8080`（Go 后端）
- 构建过程在 Vite 构建前包含 TypeScript 编译检查

### 环境配置

开发环境 API 代理在 `vite.config.ts` 中配置：
```typescript
server: {
  host: '0.0.0.0',
  port: 5173,
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true
    }
  }
}
```

生产环境需要设置 `VITE_API_BASE_URL` 环境变量，默认为 `/api`。

### 关键依赖

**核心框架：**
- Vue 3.5.22 - 渐进式 JavaScript 框架
- TypeScript 5.9.3 - JavaScript 的超集，提供静态类型检查
- Vite 7.1.7 - 现代化前端构建工具

**UI 和样式：**
- Element Plus 2.11.5 - Vue 3 UI 组件库
- Sass 1.93.2 - CSS 预处理器

**状态管理和路由：**
- Pinia 3.0.3 - Vue 3 状态管理库
- Vue Router 4.6.3 - Vue 官方路由管理器

**HTTP 客户端：**
- Axios 1.12.2 - 基于 Promise 的 HTTP 客户端