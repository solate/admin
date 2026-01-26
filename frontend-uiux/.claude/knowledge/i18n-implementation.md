# 多语言 (i18n) 实现文档

## 概述

本项目使用 Vue I18n 实现多语言功能，支持简体中文 (zh-CN) 和英文 (en-US)，并可轻松扩展更多语言。

## 技术栈

- **Vue I18n 9.14.5** - 国际化核心库
- **Element Plus** - UI 组件库，同步多语言
- **TypeScript** - 类型安全的翻译键

## 架构设计

### 文件结构

```
src/locales/
├── index.ts              # i18n 配置、导出和工具函数
├── types.ts              # TypeScript 类型定义
├── composables.ts        # 类型安全的 useI18n() composable
├── zh-CN.ts              # 中文语言包入口
├── en-US.ts              # 英文语言包入口
├── modules/              # 中文翻译模块（14个）
│   ├── common.json       # 通用翻译（按钮、状态、操作、404等）
│   ├── nav.json          # 导航菜单
│   ├── auth.json         # 认证相关（登录、注册、密码要求等）
│   ├── user.json         # 用户相关（用户管理、角色等）
│   ├── tenant.json       # 租户相关（租户管理、套餐、状态等）
│   ├── service.json      # 服务相关（服务管理、分类、状态等）
│   ├── settings.json     # 设置相关（系统设置各模块）
│   ├── analytics.json    # 数据分析（统计、图表、趋势等）
│   ├── userMenu.json     # 用户菜单（账户、工作区、登出等）
│   ├── dashboard.json    # Dashboard 概览（快捷操作、活动等）
│   ├── business.json     # 业务管理（收入、订单、统计等）
│   ├── notification.json # 通知中心（筛选、时间格式等）
│   └── profile.json      # 个人资料（表单、安全、会话等）
└── en-US/modules/        # 英文翻译模块（相同结构）
```

### 模块化设计原则

1. **按功能模块划分** - 每个业务功能对应一个 JSON 模块
2. **模块前缀访问** - 使用 `t('模块名.键名')` 格式访问翻译
3. **易于扩展** - 新增语言只需创建新的语言目录和模块文件

## 核心实现

### 1. i18n 配置 (`index.ts`)

```typescript
import { createI18n } from 'vue-i18n'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'

// Element Plus locale 映射
export const elementLocales = {
  'zh-CN': zhCn,
  'en-US': en
}

// 支持的语言列表
export const SUPPORT_LOCALES = ['zh-CN', 'en-US'] as const
export type SupportedLocale = typeof SUPPORT_LOCALES[number]

// 创建 i18n 实例
export function createI18nInstance(): I18n {
  return createI18n({
    legacy: false,                    // 使用 Composition API
    locale: getInitialLocale(),        // 初始语言
    fallbackLocale: 'en-US',          // 回退语言
    messages: {
      'zh-CN': zhCNMessages,
      'en-US': enUSMessages
    },
    globalInjection: true             // 全局注入 $t
  })
}

// 设置语言（同步 Vue I18n、Element Plus 和 localStorage）
export async function setLocale(locale: SupportedLocale) {
  i18n.global.locale.value = locale
  localStorage.setItem('locale', locale)
  document.documentElement.lang = locale

  // 同步 Element Plus locale
  const { useUiStore } = await import('@/stores/modules/ui')
  const uiStore = useUiStore()
  uiStore.setLocale(locale)
}
```

### 2. 类型定义 (`types.ts`)

```typescript
// 支持的语言配置
export const LOCALE_CONFIGS: Record<SupportedLocale, LocaleConfig> = {
  'zh-CN': { code: 'zh-CN', name: '简体中文', flag: '🇨🇳' },
  'en-US': { code: 'en-US', name: 'English', flag: '🇺🇸' }
}

// 翻译消息 Schema（用于类型推导）
export interface MessageSchema {
  common: typeof import('./modules/common.json').default
  nav: typeof import('./modules/nav.json').default
  auth: typeof import('./modules/auth.json').default
  // ... 其他模块
}

// 类型安全的翻译键
export type TranslationKey = Paths<MessageSchema>
```

