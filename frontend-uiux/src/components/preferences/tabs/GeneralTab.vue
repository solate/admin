<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePreferencesStore } from '@/stores/modules/preferences'
import { Settings2, Globe, Zap, Sparkles, ChevronDown, ChevronUp } from 'lucide-vue-next'

const { t } = useI18n()
const { locale } = useI18n()
const preferencesStore = usePreferencesStore()

// 通用设置
const general = computed(() => preferencesStore.general)

// 折叠面板状态
const animationExpanded = ref(false)

// 语言选项
const languageOptions = computed(() => [
  { value: 'zh-CN' as const, label: '简体中文', flag: '🇨🇳' },
  { value: 'en-US' as const, label: 'English', flag: '🇺🇸' }
])

// 页面过渡选项
const pageTransitionOptions = computed(() => [
  { value: 'fade' as const, label: '淡入淡出', icon: 'fade' },
  { value: 'slide' as const, label: '滑动', icon: 'slide' },
  { value: 'scale' as const, label: '缩放', icon: 'scale' },
  { value: 'zoom' as const, label: '缩放', icon: 'zoom' },
  { value: 'fade-slide' as const, label: '淡入滑动', icon: 'fade-slide' }
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

// 切换页面进度条
function togglePageProgress() {
  preferencesStore.updateGeneral('pageProgress', !general.value.pageProgress)
}

// 切换页面加载
function togglePageLoading() {
  preferencesStore.updateGeneral('pageLoading', !general.value.pageLoading)
}

// 切换检查更新
function toggleCheckUpdates() {
  preferencesStore.updateGeneral('enableCheckUpdates', !general.value.enableCheckUpdates)
}

// 更新页面过渡
function updatePageTransition(transition: 'fade' | 'slide' | 'scale' | 'zoom' | 'fade-slide') {
  preferencesStore.updateGeneral('pageTransition', transition)
}

// 更新检查更新间隔
function updateUpdateInterval(value: number) {
  preferencesStore.updateGeneral('checkUpdatesInterval', value)
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

    <!-- 语言设置 -->
    <section>
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">
        语言 / Language
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

    <!-- 页面过渡效果 -->
    <section>
      <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3">
        页面切换动画
      </h4>
      <div class="grid grid-cols-5 gap-2">
        <button
          v-for="option in pageTransitionOptions"
          :key="option.value"
          class="p-3 border-2 bg-white dark:bg-slate-700/50 transition-all duration-200 cursor-pointer hover:shadow-md rounded-xl text-center"
          :class="general.pageTransition === option.value
            ? 'border-primary-500 shadow-md shadow-primary-500/10'
            : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600'"
          @click="updatePageTransition(option.value)"
        >
          <div class="w-6 h-6 mx-auto mb-2 flex items-center justify-center">
            <!-- Fade Icon -->
            <svg v-if="option.icon === 'fade'" class="w-5 h-5 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <circle cx="12" cy="12" r="3" stroke-width="2" opacity="1" />
              <circle cx="12" cy="12" r="6" stroke-width="2" opacity="0.5" />
              <circle cx="12" cy="12" r="9" stroke-width="2" opacity="0.25" />
            </svg>
            <!-- Slide Icon -->
            <svg v-else-if="option.icon === 'slide'" class="w-5 h-5 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
            <!-- Scale Icon -->
            <svg v-else-if="option.icon === 'scale'" class="w-5 h-5 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <circle cx="12" cy="12" r="3" stroke-width="2" />
              <circle cx="12" cy="12" r="8" stroke-width="2" stroke-dasharray="2 2" />
            </svg>
            <!-- Zoom Icon -->
            <svg v-else-if="option.icon === 'zoom'" class="w-5 h-5 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <circle cx="11" cy="11" r="6" stroke-width="2" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-4.35-4.35" />
            </svg>
            <!-- Fade-Slide Icon -->
            <svg v-else class="w-5 h-5 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" opacity="0.5" />
              <circle cx="12" cy="12" r="8" stroke-width="1" opacity="0.3" />
            </svg>
          </div>
          <span class="text-xs font-medium" :class="general.pageTransition === option.value ? 'text-primary-700 dark:text-primary-300' : 'text-slate-600 dark:text-slate-400'">
            {{ option.label }}
          </span>
        </button>
      </div>
    </section>

    <!-- 动画设置折叠面板 -->
    <section class="border-2 border-slate-200 dark:border-slate-700 rounded-2xl overflow-hidden">
      <button
        class="w-full flex items-center justify-between p-4 bg-slate-50 dark:bg-slate-800/50 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
        @click="animationExpanded = !animationExpanded"
      >
        <div class="flex items-center gap-3">
          <div class="p-2 bg-gradient-to-br from-green-500 to-emerald-600 rounded-lg">
            <Sparkles :size="16" class="text-white" />
          </div>
          <span class="text-sm font-semibold text-slate-700 dark:text-slate-300">动画效果</span>
        </div>
        <component :is="animationExpanded ? ChevronUp : ChevronDown" :size="18" class="text-slate-500" />
      </button>

      <Transition
        enter-active-class="transition-all duration-200 ease-out"
        enter-from-class="opacity-0 -translate-y-2"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition-all duration-150 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 -translate-y-2"
      >
        <div v-if="animationExpanded" class="p-4 space-y-3 bg-white dark:bg-slate-900/30">
          <!-- 启用动画 -->
          <div class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-green-100 dark:bg-green-900/30 rounded-lg">
                <Zap :size="16" class="text-green-600 dark:text-green-400" />
              </div>
              <div>
                <h5 class="text-sm font-medium text-slate-800 dark:text-slate-200">启用动画</h5>
                <p class="text-xs text-slate-500 dark:text-slate-400">全局动画效果</p>
              </div>
            </div>
            <button
              class="relative w-12 h-6 rounded-full transition-all duration-300 cursor-pointer"
              :class="general.enableAnimations ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
              @click="toggleEnableAnimations"
            >
              <span
                class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
                :class="general.enableAnimations ? 'translate-x-6' : ''"
              />
            </button>
          </div>

          <!-- 页面切换进度条 -->
          <div class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
                <svg class="w-4 h-4 text-blue-600 dark:text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 12h16" />
                </svg>
              </div>
              <div>
                <h5 class="text-sm font-medium text-slate-800 dark:text-slate-200">页面切换进度条</h5>
                <p class="text-xs text-slate-500 dark:text-slate-400">顶部加载进度条</p>
              </div>
            </div>
            <button
              class="relative w-12 h-6 rounded-full transition-all duration-300 cursor-pointer"
              :class="general.pageProgress ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
              @click="togglePageProgress"
            >
              <span
                class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
                :class="general.pageProgress ? 'translate-x-6' : ''"
              />
            </button>
          </div>

          <!-- 页面切换 Loading -->
          <div class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-purple-100 dark:bg-purple-900/30 rounded-lg">
                <svg class="w-4 h-4 text-purple-600 dark:text-purple-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
              </div>
              <div>
                <h5 class="text-sm font-medium text-slate-800 dark:text-slate-200">页面切换 Loading</h5>
                <p class="text-xs text-slate-500 dark:text-slate-400">页面加载遮罩层</p>
              </div>
            </div>
            <button
              class="relative w-12 h-6 rounded-full transition-all duration-300 cursor-pointer"
              :class="general.pageLoading ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
              @click="togglePageLoading"
            >
              <span
                class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
                :class="general.pageLoading ? 'translate-x-6' : ''"
              />
            </button>
          </div>

          <!-- 动态标题 -->
          <div class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-violet-100 dark:bg-violet-900/30 rounded-lg">
                <svg class="w-4 h-4 text-violet-600 dark:text-violet-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
              </div>
              <div>
                <h5 class="text-sm font-medium text-slate-800 dark:text-slate-200">动态标题</h5>
                <p class="text-xs text-slate-500 dark:text-slate-400">页面标题跟随内容变化</p>
              </div>
            </div>
            <button
              class="relative w-12 h-6 rounded-full transition-all duration-300 cursor-pointer"
              :class="general.dynamicTitle ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
              @click="toggleDynamicTitle"
            >
              <span
                class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
                :class="general.dynamicTitle ? 'translate-x-6' : ''"
              />
            </button>
          </div>

          <!-- 检查更新 -->
          <div class="p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl">
            <div class="flex items-center justify-between mb-3">
              <div class="flex items-center gap-3">
                <div class="p-2 bg-amber-100 dark:bg-amber-900/30 rounded-lg">
                  <svg class="w-4 h-4 text-amber-600 dark:text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                </div>
                <div>
                  <h5 class="text-sm font-medium text-slate-800 dark:text-slate-200">检查更新</h5>
                  <p class="text-xs text-slate-500 dark:text-slate-400">定期检查新版本</p>
                </div>
              </div>
              <button
                class="relative w-12 h-6 rounded-full transition-all duration-300 cursor-pointer"
                :class="general.enableCheckUpdates ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
                @click="toggleCheckUpdates"
              >
                <span
                  class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
                  :class="general.enableCheckUpdates ? 'translate-x-6' : ''"
                />
              </button>
            </div>
            <div v-if="general.enableCheckUpdates">
              <label class="text-xs text-slate-600 dark:text-slate-400 mb-2 block">检查间隔：{{ general.checkUpdatesInterval }} 天</label>
              <input
                type="range"
                :value="general.checkUpdatesInterval"
                min="1"
                max="30"
                step="1"
                class="w-full h-2 bg-slate-200 dark:bg-slate-700 rounded-lg appearance-none cursor-pointer accent-primary-500"
                @input="(e) => updateUpdateInterval(Number((e.target as HTMLInputElement).value))"
              />
            </div>
          </div>
        </div>
      </Transition>
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
