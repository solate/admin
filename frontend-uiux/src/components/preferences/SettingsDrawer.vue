<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePreferencesStore } from '@/stores/modules/preferences'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  X,
  Palette,
  Layout,
  Keyboard,
  Settings2,
  RotateCcw,
  Download,
  Upload
} from 'lucide-vue-next'
import AppearanceTab from './tabs/AppearanceTab.vue'
import LayoutTab from './tabs/LayoutTab.vue'
import ShortcutsTab from './tabs/ShortcutsTab.vue'
import GeneralTab from './tabs/GeneralTab.vue'

// Props
const props = defineProps<{
  visible: boolean
}>()

// Emits
const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

const { t } = useI18n()
const preferencesStore = usePreferencesStore()

// 当前激活的标签页
const activeTab = ref<'appearance' | 'layout' | 'shortcuts' | 'general'>('appearance')

// 导入设置文件输入
const fileInputRef = ref<HTMLInputElement | null>(null)

// 计算属性 - 便于访问
const appearance = computed(() => preferencesStore.appearance)
const layout = computed(() => preferencesStore.layout)
const general = computed(() => preferencesStore.general)

// 标签页选项
const tabs = computed(() => [
  {
    id: 'appearance' as const,
    icon: Palette,
    label: t('preferences.tabs.appearance')
  },
  {
    id: 'layout' as const,
    icon: Layout,
    label: t('preferences.tabs.layout')
  },
  {
    id: 'shortcuts' as const,
    icon: Keyboard,
    label: t('preferences.tabs.shortcuts')
  },
  {
    id: 'general' as const,
    icon: Settings2,
    label: t('preferences.tabs.general')
  }
])

// 关闭抽屉
function close() {
  emit('update:visible', false)
}

