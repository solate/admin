<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePreferencesStore } from '@/stores/modules/preferences'
import { useTheme } from '@/composables/useTheme'
import { THEME_COLORS } from '@/types/preferences'
import {
  Sun,
  Moon,
  Monitor,
  Palette
} from 'lucide-vue-next'

const { t } = useI18n()
const preferencesStore = usePreferencesStore()

// 使用 useTheme composable 获取主题相关功能
const {
  primaryColor,
  isDark,
  themeMode,
  borderRadius,
  colorBlindMode,
  highContrast,
  setThemeColor,
  setThemeMode,
  setBorderRadius,
  setColorBlindMode,
  toggleHighContrast
} = useTheme()

// 外观设置（保留以兼容现有代码）
const appearance = computed(() => preferencesStore.appearance)

// 主题模式选项
const themeModeOptions = computed(() => [
  {
    value: 'light' as const,
    label: t('preferences.appearance.themeMode.light'),
    icon: Sun,
    description: '浅色模式'
  },
  {
    value: 'dark' as const,
    label: t('preferences.appearance.themeMode.dark'),
    icon: Moon,
    description: '深色模式'
  },
  {
    value: 'auto' as const,
    label: t('preferences.appearance.themeMode.auto'),
    icon: Monitor,
    description: '跟随系统'
  }
])

// 圆角选项
const borderRadiusOptions = computed(() => [
  { value: 'none' as const, label: '无' },
  { value: 'small' as const, label: '小' },
  { value: 'medium' as const, label: '中' },
  { value: 'large' as const, label: '大' }
])

// 色盲模式选项
const colorBlindOptions = computed(() => [
  { value: 'none' as const, label: t('preferences.appearance.colorBlindMode.none') },
  { value: 'protanopia' as const, label: t('preferences.appearance.colorBlindMode.protanopia') },
  { value: 'deuteranopia' as const, label: t('preferences.appearance.colorBlindMode.deuteranopia') },
  { value: 'tritanopia' as const, label: t('preferences.appearance.colorBlindMode.tritanopia') }
])

// 更新主题模式
function updateThemeMode(mode: 'light' | 'dark' | 'auto') {
  setThemeMode(mode)
}

// 更新主题色
function updateThemeColor(color: string) {
  setThemeColor(color)
}

// 更新圆角大小
function updateBorderRadius(radius: 'none' | 'small' | 'medium' | 'large') {
  setBorderRadius(radius)
}

// 切换深色侧边栏（保留原有功能）
function toggleDarkSidebar() {
  preferencesStore.updateAppearance('darkSidebar', !appearance.value.darkSidebar)
}

// 更新色盲模式
function updateColorBlindMode(mode: 'none' | 'protanopia' | 'deuteranopia' | 'tritanopia') {
  setColorBlindMode(mode)
}
</script>

