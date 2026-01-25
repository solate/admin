<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePreferencesStore } from '@/stores/modules/preferences'
import { Keyboard, RotateCcw } from 'lucide-vue-next'
import { ElMessage } from 'element-plus'

const { t } = useI18n()
const preferencesStore = usePreferencesStore()

// 录制状态
const isRecording = ref(false)
const recordingActionId = ref<string | null>(null)
const recordedKeys = ref<string[]>([])

// 默认快捷键配置
const defaultShortcuts: Record<string, string> = {
  search: 'Cmd+K',
  notifications: 'Cmd+N',
  settings: 'Cmd+,',
  darkMode: 'Cmd+Shift+D',
  sidebar: 'Cmd+B'
}

// 当前快捷键配置
const shortcuts = computed(() => preferencesStore.shortcuts)

// 快捷键配置列表
const shortcutConfigs = computed(() => [
  {
    id: 'search',
    name: t('preferences.shortcuts.search'),
    iconBg: 'bg-blue-100 dark:bg-blue-900/30',
    iconColor: 'text-blue-600 dark:text-blue-400',
    currentValue: shortcuts.value.search || defaultShortcuts.search
  },
  {
    id: 'notifications',
    name: t('preferences.shortcuts.notifications'),
    iconBg: 'bg-purple-100 dark:bg-purple-900/30',
    iconColor: 'text-purple-600 dark:text-purple-400',
    currentValue: shortcuts.value.notifications || defaultShortcuts.notifications
  },
  {
    id: 'settings',
    name: t('preferences.shortcuts.settings'),
    iconBg: 'bg-slate-100 dark:bg-slate-700/50',
    iconColor: 'text-slate-600 dark:text-slate-400',
    currentValue: shortcuts.value.settings || defaultShortcuts.settings
  },
  {
    id: 'darkMode',
    name: t('preferences.shortcuts.darkMode'),
    iconBg: 'bg-indigo-100 dark:bg-indigo-900/30',
    iconColor: 'text-indigo-600 dark:text-indigo-400',
    currentValue: shortcuts.value.darkMode || defaultShortcuts.darkMode
  },
  {
    id: 'sidebar',
    name: t('preferences.shortcuts.sidebar'),
    iconBg: 'bg-emerald-100 dark:bg-emerald-900/30',
    iconColor: 'text-emerald-600 dark:text-emerald-400',
    currentValue: shortcuts.value.sidebar || defaultShortcuts.sidebar
  }
])

// 快捷键图标
const shortcutIcons: Record<string, string> = {
  search: '<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>',
  notifications: '<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4h.006c0 2.692.002 5.32.497 7.578 1.035" /></svg>',
  settings: '<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 2.924-1.756.426 1.756 2.924 1.756 2.924 1.756a2.924 2.924 0 11-5.849 0c0-1.756 2.924-1.756 2.924-1.756zm0 0a2.924 2.924 0 115.849 0 2.924 2.924 0 01-5.849 0zM8.75 6h7.5M12 12h7.5m-7.5 3h7.5m-7.5 3h7.5" /></svg>',
  darkMode: '<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" /></svg>',
  sidebar: '<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" /></svg>'
}

// 检查快捷键冲突
function checkConflict(actionId: string, keyCombo: string): string | null {
  for (const [id, config] of Object.entries(shortcutConfigs.value)) {
    if (id !== actionId && config.currentValue === keyCombo) {
      return config.name
    }
  }
  return null
}

// 格式化按键显示
function formatKeyDisplay(keyCombo: string): string {
  return keyCombo
    .replace(/Cmd/g, '⌘')
    .replace(/Ctrl/g, '⌃')
    .replace(/Alt/g, '⌥')
    .replace(/Shift/g, '⇧')
    .replace(/\+/g, ' + ')
}

// 开始录制快捷键
function startRecording(actionId: string) {
  isRecording.value = true
  recordingActionId.value = actionId
  recordedKeys.value = []
}

// 停止录制
function stopRecording() {
  isRecording.value = false
  recordingActionId.value = null
  recordedKeys.value = []
}

