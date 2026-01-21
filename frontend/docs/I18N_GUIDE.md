# 前端多语言使用指南

## 概述

项目已集成 `vue-i18n` 实现多语言支持,目前支持简体中文和英文两种语言。

## 文件结构

```
frontend/
├── src/
│   ├── locales/
│   │   ├── index.ts       # i18n 配置入口
│   │   ├── zh-CN.ts       # 简体中文语言包
│   │   └── en-US.ts       # 英文语言包
│   ├── stores/
│   │   └── locale.ts      # 语言设置 store
│   └── components/
│       └── LocaleSelector.vue  # 语言切换组件
```

## 使用方法

### 1. 在模板中使用

```vue
<template>
  <div>{{ $t('common.save') }}</div>
  <div>{{ $t('login.title') }}</div>
</template>
```

### 2. 在 script 中使用

```vue
<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

// 显示消息
ElMessage.success(t('user.addSuccess'))
</script>
```

### 3. 带参数的翻译

```vue
<template>
  <div>{{ $t('common.total', { n: 10 }) }}</div>
  <!-- 输出: 共 10 条 -->
</template>
```

## 语言切换

用户可以通过顶部导航栏的语言切换器在中文和英文之间切换。语言设置会保存在 localStorage 中,下次访问时自动恢复。

## 添加新的翻译

### 1. 在语言文件中添加翻译

在 `src/locales/zh-CN.ts` 和 `src/locales/en-US.ts` 中添加对应的翻译:

```typescript
// zh-CN.ts
export default {
  // ...existing code
  myModule: {
    myKey: '我的翻译'
  }
}
```

```typescript
// en-US.ts
export default {
  // ...existing code
  myModule: {
    myKey: 'My Translation'
  }
}
```

### 2. 使用新添加的翻译

```vue
<template>
  <div>{{ $t('myModule.myKey') }}</div>
</template>
```

## Element Plus 组件语言

Element Plus 组件的语言会自动跟随应用语言设置,无需单独处理。

## 当前支持的语言

- `zh-CN`: 简体中文 (默认)
- `en-US`: English

## 添加新语言

如需添加更多语言支持:

1. 在 `src/locales/` 目录下创建新的语言文件 (如 `ja-JP.ts`)
2. 在 `src/locales/index.ts` 中导入并注册新语言
3. 在 `src/stores/locale.ts` 中更新支持的语言列表
4. 在 `LocaleSelector.vue` 中添加语言选项
5. 在 `App.vue` 中添加对应的 Element Plus 语言包
