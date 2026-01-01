# 后台管理系统 - 前端

基于 Vue3 + TypeScript + Element Plus 的现代化后台管理系统前端。

## 🚀 技术栈

- **框架**: Vue 3 + TypeScript
- **构建工具**: Vite
- **UI组件库**: Element Plus
- **状态管理**: Pinia
- **路由**: Vue Router
- **HTTP客户端**: Axios
- **样式**: SCSS

## 📁 项目结构

```
src/
├── api/                 # API接口定义
│   ├── auth.ts         # 认证相关API
│   ├── user.ts         # 用户管理API
│   ├── role.ts         # 角色管理API
│   ├── menu.ts         # 菜单管理API
│   ├── tenant.ts       # 租户管理API
│   ├── auditLog.ts     # 审计日志API
│   ├── user_menu.ts    # 用户菜单API
│   └── http.ts         # Axios封装
├── components/         # 公共组件
├── router/            # 路由配置
│   └── index.ts
├── stores/            # Pinia状态管理
├── styles/            # 全局样式
│   └── index.scss
├── utils/             # 工具函数
├── views/             # 页面组件
│   ├── Login.vue      # 登录页
│   ├── Layout.vue     # 主布局
│   ├── Dashboard.vue  # 首页仪表板
│   └── system/        # 系统管理页面
│       ├── users/     # 用户管理
│       ├── roles/     # 角色管理
│       ├── menus/     # 菜单管理
│       └── tenants/   # 租户管理
├── App.vue            # 根组件
└── main.ts            # 入口文件
```

## 🛠️ 开发环境

### 环境要求

- Node.js >= 16.0.0
- npm >= 8.0.0

### 安装依赖

```bash
npm install
```

### 启动开发服务器

```bash
npm run dev
```

访问 http://localhost:5173

### 构建生产版本

```bash
npm run build
```

## 🔧 配置说明

### API配置

API基础配置在 `src/api/http.ts` 中：

- 基础URL: 开发环境使用 Vite 代理，生产环境由 Nginx 处理
- 超时时间: 15秒
- 请求拦截器: 自动添加 Authorization 头
- 响应拦截器: 统一错误处理和成功响应处理

### Vite 代理配置

开发环境使用 Vite 代理转发 API 请求到后端（`vite.config.ts`）：

```typescript
proxy: {
  '/api': {
    target: 'http://localhost:8080',
    changeOrigin: true,
  }
}
```

## 📱 功能模块

### 1. 用户认证
- 用户名/密码登录
- 验证码验证
- JWT Token 管理
- 自动刷新 Token
- 登出

### 2. 租户管理
- 租户列表
- 创建/编辑/删除租户
- 租户状态管理

### 3. 用户管理
- 用户列表（分页、搜索）
- 创建/编辑/删除用户
- 用户状态管理
- 分配角色

### 4. 角色管理
- 角色列表
- 创建/编辑/删除角色
- 权限配置

### 5. 菜单管理
- 菜单树结构
- 创建/编辑/删除菜单
- 菜单排序

### 6. 审计日志
- 登录日志
- 操作日志

## 🔌 API集成

### 接口文件

所有API接口定义在 `src/api/` 目录下：

- `auth.ts`: 认证相关接口
- `user.ts`: 用户管理接口
- `role.ts`: 角色管理接口
- `menu.ts`: 菜单管理接口
- `tenant.ts`: 租户管理接口
- `auditLog.ts`: 审计日志接口

### 使用示例

```typescript
import { authApi, userApi } from '@/api'

// 登录
const { access_token, refresh_token } = await authApi.login('default', {
  username: 'admin',
  password: 'Admin@123',
  captcha_id: 'xxx',
  captcha: '1234'
})

// 获取用户列表
const { list, total } = await userApi.getList({
  page: 1,
  pageSize: 10,
  keyword: '搜索关键词'
})
```

## 🎨 样式定制

### 主题色配置

在 `src/styles/index.scss` 中定义全局样式变量：

```scss
:root {
  --el-color-primary: #409eff;
  --el-color-success: #67c23a;
  --el-color-warning: #e6a23c;
  --el-color-danger: #f56c6c;
}
```

## 🚀 部署

### 构建

```bash
npm run build
```

构建产物在 `dist/` 目录。

### 部署到Nginx

```nginx
server {
    listen 80;
    server_name your-domain.com;
    root /path/to/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 🔍 开发指南

### 添加新页面

1. 在 `src/views/` 或 `src/views/system/` 创建页面组件
2. 在 `src/router/index.ts` 添加路由配置
3. 在 `src/views/Layout.vue` 添加菜单项（如果需要）

### 添加新API

1. 在 `src/api/` 创建接口文件
2. 定义TypeScript类型
3. 导出API函数
4. 在 `src/api/index.ts` 中导出

### 代码规范

- 使用 TypeScript 严格模式
- 组件名使用 PascalCase
- 文件名使用 kebab-case 或 PascalCase（组件文件）

## 🐛 常见问题

### 1. 登录后页面空白

检查路由配置和组件导入是否正确。

### 2. API请求失败

检查后端服务是否启动，API地址配置是否正确。

### 3. 样式不生效

检查样式文件是否正确导入，scoped属性是否正确使用。

### 4. Vite 代理问题（API 返回 HTML 而不是 JSON）

**症状**：API 请求返回 HTML 页面而不是预期的 JSON 数据。

**原因**：Vite 开发服务器缓存或 HMR 问题导致代理配置没有正确生效。

**快速解决方案**：

```bash
# 1. 完全停止所有 Node 进程
killall -9 node

# 2. 清除 Vite 缓存
rm -rf node_modules/.vite

# 3. 重新启动开发服务器
npm run dev

# 4. 浏览器强制刷新（Cmd+Shift+R 或 Ctrl+Shift+R）
```

**或使用直连方案**（如果后端已配置 CORS）：

修改 `src/api/http.ts`：
```typescript
// 开发环境直接连接后端
const baseURL = import.meta.env.DEV ? 'http://localhost:8080' : ''
```

## 📞 技术支持

如有问题，请查看：
- [Vue 3 文档](https://vuejs.org/)
- [Element Plus 文档](https://element-plus.org/)
- [Vite 文档](https://vitejs.dev/)
