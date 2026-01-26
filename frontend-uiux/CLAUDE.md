# CLAUDE.md

> 本文件为 Claude Code 提供项目指引，详细文档请查阅知识库。

## 语言规范
- 所有对话和文档使用中文
- 文档使用 Markdown 格式

## 项目概述

基于 **Vue 3 + Vite + Tailwind CSS + Element Plus** 的多租户 SaaS 管理平台，采用 Glassmorphism 设计风格。

### 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Vue | 3.4.21 | 渐进式框架 |
| Vite | 7.3.1 | 构建工具 |
| Vue Router | 4.3.0 | 路由管理 |
| Pinia | 2.1.7 | 状态管理 |
| Vue I18n | 9.14.5 | 国际化 |
| Element Plus | 2.13.1 | UI 组件库 |
| Tailwind CSS | 3.4.1 | CSS 框架 |

---

## 快速开始

```bash
npm install          # 安装依赖
npm run dev          # 开发模式 (端口 3000)
npm run build        # 生产构建
npm run lint         # 代码检查
```

### 环境变量

```bash
# .env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

---

## 项目结构

```
src/
├── api/            # API 接口层
├── components/     # 组件 (business/forms/tables/layout/ui)
├── composables/    # 组合式函数
├── config/         # 配置文件
├── lib/            # 业务库 (auth/tenant/validators)
├── router/         # 路由 (含 guards/)
├── stores/         # Pinia 状态管理
├── styles/         # 样式文件
├── types/          # TypeScript 类型
├── utils/          # 工具函数
└── views/          # 页面组件
```

### 目录设计原则

| 目录 | 用途 |
|------|------|
| `api/` | API 接口层，按业务模块组织 |
| `components/` | 按功能分组 (forms/tables/business) |
| `composables/` | Vue 3 组合式函数 |
| `config/` | 集中管理配置和常量 |
| `lib/` | 业务相关的共享类库 |
| `utils/` | 纯工具函数，无业务逻辑 |

---

## 开发规范

### 命名约定

| 类型 | 规范 | 示例 |
|------|------|------|
| 组件文件 | PascalCase | `UserProfile.vue` |
| 工具文件 | camelCase | `formatUtils.ts` |
| 常量 | UPPER_SNAKE_CASE | `API_BASE_URL` |
| CSS 类名 | kebab-case | `user-card` |

### 代码风格

- 使用 2 空格缩进
- 使用单引号
- 语句末尾添加分号
- 组件名使用多单词

### 组件开发

使用 **Composition API + `<script setup>`**：

```vue
<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  variant: { type: String, default: 'primary' }
})

const emit = defineEmits(['click'])
</script>
```

---

## 常见任务

| 任务 | 操作 |
|------|------|
| 添加页面 | 在 `src/views/` 创建 `.vue`，在 `src/router/` 添加路由 |
| 添加组件 | 优先使用 Element Plus，自定义组件放 `src/components/` |
| 添加 API | 在 `src/api/modules/` 添加模块 |
| 添加翻译 | 在 `src/locales/modules/` 添加中英文翻译 |

---

## 核心架构概述

| 模块 | 说明 | 详细文档 |
|------|------|----------|
| **多租户** | 租户隔离、上下文管理 | [multi-tenant-architecture.md](.claude/knowledge/multi-tenant-architecture.md) |
| **国际化** | 中文/英文切换 | [i18n-implementation.md](.claude/knowledge/i18n-implementation.md) |
| **设计系统** | 颜色、字体、组件 | [design-system.md](.claude/knowledge/design-system.md) |
| **项目结构** | 目录组织、最佳实践 | [006-project-structure-optimization.md](.claude/knowledge/006-project-structure-optimization.md) |

### 核心状态管理

| Store | 职责 |
|-------|------|
| `useAuthStore` | 用户认证状态 |
| `useTenantsStore` | 租户数据管理 |
| `useServicesStore` | 服务状态管理 |
| `useUIStore` | UI 状态 (侧边栏、主题) |

---

## 注意事项

1. **认证**: 当前使用 Mock 认证，需对接真实后端 API
2. **API 响应格式**: `{ code: 200, data: {...}, message: "success" }`
3. **租户隔离**: 所有 API 请求自动携带 `X-Tenant-ID` 请求头
4. **Token 管理**: 存储 localStorage，401 时自动清除
5. **国际化**: 新增文本必须同时添加中英文翻译
6. **Element Plus**: 图标已全局注册，使用 `<el-icon><IconName /></el-icon>`
7. **图标库**: 使用 `lucide-vue-next`，见 [003-icon-library-migration.md](.claude/knowledge/003-icon-library-migration.md)

---

## 知识库

详细文档位于 `.claude/knowledge/`：

### 问题索引
| ID | 问题 | 状态 |
|----|------|------|
| [001](.claude/knowledge/001-text-selection-invisible.md) | 文本选中后不可见 | ✅ |
| [002](.claude/knowledge/002-element-plus-double-border.md) | Element Plus 双层边框 | ✅ |
| [003](.claude/knowledge/003-icon-library-migration.md) | 图标库迁移 | ✅ |
| [004](.claude/knowledge/004-settings-drawer-positioning.md) | 设置抽屉定位 | ✅ |
| [005](.claude/knowledge/005-html-nesting-warning.md) | HTML 嵌套警告 | ✅ |
| [006](.claude/knowledge/006-project-structure-optimization.md) | 项目结构优化 | ✅ |
| [007](.claude/knowledge/007-claudemd-simplification.md) | CLAUDE.md 精简 | ✅ |

### 实现文档
| 文档 | 说明 |
|-----|------|
| [multi-tenant-architecture.md](.claude/knowledge/multi-tenant-architecture.md) | 多租户架构 |
| [design-system.md](.claude/knowledge/design-system.md) | 设计系统 |
| [i18n-implementation.md](.claude/knowledge/i18n-implementation.md) | 国际化实现 |

### 查看完整知识库
```bash
cat .claude/knowledge/README.md
```