### 3. 类型安全的 Composable (`composables.ts`)

```typescript
import { useI18n as useVueI18n } from 'vue-i18n'

export function useI18n() {
  const i18n = useVueI18n()

  const t: TranslateFunction = (key: any, ...args: any[]) => {
    return i18n.t(key, ...args) as string
  }

  const locale = computed(() => i18n.locale.value as SupportedLocale)

  return {
    t,
    locale,
    setLocale,
    availableLocales: ['zh-CN', 'en-US'],
    i18n
  }
}
```

### 4. 语言包入口 (`zh-CN.ts` / `en-US.ts`)

```typescript
// 导入各模块翻译
import common from './modules/common.json'
import nav from './modules/nav.json'
import auth from './modules/auth.json'
// ... 其他模块

export default {
  common,
  nav,
  auth,
  // ... 其他模块
}
```

## 使用方式

### 在组件中使用

```vue
<script setup>
import { useI18n } from '@/locales/composables'

const { t, locale } = useI18n()
</script>

<template>
  <!-- 导航菜单 -->
  <span>{{ t('nav.dashboard') }}</span>

  <!-- 通用按钮 -->
  <button>{{ t('common.save') }}</button>

  <!-- 用户菜单 -->
  <span>{{ t('userMenu.profile') }}</span>
</template>
```

### 在模板中直接使用 `$t`

```vue
<template>
  <!-- 需要使用完整路径（带模块前缀） -->
  <button :aria-label="$t('common.language.title')">
    切换语言
  </button>
</template>
```

### 动态翻译键

```vue
<script setup>
const { t } = useI18n()
const authStore = useAuthStore()

// 使用模板字符串
const userRoleLabel = computed(() => {
  return t(`user.roles.${authStore.userRole}`)
})
</script>
```

## 添加新翻译

### 1. 添加新的翻译键

在对应的模块 JSON 文件中添加：

**中文** (`src/locales/modules/common.json`):
```json
{
  "newFeature": "新功能",
  "newFeatureDesc": "这是一个新功能的描述"
}
```

**英文** (`src/locales/en-US/modules/common.json`):
```json
{
  "newFeature": "New Feature",
  "newFeatureDesc": "This is a description of a new feature"
}
```

### 2. 在组件中使用

```vue
<template>
  <h1>{{ t('common.newFeature') }}</h1>
  <p>{{ t('common.newFeatureDesc') }}</p>
</template>
```

## 添加新语言

### 1. 创建语言目录和模块

```bash
# 创建新的语言目录
mkdir -p src/locales/ja-JP/modules

# 复制并翻译所有模块
cp src/locales/modules/*.json src/locales/ja-JP/modules/
# 然后翻译每个 JSON 文件
```

### 2. 创建语言入口文件

**`src/locales/ja-JP.ts`**:
```typescript
import common from './ja-JP/modules/common.json'
import nav from './ja-JP/modules/nav.json'
// ... 其他模块

export default {
  common,
  nav,
  // ... 其他模块
}
```

### 3. 更新配置

**`src/locales/index.ts`**:
```typescript
import jaJP from './ja-JP'

export const SUPPORT_LOCALES = ['zh-CN', 'en-US', 'ja-JP'] as const

export const elementLocales = {
  'zh-CN': zhCn,
  'en-US': en,
  'ja-JP': ja  // 需要从 element-plus 导入
}

messages: {
  'zh-CN': zhCNMessages,
  'en-US': enUSMessages,
  'ja-JP': jaJP  // 新增
}
```

**`src/locales/types.ts`**:
```typescript
export const LOCALE_CONFIGS: Record<SupportedLocale, LocaleConfig> = {
  'zh-CN': { code: 'zh-CN', name: '简体中文', flag: '🇨🇳' },
  'en-US': { code: 'en-US', name: 'English', flag: '🇺🇸' },
  'ja-JP': { code: 'ja-JP', name: '日本語', flag: '🇯🇵' }
}
```

## Element Plus 多语言同步

