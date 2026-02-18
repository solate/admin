<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePreferencesStore } from '@/stores/modules/preferences'
import { Layout, ChevronDown, ChevronUp } from 'lucide-vue-next'

const { t } = useI18n()
const preferencesStore = usePreferencesStore()

// 布局设置
const layout = computed(() => preferencesStore.layout)

// 折叠面板状态 - 默认全部展开
const sidebarExpanded = ref(true)
const headerExpanded = ref(true)
const tabsExpanded = ref(true)
const breadcrumbExpanded = ref(true)
const widgetsExpanded = ref(true)

// 布局模式选项
const layoutModeOptions = computed(() => [
  {
    value: 'sidebar' as const,
    label: '侧边栏模式',
    description: '左侧导航栏布局',
    icon: 'sidebar'
  },
  {
    value: 'double-sidebar' as const,
    label: '双列菜单',
    description: '图标主菜单+子菜单',
    icon: 'double-sidebar'
  },
  {
    value: 'topbar' as const,
    label: '顶部模式',
    description: '顶部导航栏布局',
    icon: 'topbar'
  },
  {
    value: 'mixed' as const,
    label: '混合模式',
    description: '一级顶部，二级侧边',
    icon: 'mixed'
  },
  {
    value: 'horizontal' as const,
    label: '水平模式',
    description: '完全水平导航布局',
    icon: 'horizontal'
  }
])

// 导航样式选项
const navStyleOptions = computed(() => [
  { value: 'icon-text' as const, label: '图标 + 文字' },
  { value: 'icon-only' as const, label: '仅图标' }
])

// 顶栏模式选项
const headerModeOptions = computed(() => [
  { value: 'static' as const, label: '静止' },
  { value: 'fixed' as const, label: '固定' },
  { value: 'auto-hide' as const, label: '自动隐藏' }
])

// 内容宽度模式选项
const contentWidthOptions = computed(() => [
  { value: 'fluid' as const, label: '流式' },
  { value: 'fixed' as const, label: '定宽' }
])

// 面包屑样式选项
const breadcrumbStyleOptions = computed(() => [
  { value: 'normal' as const, label: '常规' },
  { value: 'background' as const, label: '背景' }
])

// 标签页样式选项
const tabsStyleOptions = computed(() => [
  { value: 'chrome' as const, label: '谷歌' },
  { value: 'plain' as const, label: '朴素' },
  { value: 'card' as const, label: '卡片' },
  { value: 'smart' as const, label: '轻快' }
])

// 小部件位置选项
const widgetsPositionOptions = computed(() => [
  { value: 'auto' as const, label: '自动' },
  { value: 'header' as const, label: '顶栏' },
  { value: 'sidebar' as const, label: '侧边栏' }
])

// 双列菜单样式选项
const doubleSidebarStyleOptions = computed(() => [
  { value: 'icon-left' as const, label: '图标 + 文字', description: '第一列图标，第二列文字' },
  { value: 'text-left' as const, label: '文字 + 图标', description: '第一列文字，第二列图标' }
])

// 更新布局模式
function updateLayoutMode(mode: 'sidebar' | 'topbar' | 'mixed' | 'horizontal' | 'double-sidebar') {
  preferencesStore.updateLayout('layoutMode', mode)
}

// 更新导航样式
function updateNavStyle(style: 'icon-text' | 'icon-only') {
  preferencesStore.updateLayout('navStyle', style)
}

// 更新顶栏模式
function updateHeaderMode(mode: 'static' | 'fixed' | 'auto-hide') {
  preferencesStore.updateLayout('headerMode', mode)
}

// 更新内容宽度模式
function updateContentWidth(mode: 'fluid' | 'fixed') {
  preferencesStore.updateLayout('contentWidthMode', mode)
}

// 更新面包屑样式
function updateBreadcrumbStyle(style: 'normal' | 'background') {
  preferencesStore.updateLayout('breadcrumbStyle', style)
}

// 更新标签页样式
function updateTabsStyle(style: 'chrome' | 'plain' | 'card' | 'smart') {
  preferencesStore.updateLayout('tabsStyle', style)
}

// 更新小部件位置
function updateWidgetsPosition(position: 'auto' | 'header' | 'sidebar') {
  preferencesStore.updateLayout('widgetsPosition', position)
}

