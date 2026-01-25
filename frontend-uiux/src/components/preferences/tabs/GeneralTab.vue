<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePreferencesStore } from '@/stores/modules/preferences'
import { Settings2, Globe, Zap, Animation } from 'lucide-vue-next'

const { t } = useI18n()
const { locale } = useI18n()
const preferencesStore = usePreferencesStore()

// 通用设置
const general = computed(() => preferencesStore.general)

// 语言选项
const languageOptions = computed(() => [
  { value: 'zh-CN' as const, label: '简体中文', flag: '🇨🇳' },
  { value: 'en-US' as const, label: 'English', flag: '🇺🇸' }
])

// 页面过渡选项
const pageTransitionOptions = computed(() => [
  { value: 'fade' as const, label: t('preferences.general.pageTransition.fade') },
  { value: 'slide' as const, label: t('preferences.general.pageTransition.slide') },
  { value: 'scale' as const, label: t('preferences.general.pageTransition.scale') }
])

// 更新语言
function updateLanguage(lang: 'zh-CN' | 'en-US') {
  preferencesStore.updateGeneral('language', lang)
  locale.value = lang
}

// 切换动态标题
function toggleDynamicTitle() {
  preferencesStore.updateGeneral('dynamicTitle', !general.value.dynamicTitle)
}

// 切换启用动画
function toggleEnableAnimations() {
  preferencesStore.updateGeneral('enableAnimations', !general.value.enableAnimations)
}

// 更新页面过渡
function updatePageTransition(transition: 'fade' | 'slide' | 'scale') {
  preferencesStore.updateGeneral('pageTransition', transition)
}
</script>

