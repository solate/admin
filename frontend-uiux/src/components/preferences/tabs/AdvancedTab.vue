<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePreferencesStore } from '@/stores/modules/preferences'
import {
  Settings,
  LayoutGrid,
  Building2,
  ChevronDown,
  ChevronUp,
  Search,
  Palette,
  Globe,
  Maximize,
  Bell,
  Lock,
  PanelLeft,
  RefreshCw
} from 'lucide-vue-next'

const { t } = useI18n()
const preferencesStore = usePreferencesStore()

// 小部件和版权设置
const widgets = computed(() => preferencesStore.widgets)
const copyright = computed(() => preferencesStore.copyright)

// 折叠面板状态
const widgetsExpanded = ref(true)
const copyrightExpanded = ref(true)

// 小部件位置选项
const positionOptions = computed(() => [
  { value: 'auto' as const, label: '自动', description: '根据布局自动选择' },
  { value: 'header' as const, label: '顶栏', description: '显示在顶部导航栏' },
  { value: 'sidebar' as const, label: '侧边栏', description: '显示在侧边栏底部' }
])

// 小部件定义
const widgetItems = computed(() => [
  { key: 'globalSearch' as const, label: '全局搜索', icon: Search, description: '快捷搜索功能' },
  { key: 'themeToggle' as const, label: '主题切换', icon: Palette, description: '明暗主题切换' },
  { key: 'languageToggle' as const, label: '语言切换', icon: Globe, description: '中英文切换' },
  { key: 'fullscreen' as const, label: '全屏', icon: Maximize, description: '全屏显示' },
  { key: 'notification' as const, label: '通知', icon: Bell, description: '消息通知' },
  { key: 'lockScreen' as const, label: '锁屏', icon: Lock, description: '屏幕锁定' },
  { key: 'sidebarToggle' as const, label: '侧边栏切换', icon: PanelLeft, description: '显示/隐藏侧边栏' },
  { key: 'refresh' as const, label: '刷新', icon: RefreshCw, description: '刷新当前页面' }
])

// 切换小部件显示
function toggleWidget(key: keyof typeof preferencesStore.widgets) {
  const currentValue = widgets.value[key]
  preferencesStore.updateWidgets(key, !currentValue as never)
}

// 更新小部件位置
function updateWidgetPosition(position: 'auto' | 'header' | 'sidebar') {
  preferencesStore.updateWidgets('position', position)
}

// 切换版权启用
function toggleCopyrightEnable() {
  preferencesStore.updateCopyright('enable', !copyright.value.enable)
}

// 更新版权信息
function updateCopyrightField<K extends keyof typeof preferencesStore.copyright>(
  key: K,
  value: typeof preferencesStore.copyright[K]
) {
  preferencesStore.updateCopyright(key, value as never)
}
</script>

