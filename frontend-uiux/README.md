# Multi-Tenant SaaS Platform

一个使用 Vue3 + Vite 构建的多租户 SaaS 管理平台。

## 设计系统

### 颜色方案
- **主色**: #2563EB (Primary Blue)
- **辅助色**: #F97316 (Secondary Orange)
- **背景**: #F8FAFC (Slate 50)
- **文字**: #1E293B (Slate 800)

### 字体
- **标题**: Poppins
- **正文**: Open Sans

### 风格
- Glassmorphism (毛玻璃效果)
- 圆角卡片设计
- 平滑过渡动画
- 响应式布局

## 项目结构

```
frontend-uiux/
├── src/
│   ├── assets/          # 静态资源
│   │   └── main.css     # 全局样式
│   ├── components/      # 组件
│   │   └── icons/       # 图标组件
│   ├── layouts/         # 布局组件
│   │   └── DashboardLayout.vue
│   ├── router/          # 路由配置
│   │   └── index.js
│   ├── stores/          # Pinia 状态管理
│   │   ├── auth.js      # 认证状态
│   │   ├── tenants.js   # 租户状态
│   │   ├── services.js  # 服务状态
│   │   └── ui.js        # UI 状态
│   ├── views/           # 页面组件
│   │   ├── auth/        # 认证页面
│   │   ├── dashboard/   # 仪表板
│   │   ├── tenants/     # 租户管理
│   │   ├── services/    # 服务管理
│   │   ├── users/       # 用户管理
│   │   ├── business/    # 业务管理
│   │   ├── analytics/   # 数据分析
│   │   ├── settings/    # 系统设置
│   │   └── profile/     # 个人中心
│   ├── App.vue
│   └── main.js
├── index.html
├── package.json
├── vite.config.js
├── tailwind.config.js
└── postcss.config.js
```

## 功能模块

### 基础服务模块
- 租户管理 - 多租户账户管理
- 服务管理 - 基础服务配置
- 用户管理 - 平台用户管理

### 业务模块
- 仪表板 - 数据概览和统计
- 业务管理 - 订单和收入管理
- 数据分析 - 运营数据分析

### 系统设置
- 系统设置 - 平台配置管理
- 个人中心 - 用户资料管理

## 快速开始

### 安装依赖

```bash
cd frontend-uiux
npm install
```

### 启动开发服务器

```bash
npm run dev
```

### 构建生产版本

```bash
npm run build
```

## 技术栈

- **Vue 3** - 渐进式 JavaScript 框架
- **Vite** - 下一代前端构建工具
- **Vue Router** - 官方路由管理器
- **Pinia** - Vue 状态管理库
- **Tailwind CSS** - 实用优先的 CSS 框架

## 特性

- Composition API
- Glassmorphism 设计风格
- 完全响应式布局
- 暗色模式支持
- 可折叠侧边栏
- 移动端友好
- SVG 图标系统
- 平滑页面过渡
- 可访问性优化

## 许可证

MIT
