# 图标库迁移到 lucide-vue-next

> 问题描述：项目从 `@element-plus/icons-vue` 迁移到 `lucide-vue-next`，需要修复所有图标导入和使用问题。

---

## 完整图标名称对照表

| 错误/废弃名称 | 正确名称 | 说明 |
|--------------|----------|------|
| `OfficeBuilding` | `Building` | 建筑图标 |
| `Close` | `X` | 关闭图标 |
| `View` | `Eye` | 查看图标 |
| `Hide` | `EyeOff` | 隐藏图标 |
| `Edit` | `Pencil` | 编辑图标 |
| `Delete` | `Trash2` | 删除图标 |
| `Coin` | `Coins` | 金币图标 |
| `DataAnalysis` | `BarChart3` | 数据分析图表 |
| `Message` | `Mail` | 邮件图标 |
| `Setting` | `Settings` | 设置图标（单数） |
| `ArrowRight` | `ChevronRight` | 右箭头 |
| `ArrowLeft` | `ChevronLeft` | 左箭头 |
| `ArrowDown` | `ChevronDown` | 下箭头 |
| `Lightning` | `Zap` | 闪电图标 |
| `CheckCircle` | `CircleCheck` | 勾选圆圈 |
| `InformationCircle` | `Info` | 信息图标 |
| `AlertCircle` | `TriangleAlert` | 警告圆圈 |
| `ExclamationTriangle` | `AlertTriangle` | 感叹三角 |
| `Envelope` | `Mail` | 信封图标 |
| `Cube` | `Box` | 盒子图标 |
| `ChartBar` | `BarChart3` | 柱状图 |
| `XMark` | `X` | X 标记 |
| `DevicePhoneMobile` | `Smartphone` | 手机图标 |
| `IdentificationCard` | `Badge` | 徽章图标 |
| `InfoFilled` | `Info` | 信息图标 |
| `MagicStick` | - | 不存在，移除 |
| `SwitchButton` | `LogOut` | 登出按钮（上下文相关） |

---

## 正确的图标导入方式

```vue
<script setup>
// ✅ 正确 - 直接从 lucide-vue-next 导入
import { Search, User, Mail, X, ChevronLeft, Building } from 'lucide-vue-next'
</script>

<template>
  <!-- 两种方式都可以 -->
  <Search :size="20" />
  <Mail :size="16" class="text-slate-400" />

  <!-- 在 el-icon 中使用也正确 -->
  <el-icon :size="20">
    <Building />
  </el-icon>
</template>
```

---

## 常见错误

### 错误 1：使用未定义的 `icons` 变量

```vue
<script setup>
// ❌ 错误 - icons 未定义
const { Building, User, Mail } = icons

// ✅ 正确
import { Building, User, Mail } from 'lucide-vue-next'
</script>
```

### 错误 2：使用不存在的图标名称

```vue
<script setup>
// ❌ 错误 - lucide-vue-next 中没有 OfficeBuilding
import { OfficeBuilding } from 'lucide-vue-next'

// ✅ 正确
import { Building } from 'lucide-vue-next'
</script>
```

---

## 已修复的文件列表

- `src/views/users/UserListView.vue`
- `src/views/services/ServiceListView.vue`
- `src/views/tenants/TenantListView.vue`
- `src/views/analytics/AnalyticsView.vue`
- `src/components/user/UserMenu.vue`
- `src/components/notification/NotificationCenter.vue`
- `src/components/tenant/TenantSelector.vue`
- `src/views/settings/SettingsView.vue`
- `src/views/LandingView.vue`
- `src/views/tenants/TenantDetailView.vue`
- `src/views/services/ServiceDetailView.vue`
- `src/views/auth/RegisterView.vue`
- `src/views/NotFoundView.vue`
- `src/views/notifications/NotificationView.vue`
- `src/views/business/BusinessView.vue`
- `src/views/users/UserDetailView.vue`
- `src/views/profile/ProfileView.vue`

---

## 查询可用图标

访问 [lucide.dev/icons](https://lucide.dev/icons) 查看所有可用的图标及其正确名称。

---

| 状态 | ✅ 已完成 |
|------|----------|
| 完成日期 | 2025-01-23 |
