# StatusUtils 状态工具类使用说明

## 简介

`StatusUtils` 是一个统一处理状态值（1=激活/正常，2=停用/禁用）的工具类，避免在代码中散落各种 `status === 1` 的判断。

## 导入

```typescript
import { StatusUtils } from '@/utils/status'
// 或者使用组合式函数
import { useStatus } from '@/utils/status'

const statusUtils = useStatus()
```

## 常用方法

### 1. 判断状态

```typescript
// 判断是否为激活状态 (status === 1)
StatusUtils.isActive(row.status)  // boolean

// 判断是否为停用状态 (status === 2)
StatusUtils.isInactive(row.status)  // boolean
```

### 2. 获取 Element Plus 组件类型

```vue
<!-- el-tag 的 type -->
<el-tag :type="StatusUtils.getTagType(row.status)">
  {{ StatusUtils.getStatusText(row.status) }}
</el-tag>

<!-- el-button 的 type -->
<el-button :type="StatusUtils.getButtonType(row.status, 'warning', 'success')">
  按钮
</el-button>
```

### 3. 获取状态文本

```typescript
// 默认: 正常/禁用
StatusUtils.getStatusText(row.status)  // '正常' | '禁用'

// 自定义文本
StatusUtils.getStatusText(row.status, '启用', '停用')  // '启用' | '停用'
StatusUtils.getStatusText(row.status, '上架', '下架')  // '上架' | '下架'
```

### 4. 切换状态

```typescript
// 切换状态值 (1 -> 2, 2 -> 1)
const newStatus = StatusUtils.toggleStatus(row.status)

// 获取切换操作文本
const action = StatusUtils.getToggleActionText(row.status)
// 默认返回: '禁用' (当 status=1) 或 '启用' (当 status=2)

// 自定义操作文本
const action = StatusUtils.getToggleActionText(row.status, '停用', '启用')
```

### 5. 完整示例

```vue
<template>
  <el-table :data="tableData">
    <el-table-column label="状态" width="100">
      <template #default="{ row }">
        <el-tag :type="StatusUtils.getTagType(row.status)">
          {{ StatusUtils.getStatusText(row.status) }}
        </el-tag>
      </template>
    </el-table-column>

    <el-table-column label="操作" width="200">
      <template #default="{ row }">
        <el-button
          :type="StatusUtils.getButtonType(row.status, 'warning', 'success')"
          @click="handleToggle(row)"
        >
          <el-icon>
            <component :is="StatusUtils.isActive(row.status) ? 'Lock' : 'Unlock'" />
          </el-icon>
          {{ StatusUtils.getToggleActionText(row.status) }}
        </el-button>
      </template>
    </el-table-column>
  </el-table>
</template>

<script setup lang="ts">
import { StatusUtils } from '@/utils/status'

const handleToggle = (row: any) => {
  const newStatus = StatusUtils.toggleStatus(row.status)
  const action = StatusUtils.getToggleActionText(newStatus, '启用', '禁用')
  console.log(`${action}后状态: ${newStatus}`)
}
</script>
```

## 状态常量

```typescript
import { STATUS } from '@/utils/status'

// 使用常量避免魔法数字
const activeStatus = STATUS.ACTIVE  // 1
const inactiveStatus = STATUS.INACTIVE  // 2
```

## 注意事项

1. `StatusUtils` 会自动处理 `null`、`undefined` 和字符串类型的 status，统一转换为数字判断
2. 所有方法都是静态方法，无需实例化
3. 默认状态: 1=激活/正常，2=停用/禁用