// 处理键盘事件
function handleKeyDown(event: KeyboardEvent) {
  if (!isRecording.value || !recordingActionId.value) return

  event.preventDefault()
  event.stopPropagation()

  const keys: string[] = []

  // 检测修饰键
  if (event.metaKey) keys.push('Cmd')
  if (event.ctrlKey) keys.push('Ctrl')
  if (event.altKey) keys.push('Alt')
  if (event.shiftKey) keys.push('Shift')

  // 主键
  if (event.key && !['Meta', 'Control', 'Alt', 'Shift'].includes(event.key)) {
    keys.push(event.key.toUpperCase())
  }

  // 必须有主键
  const mainKey = keys.find(k => !['Cmd', 'Ctrl', 'Alt', 'Shift'].includes(k))
  if (!mainKey) return

  const keyCombo = keys.join('+')
  recordedKeys.value = keys

  // 检查冲突
  const conflict = checkConflict(recordingActionId.value, keyCombo)
  if (conflict) {
    ElMessage.warning(`${t('preferences.shortcuts.conflictDesc')}: ${conflict}`)
    stopRecording()
    return
  }

  // 保存快捷键
  preferencesStore.updateShortcut(recordingActionId.value, keyCombo)
  ElMessage.success(`${shortcutConfigs.value.find(c => c.id === recordingActionId.value)?.name}: ${formatKeyDisplay(keyCombo)}`)
  stopRecording()
}

// 清除快捷键
function clearShortcut(actionId: string) {
  preferencesStore.updateShortcut(actionId, '')
  ElMessage.success(t('common.success'))
}

// 重置快捷键
function resetShortcut(actionId: string) {
  const defaultValue = defaultShortcuts[actionId]
  if (defaultValue) {
    preferencesStore.updateShortcut(actionId, defaultValue)
    ElMessage.success(t('common.success'))
  }
}

// 重置所有快捷键
function resetAllShortcuts() {
  Object.entries(defaultShortcuts).forEach(([id, value]) => {
    preferencesStore.updateShortcut(id, value)
  })
  ElMessage.success(t('common.success'))
}

// 生命周期
onMounted(() => {
  // 初始化默认快捷键
  Object.entries(defaultShortcuts).forEach(([id, value]) => {
    if (!shortcuts.value[id]) {
      preferencesStore.updateShortcut(id, value)
    }
  })
  document.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeyDown)
})
</script>

