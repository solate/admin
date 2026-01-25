<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePreferencesStore } from '@/stores/modules/preferences'
import { Layout } from 'lucide-vue-next'

const { t } = useI18n()
const preferencesStore = usePreferencesStore()

// 布局设置
const layout = computed(() => preferencesStore.layout)

// 布局模式选项
const layoutModeOptions = computed(() => [
  {
    value: 'sidebar' as const,
    label: t('preferences.layout.layoutMode.sidebar'),
    description: '左侧导航栏布局'
  },
  {
    value: 'topbar' as const,
    label: t('preferences.layout.layoutMode.topbar'),
    description: '顶部导航栏布局'
  }
])

// 侧边栏宽度选项
const sidebarWidthOptions = computed(() => [
  { value: 'narrow' as const, label: '窄 (64px)' },
  { value: 'medium' as const, label: '中 (256px)' },
  { value: 'wide' as const, label: '宽 (320px)' }
])

// 导航样式选项
const navStyleOptions = computed(() => [
  { value: 'icon-text' as const, label: '图标+文字' },
  { value: 'icon-only' as const, label: '仅图标' }
])

// 更新布局模式
function updateLayoutMode(mode: 'sidebar' | 'topbar') {
  preferencesStore.updateLayout('layoutMode', mode)
}

// 更新侧边栏宽度
function updateSidebarWidth(width: 'narrow' | 'medium' | 'wide') {
  preferencesStore.updateLayout('sidebarWidth', width)
}

// 更新导航样式
function updateNavStyle(style: 'icon-text' | 'icon-only') {
  preferencesStore.updateLayout('navStyle', style)
}

