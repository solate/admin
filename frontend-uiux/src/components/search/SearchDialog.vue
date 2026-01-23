<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  Search,
  X,
  CornerDownLeft,
  ArrowUp,
  ArrowDown,
  LayoutDashboard,
  Building,
  Grid,
  Users,
  Briefcase,
  BarChart3,
  Settings,
  Bell,
  User,
  FileText
} from 'lucide-vue-next'
import { dashboardRoutes } from '@/router/routes/dashboard'

// 图标映射
const iconMap = {
  Dashboard: LayoutDashboard,
  Building,
  Grid,
  Users,
  Briefcase,
  BarChart3,
  Settings,
  Bell,
  User,
  FileText
}

// Props
const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  }
})

// Emits
const emit = defineEmits(['update:visible', 'close'])

// Router
const router = useRouter()

// State
const searchQuery = ref('')
const selectedIndex = ref(0)
const searchInput = ref(null)
const searchHistory = ref(JSON.parse(localStorage.getItem('searchHistory') || '[]'))

// Mock 菜单数据（包含中文标题）
const mockMenuItems = [
  { name: 'overview', path: 'overview', title: '概览', icon: 'Dashboard' },
  { name: 'tenants', path: 'tenants', title: '租户管理', icon: 'Building' },
  { name: 'services', path: 'services', title: '服务管理', icon: 'Grid' },
  { name: 'users', path: 'users', title: '用户管理', icon: 'Users' },
  { name: 'business', path: 'business', title: '业务管理', icon: 'Briefcase' },
  { name: 'analytics', path: 'analytics', title: '数据分析', icon: 'BarChart3' },
  { name: 'settings', path: 'settings', title: '系统设置', icon: 'Settings' },
  { name: 'profile', path: 'profile', title: '个人中心', icon: 'User' },
  { name: 'notifications', path: 'notifications', title: '通知中心', icon: 'Bell' }
]

// 获取可搜索的菜单项
const menuItems = computed(() => {
  return mockMenuItems
})

// 搜索结果
const searchResults = computed(() => {
  if (!searchQuery.value.trim()) {
    // 显示搜索历史
    return searchHistory.value
      .map(path => menuItems.value.find(item => item.path === path))
      .filter(Boolean)
  }

  const query = searchQuery.value.toLowerCase()
  return menuItems.value.filter(item => {
    return (
      item.title.toLowerCase().includes(query) ||
      item.name.toLowerCase().includes(query)
    )
  })
})

// 是否有搜索结果
const hasResults = computed(() => searchResults.value.length > 0)

// 更新选中索引
watch(searchResults, () => {
  selectedIndex.value = 0
})

// 监听 visible 变化，自动聚焦输入框
watch(() => props.visible, (visible) => {
  if (visible) {
    nextTick(() => {
      searchInput.value?.focus()
    })
  } else {
    searchQuery.value = ''
    selectedIndex.value = 0
  }
})

// 添加到搜索历史
function addToHistory(path) {
  const history = searchHistory.value.filter(p => p !== path)
  history.unshift(path)
  if (history.length > 5) history.pop()
  searchHistory.value = history
  localStorage.setItem('searchHistory', JSON.stringify(history))
}

// 导航到选中项
function navigateTo(item) {
  if (!item) return
  addToHistory(item.path)
  emit('update:visible', false)
  router.push(`/dashboard/${item.path}`)
}

// 键盘事件处理
function handleKeydown(e) {
  switch (e.key) {
    case 'ArrowDown':
      e.preventDefault()
      selectedIndex.value = Math.min(selectedIndex.value + 1, searchResults.value.length - 1)
      break
    case 'ArrowUp':
      e.preventDefault()
      selectedIndex.value = Math.max(selectedIndex.value - 1, 0)
      break
    case 'Enter':
      e.preventDefault()
      if (searchResults.value[selectedIndex.value]) {
        navigateTo(searchResults.value[selectedIndex.value])
      }
      break
    case 'Escape':
      e.preventDefault()
      emit('update:visible', false)
      break
  }
}

// 移除历史记录项
function removeFromHistory(e, path) {
  e.stopPropagation()
  searchHistory.value = searchHistory.value.filter(p => p !== path)
  localStorage.setItem('searchHistory', JSON.stringify(searchHistory.value))
}

// 获取图标组件
function getIconComponent(iconName) {
  return iconMap[iconName] || FileText
}