<template>
  <div class="space-y-6">
    <!-- 标题 -->
    <div class="flex items-center gap-3 pb-2">
      <div class="p-2.5 bg-gradient-to-br from-violet-500 to-purple-600 rounded-xl shadow-lg shadow-violet-500/25">
        <Palette :size="20" class="text-white" />
      </div>
      <div>
        <h3 class="text-lg font-semibold text-slate-900 dark:text-slate-100">
          {{ t('preferences.appearance.title') }}
        </h3>
        <p class="text-sm text-slate-500 dark:text-slate-400">
          {{ t('preferences.appearance.description') }}
        </p>
      </div>
    </div>

    <!-- 主题模式 - 卡片式选择 -->
    <section>
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">
        {{ t('preferences.appearance.themeMode.label') }}
      </h4>
      <div class="grid grid-cols-3 gap-3">
        <button
          v-for="option in themeModeOptions"
          :key="option.value"
          class="group relative flex flex-col items-center gap-3 p-4 border-2 rounded-2xl transition-all duration-200 cursor-pointer"
          :class="appearance.themeMode === option.value
            ? 'border-primary-500 bg-gradient-to-br from-primary-50 to-primary-100/50 dark:from-primary-900/30 dark:to-primary-800/20 shadow-lg shadow-primary-500/10'
            : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600 hover:bg-white dark:hover:bg-slate-700/30'"
          @click="updateThemeMode(option.value)"
        >
          <div
            class="p-3 rounded-xl transition-all duration-200"
            :class="appearance.themeMode === option.value
              ? 'bg-gradient-to-br from-primary-500 to-primary-600 text-white shadow-lg'
              : 'bg-slate-100 dark:bg-slate-700/50 text-slate-500 dark:text-slate-400 group-hover:bg-slate-200 dark:group-hover:bg-slate-600'"
          >
            <component :is="option.icon" :size="22" />
          </div>
          <span
            class="text-sm font-medium"
            :class="appearance.themeMode === option.value
              ? 'text-primary-700 dark:text-primary-300'
              : 'text-slate-600 dark:text-slate-400'"
          >
            {{ option.label }}
          </span>
          <!-- 选中指示器 -->
          <div
            v-if="appearance.themeMode === option.value"
            class="absolute top-3 right-3 w-5 h-5 bg-primary-500 rounded-full flex items-center justify-center text-white"
          >
            <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
            </svg>
          </div>
        </button>
      </div>
    </section>

    <!-- 主题色 - 精美色板 -->
    <section>
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">
        {{ t('preferences.appearance.themeColor.label') }}
      </h4>
      <div class="flex flex-wrap gap-3">
        <button
          v-for="color in THEME_COLORS"
          :key="color.name"
          class="group relative w-12 h-12 rounded-2xl transition-all duration-200 cursor-pointer shadow-md hover:shadow-lg hover:scale-110"
          :class="appearance.primaryColor === color.value
            ? 'ring-2 ring-offset-2 ring-primary-500 dark:ring-offset-slate-800 scale-110'
            : 'ring-offset-2 dark:ring-offset-slate-800'"
          :style="{ backgroundColor: color.value }"
          :title="color.label"
          @click="updateThemeColor(color.value)"
        >
          <!-- 选中指示器 -->
          <span
            v-if="appearance.primaryColor === color.value"
            class="flex items-center justify-center h-full text-white"
          >
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
            </svg>
          </span>
        </button>
      </div>
    </section>

    <!-- 圆角大小 - 精美卡片 -->
    <section>
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">
        {{ t('preferences.appearance.borderRadius.label') }}
      </h4>
      <div class="grid grid-cols-4 gap-3">
        <button
          v-for="option in borderRadiusOptions"
          :key="option.value"
          class="p-3 border-2 bg-white dark:bg-slate-700/50 transition-all duration-200 cursor-pointer hover:shadow-md"
          :class="[
            appearance.borderRadius === option.value
              ? 'border-primary-500 shadow-md shadow-primary-500/10'
              : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600',
            option.value === 'none' ? 'rounded-none' : option.value === 'small' ? 'rounded-lg' : option.value === 'medium' ? 'rounded-xl' : 'rounded-2xl'
          ]"
          @click="updateBorderRadius(option.value)"
        >
          <span class="text-sm font-medium" :class="appearance.borderRadius === option.value ? 'text-primary-700 dark:text-primary-300' : 'text-slate-600 dark:text-slate-400'">
            {{ option.label }}
          </span>
        </button>
      </div>
    </section>

    <!-- 切换开关组 -->
    <section class="space-y-3">
      <!-- 深色侧边栏 -->
      <div class="flex items-center justify-between p-4 bg-white dark:bg-slate-700/30 rounded-2xl border border-slate-200 dark:border-slate-700 transition-all duration-200 hover:shadow-md">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-indigo-100 dark:bg-indigo-900/30 rounded-xl">
            <svg class="w-5 h-5 text-indigo-600 dark:text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
            </svg>
          </div>
          <div>
            <h5 class="text-sm font-semibold text-slate-800 dark:text-slate-200">
              {{ t('preferences.appearance.darkSidebar.label') }}
            </h5>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              {{ t('preferences.appearance.darkSidebar.description') }}
            </p>
          </div>
        </div>
        <button
          class="relative w-14 h-7 rounded-full transition-all duration-300 cursor-pointer"
          :class="appearance.darkSidebar ? 'bg-primary-500 shadow-lg shadow-primary-500/30' : 'bg-slate-300 dark:bg-slate-600'"
          @click="toggleDarkSidebar"
        >
          <span
            class="absolute top-1 left-1 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
            :class="appearance.darkSidebar ? 'translate-x-7' : ''"
          />
        </button>
      </div>

      <!-- 高对比度 -->
      <div class="flex items-center justify-between p-4 bg-white dark:bg-slate-700/30 rounded-2xl border border-slate-200 dark:border-slate-700 transition-all duration-200 hover:shadow-md">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-amber-100 dark:bg-amber-900/30 rounded-xl">
            <svg class="w-5 h-5 text-amber-600 dark:text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
            </svg>
          </div>
          <div>
            <h5 class="text-sm font-semibold text-slate-800 dark:text-slate-200">
              {{ t('preferences.appearance.highContrast.label') }}
            </h5>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              {{ t('preferences.appearance.highContrast.description') }}
            </p>
          </div>
        </div>
        <button
          class="relative w-14 h-7 rounded-full transition-all duration-300 cursor-pointer"
          :class="appearance.highContrast ? 'bg-primary-500 shadow-lg shadow-primary-500/30' : 'bg-slate-300 dark:bg-slate-600'"
          @click="toggleHighContrast"
        >
          <span
            class="absolute top-1 left-1 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
            :class="appearance.highContrast ? 'translate-x-7' : ''"
          />
        </button>
      </div>
    </section>

    <!-- 色盲模式 - 下拉选择 -->
    <section>
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">
        {{ t('preferences.appearance.colorBlindMode.label') }}
      </h4>
      <div class="relative">
        <select
          :value="appearance.colorBlindMode"
          class="w-full px-4 py-3 bg-white dark:bg-slate-700/50 border-2 border-slate-200 dark:border-slate-700 rounded-xl text-sm text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none cursor-pointer appearance-none transition-all duration-200 hover:border-slate-300 dark:hover:border-slate-600"
          @change="updateColorBlindMode($event.target.value as any)"
        >
          <option
            v-for="option in colorBlindOptions"
            :key="option.value"
            :value="option.value"
          >
            {{ option.label }}
          </option>
        </select>
        <div class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none">
          <svg class="w-5 h-5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </div>
      </div>
      <div v-if="appearance.colorBlindMode !== 'none'" class="mt-3 p-3 bg-amber-50 dark:bg-amber-900/20 rounded-xl border border-amber-200 dark:border-amber-800">
        <p class="text-sm text-amber-800 dark:text-amber-300 flex items-center gap-2">
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          色盲模式已启用，页面色彩已调整
        </p>
      </div>
    </section>
  </div>
</template>