<template>
  <div class="space-y-6">
    <!-- 标题 -->
    <div class="flex items-center gap-3 pb-2">
      <div class="p-2.5 bg-gradient-to-br from-blue-500 to-cyan-600 rounded-xl shadow-lg shadow-blue-500/25">
        <Settings2 :size="20" class="text-white" />
      </div>
      <div>
        <h3 class="text-lg font-semibold text-slate-900 dark:text-slate-100">
          {{ t('preferences.general.title') }}
        </h3>
        <p class="text-sm text-slate-500 dark:text-slate-400">
          {{ t('preferences.general.description') }}
        </p>
      </div>
    </div>

    <!-- 语言设置 - 卡片式选择 -->
    <section>
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">
        {{ t('preferences.general.language.label') }}
      </h4>
      <div class="grid grid-cols-2 gap-4">
        <button
          v-for="option in languageOptions"
          :key="option.value"
          class="group relative flex flex-col items-center gap-3 p-4 border-2 rounded-2xl transition-all duration-200 cursor-pointer"
          :class="general.language === option.value
            ? 'border-primary-500 bg-gradient-to-br from-primary-50 to-primary-100/50 dark:from-primary-900/30 dark:to-primary-800/20 shadow-lg shadow-primary-500/10'
            : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600 hover:bg-white dark:hover:bg-slate-700/30'"
          @click="updateLanguage(option.value)"
        >
          <span class="text-3xl">{{ option.flag }}</span>
          <span
            class="text-sm font-medium"
            :class="general.language === option.value
              ? 'text-primary-700 dark:text-primary-300'
              : 'text-slate-600 dark:text-slate-400'"
          >
            {{ option.label }}
          </span>
          <!-- 选中指示器 -->
          <div
            v-if="general.language === option.value"
            class="absolute top-3 right-3 w-5 h-5 bg-primary-500 rounded-full flex items-center justify-center text-white"
          >
            <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
            </svg>
          </div>
        </button>
      </div>
    </section>

    <!-- 页面过渡效果 - 可视化选择 -->
    <section>
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">
        {{ t('preferences.general.pageTransition.label') }}
      </h4>
      <div class="grid grid-cols-3 gap-3">
        <button
          v-for="option in pageTransitionOptions"
          :key="option.value"
          class="p-4 border-2 bg-white dark:bg-slate-700/50 transition-all duration-200 cursor-pointer hover:shadow-md rounded-xl"
          :class="general.pageTransition === option.value
            ? 'border-primary-500 shadow-md shadow-primary-500/10'
            : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600'"
          @click="updatePageTransition(option.value)"
        >
          <!-- 过渡动画预览图标 -->
          <div class="flex justify-center mb-2">
            <div class="flex gap-1">
              <div
                class="w-2 h-2 rounded-full"
                :class="option.value === 'fade' ? 'opacity-100' : option.value === 'slide' ? '-translate-x-1' : 'scale-75'"
                :style="{ transition: 'all 0.3s' }"
              />
              <div
                class="w-2 h-2 rounded-full"
                :class="option.value === 'fade' ? 'opacity-50' : option.value === 'slide' ? '' : 'scale-50'"
                :style="{ transition: 'all 0.3s' }"
              />
              <div
                class="w-2 h-2 rounded-full"
                :class="option.value === 'fade' ? 'opacity-0' : option.value === 'slide' ? 'translate-x-1' : 'scale-25'"
                :style="{ transition: 'all 0.3s' }"
              />
            </div>
          </div>
          <span class="text-sm font-medium block" :class="general.pageTransition === option.value ? 'text-primary-700 dark:text-primary-300' : 'text-slate-600 dark:text-slate-400'">
            {{ option.label }}
          </span>
        </button>
      </div>
    </section>

    <!-- 开关选项组 -->
    <section class="space-y-3">
      <!-- 动态标题 -->
      <div class="flex items-center justify-between p-4 bg-white dark:bg-slate-700/30 rounded-2xl border border-slate-200 dark:border-slate-700 transition-all duration-200 hover:shadow-md">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-violet-100 dark:bg-violet-900/30 rounded-xl">
            <svg class="w-5 h-5 text-violet-600 dark:text-violet-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <div>
            <h5 class="text-sm font-semibold text-slate-800 dark:text-slate-200">
              {{ t('preferences.general.dynamicTitle.label') }}
            </h5>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              {{ t('preferences.general.dynamicTitle.description') }}
            </p>
          </div>
        </div>
        <button
          class="relative w-14 h-7 rounded-full transition-all duration-300 cursor-pointer"
          :class="general.dynamicTitle ? 'bg-primary-500 shadow-lg shadow-primary-500/30' : 'bg-slate-300 dark:bg-slate-600'"
          @click="toggleDynamicTitle"
        >
          <span
            class="absolute top-1 left-1 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
            :class="general.dynamicTitle ? 'translate-x-7' : ''"
          />
        </button>
      </div>

      <!-- 启用动画 -->
      <div class="flex items-center justify-between p-4 bg-white dark:bg-slate-700/30 rounded-2xl border border-slate-200 dark:border-slate-700 transition-all duration-200 hover:shadow-md">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-green-100 dark:bg-green-900/30 rounded-xl">
            <Zap :size="20" class="text-green-600 dark:text-green-400" />
          </div>
          <div>
            <h5 class="text-sm font-semibold text-slate-800 dark:text-slate-200">
              {{ t('preferences.general.enableAnimations.label') }}
            </h5>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              {{ t('preferences.general.enableAnimations.description') }}
            </p>
          </div>
        </div>
        <button
          class="relative w-14 h-7 rounded-full transition-all duration-300 cursor-pointer"
          :class="general.enableAnimations ? 'bg-primary-500 shadow-lg shadow-primary-500/30' : 'bg-slate-300 dark:bg-slate-600'"
          @click="toggleEnableAnimations"
        >
          <span
            class="absolute top-1 left-1 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
            :class="general.enableAnimations ? 'translate-x-7' : ''"
          />
        </button>
      </div>
    </section>

    <!-- 关于信息 -->
    <section class="p-5 bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-800 dark:to-slate-800/50 rounded-2xl border border-slate-200 dark:border-slate-700">
      <div class="flex items-center gap-3 mb-4">
        <div class="p-2.5 bg-gradient-to-br from-slate-600 to-slate-700 rounded-xl">
          <Globe :size="20" class="text-white" />
        </div>
        <div>
          <h4 class="text-sm font-semibold text-slate-800 dark:text-slate-200">关于</h4>
        </div>
      </div>
      <div class="grid grid-cols-2 gap-3 text-center">
        <div class="p-3 bg-white dark:bg-slate-700/50 rounded-xl">
          <p class="text-xs text-slate-500 dark:text-slate-400 mb-1">版本</p>
          <p class="text-sm font-mono text-slate-700 dark:text-slate-300">1.0.0</p>
        </div>
        <div class="p-3 bg-white dark:bg-slate-700/50 rounded-xl">
          <p class="text-xs text-slate-500 dark:text-slate-400 mb-1">技术栈</p>
          <p class="text-sm text-slate-700 dark:text-slate-300">Vue 3</p>
        </div>
      </div>
      <div class="mt-3 p-3 bg-primary-50 dark:bg-primary-900/20 rounded-xl border border-primary-200 dark:border-primary-800">
        <p class="text-xs text-primary-700 dark:text-primary-300 text-center">
          Built with Vue 3 + Vite + Tailwind CSS + Element Plus
        </p>
      </div>
    </section>
  </div>
</template>