Element Plus 组件库需要单独设置语言：

**`src/stores/modules/ui.ts`**:
```typescript
import { elementLocales } from '@/locales'

const elementLocale = computed(() => {
  return elementLocales[locale.value as keyof typeof elementLocales]
})
```

**在模板中使用**:
```vue
<template>
  <el-config-provider :locale="elementLocale">
    <App />
  </el-config-provider>
</template>
```

## 常见问题

### Q1: 翻译键找不到，显示 key 本身

**原因**: 翻译键路径不正确

**解决**: 确保使用完整的模块路径，例如 `t('common.save')` 而不是 `t('save')`

### Q2: Element Plus 组件还是中文

**原因**: Element Plus locale 未同步

**解决**: 确保在 `setLocale` 函数中调用了 `uiStore.setLocale(locale)`

### Q3: 新增翻译后不生效

**原因**: 开发服务器缓存

**解决**: 重启开发服务器或硬刷新浏览器 (Ctrl+Shift+R)

### Q4: 类型错误 "key is not assignable to parameter"

**原因**: TypeScript 类型定义不完整

**解决**: 在 `types.ts` 的 `MessageSchema` 中添加对应模块的类型定义

### Q5: computed 属性在模板中访问报错

**错误**: `Cannot read properties of undefined (reading 'length')`

**原因**: 在模板中使用 `categories.value.length`，但 computed 属性会自动解包

**解决**: 模板中使用 `categories.length` 而不是 `categories.value.length`

## 最佳实践

1. **始终使用模块前缀**: `t('common.save')` 而非 `t('save')`
2. **中英文同步添加**: 每次添加翻译时同时更新中英文文件
3. **使用类型安全的 `useI18n`**: 获得自动补全和类型检查
4. **复数形式**: 使用 `tc` 函数处理复数
5. **日期/数字格式化**: 使用 `d` 和 `n` 函数保持格式一致

## 已实现多语言的页面

| 页面类型 | 文件 |
|---------|------|
| **认证** | LoginView.vue, RegisterView.vue |
| **Dashboard** | OverviewView.vue |
| **租户管理** | TenantListView.vue, TenantDetailView.vue |
| **服务管理** | ServiceListView.vue, ServiceDetailView.vue |
| **用户管理** | UserListView.vue, UserDetailView.vue |
| **数据分析** | AnalyticsView.vue |
| **业务管理** | BusinessView.vue |
| **通知** | NotificationView.vue |
| **个人资料** | ProfileView.vue |
| **系统设置** | SettingsView.vue |
| **404页面** | NotFoundView.vue |

## 实现历程

### 第一阶段：框架搭建
- 创建 i18n 配置和类型系统
- 建立模块化 JSON 语言包结构
- 实现类型安全的 useI18n composable
- 同步 Element Plus 多语言

### 第二阶段：核心页面改造
- Dashboard 概览页
- 认证页面（登录/注册）
- 导航和布局组件

### 第三阶段：业务页面全面覆盖
- 租户管理（列表/详情）
- 服务管理（列表/详情）
- 用户管理（列表/详情）
- 数据分析页面
- 系统设置页面

### 第四阶段：功能页面补充
- 业务管理页面
- 通知中心
- 个人资料
- 404 错误页面

## 相关文件

### 语言包
- `src/locales/index.ts` - i18n 配置
- `src/locales/types.ts` - 类型定义
- `src/locales/composables.ts` - Composable
- `src/locales/zh-CN.ts` - 中文入口
- `src/locales/en-US.ts` - 英文入口
- `src/locales/modules/*.json` - 中文翻译（14个模块）
- `src/locales/en-US/modules/*.json` - 英文翻译（14个模块）

### 组件
- `src/components/language/LanguageSwitcher.vue` - 语言切换器
- `src/layouts/DashboardLayout.vue` - 布局组件
- `src/components/layout/TopNavbar.vue` - 顶部导航栏

### 状态管理
- `src/stores/modules/ui.ts` - UI 状态（Element Plus locale）
- `src/stores/modules/auth.ts` - 认证状态
