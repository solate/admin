<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePreferencesStore } from '@/stores/modules/preferences'
import { useTheme } from '@/composables/useTheme'
import { THEME_COLORS } from '@/types/preferences'
import {
  Sun,
  Moon,
  Monitor,
  Palette,
  ChevronDown,
  ChevronUp
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

// 折叠面板状态
const advancedExpanded = ref(false)

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
  { value: 'large' as const, label: '大' },
  { value: 'custom' as const, label: '自定义' }
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
function updateBorderRadius(radius: 'none' | 'small' | 'medium' | 'large' | 'custom') {
  setBorderRadius(radius)
}

// 切换深色侧边栏
function toggleDarkSidebar() {
  preferencesStore.updateAppearance('darkSidebar', !appearance.value.darkSidebar)
}

// 切换深色顶栏
function toggleDarkHeader() {
  preferencesStore.updateAppearance('darkHeader', !appearance.value.darkHeader)
}

// 切换灰色模式
function toggleGrayMode() {
  preferencesStore.updateAppearance('grayMode', !appearance.value.grayMode)
}

// 更新色盲模式
function updateColorBlindMode(mode: 'none' | 'protanopia' | 'deuteranopia' | 'tritanopia') {
  setColorBlindMode(mode)
}

// 更新自定义圆角值
function updateCustomBorderRadius(value: string) {
  preferencesStore.updateAppearance('customBorderRadius', value)
}

// 切换高级选项展开状态
function toggleAdvanced() {
  advancedExpanded.value = !advancedExpanded.value
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

    <!-- 基础主题设置 -->
    <section>
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">
        主题模式
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

    <!-- 主题色 -->
    <section>
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">
        主题色
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

    <!-- 圆角大小 -->
    <section>
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">
        圆角大小
      </h4>
      <div class="grid grid-cols-5 gap-3">
        <button
          v-for="option in borderRadiusOptions"
          :key="option.value"
          class="p-3 border-2 bg-white dark:bg-slate-700/50 transition-all duration-200 cursor-pointer hover:shadow-md"
          :class="[
            appearance.borderRadius === option.value
              ? 'border-primary-500 shadow-md shadow-primary-500/10'
              : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600',
            option.value === 'none' ? 'rounded-none' : option.value === 'small' ? 'rounded-lg' : option.value === 'medium' ? 'rounded-xl' : option.value === 'large' ? 'rounded-2xl' : 'rounded-xl'
          ]"
          @click="updateBorderRadius(option.value)"
        >
          <span class="text-sm font-medium" :class="appearance.borderRadius === option.value ? 'text-primary-700 dark:text-primary-300' : 'text-slate-600 dark:text-slate-400'">
            {{ option.label }}
          </span>
        </button>
      </div>

      <!-- 自定义圆角输入 -->
      <div v-if="appearance.borderRadius === 'custom'" class="mt-3">
        <input
          type="text"
          :value="appearance.customBorderRadius"
          placeholder="例如: 0.25rem 或 4px"
          class="w-full px-4 py-2.5 bg-white dark:bg-slate-700/50 border-2 border-slate-200 dark:border-slate-700 rounded-xl text-sm text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition-all duration-200"
          @input="(e) => updateCustomBorderRadius((e.target as HTMLInputElement).value)"
        />
      </div>
    </section>

    <!-- 高级选项折叠面板 -->
    <section class="border-2 border-slate-200 dark:border-slate-700 rounded-2xl overflow-hidden">
      <button
        class="w-full flex items-center justify-between p-4 bg-slate-50 dark:bg-slate-800/50 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
        @click="toggleAdvanced"
      >
        <div class="flex items-center gap-3">
          <div class="p-2 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-lg">
            <svg class="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4" />
            </svg>
          </div>
          <span class="text-sm font-semibold text-slate-700 dark:text-slate-300">高级选项</span>
        </div>
        <component :is="advancedExpanded ? ChevronUp : ChevronDown" :size="18" class="text-slate-500" />
      </button>

      <Transition
        enter-active-class="transition-all duration-200 ease-out"
        enter-from-class="opacity-0 -translate-y-2"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition-all duration-150 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 -translate-y-2"
      >
        <div v-if="advancedExpanded" class="p-4 space-y-3 bg-white dark:bg-slate-900/30">
          <!-- 深色侧边栏 -->
          <div class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-indigo-100 dark:bg-indigo-900/30 rounded-lg">
                <svg class="w-4 h-4 text-indigo-600 dark:text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
                </svg>
              </div>
              <div>
                <h5 class="text-sm font-medium text-slate-800 dark:text-slate-200">深色侧边栏</h5>
                <p class="text-xs text-slate-500 dark:text-slate-400">浅色模式下使用深色侧边栏</p>
              </div>
            </div>
            <button
              class="relative w-12 h-6 rounded-full transition-all duration-300 cursor-pointer"
              :class="appearance.darkSidebar ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
              @click="toggleDarkSidebar"
            >
              <span
                class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
                :class="appearance.darkSidebar ? 'translate-x-6' : ''"
              />
            </button>
          </div>

          <!-- 深色顶栏 -->
          <div class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-cyan-100 dark:bg-cyan-900/30 rounded-lg">
                <svg class="w-4 h-4 text-cyan-600 dark:text-cyan-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
                </svg>
              </div>
              <div>
                <h5 class="text-sm font-medium text-slate-800 dark:text-slate-200">深色顶栏</h5>
                <p class="text-xs text-slate-500 dark:text-slate-400">浅色模式下使用深色顶栏</p>
              </div>
            </div>
            <button
              class="relative w-12 h-6 rounded-full transition-all duration-300 cursor-pointer"
              :class="appearance.darkHeader ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
              @click="toggleDarkHeader"
            >
              <span
                class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
                :class="appearance.darkHeader ? 'translate-x-6' : ''"
              />
            </button>
          </div>

          <!-- 灰色模式 -->
          <div class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-slate-200 dark:bg-slate-700 rounded-lg">
                <svg class="w-4 h-4 text-slate-600 dark:text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
                </svg>
              </div>
              <div>
                <h5 class="text-sm font-medium text-slate-800 dark:text-slate-200">灰色模式</h5>
                <p class="text-xs text-slate-500 dark:text-slate-400">将界面转换为灰色调</p>
              </div>
            </div>
            <button
              class="relative w-12 h-6 rounded-full transition-all duration-300 cursor-pointer"
              :class="appearance.grayMode ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
              @click="toggleGrayMode"
            >
              <span
                class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
                :class="appearance.grayMode ? 'translate-x-6' : ''"
              />
            </button>
          </div>

          <!-- 色盲模式 -->
          <div class="p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl">
            <div class="flex items-center gap-2 mb-2">
              <div class="p-2 bg-amber-100 dark:bg-amber-900/30 rounded-lg">
                <svg class="w-4 h-4 text-amber-600 dark:text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                </svg>
              </div>
              <h5 class="text-sm font-medium text-slate-800 dark:text-slate-200">色盲模式</h5>
            </div>
            <select
              :value="appearance.colorBlindMode"
              class="w-full px-3 py-2 bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 rounded-lg text-sm text-slate-900 dark:text-slate-100 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none cursor-pointer"
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
            <div v-if="appearance.colorBlindMode !== 'none'" class="mt-2 p-2 bg-amber-50 dark:bg-amber-900/20 rounded-lg border border-amber-200 dark:border-amber-800">
              <p class="text-xs text-amber-800 dark:text-amber-300">
                色盲模式已启用，页面色彩已调整以辅助视觉
              </p>
            </div>
          </div>
        </div>
      </Transition>
    </section>
  </div>
</template>
