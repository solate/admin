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
│   ├── factory.ts      # 工厂管理API
│   ├── product.ts      # 商品管理API
│   ├── stats.ts        # 统计API
│   └── http.ts         # Axios封装
├── components/         # 公共组件
├── router/            # 路由配置
│   └── index.ts
├── styles/            # 全局样式
│   └── index.scss
├── utils/             # 工具函数
├── views/             # 页面组件
│   ├── Login.vue      # 登录页
│   ├── Layout.vue     # 主布局
│   ├── Dashboard.vue  # 首页仪表板
│   ├── Factories.vue  # 工厂管理
│   ├── Products.vue   # 商品管理
│   └── Statistics.vue # 数据统计
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

### 预览生产构建

```bash
npm run preview
```

## 🔧 配置说明

### 环境变量

创建 `.env.development` 文件：

```env
VITE_API_BASE_URL=http://localhost:8080/api
```

创建 `.env.production` 文件：

```env
VITE_API_BASE_URL=https://your-api-domain.com/api
```

### API配置

API基础配置在 `src/api/http.ts` 中：

- 基础URL: 通过环境变量 `VITE_API_BASE_URL` 配置
- 超时时间: 15秒
- 请求拦截器: 自动添加 Authorization 头
- 响应拦截器: 统一错误处理和成功响应处理

## 📱 功能模块

### 1. 登录认证
- 用户名/密码登录
- 密码显示切换
- 登录状态保持
- 自动跳转和鉴权

### 2. 主布局
- 响应式侧边栏（可折叠）
- 顶部导航栏
- 面包屑导航
- 用户下拉菜单

### 3. 工厂管理
- 工厂列表展示
- 搜索和分页
- 新增/编辑/删除工厂
- 批量删除

### 4. 商品管理
- 商品列表展示
- 商品CRUD操作
- 库存入库/出库
- 实时库存显示

### 5. 数据统计
- 概览统计卡片
- 趋势图表（待开发）
- 分类统计表格

## 🔌 API集成

### 接口文件

所有API接口定义在 `src/api/` 目录下：

- `auth.ts`: 认证相关接口
- `factory.ts`: 工厂管理接口
- `product.ts`: 商品管理接口
- `stats.ts`: 统计接口

### 使用示例

```typescript
import { factoryApi } from '@/api/factory'

// 获取工厂列表
const { list, total } = await factoryApi.getList({
  page: 1,
  pageSize: 10,
  keyword: '搜索关键词'
})

// 创建工厂
const newFactory = await factoryApi.create({
  name: '新工厂',
  address: '工厂地址',
  owner: '负责人'
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

### 组件样式

每个组件使用 `<style scoped>` 定义局部样式，避免样式污染。

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

1. 在 `src/views/` 创建页面组件
2. 在 `src/router/index.ts` 添加路由配置
3. 在 `src/views/Layout.vue` 添加菜单项

### 添加新API

1. 在 `src/api/` 创建接口文件
2. 定义TypeScript类型
3. 导出API函数
4. 在页面中导入使用

### 代码规范

- 使用 TypeScript 严格模式
- 组件名使用 PascalCase
- 文件名使用 kebab-case
- 使用 ESLint 进行代码检查

## 🐛 常见问题

### 1. 登录后页面空白

检查路由配置和组件导入是否正确。

### 2. API请求失败

检查后端服务是否启动，API地址配置是否正确。

### 3. 样式不生效

检查样式文件是否正确导入，scoped属性是否正确使用。

## 📞 技术支持

如有问题，请查看：
- [Vue 3 文档](https://vuejs.org/)
- [Element Plus 文档](https://element-plus.org/)
- [Vite 文档](https://vitejs.dev/)