// 切换显示选项
function toggleShowOption(key: keyof typeof layout.value) {
  const keyMap: Record<string, keyof typeof layout.value> = {
    showBreadcrumbs: 'showBreadcrumbs',
    showTabs: 'showTabs',
    showWidgets: 'showWidgets',
    showFooter: 'showFooter',
    showCopyright: 'showCopyright'
  }
  if (keyMap[key]) {
    const prop = keyMap[key]
    preferencesStore.updateLayout(prop, !layout.value[prop])
  }
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

    <!-- 布局模式 - 可视化选择 -->
    <section>
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">
        {{ t('preferences.layout.layoutMode.label') }}
      </h4>
      <div class="grid grid-cols-2 gap-4">
        <button
          v-for="option in layoutModeOptions"
          :key="option.value"
          class="group relative p-4 border-2 rounded-2xl transition-all duration-200 cursor-pointer"
          :class="layout.layoutMode === option.value
            ? 'border-primary-500 bg-gradient-to-br from-primary-50 to-primary-100/50 dark:from-primary-900/30 dark:to-primary-800/20 shadow-lg shadow-primary-500/10'
            : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600 hover:bg-white dark:hover:bg-slate-700/30'"
          @click="updateLayoutMode(option.value)"
        >
          <!-- 布局预览图标 -->
          <div class="flex justify-center mb-3">
            <div
              class="w-20 h-14 border-2 border-slate-300 dark:border-slate-600 rounded-xl transition-colors"
              :class="layout.layoutMode === option.value ? 'border-primary-500' : ''"
            >
              <div
                v-if="option.value === 'sidebar'"
                class="h-full w-5 border-r-2 border-slate-300 dark:border-slate-600 transition-colors"
                :class="layout.layoutMode === option.value ? 'border-primary-500' : ''"
              />
              <div
                v-else
                class="h-2.5 w-full border-b-2 border-slate-300 dark:border-slate-600 transition-colors"
                :class="layout.layoutMode === option.value ? 'border-primary-500' : ''"
              />
            </div>
          </div>
          <span
            class="text-sm font-medium block"
            :class="layout.layoutMode === option.value
              ? 'text-primary-700 dark:text-primary-300'
              : 'text-slate-600 dark:text-slate-400'"
          >
            {{ option.label }}
          </span>
          <!-- 选中指示器 -->
          <div
            v-if="layout.layoutMode === option.value"
            class="absolute top-3 right-3 w-5 h-5 bg-primary-500 rounded-full flex items-center justify-center text-white"
          >
            <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
            </svg>
          </div>
        </button>
      </div>
    </section>

    <!-- 侧边栏宽度 -->
    <section>
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">
        {{ t('preferences.layout.sidebarWidth.label') }}
      </h4>
      <div class="grid grid-cols-3 gap-3">
        <button
          v-for="option in sidebarWidthOptions"
          :key="option.value"
          class="p-3 border-2 bg-white dark:bg-slate-700/50 transition-all duration-200 cursor-pointer hover:shadow-md rounded-xl"
          :class="layout.sidebarWidth === option.value
            ? 'border-primary-500 shadow-md shadow-primary-500/10'
            : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600'"
          @click="updateSidebarWidth(option.value)"
        >
          <span class="text-sm font-medium" :class="layout.sidebarWidth === option.value ? 'text-primary-700 dark:text-primary-300' : 'text-slate-600 dark:text-slate-400'">
            {{ option.label }}
          </span>
        </button>
      </div>
    </section>

    <!-- 导航样式 -->
    <section>
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">
        {{ t('preferences.layout.navStyle.label') }}
      </h4>
      <div class="grid grid-cols-2 gap-3">
        <button
          v-for="option in navStyleOptions"
          :key="option.value"
          class="p-4 border-2 bg-white dark:bg-slate-700/50 transition-all duration-200 cursor-pointer hover:shadow-md rounded-xl"
          :class="layout.navStyle === option.value
            ? 'border-primary-500 shadow-md shadow-primary-500/10'
            : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600'"
          @click="updateNavStyle(option.value)"
        >
          <div class="flex items-center justify-center gap-2">
            <div class="w-8 h-8 rounded-lg bg-slate-200 dark:bg-slate-600 flex items-center justify-center">
              <svg class="w-4 h-4 text-slate-500 dark:text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </div>
            <div v-if="option.value === 'icon-text'" class="h-2 w-12 bg-slate-200 dark:bg-slate-600 rounded"></div>
          </div>
          <span class="text-sm font-medium mt-3 block" :class="layout.navStyle === option.value ? 'text-primary-700 dark:text-primary-300' : 'text-slate-600 dark:text-slate-400'">
            {{ option.label }}
          </span>
        </button>
      </div>
    </section>

    <!-- 显示选项 - 开关组 -->
    <section class="space-y-3">
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300">
        界面元素显示
      </h4>

      <!-- 面包屑 -->
      <div class="flex items-center justify-between p-4 bg-white dark:bg-slate-700/30 rounded-2xl border border-slate-200 dark:border-slate-700 transition-all duration-200 hover:shadow-md">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-blue-100 dark:bg-blue-900/30 rounded-xl">
            <svg class="w-5 h-5 text-blue-600 dark:text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
            </svg>
          </div>
          <div>
            <h5 class="text-sm font-semibold text-slate-800 dark:text-slate-200">
              {{ t('preferences.layout.showBreadcrumbs.label') }}
            </h5>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              {{ t('preferences.layout.showBreadcrumbs.description') }}
            </p>
          </div>
        </div>
        <button
          class="relative w-14 h-7 rounded-full transition-all duration-300 cursor-pointer"
          :class="layout.showBreadcrumbs ? 'bg-primary-500 shadow-lg shadow-primary-500/30' : 'bg-slate-300 dark:bg-slate-600'"
          @click="toggleShowOption('showBreadcrumbs')"
        >
          <span
            class="absolute top-1 left-1 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
            :class="layout.showBreadcrumbs ? 'translate-x-7' : ''"
          />
        </button>
      </div>

      <!-- 标签页 -->
      <div class="flex items-center justify-between p-4 bg-white dark:bg-slate-700/30 rounded-2xl border border-slate-200 dark:border-slate-700 transition-all duration-200 hover:shadow-md">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-purple-100 dark:bg-purple-900/30 rounded-xl">
            <svg class="w-5 h-5 text-purple-600 dark:text-purple-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <div>
            <h5 class="text-sm font-semibold text-slate-800 dark:text-slate-200">
              {{ t('preferences.layout.showTabs.label') }}
            </h5>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              {{ t('preferences.layout.showTabs.description') }}
            </p>
          </div>
        </div>
        <button
          class="relative w-14 h-7 rounded-full transition-all duration-300 cursor-pointer"
          :class="layout.showTabs ? 'bg-primary-500 shadow-lg shadow-primary-500/30' : 'bg-slate-300 dark:bg-slate-600'"
          @click="toggleShowOption('showTabs')"
        >
          <span
            class="absolute top-1 left-1 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
            :class="layout.showTabs ? 'translate-x-7' : ''"
          />
        </button>
      </div>

      <!-- 小部件 -->
      <div class="flex items-center justify-between p-4 bg-white dark:bg-slate-700/30 rounded-2xl border border-slate-200 dark:border-slate-700 transition-all duration-200 hover:shadow-md">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-cyan-100 dark:bg-cyan-900/30 rounded-xl">
            <svg class="w-5 h-5 text-cyan-600 dark:text-cyan-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
            </svg>
          </div>
          <div>
            <h5 class="text-sm font-semibold text-slate-800 dark:text-slate-200">
              {{ t('preferences.layout.showWidgets.label') }}
            </h5>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              {{ t('preferences.layout.showWidgets.description') }}
            </p>
          </div>
        </div>
        <button
          class="relative w-14 h-7 rounded-full transition-all duration-300 cursor-pointer"
          :class="layout.showWidgets ? 'bg-primary-500 shadow-lg shadow-primary-500/30' : 'bg-slate-300 dark:bg-slate-600'"
          @click="toggleShowOption('showWidgets')"
        >
          <span
            class="absolute top-1 left-1 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
            :class="layout.showWidgets ? 'translate-x-7' : ''"
          />
        </button>
      </div>

      <!-- 页脚 -->
      <div class="flex items-center justify-between p-4 bg-white dark:bg-slate-700/30 rounded-2xl border border-slate-200 dark:border-slate-700 transition-all duration-200 hover:shadow-md">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-pink-100 dark:bg-pink-900/30 rounded-xl">
            <svg class="w-5 h-5 text-pink-600 dark:text-pink-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <div>
            <h5 class="text-sm font-semibold text-slate-800 dark:text-slate-200">
              {{ t('preferences.layout.showFooter.label') }}
            </h5>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              {{ t('preferences.layout.showFooter.description') }}
            </p>
          </div>
        </div>
        <button
          class="relative w-14 h-7 rounded-full transition-all duration-300 cursor-pointer"
          :class="layout.showFooter ? 'bg-primary-500 shadow-lg shadow-primary-500/30' : 'bg-slate-300 dark:bg-slate-600'"
          @click="toggleShowOption('showFooter')"
        >
          <span
            class="absolute top-1 left-1 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
            :class="layout.showFooter ? 'translate-x-7' : ''"
          />
        </button>
      </div>

      <!-- 版权信息 -->
      <div class="flex items-center justify-between p-4 bg-white dark:bg-slate-700/30 rounded-2xl border border-slate-200 dark:border-slate-700 transition-all duration-200 hover:shadow-md">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-orange-100 dark:bg-orange-900/30 rounded-xl">
            <svg class="w-5 h-5 text-orange-600 dark:text-orange-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <div>
            <h5 class="text-sm font-semibold text-slate-800 dark:text-slate-200">
              {{ t('preferences.layout.showCopyright.label') }}
            </h5>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              {{ t('preferences.layout.showCopyright.description') }}
            </p>
          </div>
        </div>
        <button
          class="relative w-14 h-7 rounded-full transition-all duration-300 cursor-pointer"
          :class="layout.showCopyright ? 'bg-primary-500 shadow-lg shadow-primary-500/30' : 'bg-slate-300 dark:bg-slate-600'"
          @click="toggleShowOption('showCopyright')"
        >
          <span
            class="absolute top-1 left-1 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
            :class="layout.showCopyright ? 'translate-x-7' : ''"
          />
        </button>
      </div>
    </section>
  </div>
</template>
