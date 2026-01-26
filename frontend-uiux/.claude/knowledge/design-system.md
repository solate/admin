# 设计系统

> 本文档详细说明项目的设计系统，包括颜色、字体、组件规范和 Glassmorphism 风格

---

## 概述

项目采用 **Glassmorphism（毛玻璃）** 设计风格，结合 Tailwind CSS 和 Element Plus 组件库，打造现代化的多租户 SaaS 管理平台。

---

## 颜色系统

### 主色调

```javascript
// 主色 - 专业与信任
primary: {
  DEFAULT: '#2563eb',  // blue-600
  50:   'rgb(37 99 235 / 0.05)',   // #eff6ff
  100:  'rgb(37 99 235 / 0.1)',    // #dbeafe
  200:  'rgb(37 99 235 / 0.2)',    // #bfdbfe
  300:  'rgb(37 99 235 / 0.35)',   // #93c5fd
  400:  'rgb(37 99 235 / 0.5)',    // #60a5fa
  500:  'rgb(37 99 235 / 0.75)',   // #3b82f6 (鲜艳)
  600:  'rgb(37 99 235 / 1)',      // #2563eb (最鲜艳)
  700:  '#1d4ed8',
  800:  '#1e40af',
  900:  '#1e3a8a',
  950:  '#172554',
}

// CTA 色 - 对比与行动
cta: {
  DEFAULT: '#f97316',  // orange-500
  500: '#f97316',
  600: '#ea580c',
}
```

### 语义色

```javascript
success: '#22c55e'  // green-500
warning: '#f59e0b'  // amber-500
error:   '#ef4444'  // red-500
info:    '#0ea5e9'  // sky-500
```

### Element Plus 同步

```css
/* src/styles/variables.css */
:root {
  --el-color-primary: #2563eb;
  --el-color-success: #22c55e;
  --el-color-warning: #f59e0b;
  --el-color-danger: #ef4444;
}
```

---

## 字体系统

### 字体族

```css
/* src/styles/variables.css */
:root {
  --font-sans: 'Fira Sans', system-ui, -apple-system, sans-serif;
  --font-mono: 'Fira Code', 'Courier New', monospace;
}
```

### 使用规范

| 场景 | 字体 | 示例 |
|------|------|------|
| 正文、标题 | Sans | 用户管理、系统设置 |
| 代码、数据 | Mono | `user_id = 123` |
| 数字、金额 | Mono | ¥1,234.56 |

---

## Glassmorphism 风格

### 毛玻璃效果

```vue
<template>
  <div class="glass-card">
    <h3>卡片标题</h3>
    <p>卡片内容</p>
  </div>
</template>

<style scoped>
.glass-card {
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 16px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.05);
}
</style>
```

### 实用类

```css
/* 毛玻璃卡片 */
.glass {
  @apply bg-white/70 backdrop-blur-md border border-white/30 rounded-2xl;
}

/* 浮动阴影 */
.floating {
  @apply shadow-lg shadow-primary/5;
}

/* 渐变背景 */
.gradient-bg {
  @apply bg-gradient-to-br from-primary/10 via-transparent to-cta/10;
}
```

---

## 组件规范

### 组件选择优先级

1. **Element Plus 组件** (优先)
   - `el-button` - 按钮
   - `el-input` - 输入框
   - `el-table` - 表格
   - `el-dialog` - 对话框
   - `el-form` - 表单
   - `el-dropdown` - 下拉菜单

2. **Tailwind CSS** (布局/样式)
   - 布局: `flex`, `grid`, `gap-4`
   - 间距: `p-4`, `m-2`, `space-y-2`
   - 颜色: `text-primary-600`, `bg-gray-100`

3. **自定义组件** (`src/components/ui/`)
   - 需要特殊交互时创建

### 按钮规范

```vue
<!-- 主按钮 -->
<el-button type="primary">确认</el-button>

<!-- 次要按钮 -->
<el-button>取消</el-button>

<!-- CTA 按钮 -->
<el-button type="warning">立即升级</el-button>

<!-- 危险操作 -->
<el-button type="danger">删除</el-button>
```

### 表单规范

```vue
<el-form :model="form" :rules="rules" label-width="100px">
  <el-form-item label="用户名" prop="username">
    <el-input v-model="form.username" />
  </el-form-item>

  <el-form-item label="邮箱" prop="email">
    <el-input v-model="form.email" type="email" />
  </el-form-item>
</el-form>
```

---

## 间距系统

| Tailwind 类 | 值 | 使用场景 |
|-------------|-----|----------|
| `gap-2` | 8px | 紧密排列 |
| `gap-4` | 16px | 默认间距 |
| `gap-6` | 24px | 宽松间距 |
| `p-4` | 16px | 卡片内边距 |
| `p-6` | 24px | 大卡片内边距 |
| `px-4 py-2` | 16px 8px | 按钮内边距 |

---

## 圆角系统

| Tailwind 类 | 值 | 使用场景 |
|-------------|-----|----------|
| `rounded` | 4px | 小元素 |
| `rounded-lg` | 8px | 卡片 |
| `rounded-xl` | 12px | 对话框 |
| `rounded-2xl` | 16px | 毛玻璃卡片 |
| `rounded-full` | 50% | 头像、徽章 |

---

## 阴影系统

| Tailwind 类 | 使用场景 |
|-------------|----------|
| `shadow-sm` | 轻微阴影 |
| `shadow` | 默认阴影 |
| `shadow-lg` | 卡片阴影 |
| `shadow-xl` | 弹出层 |
| `shadow-primary/5` | 彩色阴影 |

---

## 响应式设计

### 断点

```javascript
// tailwind.config.js
screens: {
  'sm': '640px',
  'md': '768px',
  'lg': '1024px',
  'xl': '1280px',
  '2xl': '1536px',
}
```

### 使用示例

```vue
<template>
  <!-- 移动端垂直，桌面端水平 -->
  <div class="flex flex-col md:flex-row gap-4">
    <div>内容 1</div>
    <div>内容 2</div>
  </div>
</template>
```

---

## 深色模式

### 切换方式

```typescript
import { useTheme } from '@/composables/useTheme'

const { isDark, toggleDarkMode } = useTheme()
```

### 样式适配

```vue
<template>
  <div class="bg-white dark:bg-gray-900">
    <p class="text-gray-900 dark:text-white">内容</p>
  </div>
</template>
```

---

## 图标系统

项目使用 `lucide-vue-next` 图标库：

```vue
<template>
  <el-icon>
    <Plus />
  </el-icon>
</template>

<script setup>
import { Plus, Edit, Trash, Search } from 'lucide-vue-next'
</script>
```

详细图标规范见 [003-icon-library-migration.md](003-icon-library-migration.md)

---

## 相关文件

| 文件 | 说明 |
|------|------|
| `tailwind.config.js` | Tailwind 配置 |
| `src/styles/variables.css` | CSS 变量 |
| `src/styles/base.css` | 基础样式 |
| `src/config/theme.ts` | 主题配置 |
| `src/plugins/theme.ts` | 主题初始化 |

---

| 状态 | ✅ 已实现 |
|------|----------|
| 日期 | 2026-01-26 |