<template>
  <div class="space-y-6">
    <!-- 标题 -->
    <div class="flex items-center gap-3 pb-2">
      <div class="p-2.5 bg-gradient-to-br from-violet-500 to-purple-600 rounded-xl shadow-lg shadow-violet-500/25">
        <Settings :size="20" class="text-white" />
      </div>
      <div>
        <h3 class="text-lg font-semibold text-slate-900 dark:text-slate-100">
          {{ t('preferences.advanced.title') }}
        </h3>
        <p class="text-sm text-slate-500 dark:text-slate-400">
          {{ t('preferences.advanced.description') }}
        </p>
      </div>
    </div>

    <!-- 小部件设置 -->
    <section class="border-2 border-slate-200 dark:border-slate-700 rounded-2xl overflow-hidden">
      <button
        class="w-full flex items-center justify-between p-4 bg-slate-50 dark:bg-slate-800/50 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
        @click="widgetsExpanded = !widgetsExpanded"
      >
        <div class="flex items-center gap-3">
          <div class="p-2 bg-gradient-to-br from-blue-500 to-cyan-600 rounded-lg">
            <LayoutGrid :size="16" class="text-white" />
          </div>
          <span class="text-sm font-semibold text-slate-700 dark:text-slate-300">小部件设置</span>
        </div>
        <component :is="widgetsExpanded ? ChevronUp : ChevronDown" :size="18" class="text-slate-500" />
      </button>

      <Transition
        enter-active-class="transition-all duration-200 ease-out"
        enter-from-class="opacity-0 -translate-y-2"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition-all duration-150 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 -translate-y-2"
      >
        <div v-if="widgetsExpanded" class="p-4 space-y-4 bg-white dark:bg-slate-900/30">
          <!-- 小部件位置 -->
          <div class="mb-4">
            <h5 class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-3">小部件位置</h5>
            <div class="grid grid-cols-3 gap-2">
              <button
                v-for="option in positionOptions"
                :key="option.value"
                class="p-3 border-2 bg-white dark:bg-slate-700/50 transition-all duration-200 cursor-pointer hover:shadow-md rounded-xl"
                :class="widgets.position === option.value
                  ? 'border-primary-500 shadow-md shadow-primary-500/10'
                  : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600'"
                @click="updateWidgetPosition(option.value)"
              >
                <p class="text-sm font-medium" :class="widgets.position === option.value ? 'text-primary-700 dark:text-primary-300' : 'text-slate-600 dark:text-slate-400'">
                  {{ option.label }}
                </p>
                <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
                  {{ option.description }}
                </p>
              </button>
            </div>
          </div>

          <!-- 小部件开关列表 -->
          <div class="space-y-2">
            <h5 class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-3">显示小部件</h5>
            <div
              v-for="item in widgetItems"
              :key="item.key"
              class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
              <div class="flex items-center gap-3">
                <div class="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
                  <component :is="item.icon" :size="16" class="text-blue-600 dark:text-blue-400" />
                </div>
                <div>
                  <h5 class="text-sm font-medium text-slate-800 dark:text-slate-200">{{ item.label }}</h5>
                  <p class="text-xs text-slate-500 dark:text-slate-400">{{ item.description }}</p>
                </div>
              </div>
              <button
                class="relative w-12 h-6 rounded-full transition-all duration-300 cursor-pointer"
                :class="widgets[item.key] ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
                @click="toggleWidget(item.key)"
              >
                <span
                  class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
                  :class="widgets[item.key] ? 'translate-x-6' : ''"
                />
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </section>

    <!-- 版权设置 -->
    <section class="border-2 border-slate-200 dark:border-slate-700 rounded-2xl overflow-hidden">
      <button
        class="w-full flex items-center justify-between p-4 bg-slate-50 dark:bg-slate-800/50 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
        @click="copyrightExpanded = !copyrightExpanded"
      >
        <div class="flex items-center gap-3">
          <div class="p-2 bg-gradient-to-br from-amber-500 to-orange-600 rounded-lg">
            <Building2 :size="16" class="text-white" />
          </div>
          <span class="text-sm font-semibold text-slate-700 dark:text-slate-300">版权信息</span>
        </div>
        <component :is="copyrightExpanded ? ChevronUp : ChevronDown" :size="18" class="text-slate-500" />
      </button>

      <Transition
        enter-active-class="transition-all duration-200 ease-out"
        enter-from-class="opacity-0 -translate-y-2"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition-all duration-150 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 -translate-y-2"
      >
        <div v-if="copyrightExpanded" class="p-4 space-y-4 bg-white dark:bg-slate-900/30">
          <!-- 启用版权 -->
          <div class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-amber-100 dark:bg-amber-900/30 rounded-lg">
                <Building2 :size="16" class="text-amber-600 dark:text-amber-400" />
              </div>
              <div>
                <h5 class="text-sm font-medium text-slate-800 dark:text-slate-200">显示版权信息</h5>
                <p class="text-xs text-slate-500 dark:text-slate-400">在页脚显示版权声明</p>
              </div>
            </div>
            <button
              class="relative w-12 h-6 rounded-full transition-all duration-300 cursor-pointer"
              :class="copyright.enable ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
              @click="toggleCopyrightEnable"
            >
              <span
                class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
                :class="copyright.enable ? 'translate-x-6' : ''"
              />
            </button>
          </div>

          <!-- 版权信息表单 -->
          <div v-if="copyright.enable" class="space-y-3">
            <!-- 公司名称 -->
            <div>
              <label class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-2 block">公司名称</label>
              <input
                type="text"
                :value="copyright.companyName"
                placeholder="输入公司名称"
                class="w-full px-4 py-2.5 bg-white dark:bg-slate-800 border-2 border-slate-200 dark:border-slate-700 rounded-xl focus:outline-none focus:border-primary-500 dark:focus:border-primary-400 transition-colors text-sm"
                @input="(e) => updateCopyrightField('companyName', (e.target as HTMLInputElement).value)"
              />
            </div>

            <!-- 公司网站链接 -->
            <div>
              <label class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-2 block">公司网站</label>
              <input
                type="url"
                :value="copyright.companySiteLink"
                placeholder="https://example.com"
                class="w-full px-4 py-2.5 bg-white dark:bg-slate-800 border-2 border-slate-200 dark:border-slate-700 rounded-xl focus:outline-none focus:border-primary-500 dark:focus:border-primary-400 transition-colors text-sm"
                @input="(e) => updateCopyrightField('companySiteLink', (e.target as HTMLInputElement).value)"
              />
            </div>

            <!-- 版权年份 -->
            <div>
              <label class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-2 block">版权年份</label>
              <input
                type="text"
                :value="copyright.date"
                placeholder="2024 或 2020-2024"
                class="w-full px-4 py-2.5 bg-white dark:bg-slate-800 border-2 border-slate-200 dark:border-slate-700 rounded-xl focus:outline-none focus:border-primary-500 dark:focus:border-primary-400 transition-colors text-sm"
                @input="(e) => updateCopyrightField('date', (e.target as HTMLInputElement).value)"
              />
            </div>

            <!-- ICP 备案号 -->
            <div>
              <label class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-2 block">ICP 备案号</label>
              <input
                type="text"
                :value="copyright.icp"
                placeholder="京ICP备XXXXXXXX号"
                class="w-full px-4 py-2.5 bg-white dark:bg-slate-800 border-2 border-slate-200 dark:border-slate-700 rounded-xl focus:outline-none focus:border-primary-500 dark:focus:border-primary-400 transition-colors text-sm"
                @input="(e) => updateCopyrightField('icp', (e.target as HTMLInputElement).value)"
              />
            </div>

            <!-- ICP 链接 -->
            <div>
              <label class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-2 block">ICP 查询链接</label>
              <input
                type="url"
                :value="copyright.icpLink"
                placeholder="https://beian.miit.gov.cn"
                class="w-full px-4 py-2.5 bg-white dark:bg-slate-800 border-2 border-slate-200 dark:border-slate-700 rounded-xl focus:outline-none focus:border-primary-500 dark:focus:border-primary-400 transition-colors text-sm"
                @input="(e) => updateCopyrightField('icpLink', (e.target as HTMLInputElement).value)"
              />
            </div>

            <!-- 在设置中显示 -->
            <div class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-800/50 rounded-xl">
              <div class="flex items-center gap-3">
                <div class="p-2 bg-violet-100 dark:bg-violet-900/30 rounded-lg">
                  <Settings :size="16" class="text-violet-600 dark:text-violet-400" />
                </div>
                <div>
                  <h5 class="text-sm font-medium text-slate-800 dark:text-slate-200">在设置中显示</h5>
                  <p class="text-xs text-slate-500 dark:text-slate-400">允许在系统设置中修改版权信息</p>
                </div>
              </div>
              <button
                class="relative w-12 h-6 rounded-full transition-all duration-300 cursor-pointer"
                :class="copyright.settingShow ? 'bg-primary-500' : 'bg-slate-300 dark:bg-slate-600'"
                @click="() => updateCopyrightField('settingShow', !copyright.settingShow)"
              >
                <span
                  class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-md transition-transform duration-300"
                  :class="copyright.settingShow ? 'translate-x-6' : ''"
                />
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </section>

    <!-- 预览区域 -->
    <section class="p-5 bg-gradient-to-br from-violet-50 to-purple-100 dark:from-violet-900/20 dark:to-purple-900/10 rounded-2xl border border-violet-200 dark:border-violet-800">
      <div class="flex items-center gap-3 mb-4">
        <div class="p-2.5 bg-gradient-to-br from-violet-600 to-purple-700 rounded-xl">
          <Settings :size="20" class="text-white" />
        </div>
        <div>
          <h4 class="text-sm font-semibold text-slate-800 dark:text-slate-200">设置预览</h4>
          <p class="text-xs text-slate-600 dark:text-slate-400">实时预览高级设置效果</p>
        </div>
      </div>

      <!-- 版权信息预览 -->
      <div v-if="copyright.enable" class="p-4 bg-white dark:bg-slate-800/50 rounded-xl border border-slate-200 dark:border-slate-700">
        <p class="text-sm text-slate-600 dark:text-slate-400 text-center">
          &copy; {{ copyright.date || '2024' }} {{ copyright.companyName || '公司名称' }}
        </p>
        <p v-if="copyright.icp" class="text-xs text-slate-500 dark:text-slate-500 text-center mt-1">
          <a
            v-if="copyright.icpLink"
            :href="copyright.icpLink"
            target="_blank"
            rel="noopener"
            class="hover:text-primary-500 dark:hover:text-primary-400 transition-colors"
          >
            {{ copyright.icp }}
          </a>
          <span v-else>{{ copyright.icp }}</span>
        </p>
      </div>

      <!-- 小部件预览提示 -->
      <div v-else class="p-4 bg-white dark:bg-slate-800/50 rounded-xl border border-slate-200 dark:border-slate-700">
        <p class="text-sm text-slate-500 dark:text-slate-400 text-center">
          版权信息已禁用，启用后将在此处显示预览
        </p>
      </div>
    </section>
  </div>
</template>