// 键盘快捷键监听
function handleGlobalKeydown(e) {
  // ⌘K 或 Ctrl+K 打开搜索
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    emit('update:visible', !props.visible)
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleGlobalKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleGlobalKeydown)
})
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-all duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-all duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="visible"
        class="fixed inset-0 z-[1000] flex items-start justify-center pt-[12vh] bg-black/40 backdrop-blur-sm"
        @click.self="emit('update:visible', false)"
      >
        <Transition
          enter-active-class="transition-all duration-200 ease-out"
          enter-from-class="opacity-0 scale-95 translate-y-[-20px]"
          enter-to-class="opacity-100 scale-100 translate-y-0"
          leave-active-class="transition-all duration-150 ease-in"
          leave-from-class="opacity-100 scale-100 translate-y-0"
          leave-to-class="opacity-0 scale-95 translate-y-[-20px]"
        >
          <div
            v-if="visible"
            class="relative w-full max-w-[560px] mx-4 bg-white/95 dark:bg-slate-900/95 backdrop-blur-xl rounded-2xl shadow-2xl overflow-hidden ring-1 ring-slate-900/10 dark:ring-white/10 outline-none focus:outline-none"
            @click.stop
          >
            <!-- Header -->
            <div class="flex items-center gap-2 px-5 py-4">
              <Search :size="20" class="text-slate-400 flex-shrink-0" />
              <input
                ref="searchInput"
                v-model="searchQuery"
                type="text"
                placeholder="搜索菜单、页面..."
                class="flex-1 bg-transparent outline-none focus:outline-none focus:shadow-none text-base text-slate-900 dark:text-slate-100 placeholder-slate-400"
                style="box-shadow: none !important;"
                @keydown="handleKeydown"
              />
            </div>

            <!-- 分隔线 -->
            <div class="h-px bg-slate-100 dark:bg-slate-800 mx-5" />

            <!-- Results -->
            <div class="max-h-[360px] overflow-y-auto p-2">
              <!-- 搜索历史标题 -->
              <p
                v-if="!searchQuery && searchHistory.length > 0"
                class="px-3 mb-2 text-xs font-medium text-slate-400 dark:text-slate-500 uppercase tracking-wide"
              >
                最近搜索
              </p>

              <!-- 空状态 -->
              <div
                v-if="!hasResults"
                class="flex flex-col items-center justify-center py-10 text-slate-400 dark:text-slate-500"
              >
                <div class="w-12 h-12 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center mb-3">
                  <Search :size="24" class="opacity-50" />
                </div>
                <p class="text-sm">没有找到相关结果</p>
                <p class="text-xs mt-1 text-slate-300 dark:text-slate-600">试试搜索"租户"或"设置"</p>
              </div>

              <!-- 结果列表 -->
              <ul v-else class="space-y-0.5">
                <li
                  v-for="(item, index) in searchResults"
                  :key="item.path"
                  :class="[
                    'flex items-center gap-3 px-3 py-2.5 rounded-xl cursor-pointer transition-all duration-150 group',
                    selectedIndex === index
                      ? 'bg-primary-500 text-white shadow-md shadow-primary-500/25'
                      : 'hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-200'
                  ]"
                  @click="navigateTo(item)"
                  @mouseenter="selectedIndex = index"
                >
                  <!-- 图标 -->
                  <component
                    :is="getIconComponent(item.icon)"
                    :size="18"
                    :class="selectedIndex === index ? 'text-white' : 'text-slate-400'"
                  />

                  <!-- 标题 -->
                  <span class="flex-1 text-sm font-medium">{{ item.title }}</span>

                  <!-- 移除历史记录按钮 -->
                  <button
                    v-if="!searchQuery && searchHistory.includes(item.path)"
                    class="flex-shrink-0 p-1 hover:bg-slate-200 dark:hover:bg-slate-700 rounded-lg transition-all"
                    @click="removeFromHistory($event, item.path)"
                  >
                    <X :size="14" />
                  </button>
                </li>
              </ul>
            </div>

            <!-- Footer -->
            <div class="flex items-center justify-between px-4 py-2.5 bg-slate-50/50 dark:bg-slate-800/50">
              <div class="flex items-center gap-4 text-xs text-slate-400 dark:text-slate-500">
                <div class="flex items-center gap-1.5">
                  <span class="flex items-center gap-0.5 px-1.5 py-0.5 rounded bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-[10px] font-medium">
                    <CornerDownLeft :size="10" />
                  </span>
                  <span>选择</span>
                </div>
                <div class="flex items-center gap-1.5">
                  <span class="flex items-center gap-0.5 px-1.5 py-0.5 rounded bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-[10px] font-medium">
                    <ArrowUp :size="10" />
                    <ArrowDown :size="10" />
                  </span>
                  <span>导航</span>
                </div>
                <div class="flex items-center gap-1.5">
                  <span class="px-1.5 py-0.5 rounded bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-[10px] font-medium">Esc</span>
                  <span>关闭</span>
                </div>
              </div>
            </div>

            <!-- 关闭按钮 -->
            <button
              class="absolute top-3 right-3 p-1.5 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
              @click="emit('update:visible', false)"
            >
              <X :size="16" class="text-slate-400" />
            </button>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>
