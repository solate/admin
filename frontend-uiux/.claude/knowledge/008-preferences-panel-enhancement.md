# 设置面板系统增强

> 需求描述：用户希望参考业界 SaaS 系统（如 Wise、Superlist、Cal.com 等）的设计，增强设置面板的功能和交互体验，支持更丰富的配置选项和更好的分类组织。

---

## 背景调研

### 业界 SaaS 设置面板设计模式

通过调研 [SaaSFrame](https://www.saasframe.io/categories/settings) 的 167 个设置设计案例，发现主流模式：

| UI 模式 | 代表产品 | 特点 |
|---------|----------|------|
| **右侧抽屉** | Wise, Superlist, Cal.com | 设置时不离开当前页面，支持实时预览 |
| **独立页面** | Intercom, Vercel, Supabase | 设置项很多时使用，完整的导航体系 |
| **弹窗 Modal** | 轻量级 SaaS | 设置项较少时使用 |

### Vben Admin 配置系统参考

Vben Admin 提供了完整的 preferences 配置系统，支持：
- 12 个配置模块（app、theme、layout、widget 等）
- 配置覆盖机制（`overridesPreferences`）
- 完整的 TypeScript 类型定义

---

## 设计决策

### 1. 保持现有架构

**决策**：保持 4 个 Tab 的分类结构（外观、布局、通用、快捷键），通过折叠面板组织复杂选项。

**原因**：
- 用户已熟悉的交互模式
- 避免一次性展示过多选项造成认知负担
- 符合"渐进式增强"原则

### 2. 使用折叠面板

**决策**：将高级选项收纳到可折叠的面板中。

**原因**：
- 保持界面整洁
- 降低复杂度的感知
- 让基础用户不被高级选项干扰

### 3. 精确数值控制

**决策**：侧边栏宽度、顶栏高度等使用滑块提供精确数值控制。

**原因**：
- 比"窄/中/宽"更灵活
- 实时反馈数值
- 满足高级用户需求

---

## 实现内容

### 类型系统扩展

`src/types/preferences.ts` - 新增类型：

```typescript
// 外观设置新增
export interface AppearancePreferences {
  darkHeader: boolean           // 深色顶栏
  customBorderRadius?: string  // 自定义圆角
  grayMode: boolean             // 灰色模式
}

// 布局设置大幅扩展
export interface LayoutPreferences {
  layoutMode: 'sidebar' | 'topbar' | 'mixed' | 'horizontal'
  sidebarWidth: number                      // 精确像素值
  sidebarCollapsedWidth: number
  sidebarCollapsible: boolean
  headerMode: 'static' | 'fixed' | 'auto-hide'
  contentWidthMode: 'fluid' | 'fixed'
  tabsStyle: 'chrome' | 'plain' | 'card' | 'smart'
  // ... 更多选项
}

// 新增快捷键偏好
export interface ShortcutPreferences {
  enable: boolean
  globalSearch: boolean
  globalLogout: boolean
  // ...
}

// 新增小部件偏好
export interface WidgetPreferences {
  globalSearch: boolean
  themeToggle: boolean
  position: 'auto' | 'header' | 'sidebar'
  // ...
}

// 新增版权偏好
export interface CopyrightPreferences {
  companyName: string
  companySiteLink: string
  icp: string
  // ...
}
```

### Store 增强

`src/stores/modules/preferences.ts` - 新增功能：

```typescript
// 新增计算属性
const widgets = computed(() => preferences.value.widgets)
const copyright = computed(() => preferences.value.copyright)

// 新增更新方法
function updateWidgets<K extends keyof UserPreferences['widgets']>(...)
function updateCopyright<K extends keyof UserPreferences['copyright']>(...)

// 增强外观应用
function applyAppearanceSettings() {
  // 新增灰色模式支持
  if (appearance.value.grayMode) {
    root.setAttribute('data-gray-mode', 'true')
  }
  // 支持自定义圆角
  const borderRadiusMap = {
    custom: appearance.value.customBorderRadius || '0.5rem'
  }
}
```

### AppearanceTab 重构

**新增功能**：
- 高级选项折叠面板
- 深色顶栏开关
- 灰色模式开关
- 自定义圆角输入（当选择"自定义"时显示）

```vue
<!-- 折叠面板结构 -->
<section class="border-2 rounded-2xl overflow-hidden">
  <button @click="advancedExpanded = !advancedExpanded">
    高级选项
    <ChevronDown / ChevronUp>
  </button>
  <Transition>
    <div v-if="advancedExpanded">
      <!-- 高级选项内容 -->
    </div>
  </Transition>
</section>
```

### LayoutTab 重构

**新增功能**：
- 4 种布局模式（侧边栏/顶部/混合/水平）
- 侧边栏设置折叠面板：
  - 宽度滑块（180-400px）
  - 折叠宽度滑块（48-80px）
  - 可折叠、默认折叠、手风琴模式
- 顶栏设置折叠面板：
  - 高度滑块（48-80px）
  - 顶栏模式（静止/固定/自动隐藏）
- 内容区域模式（流式/定宽 + 自定义宽度）

```vue
<!-- 滑块示例 -->
<label>侧边栏宽度: {{ layout.sidebarWidth }}px</label>
<input
  type="range"
  :value="layout.sidebarWidth"
  min="180" max="400" step="4"
  @input="updateSidebarWidth($event)"
/>
```

### GeneralTab 增强

**新增功能**：
- 5 种页面切换动画效果（淡入淡出/滑动/缩放/缩放/淡入滑动）
- 动画效果折叠面板：
  - 页面切换进度条
  - 页面切换 Loading
  - 动态标题
  - 检查更新（间隔天数滑块）

```vue
<!-- 动画选项 -->
const pageTransitionOptions = [
  { value: 'fade', label: '淡入淡出', icon: 'fade' },
  { value: 'slide', label: '滑动', icon: 'slide' },
  { value: 'scale', label: '缩放', icon: 'scale' },
  { value: 'zoom', label: '缩放', icon: 'zoom' },
  { value: 'fade-slide', label: '淡入滑动', icon: 'fade-slide' }
]
```

---

## UI/UX 设计亮点

### 1. 视觉层次清晰

```
基础选项（始终可见）
  ├─ 主题模式（3 个卡片）
  ├─ 主题色（颜色按钮）
  └─ 圆角大小（5 个按钮）

高级选项（折叠面板）
  └─ 深色侧边栏、深色顶栏、灰色模式...
```

### 2. 交互反馈及时

- 滑块实时显示数值
- 开关按钮立即生效
- 选中状态清晰指示（边框 + 勾选图标）

### 3. 渐进式复杂度

```
┌─────────────────────────────────────┐
│ 基础用户：3 个主题模式 + 1 个圆角设置    │
├─────────────────────────────────────┤
│ 进阶用户：展开高级选项，开启深色模式    │
├─────────────────────────────────────┤
│ 高级用户：调整精确像素值、自定义圆角    │
└─────────────────────────────────────┘
```

### 4. 保持 Glassmorphism 风格

- 渐变背景：`bg-gradient-to-br from-primary-50 to-primary-100/50`
- 毛玻璃效果：`backdrop-blur-xl`
- 柔和阴影：`shadow-lg shadow-primary-500/25`

---

## 技术细节

### 折叠面板实现

使用 Vue 3 Transition + 响应式状态：

```typescript
const advancedExpanded = ref(false)

function toggleAdvanced() {
  advancedExpanded.value = !advancedExpanded.value
}
```

```vue
<Transition
  enter-active-class="transition-all duration-200 ease-out"
  enter-from-class="opacity-0 -translate-y-2"
  enter-to-class="opacity-100 translate-y-0"
  leave-active-class="transition-all duration-150 ease-in"
  leave-from-class="opacity-100 translate-y-0"
  leave-to-class="opacity-0 -translate-y-2"
>
  <div v-if="advancedExpanded">
    <!-- 内容 -->
  </div>
</Transition>
```

### 滑块实现

使用原生 range input + Tailwind 样式：

```vue
<input
  type="range"
  :value="layout.sidebarWidth"
  min="180" max="400" step="4"
  class="w-full h-2 bg-slate-200 dark:bg-slate-700
         rounded-lg appearance-none cursor-pointer
         accent-primary-500"
/>
```

### 配置持久化

```typescript
// 监听变化自动保存到 localStorage
watch(
  preferences,
  (newPreferences) => {
    localStorage.setItem(
      PREFERENCES_STORAGE_KEY,
      JSON.stringify(newPreferences)
    )
  },
  { deep: true }
)
```

---

## 配置选项完整列表

### 外观设置

| 选项 | 类型 | 可选值/范围 |
|------|------|-------------|
| 主题模式 | 枚举 | light / dark / auto |
| 主题色 | 字符串 | 8 种预设颜色 |
| 圆角大小 | 枚举 | none / small / medium / large / custom |
| 自定义圆角 | 字符串 | CSS 值 (如 "0.25rem") |
| 深色侧边栏 | 布尔 | true / false |
| 深色顶栏 | 布尔 | true / false |
| 灰色模式 | 布尔 | true / false |
| 色盲模式 | 枚举 | none / protanopia / deuteranopia / tritanopia |
| 高对比度 | 布尔 | true / false |

### 布局设置

| 选项 | 类型 | 可选值/范围 |
|------|------|-------------|
| 布局模式 | 枚举 | sidebar / topbar / mixed / horizontal |
| 侧边栏宽度 | 数值 | 180-400 (px) |
| 侧边栏折叠宽度 | 数值 | 48-80 (px) |
| 侧边栏可折叠 | 布尔 | true / false |
| 侧边栏默认折叠 | 布尔 | true / false |
| 导航样式 | 枚举 | icon-text / icon-only |
| 导航手风琴模式 | 布尔 | true / false |
| 顶栏模式 | 枚举 | static / fixed / auto-hide |
| 顶栏高度 | 数值 | 48-80 (px) |
| 内容宽度模式 | 枚举 | fluid / fixed |
| 内容定宽 | 数值 | 800-1920 (px) |
| 显示面包屑 | 布尔 | true / false |
| 显示标签页 | 布尔 | true / false |
| 标签页样式 | 枚举 | chrome / plain / card / smart |
| 标签页持久化 | 布尔 | true / false |
| 标签页可拖拽 | 布尔 | true / false |
| 显示小部件 | 布尔 | true / false |
| 显示页脚 | 布尔 | true / false |

### 通用设置

| 选项 | 类型 | 可选值/范围 |
|------|------|-------------|
| 语言 | 枚举 | zh-CN / en-US |
| 动态标题 | 布尔 | true / false |
| 启用动画 | 布尔 | true / false |
| 页面切换动画 | 枚举 | fade / slide / scale / zoom / fade-slide |
| 页面切换进度条 | 布尔 | true / false |
| 页面切换 Loading | 布尔 | true / false |
| 检查更新间隔 | 数值 | 1-30 (天) |
| 启用检查更新 | 布尔 | true / false |

---

## 文件变更清单

### 修改的文件

1. **src/types/preferences.ts**
   - 扩展 `AppearancePreferences` 接口
   - 重构 `LayoutPreferences` 接口
   - 新增 `ShortcutPreferences`、`WidgetPreferences`、`CopyrightPreferences`
   - 更新 `DEFAULT_PREFERENCES` 默认值

2. **src/stores/modules/preferences.ts**
   - 新增 `widgets`、`copyright` 计算属性
   - 新增 `updateWidgets`、`updateCopyright` 方法
   - 增强 `applyAppearanceSettings` 函数
   - 更新 `updateShortcut` 为类型安全的方法

3. **src/components/preferences/tabs/AppearanceTab.vue**
   - 新增高级选项折叠面板
   - 新增深色顶栏、灰色模式开关
   - 圆角选项新增"自定义"
   - 优化布局和视觉层次

4. **src/components/preferences/tabs/LayoutTab.vue**
   - 布局模式新增 4 种选项
   - 新增侧边栏设置折叠面板（宽度、折叠、手风琴）
   - 新增顶栏设置折叠面板（高度、模式）
   - 新增内容区域设置（流式/定宽）

5. **src/components/preferences/tabs/GeneralTab.vue**
   - 页面切换动画新增 5 种效果
   - 新增动画效果折叠面板
   - 新增页面进度条、Loading、动态标题开关
   - 新增检查更新设置（开关 + 滑块）

---

## 使用示例

### 在组件中使用设置

```typescript
import { usePreferencesStore } from '@/stores/modules/preferences'

const preferencesStore = usePreferencesStore()

// 初始化（通常在 App.vue 中调用一次）
preferencesStore.initialize()

// 更新设置
preferencesStore.updateAppearance('themeMode', 'dark')
preferencesStore.updateLayout('sidebarWidth', 300)
preferencesStore.updateGeneral('pageProgress', true)

// 监听设置变化
watch(() => preferencesStore.appearance.themeMode, (newMode) => {
  console.log('主题模式已变:', newMode)
})
```

### 应用 CSS 变量到组件

```vue
<script setup>
import { usePreferencesStore } from '@/stores/modules/preferences'

const preferencesStore = usePreferencesStore()
const layout = computed(() => preferencesStore.layout)
</script>

<template>
  <aside
    class="transition-all duration-300"
    :style="{
      width: `${layout.sidebarWidth}px`,
      transform: layout.sidebarCollapsed ? `translateX(-${layout.sidebarWidth - layout.sidebarCollapsedWidth}px)` : 'translateX(0)'
    }"
  >
    <!-- 侧边栏内容 -->
  </aside>
</template>
```

---

## 未来扩展方向

### 1. 添加 CSS 样式支持

需要为新增的数据属性添加样式：

```css
/* 灰色模式 */
[data-gray-mode="true"] * {
  filter: grayscale(100%);
}

/* 深色顶栏 */
[data-dark-header="true"] .topbar {
  background: dark-slate-800;
}
```

### 2. 实现高级 Tab（可选）

如果需要更复杂的设置，可以添加：
- 小部件配置（控制各个小部件的显示/隐藏）
- 快捷键绑定（自定义快捷键）
- 版权信息配置

### 3. 配置导入导出优化

当前已支持 JSON 格式，可以考虑：
- 添加配置文件命名（如 `preferences-backup-2025-01-26.json`）
- 支持配置分享（生成分享链接）
- 配置验证（防止导入无效配置）

### 4. 实时预览功能

在设置面板中实时预览效果：
- 主题切换时立即显示效果
- 布局调整时显示示意图
- 圆角调整时实时预览

---

## 相关资源

- [SaaSFrame - Settings](https://www.saasframe.io/categories/settings) - 167 个 SaaS 设置设计案例
- [Vben Admin - Configuration](https://doc.vben.pro/en/guide/essentials/settings.html) - 完整配置系统文档
- [Vben Admin GitHub](https://github.com/vbenjs/vue-vben-admin) - 开源代码参考

---

## 设计原则总结

1. **渐进式复杂度**：基础选项始终可见，高级选项折叠收纳
2. **即时反馈**：所有设置变更立即生效，无需保存按钮
3. **类型安全**：完整的 TypeScript 类型定义，编译时检查
4. **可扩展性**：模块化设计，易于添加新选项
5. **用户友好**：清晰的视觉层次，直观的交互控件

---

| 状态 | ✅ 已实现 |
|------|----------|
| 实现日期 | 2026-01-26 |
| 版本 | v1.0.0 |
