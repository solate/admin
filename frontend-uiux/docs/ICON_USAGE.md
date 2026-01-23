# 图标使用指南

## 🎉 已配置 unplugin-icons

**重要提示**：unplugin-icons 自动解析的组件名格式为 `ILucide{图标名}`（驼峰式）

## 使用方式

### 方式一：使用自动解析的组件名（推荐）

```vue
<template>
  <!-- 无需导入，直接使用 -->
  <ILucideLanguages :size="20" />
  <ILucideSettings :size="20" />
  <ILucideUser />
  <ILucideSearch />
</template>
```

### 方式二：手动导入（如果自动解析不工作）

```vue
<script setup>
// 仍然可以手动导入
import { Languages, Settings, User } from 'lucide-vue-next'
</script>

<template>
  <Languages :size="20" />
  <Settings :size="20" />
  <User />
</template>
```

## 常用图标对照表

| 功能 | 自动解析名 | 手动导入 |
|------|------------|----------|
| 语言 | `ILucideLanguages` | `Languages` |
| 设置 | `ILucideSettings` | `Settings` |
| 用户 | `ILucideUser` | `User` |
| 首页 | `ILucideHome` | `Home` |
| 搜索 | `ILucideSearch` | `Search` |
| 通知 | `ILucideBell` | `Bell` |
| 太阳 | `ILucideSun` | `Sun` |
| 月亮 | `ILucideMoon` | `Moon` |
| 箭头右 | `ILucideChevronRight` | `ChevronRight` |
| 箭头左 | `ILucideChevronLeft` | `ChevronLeft` |
| 建筑 | `ILucideBuilding` | `Building` |
| 盒子 | `ILucideBox` | `Box` |
| 锁 | `ILucideLock` | `Lock` |
| 登出 | `ILucideLogOut` | `LogOut` |
| 添加 | `ILucidePlus` | `Plus` |
| 删除 | `ILucideTrash2` | `Trash2` |
| 编辑 | `ILucidePencil` | `Pencil` |
| 眼睛 | `ILucideEye` | `Eye` |
| 关闭 | `ILucideX` | `X` |

## 命名规则

**Lucide 图标自动解析规则**：
- 图标名 `languages` → 组件名 `ILucideLanguages`
- 图标名 `search` → 组件名 `ILucideSearch`
- 图标名 `chevron-right` → 组件名 `ILucideChevronRight`

**规则**：`ILucide` + 图标名（首字母大写，其余驼峰）

## 查找图标

访问 https://lucide.dev/icons/ 查看所有图标
- 图标页面会显示组件名，如 `Languages`
- 在模板中加上 `ILucide` 前缀即可

## 示例

```vue
<template>
  <div class="flex gap-4">
    <!-- 直接使用，无需 import -->
    <ILucideLanguages class="w-5 h-5" />
    <ILucideSettings :size="24" class="text-blue-500" />
    <ILucideBell />
  </div>
</template>
```

## 迁移建议

**对于新代码**：
- 直接使用 `ILucide{图标名}` 格式
- 无需手动 import

**对于已有代码**：
- 可以保持手动导入，继续使用
- 或者逐步迁移到自动解析格式