// 重置所有设置
async function resetToDefaults() {
  try {
    await ElMessageBox.confirm(
      t('preferences.actions.resetConfirm'),
      t('common.confirm'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
    preferencesStore.resetToDefaults()
    ElMessage.success(t('common.success'))
  } catch {
    // 用户取消
  }
}

// 导出设置
function exportSettings() {
  const settingsJson = preferencesStore.exportSettings()
  const blob = new Blob([settingsJson], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `preferences-${new Date().toISOString().split('T')[0]}.json`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
  ElMessage.success(t('preferences.actions.copySuccess'))
}

// 触发导入设置
function triggerImport() {
  fileInputRef.value?.click()
}

// 导入设置
function importSettings(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  const reader = new FileReader()
  reader.onload = (e) => {
    const content = e.target?.result as string
    const success = preferencesStore.importSettings(content)
    if (success) {
      ElMessage.success(t('preferences.actions.importSuccess'))
    } else {
      ElMessage.error(t('preferences.actions.importError'))
    }
  }
  reader.readAsText(file)

  // 重置文件输入
  target.value = ''
}

// 抽屉打开时初始化 store
watch(() => props.visible, (visible) => {
  if (visible) {
    preferencesStore.initialize()
  }
})
</script>

<template>
  <!-- 遮罩层 - 增强模糊效果 -->
  <Transition
    enter-active-class="transition-opacity duration-300 ease-out"
    enter-from-class="opacity-0"
    enter-to-class="opacity-100"
    leave-active-class="transition-opacity duration-200 ease-in"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div
      v-if="visible"
      class="fixed inset-0 bg-black/50 backdrop-blur-sm z-40"
      @click="close"
    />
  </Transition>

  <!-- 抽屉 - Glassmorphism 设计 -->
  <Transition
    enter-active-class="transition-transform duration-300 ease-out"
    enter-from-class="translate-x-full"
    enter-to-class="translate-x-0"
    leave-active-class="transition-transform duration-200 ease-in"
    leave-from-class="translate-x-0"
    leave-to-class="translate-x-full"
  >
    <div
      v-if="visible"
      class="fixed right-0 top-0 bottom-0 w-full max-w-lg bg-white/95 dark:bg-slate-900/95 backdrop-blur-xl shadow-2xl z-50 flex flex-col"
    >
      <!-- 头部 - 优化的视觉层次 -->
      <div class="relative px-6 py-5 border-b border-slate-200/60 dark:border-slate-700/60 bg-gradient-to-r from-slate-50/50 to-transparent dark:from-slate-800/50">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div class="p-2.5 bg-gradient-to-br from-primary-500 to-primary-600 rounded-xl shadow-lg shadow-primary-500/25">
              <Settings2 :size="20" class="text-white" />
            </div>
            <div>
              <h2 class="text-xl font-bold text-slate-900 dark:text-slate-100 tracking-tight">
                {{ t('preferences.title') }}
              </h2>
              <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
                {{ t('preferences.description') }}
              </p>
            </div>
          </div>
          <button
            class="p-2.5 hover:bg-slate-200/80 dark:hover:bg-slate-700/80 rounded-xl transition-all duration-200 cursor-pointer group"
            :aria-label="t('common.close')"
            @click="close"
          >
            <X :size="20" class="text-slate-500 dark:text-slate-400 group-hover:text-slate-700 dark:group-hover:text-slate-200 transition-colors" />
          </button>
        </div>
      </div>

      <!-- Tab 导航 - 悬浮卡片风格 -->
      <div class="px-6 py-4 border-b border-slate-200/60 dark:border-slate-700/60 bg-slate-50/50 dark:bg-slate-800/30">
        <div class="flex items-center gap-1.5 overflow-x-auto scrollbar-hide">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            class="flex items-center gap-2 px-3.5 py-2 rounded-lg text-sm font-medium transition-all duration-200 whitespace-nowrap cursor-pointer"
            :class="activeTab === tab.id
              ? 'bg-gradient-to-r from-primary-500 to-primary-600 text-white shadow-md shadow-primary-500/20'
              : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100/80 dark:hover:bg-slate-700/50'"
            @click="activeTab = tab.id"
          >
            <component :is="tab.icon" :size="15" />
            <span>{{ tab.label }}</span>
          </button>
        </div>
      </div>

      <!-- 内容区域 -->
      <div class="flex-1 overflow-y-auto">
        <div class="p-6">
          <!-- 外观设置 -->
          <AppearanceTab v-if="activeTab === 'appearance'" />

          <!-- 布局设置 -->
          <LayoutTab v-if="activeTab === 'layout'" />

          <!-- 快捷键设置 -->
          <ShortcutsTab v-if="activeTab === 'shortcuts'" />

          <!-- 通用设置 -->
          <GeneralTab v-if="activeTab === 'general'" />

          <!-- 设置预览面板 - 优化样式 -->
          <Transition
            enter-active-class="transition-all duration-200"
            enter-from-class="opacity-0 translate-y-2"
            enter-to-class="opacity-100 translate-y-0"
            leave-active-class="transition-all duration-150"
            leave-from-class="opacity-100 translate-y-0"
            leave-to-class="opacity-0 translate-y-2"
          >
            <div
              v-if="activeTab !== 'shortcuts'"
              class="mt-8 p-5 bg-gradient-to-br from-slate-50 via-white to-slate-50 dark:from-slate-800 dark:via-slate-800/50 dark:to-slate-800/30 rounded-2xl border border-slate-200/60 dark:border-slate-700/60 shadow-sm"
            >
              <h4 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-4 flex items-center gap-2">
                <span class="w-6 h-6 rounded-lg bg-gradient-to-br from-primary-500 to-primary-600 flex items-center justify-center">
                  <svg class="w-3.5 h-3.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                </span>
                <span>设置预览</span>
              </h4>

              <!-- 主题预览 -->
              <div v-if="activeTab === 'appearance'" class="space-y-3">
                <div class="flex items-center justify-between p-3 bg-white dark:bg-slate-700/50 rounded-xl">
                  <span class="text-sm text-slate-600 dark:text-slate-400">主题色</span>
                  <div class="flex items-center gap-2.5">
                    <span
                      class="w-5 h-5 rounded-full border-2 border-white dark:border-slate-600 shadow-sm"
                      :style="{ backgroundColor: appearance.primaryColor }"
                    />
                    <span class="text-xs font-mono text-slate-500 dark:text-slate-400">
                      {{ appearance.primaryColor }}
                    </span>
                  </div>
                </div>
                <div class="flex items-center justify-between p-3 bg-white dark:bg-slate-700/50 rounded-xl">
                  <span class="text-sm text-slate-600 dark:text-slate-400">圆角大小</span>
                  <span class="text-sm text-slate-500 dark:text-slate-400">
                    {{ appearance.borderRadius === 'none' ? '无' : appearance.borderRadius === 'small' ? '小' : appearance.borderRadius === 'medium' ? '中' : '大' }}
                  </span>
                </div>
                <div class="flex items-center justify-between p-3 bg-white dark:bg-slate-700/50 rounded-xl">
                  <span class="text-sm text-slate-600 dark:text-slate-400">主题模式</span>
                  <span class="text-sm text-slate-500 dark:text-slate-400">
                    {{ appearance.themeMode === 'light' ? '浅色' : appearance.themeMode === 'dark' ? '深色' : '跟随系统' }}
                  </span>
                </div>
                <div v-if="appearance.colorBlindMode !== 'none'" class="flex items-center justify-between p-3 bg-amber-50 dark:bg-amber-900/20 rounded-xl border border-amber-200 dark:border-amber-800">
                  <span class="text-sm text-amber-700 dark:text-amber-300">色盲模式</span>
                  <span class="text-sm text-amber-600 dark:text-amber-400 font-medium">
                    {{ appearance.colorBlindMode === 'protanopia' ? '红色盲' : appearance.colorBlindMode === 'deuteranopia' ? '绿色盲' : '蓝色盲' }}
                  </span>
                </div>
              </div>

              <!-- 布局预览 -->
              <div v-if="activeTab === 'layout'" class="space-y-3">
                <div class="flex items-center justify-between p-3 bg-white dark:bg-slate-700/50 rounded-xl">
                  <span class="text-sm text-slate-600 dark:text-slate-400">布局模式</span>
                  <span class="text-sm text-slate-500 dark:text-slate-400">
                    {{ layout.layoutMode === 'sidebar' ? '侧边栏' : '顶部栏' }}
                  </span>
                </div>
                <div class="flex items-center justify-between p-3 bg-white dark:bg-slate-700/50 rounded-xl">
                  <span class="text-sm text-slate-600 dark:text-slate-400">侧边栏宽度</span>
                  <span class="text-sm text-slate-500 dark:text-slate-400">
                    {{ layout.sidebarWidth === 'narrow' ? '窄 (64px)' : layout.sidebarWidth === 'medium' ? '中 (256px)' : '宽 (320px)' }}
                  </span>
                </div>
                <div class="flex items-center justify-between p-3 bg-white dark:bg-slate-700/50 rounded-xl">
                  <span class="text-sm text-slate-600 dark:text-slate-400">导航样式</span>
                  <span class="text-sm text-slate-500 dark:text-slate-400">
                    {{ layout.navStyle === 'icon-text' ? '图标+文字' : '仅图标' }}
                  </span>
                </div>
                <div class="flex flex-wrap gap-2 mt-3 p-3 bg-white dark:bg-slate-700/50 rounded-xl">
                  <span
                    v-if="layout.showBreadcrumbs"
                    class="px-3 py-1.5 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300 text-sm rounded-lg font-medium"
                  >面包屑</span>
                  <span
                    v-if="layout.showTabs"
                    class="px-3 py-1.5 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300 text-sm rounded-lg font-medium"
                  >标签页</span>
                  <span
                    v-if="layout.showWidgets"
                    class="px-3 py-1.5 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300 text-sm rounded-lg font-medium"
                  >小部件</span>
                  <span
                    v-if="layout.showFooter"
                    class="px-3 py-1.5 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300 text-sm rounded-lg font-medium"
                  >页脚</span>
                </div>
              </div>

              <!-- 通用设置预览 -->
              <div v-if="activeTab === 'general'" class="space-y-3">
                <div class="flex items-center justify-between p-3 bg-white dark:bg-slate-700/50 rounded-xl">
                  <span class="text-sm text-slate-600 dark:text-slate-400">语言</span>
                  <span class="text-sm text-slate-500 dark:text-slate-400">
                    {{ general.language === 'zh-CN' ? '简体中文' : 'English' }}
                  </span>
                </div>
                <div class="flex items-center justify-between p-3 bg-white dark:bg-slate-700/50 rounded-xl">
                  <span class="text-sm text-slate-600 dark:text-slate-400">页面过渡</span>
                  <span class="text-sm text-slate-500 dark:text-slate-400">
                    {{ general.pageTransition === 'fade' ? '淡入淡出' : general.pageTransition === 'slide' ? '滑动' : '缩放' }}
                  </span>
                </div>
                <div class="flex items-center justify-between p-3 bg-white dark:bg-slate-700/50 rounded-xl">
                  <span class="text-sm text-slate-600 dark:text-slate-400">动画</span>
                  <span
                    class="text-sm font-medium"
                    :class="general.enableAnimations ? 'text-green-600 dark:text-green-400' : 'text-slate-400 dark:text-slate-500'"
                  >
                    {{ general.enableAnimations ? '已启用' : '已禁用' }}
                  </span>
                </div>
              </div>
            </div>
          </Transition>
        </div>
      </div>

      <!-- 底部操作栏 - 优化样式 -->
      <div class="flex items-center justify-between px-6 py-4 border-t border-slate-200/60 dark:border-slate-700/60 bg-gradient-to-r from-slate-50/80 to-slate-100/80 dark:from-slate-800/50 dark:to-slate-800/30">
        <div class="flex items-center gap-2">
          <!-- 导入设置 -->
          <button
            class="flex items-center gap-2 px-4 py-2.5 text-sm font-medium text-slate-700 dark:text-slate-300 bg-white dark:bg-slate-700 hover:bg-slate-100 dark:hover:bg-slate-600 rounded-xl transition-all duration-200 cursor-pointer shadow-sm hover:shadow border border-slate-200 dark:border-slate-600"
            :title="t('preferences.actions.import')"
            @click="triggerImport"
          >
            <Upload :size="16" />
            <span class="hidden sm:inline">{{ t('preferences.actions.import') }}</span>
          </button>
          <input
            ref="fileInputRef"
            type="file"
            accept=".json"
            class="hidden"
            @change="importSettings"
          >

          <!-- 导出设置 -->
          <button
            class="flex items-center gap-2 px-4 py-2.5 text-sm font-medium text-slate-700 dark:text-slate-300 bg-white dark:bg-slate-700 hover:bg-slate-100 dark:hover:bg-slate-600 rounded-xl transition-all duration-200 cursor-pointer shadow-sm hover:shadow border border-slate-200 dark:border-slate-600"
            :title="t('preferences.actions.export')"
            @click="exportSettings"
          >
            <Download :size="16" />
            <span class="hidden sm:inline">{{ t('preferences.actions.export') }}</span>
          </button>
        </div>

        <!-- 重置按钮 -->
        <button
          class="flex items-center gap-2 px-4 py-2.5 text-sm font-medium text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 hover:bg-red-100 dark:hover:bg-red-900/30 rounded-xl transition-all duration-200 cursor-pointer border border-red-200 dark:border-red-800"
          @click="resetToDefaults"
        >
          <RotateCcw :size="16" />
          <span>{{ t('preferences.actions.reset') }}</span>
        </button>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
/* 隐藏滚动条但保持滚动功能 */
.scrollbar-hide {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}
</style>