// 更新双列菜单样式
function updateDoubleSidebarStyle(style: 'icon-left' | 'text-left') {
  preferencesStore.updateLayout('doubleSidebarStyle', style)
}

// 切换布尔选项
function toggleOption(key: keyof typeof layout.value) {
  preferencesStore.updateLayout(key, !layout.value[key])
}

// 更新侧边栏宽度
function updateSidebarWidth(value: number) {
  preferencesStore.updateLayout('sidebarWidth', value)
}

// 更新侧边栏折叠宽度
function updateSidebarCollapsedWidth(value: number) {
  preferencesStore.updateLayout('sidebarCollapsedWidth', value)
}

// 更新顶栏高度
function updateHeaderHeight(value: number) {
  preferencesStore.updateLayout('headerHeight', value)
}

// 更新内容定宽
function updateContentFixedWidth(value: number) {
  preferencesStore.updateLayout('contentFixedWidth', value)
}
</script>

<template>
  <div class="space-y-6">
    <!-- 标题 -->
    <div class="flex items-center gap-3 pb-2">
      <div class="p-2.5 bg-gradient-to-br from-emerald-500 to-teal-600 rounded-xl shadow-lg shadow-emerald-500/25">
        <Layout :size="20" class="text-white" />
      </div>
      <div>
        <h3 class="text-lg font-semibold text-slate-900 dark:text-slate-100">
          {{ t('preferences.layout.title') }}
        </h3>
        <p class="text-sm text-slate-500 dark:text-slate-400">
          {{ t('preferences.layout.description') }}
        </p>
      </div>
    </div>

    <!-- 布局模式选择 -->
    <section>
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">
        布局模式
      </h4>
      <div class="grid grid-cols-3 lg:grid-cols-5 gap-2">
        <button
          v-for="option in layoutModeOptions"
          :key="option.value"
          class="group relative p-2.5 border-2 rounded-xl transition-all duration-200 cursor-pointer"
          :class="layout.layoutMode === option.value
            ? 'border-primary-500 bg-gradient-to-br from-primary-50 to-primary-100/50 dark:from-primary-900/30 dark:to-primary-800/20 shadow-lg shadow-primary-500/10'
            : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600 hover:bg-white dark:hover:bg-slate-700/30'"
          @click="updateLayoutMode(option.value)"
        >
          <!-- 布局预览图标 -->
          <div class="flex justify-center mb-2 h-10">
            <div class="w-14 h-full border border-slate-300 dark:border-slate-600 rounded transition-colors p-1" :class="layout.layoutMode === option.value ? 'border-primary-500' : ''">
              <!-- Sidebar Icon -->
              <div v-if="option.icon === 'sidebar'" class="h-full border-r border-slate-300 dark:border-slate-600 flex items-center justify-center" :class="layout.layoutMode === option.value ? 'border-primary-500' : ''">
                <div class="w-1 h-1.5 bg-slate-400 dark:bg-slate-500 rounded-sm" />
              </div>
              <!-- Double Sidebar Icon -->
              <div v-else-if="option.icon === 'double-sidebar'" class="h-full flex">
                <div class="w-1.5 border-r border-slate-300 dark:border-slate-600" :class="layout.layoutMode === option.value ? 'border-primary-500' : ''" />
                <div class="flex-1 border-r border-slate-300 dark:border-slate-600 flex flex-col justify-center gap-0.5 p-0.5" :class="layout.layoutMode === option.value ? 'border-primary-500' : ''">
                  <div class="w-full h-0.5 bg-slate-400 dark:bg-slate-500 rounded-sm" />
                  <div class="w-3/4 h-0.5 bg-slate-400 dark:bg-slate-500 rounded-sm" />
                </div>
              </div>
              <!-- Topbar Icon -->
              <div v-else-if="option.icon === 'topbar'" class="h-2 w-full border-b border-slate-300 dark:border-slate-600" :class="layout.layoutMode === option.value ? 'border-primary-500' : ''" />
              <!-- Mixed Icon -->
              <template v-else-if="option.icon === 'mixed'">
                <div class="h-1.5 w-full border-b border-slate-300 dark:border-slate-600" :class="layout.layoutMode === option.value ? 'border-primary-500' : ''" />
                <div class="h-full border-r border-slate-300 dark:border-slate-600 flex items-center justify-center" :class="layout.layoutMode === option.value ? 'border-primary-500' : ''">
                  <div class="w-0.5 h-1 bg-slate-400 dark:bg-slate-500 rounded-sm" />
                </div>
              </template>
              <!-- Horizontal Icon -->
              <div v-else class="flex flex-col gap-0.5 h-full justify-center">
                <div class="h-1 w-full bg-slate-400 dark:bg-slate-500 rounded-sm" />
                <div class="h-1 w-2/3 bg-slate-300 dark:bg-slate-600 rounded-sm" />
              </div>
            </div>
          </div>
          <span
            class="text-xs font-medium block text-center truncate"
            :class="layout.layoutMode === option.value
              ? 'text-primary-700 dark:text-primary-300'
              : 'text-slate-600 dark:text-slate-400'"
          >
            {{ option.label }}
          </span>
          <!-- 选中指示器 -->
          <div
            v-if="layout.layoutMode === option.value"
            class="absolute top-1.5 right-1.5 w-4 h-4 bg-primary-500 rounded-full flex items-center justify-center text-white"
          >
            <svg class="w-2.5 h-2.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
            </svg>
          </div>
        </button>
      </div>
    </section>

    <!-- 双列菜单设置 -->
    <section v-if="layout.layoutMode === 'double-sidebar'" class="border-2 border-slate-200 dark:border-slate-700 rounded-2xl overflow-hidden">
      <div class="p-4 space-y-4 bg-white dark:bg-slate-900/30">
        <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300">
          双列菜单样式
        </h4>
        <div class="grid grid-cols-2 gap-3">
          <button
            v-for="option in doubleSidebarStyleOptions"
            :key="option.value"
            class="p-4 border-2 bg-white dark:bg-slate-700/50 rounded-xl transition-all"
            :class="layout.doubleSidebarStyle === option.value ? 'border-primary-500 shadow-md' : 'border-slate-200 dark:border-slate-700'"
            @click="updateDoubleSidebarStyle(option.value)"
          >
            <div class="flex flex-col items-center gap-2">
              <!-- 预览图标 -->
              <div class="w-full h-10 flex border-2 border-slate-300 dark:border-slate-600 rounded" :class="layout.doubleSidebarStyle === option.value ? 'border-primary-500' : ''">
                <div v-if="option.value === 'icon-left'" class="flex w-full">
                  <div class="w-6 border-r-2 border-slate-300 dark:border-slate-600 flex flex-col items-center justify-center gap-0.5" :class="layout.doubleSidebarStyle === option.value ? 'border-primary-500' : ''">
                    <div class="w-2 h-2 bg-slate-400 dark:bg-slate-500 rounded-sm" />
                    <div class="w-2 h-2 bg-slate-400 dark:bg-slate-500 rounded-sm" />
                  </div>
                  <div class="flex-1 flex flex-col justify-center gap-0.5 p-1">
                    <div class="h-1.5 w-full bg-slate-300 dark:bg-slate-600 rounded-sm" />
                    <div class="h-1.5 w-3/4 bg-slate-300 dark:bg-slate-600 rounded-sm" />
                  </div>
                </div>
                <div v-else class="flex w-full">
                  <div class="flex-1 flex flex-col justify-center gap-0.5 p-1">
                    <div class="h-1.5 w-full bg-slate-300 dark:bg-slate-600 rounded-sm" />
                    <div class="h-1.5 w-3/4 bg-slate-300 dark:bg-slate-600 rounded-sm" />
                  </div>
                  <div class="w-6 border-l-2 border-slate-300 dark:border-slate-600 flex flex-col items-center justify-center gap-0.5" :class="layout.doubleSidebarStyle === option.value ? 'border-primary-500' : ''">
                    <div class="w-2 h-2 bg-slate-400 dark:bg-slate-500 rounded-sm" />
                    <div class="w-2 h-2 bg-slate-400 dark:bg-slate-500 rounded-sm" />
                  </div>
                </div>
              </div>
              <span class="text-sm font-medium" :class="layout.doubleSidebarStyle === option.value ? 'text-primary-700 dark:text-primary-300' : 'text-slate-600 dark:text-slate-400'">
                {{ option.label }}
              </span>
              <span class="text-xs text-slate-500 dark:text-slate-400">
                {{ option.description }}
              </span>
            </div>
          </button>
        </div>
      </div>
    </section>

    <!-- 侧边栏设置 -->
    <section class="border-2 border-slate-200 dark:border-slate-700 rounded-2xl overflow-hidden">
      <button
        class="w-full flex items-center justify-between p-4 bg-slate-50 dark:bg-slate-800/50 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
        @click="sidebarExpanded = !sidebarExpanded"
      >
        <div class="flex items-center gap-3">
          <div class="p-2 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-lg">
            <svg class="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" />
            </svg>
          </div>
          <span class="text-sm font-semibold text-slate-700 dark:text-slate-300">侧边栏设置</span>
        </div>
        <component :is="sidebarExpanded ? ChevronUp : ChevronDown" :size="18" class="text-slate-500" />
      </button>

      <Transition
        enter-active-class="transition-all duration-200 ease-out"
        enter-from-class="opacity-0 -translate-y-2"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition-all duration-150 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 -translate-y-2"
      >
        <div v-if="sidebarExpanded" class="p-4 space-y-4 bg-white dark:bg-slate-900/30">
          <!-- 宽度滑块 -->
          <div>
            <label class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-2 block">
              侧边栏宽度: {{ layout.sidebarWidth }}px
            </label>
            <input
              type="range"
              :value="layout.sidebarWidth"
              min="180"
              max="400"
              step="4"
              class="w-full h-2 bg-slate-200 dark:bg-slate-700 rounded-lg appearance-none cursor-pointer accent-primary-500"
              @input="(e) => updateSidebarWidth(Number((e.target as HTMLInputElement).value))"
            />
          </div>

          <!-- 折叠宽度滑块 -->
          <div>
            <label class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-2 block">
              折叠宽度: {{ layout.sidebarCollapsedWidth }}px
            </label>
            <input
              type="range"
              :value="layout.sidebarCollapsedWidth"
              min="48"
              max="80"
              step="4"
              class="w-full h-2 bg-slate-200 dark:bg-slate-700 rounded-lg appearance-none cursor-pointer accent-primary-500"
              @input="(e) => updateSidebarCollapsedWidth(Number((e.target as HTMLInputElement).value))"
            />
          </div>

          <!-- 开关选项 -->
          <div class="space-y-3">
            <div class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl">
              <span class="text-sm text-slate-700 dark:text-slate-300">可折叠</span>
              <button
                class="relative w-12 h-6 rounded-full transition-all duration-300 cursor-pointer"
                :class="layout.sidebarCollapsible ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
                @click="toggleOption('sidebarCollapsible')"
              >
                <span
                  class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
                  :class="layout.sidebarCollapsible ? 'translate-x-6' : ''"
                />
              </button>
            </div>
            <div class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl">
              <span class="text-sm text-slate-700 dark:text-slate-300">默认折叠</span>
              <button
                class="relative w-12 h-6 rounded-full transition-all duration-300 cursor-pointer"
                :class="layout.sidebarCollapsed ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
                @click="toggleOption('sidebarCollapsed')"
              >
                <span
                  class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
                  :class="layout.sidebarCollapsed ? 'translate-x-6' : ''"
                />
              </button>
            </div>
            <div class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl">
              <span class="text-sm text-slate-700 dark:text-slate-300">手风琴模式</span>
              <button
                class="relative w-12 h-6 rounded-full transition-all duration-300 cursor-pointer"
                :class="layout.navAccordion ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
                @click="toggleOption('navAccordion')"
              >
                <span
                  class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
                  :class="layout.navAccordion ? 'translate-x-6' : ''"
                />
              </button>
            </div>
          </div>

          <!-- 导航样式 -->
          <div>
            <label class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-2 block">导航样式</label>
            <div class="grid grid-cols-2 gap-2">
              <button
                v-for="option in navStyleOptions"
                :key="option.value"
                class="p-2 border-2 bg-white dark:bg-slate-700/50 rounded-lg text-sm transition-all"
                :class="layout.navStyle === option.value ? 'border-primary-500 text-primary-700 dark:text-primary-300' : 'border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400'"
                @click="updateNavStyle(option.value)"
              >
                {{ option.label }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </section>

    <!-- 顶栏设置 -->
    <section class="border-2 border-slate-200 dark:border-slate-700 rounded-2xl overflow-hidden">
      <button
        class="w-full flex items-center justify-between p-4 bg-slate-50 dark:bg-slate-800/50 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
        @click="headerExpanded = !headerExpanded"
      >
        <div class="flex items-center gap-3">
          <div class="p-2 bg-gradient-to-br from-purple-500 to-pink-600 rounded-lg">
            <svg class="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
            </svg>
          </div>
          <span class="text-sm font-semibold text-slate-700 dark:text-slate-300">顶栏设置</span>
        </div>
        <component :is="headerExpanded ? ChevronUp : ChevronDown" :size="18" class="text-slate-500" />
      </button>

      <Transition
        enter-active-class="transition-all duration-200 ease-out"
        enter-from-class="opacity-0 -translate-y-2"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition-all duration-150 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 -translate-y-2"
      >
        <div v-if="headerExpanded" class="p-4 space-y-4 bg-white dark:bg-slate-900/30">
          <!-- 顶栏高度 -->
          <div>
            <label class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-2 block">
              顶栏高度: {{ layout.headerHeight }}px
            </label>
            <input
              type="range"
              :value="layout.headerHeight"
              min="48"
              max="80"
              step="4"
              class="w-full h-2 bg-slate-200 dark:bg-slate-700 rounded-lg appearance-none cursor-pointer accent-primary-500"
              @input="(e) => updateHeaderHeight(Number((e.target as HTMLInputElement).value))"
            />
          </div>

          <!-- 顶栏模式 -->
          <div>
            <label class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-2 block">顶栏模式</label>
            <div class="grid grid-cols-3 gap-2">
              <button
                v-for="option in headerModeOptions"
                :key="option.value"
                class="p-2 border-2 bg-white dark:bg-slate-700/50 rounded-lg text-sm transition-all"
                :class="layout.headerMode === option.value ? 'border-primary-500 text-primary-700 dark:text-primary-300' : 'border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400'"
                @click="updateHeaderMode(option.value)"
              >
                {{ option.label }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </section>

    <!-- 内容设置 -->
    <section>
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">
        内容区域
      </h4>
      <div class="grid grid-cols-2 gap-3">
        <button
          v-for="option in contentWidthOptions"
          :key="option.value"
          class="p-4 border-2 bg-white dark:bg-slate-700/50 rounded-xl transition-all"
          :class="layout.contentWidthMode === option.value ? 'border-primary-500 shadow-md' : 'border-slate-200 dark:border-slate-700'"
          @click="updateContentWidth(option.value)"
        >
          <div class="flex flex-col items-center gap-2">
            <div class="w-full h-8 border-2 border-dashed rounded" :class="layout.contentWidthMode === option.value ? 'border-primary-500' : 'border-slate-300 dark:border-slate-600'" :style="option.value === 'fixed' ? 'max-width: 80%; margin: 0 auto;' : ''" />
            <span class="text-sm font-medium" :class="layout.contentWidthMode === option.value ? 'text-primary-700 dark:text-primary-300' : 'text-slate-600 dark:text-slate-400'">
              {{ option.label }}
            </span>
          </div>
        </button>
      </div>

      <!-- 定宽值输入 -->
      <div v-if="layout.contentWidthMode === 'fixed'" class="mt-3">
        <label class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-2 block">
          内容宽度: {{ layout.contentFixedWidth || 1200 }}px
        </label>
        <input
          type="number"
          :value="layout.contentFixedWidth || 1200"
          min="800"
          max="1920"
          step="50"
          class="w-full px-4 py-2.5 bg-white dark:bg-slate-700/50 border-2 border-slate-200 dark:border-slate-700 rounded-xl text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
          @input="(e) => updateContentFixedWidth(Number((e.target as HTMLInputElement).value))"
        />
      </div>
    </section>

    <!-- 界面元素显示 -->
    <section class="space-y-3">
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300">界面元素显示</h4>

      <!-- 面包屑 -->
      <div class="flex items-center justify-between p-4 bg-white dark:bg-slate-700/30 rounded-2xl border border-slate-200 dark:border-slate-700">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-blue-100 dark:bg-blue-900/30 rounded-xl">
            <svg class="w-5 h-5 text-blue-600 dark:text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
            </svg>
          </div>
          <div>
            <h5 class="text-sm font-semibold text-slate-800 dark:text-slate-200">显示面包屑</h5>
            <p class="text-xs text-slate-500 dark:text-slate-400">页面导航路径</p>
          </div>
        </div>
        <button
          class="relative w-14 h-7 rounded-full transition-all duration-300 cursor-pointer"
          :class="layout.showBreadcrumbs ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
          @click="toggleOption('showBreadcrumbs')"
        >
          <span
            class="absolute top-1 left-1 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
            :class="layout.showBreadcrumbs ? 'translate-x-7' : ''"
          />
        </button>
      </div>

      <!-- 标签页 -->
      <section class="border-2 border-slate-200 dark:border-slate-700 rounded-2xl overflow-hidden">
        <div class="flex items-center justify-between p-4 bg-slate-50 dark:bg-slate-800/50">
          <div class="flex items-center gap-3">
            <div class="p-2.5 bg-purple-100 dark:bg-purple-900/30 rounded-xl">
              <svg class="w-5 h-5 text-purple-600 dark:text-purple-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
            </div>
            <div>
              <h5 class="text-sm font-semibold text-slate-800 dark:text-slate-200">显示标签页</h5>
              <p class="text-xs text-slate-500 dark:text-slate-400">多页面标签导航</p>
            </div>
          </div>
          <button
            class="relative w-14 h-7 rounded-full transition-all duration-300 cursor-pointer"
            :class="layout.showTabs ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
            @click="toggleOption('showTabs')"
          >
            <span
              class="absolute top-1 left-1 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
              :class="layout.showTabs ? 'translate-x-7' : ''"
            />
          </button>
        </div>

        <!-- 标签页详细设置 -->
        <Transition
          enter-active-class="transition-all duration-200 ease-out"
          enter-from-class="opacity-0 max-h-0"
          enter-to-class="opacity-100 max-h-96"
          leave-active-class="transition-all duration-150 ease-in"
          leave-from-class="opacity-100 max-h-96"
          leave-to-class="opacity-0 max-h-0"
        >
          <div v-if="layout.showTabs" class="p-4 space-y-4 bg-white dark:bg-slate-900/30 border-t border-slate-200 dark:border-slate-700">
            <!-- 标签页风格 -->
            <div>
              <label class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-2 block">标签页风格</label>
              <div class="grid grid-cols-4 gap-2">
                <button
                  v-for="option in tabsStyleOptions"
                  :key="option.value"
                  class="p-2 border-2 bg-white dark:bg-slate-700/50 rounded-lg text-xs transition-all"
                  :class="layout.tabsStyle === option.value ? 'border-primary-500 text-primary-700 dark:text-primary-300' : 'border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400'"
                  @click="updateTabsStyle(option.value)"
                >
                  {{ option.label }}
                </button>
              </div>
            </div>

            <!-- 标签页选项 -->
            <div class="space-y-3">
              <div class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl">
                <span class="text-sm text-slate-700 dark:text-slate-300">持久化标签页</span>
                <button
                  class="relative w-12 h-6 rounded-full transition-all duration-300 cursor-pointer"
                  :class="layout.tabsPersistent ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
                  @click="toggleOption('tabsPersistent')"
                >
                  <span
                    class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
                    :class="layout.tabsPersistent ? 'translate-x-6' : ''"
                  />
                </button>
              </div>
              <div class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl">
                <span class="text-sm text-slate-700 dark:text-slate-300">拖拽排序</span>
                <button
                  class="relative w-12 h-6 rounded-full transition-all duration-300 cursor-pointer"
                  :class="layout.tabsDraggable ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
                  @click="toggleOption('tabsDraggable')"
                >
                  <span
                    class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
                    :class="layout.tabsDraggable ? 'translate-x-6' : ''"
                  />
                </button>
              </div>
            </div>
          </div>
        </Transition>
      </section>

      <!-- 页脚 -->
      <div class="flex items-center justify-between p-4 bg-white dark:bg-slate-700/30 rounded-2xl border border-slate-200 dark:border-slate-700">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-pink-100 dark:bg-pink-900/30 rounded-xl">
            <svg class="w-5 h-5 text-pink-600 dark:text-pink-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <div>
            <h5 class="text-sm font-semibold text-slate-800 dark:text-slate-200">显示页脚</h5>
            <p class="text-xs text-slate-500 dark:text-slate-400">页面底部信息</p>
          </div>
        </div>
        <button
          class="relative w-14 h-7 rounded-full transition-all duration-300 cursor-pointer"
          :class="layout.showFooter ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
          @click="toggleOption('showFooter')"
        >
          <span
            class="absolute top-1 left-1 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
            :class="layout.showFooter ? 'translate-x-7' : ''"
          />
        </button>
      </div>
    </section>
  </div>
</template>