<template>
  <div class="space-y-6">
    <!-- 标题 -->
    <div class="flex items-center gap-3 pb-2">
      <div class="p-2.5 bg-gradient-to-br from-amber-500 to-orange-600 rounded-xl shadow-lg shadow-amber-500/25">
        <Keyboard :size="20" class="text-white" />
      </div>
      <div>
        <h3 class="text-lg font-semibold text-slate-900 dark:text-slate-100">
          {{ t('preferences.shortcuts.title') }}
        </h3>
        <p class="text-sm text-slate-500 dark:text-slate-400">
          {{ t('preferences.shortcuts.description') }}
        </p>
      </div>
    </div>

    <!-- 录制提示 -->
    <Transition
      enter-active-class="transition-all duration-200"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition-all duration-150"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-if="isRecording"
        class="p-4 bg-gradient-to-r from-amber-50 to-orange-50 dark:from-amber-900/30 dark:to-orange-900/30 border-2 border-amber-300 dark:border-amber-700 rounded-2xl"
      >
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 bg-amber-500 rounded-full flex items-center justify-center animate-pulse">
            <Keyboard :size="18" class="text-white" />
          </div>
          <div>
            <p class="text-sm font-semibold text-amber-800 dark:text-amber-200">
              {{ t('preferences.shortcuts.pressKey') }}
            </p>
            <p class="text-xs text-amber-600 dark:text-amber-400 mt-0.5">
              按 ESC 取消录制
            </p>
          </div>
        </div>
      </div>
    </Transition>

    <!-- 快捷键列表 -->
    <section class="space-y-3">
      <div
        v-for="config in shortcutConfigs"
        :key="config.id"
        class="flex items-center justify-between p-4 bg-white dark:bg-slate-700/30 rounded-2xl border-2 border-slate-200 dark:border-slate-700 transition-all duration-200 hover:shadow-md"
        :class="{
          'ring-2 ring-amber-500 dark:ring-amber-400 shadow-lg shadow-amber-500/10': isRecording && recordingActionId === config.id
        }"
      >
        <div class="flex items-center gap-3">
          <div class="p-2.5 rounded-xl" :class="config.iconBg">
            <span class="block" :class="config.iconColor" v-html="shortcutIcons[config.id]" />
          </div>
          <div>
            <p class="text-sm font-semibold text-slate-800 dark:text-slate-200">
              {{ config.name }}
            </p>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5 font-mono">
              {{ config.currentValue ? formatKeyDisplay(config.currentValue) : t('preferences.shortcuts.clear') }}
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <!-- 当前快捷键显示 -->
          <kbd
            v-if="config.currentValue"
            class="hidden sm:block px-3 py-1.5 bg-slate-100 dark:bg-slate-700 border-2 border-slate-300 dark:border-slate-600 rounded-lg text-xs font-mono text-slate-700 dark:text-slate-300 shadow-sm"
          >
            {{ formatKeyDisplay(config.currentValue) }}
          </kbd>

          <!-- 录制按钮 -->
          <button
            class="px-4 py-2 text-sm font-medium rounded-xl transition-all duration-200 cursor-pointer"
            :class="isRecording && recordingActionId === config.id
              ? 'bg-amber-500 text-white shadow-lg shadow-amber-500/25'
              : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-600'"
            @click="isRecording ? stopRecording() : startRecording(config.id)"
          >
            {{ isRecording && recordingActionId === config.id ? t('common.cancel') : t('preferences.shortcuts.recordShortcut') }}
          </button>

          <!-- 重置按钮 -->
          <button
            v-if="config.currentValue !== defaultShortcuts[config.id]"
            class="p-2.5 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-xl transition-all duration-200 cursor-pointer"
            :title="t('preferences.shortcuts.reset')"
            @click="resetShortcut(config.id)"
          >
            <RotateCcw :size="16" />
          </button>
        </div>
      </div>
    </section>

    <!-- 全部重置 -->
    <section class="pt-2 border-t border-slate-200 dark:border-slate-700">
      <button
        class="w-full flex items-center justify-center gap-3 px-4 py-3 text-sm font-medium text-slate-700 dark:text-slate-300 bg-slate-100 dark:bg-slate-700/50 hover:bg-slate-200 dark:hover:bg-slate-700 rounded-2xl transition-all duration-200 cursor-pointer"
        @click="resetAllShortcuts"
      >
        <RotateCcw :size="18" />
        <span>{{ t('preferences.shortcuts.resetAll') }}</span>
      </button>
    </section>

    <!-- 使用说明 -->
    <section class="p-4 bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 rounded-2xl border border-blue-200 dark:border-blue-800">
      <h4 class="text-sm font-semibold text-blue-800 dark:text-blue-300 mb-3 flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>快捷键说明</span>
      </h4>
      <ul class="text-xs text-blue-700 dark:text-blue-400 space-y-2">
        <li class="flex items-start gap-2">
          <span class="text-blue-500 font-bold">1.</span>
          <span>点击"按下快捷键组合"开始录制</span>
        </li>
        <li class="flex items-start gap-2">
          <span class="text-blue-500 font-bold">2.</span>
          <span>按下想要的组合键（如 ⌘+K）</span>
        </li>
        <li class="flex items-start gap-2">
          <span class="text-blue-500 font-bold">3.</span>
          <span>按 ESC 键取消录制</span>
        </li>
        <li class="flex items-start gap-2">
          <span class="text-blue-500 font-bold">4.</span>
          <span>快捷键会自动保存到本地存储</span>
        </li>
      </ul>
    </section>
  </div>
</template>
