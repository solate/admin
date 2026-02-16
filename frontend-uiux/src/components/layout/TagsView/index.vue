<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTagsViewStore } from '@/stores/modules/tagsView'
import { usePreferencesStore } from '@/stores/modules/preferences'
import { useI18n } from '@/locales/composables'
import {
  X,
  RefreshCw,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Pin,
  Maximize2,
  Minimize2,
  ExternalLink,
  ArrowLeftToLine,
  ArrowRightToLine,
  FoldHorizontal,
  ArrowRightLeft,
  GripVertical
} from 'lucide-vue-next'
import type { TagView } from '@/types/tagsView'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const tagsViewStore = useTagsViewStore()
const preferencesStore = usePreferencesStore()

// 拖拽状态
const dragIndex = ref<number | null>(null)
const dragOverIndex = ref<number | null>(null)

// 右键菜单状态
const contextMenuVisible = ref(false)
const contextMenuPosition = ref({ x: 0, y: 0 })
const contextMenuTag = ref<TagView | null>(null)

// 下拉菜单状态
const dropdownVisible = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)
const dropdownPosition = ref({ top: 0, left: 0 })

// 滚动容器引用
const scrollContainer = ref<HTMLElement | null>(null)

// 滚动状态
const showScrollLeft = ref(false)
const showScrollRight = ref(false)

// 布局偏好
const layoutPrefs = computed(() => preferencesStore.layout)

// 标签列表
const visitedTags = computed(() => tagsViewStore.visitedTags)

// 当前激活标签
const activeTagPath = computed(() => tagsViewStore.activeTagPath)

// 最大化状态（从 store 获取，用于菜单显示）
const isMaximized = computed(() => tagsViewStore.isMaximized)

// 翻译标签标题
function getTagTitle(tag: TagView): string {
  if (tag.title && tag.title.includes('.')) {
    const translated = t(tag.title)
    if (translated !== tag.title) {
      return translated
    }
  }
  const navKey = tag.name?.replace('-', '.') || ''
  const navTranslated = t(`nav.${navKey}`)
  if (navTranslated && !navTranslated.includes(navKey)) {
    return navTranslated
  }
  return tag.title
}

// 监听路由变化，添加标签
watch(
  () => route.path,
  () => {
    if (route.path && !route.meta?.hideTag) {
      tagsViewStore.addTag(route)
    }
  },
  { immediate: true }
)

// 点击标签
function handleClick(tag: TagView) {
  if (tag.path !== route.path) {
    router.push(tag.fullPath || tag.path)
  }
}

// 双击关闭标签（非固定标签）
function handleDoubleClick(tag: TagView) {
  if (!tag.affix) {
    handleClose(tag)
  }
}

// 关闭标签
function handleClose(tag: TagView, e?: Event) {
  e?.stopPropagation()
  if (tag.affix) return

  const nextTag = tagsViewStore.removeTag(tag.path)
  if (nextTag) {
    router.push(nextTag.path)
  }
}

// 右键菜单
function handleContextMenu(e: MouseEvent, tag: TagView) {
  e.preventDefault()
  e.stopPropagation()
  contextMenuTag.value = tag

  // 计算菜单位置，确保不超出视口
  const menuWidth = 180
  const menuHeight = 350 // 估算的菜单高度
  let x = e.clientX
  let y = e.clientY

  // 右边界检测
  if (x + menuWidth > window.innerWidth) {
    x = window.innerWidth - menuWidth - 8
  }

  // 下边界检测
  if (y + menuHeight > window.innerHeight) {
    y = window.innerHeight - menuHeight - 8
  }

  contextMenuPosition.value = { x, y }
  contextMenuVisible.value = true
}

// 关闭右键菜单
function closeContextMenu() {
  contextMenuVisible.value = false
  contextMenuTag.value = null
}

// 打开下拉菜单
function openDropdown() {
  if (dropdownRef.value) {
    const rect = dropdownRef.value.getBoundingClientRect()
    const menuWidth = 180
    const menuHeight = 280 // 估算的菜单高度

    let left = rect.right - menuWidth
    let top = rect.bottom + 8

    // 右边界检测
    if (left + menuWidth > window.innerWidth) {
      left = window.innerWidth - menuWidth - 8
    }

    // 左边界检测
    if (left < 8) {
      left = 8
    }

    // 下边界检测
    if (top + menuHeight > window.innerHeight) {
      top = rect.top - menuHeight - 8
    }

    dropdownPosition.value = { top, left }
  }
  dropdownVisible.value = true
}

// 关闭下拉菜单
function closeDropdown() {
  dropdownVisible.value = false
}

// 刷新标签
function handleRefresh(tag: TagView) {
  const componentName = tagsViewStore.refreshTag(tag.path)
  if (componentName) {
    nextTick(() => {
      router.replace({ path: '/redirect' + tag.path })
    })
  }
  closeContextMenu()
  closeDropdown()
}

// 固定/取消固定标签
function handleTogglePin(tag: TagView) {
  tagsViewStore.toggleAffix(tag.path)
  closeContextMenu()
  closeDropdown()
}

// 最大化标签
function handleMaximize(tag: TagView) {
  tagsViewStore.toggleMaximized()
  closeContextMenu()
  closeDropdown()
}

// 在新窗口打开
function handleOpenInNewWindow(tag: TagView) {
  const url = window.location.origin + tag.fullPath
  window.open(url, '_blank')
  closeContextMenu()
  closeDropdown()
}

// 关闭左侧标签
function handleCloseLeft(tag: TagView) {
  const nextTag = tagsViewStore.closeLeftTags(tag.path)
  if (nextTag) {
    router.push(nextTag.path)
  }
  closeContextMenu()
  closeDropdown()
}

// 关闭右侧标签
function handleCloseRight(tag: TagView) {
  const nextTag = tagsViewStore.closeRightTags(tag.path)
  if (nextTag) {
    router.push(nextTag.path)
  }
  closeContextMenu()
  closeDropdown()
}

// 关闭其他标签
function handleCloseOther(tag: TagView) {
  tagsViewStore.closeOtherTags(tag.path)
  if (tag.path !== route.path) {
    router.push(tag.path)
  }
  closeContextMenu()
  closeDropdown()
}

// 关闭所有标签
function handleCloseAll() {
  const nextTag = tagsViewStore.closeAllTags()
  if (nextTag) {
    router.push(nextTag.path)
  }
  closeContextMenu()
  closeDropdown()
}

// 切换最大化
function toggleMaximize() {
  tagsViewStore.toggleMaximized()
}

// 拖拽开始
function handleDragStart(index: number) {
  if (!layoutPrefs.value.tabsDraggable) return
  dragIndex.value = index
}

// 拖拽经过
function handleDragOver(e: DragEvent, index: number) {
  if (!layoutPrefs.value.tabsDraggable) return
  e.preventDefault()
  dragOverIndex.value = index
}

// 拖拽结束
function handleDrop(e: DragEvent, index: number) {
  if (!layoutPrefs.value.tabsDraggable) return
  e.preventDefault()
  if (dragIndex.value !== null && dragIndex.value !== index) {
    tagsViewStore.updateTagOrder(dragIndex.value, index)
  }
  dragIndex.value = null
  dragOverIndex.value = null
}

// 拖拽离开
function handleDragEnd() {
  dragIndex.value = null
  dragOverIndex.value = null
}

// 滚动到激活标签
function scrollToActiveTag() {
  nextTick(() => {
    if (!scrollContainer.value) return
    const activeTag = scrollContainer.value.querySelector('.tag-active')
    if (activeTag) {
      const containerRect = scrollContainer.value.getBoundingClientRect()
      const tagRect = activeTag.getBoundingClientRect()
      if (tagRect.left < containerRect.left) {
        scrollContainer.value.scrollLeft -= containerRect.left - tagRect.left + 50
      } else if (tagRect.right > containerRect.right) {
        scrollContainer.value.scrollLeft += tagRect.right - containerRect.right + 50
      }
    }
  })
}

// 监听激活标签变化，自动滚动
watch(activeTagPath, scrollToActiveTag)

// 监听标签列表变化，重新检测滚动按钮
watch(visitedTags, () => {
  nextTick(checkScrollButtons)
}, { deep: true })

// 检测滚动状态
function checkScrollButtons() {
  if (!scrollContainer.value) return
  const { scrollLeft, scrollWidth, clientWidth } = scrollContainer.value
  showScrollLeft.value = scrollLeft > 0
  showScrollRight.value = scrollLeft + clientWidth < scrollWidth - 1
}

// 向左滚动
function scrollToLeft() {
  if (!scrollContainer.value) return
  scrollContainer.value.scrollBy({ left: -200, behavior: 'smooth' })
}

// 向右滚动
function scrollToRight() {
  if (!scrollContainer.value) return
  scrollContainer.value.scrollBy({ left: 200, behavior: 'smooth' })
}

// 点击外部关闭菜单（使用 mousedown 事件）
function handleMouseDownOutside(e: MouseEvent) {
  const target = e.target as HTMLElement
  // 检查是否点击在菜单内部
  if (target.closest('.tags-context-menu')) {
    return
  }
  // 检查是否点击在下拉按钮上（点击按钮会触发 openDropdown）
  if (target.closest('.tags-view__action-btn')) {
    return
  }
  if (contextMenuVisible.value) {
    closeContextMenu()
  }
  if (dropdownVisible.value) {
    closeDropdown()
  }
}

// 获取当前标签索引
const currentTagIndex = computed(() => {
  if (!contextMenuTag.value) return -1
  return visitedTags.value.findIndex(t => t.path === contextMenuTag.value?.path)
})

// 是否可以关闭左侧
const canCloseLeft = computed(() => {
  if (currentTagIndex.value <= 0) return false
  // 检查左侧是否有可关闭的标签（非固定）
  for (let i = 0; i < currentTagIndex.value; i++) {
    if (!visitedTags.value[i].affix) return true
  }
  return false
})

// 是否可以关闭右侧
const canCloseRight = computed(() => {
  if (currentTagIndex.value < 0 || currentTagIndex.value >= visitedTags.value.length - 1) return false
  // 检查右侧是否有可关闭的标签（非固定）
  for (let i = currentTagIndex.value + 1; i < visitedTags.value.length; i++) {
    if (!visitedTags.value[i].affix) return true
  }
  return false
})

// 是否可以关闭其他
const canCloseOther = computed(() => {
  if (currentTagIndex.value < 0) return false
  // 检查是否有其他可关闭的标签
  return visitedTags.value.some((t, i) => i !== currentTagIndex.value && !t.affix)
})

// 当前激活标签索引
const activeTagIndex = computed(() => {
  return visitedTags.value.findIndex(t => t.path === activeTagPath.value)
})

// 从当前激活标签计算是否可以关闭左侧（用于下拉菜单）
const canCloseLeftFromActive = computed(() => {
  if (activeTagIndex.value <= 0) return false
  for (let i = 0; i < activeTagIndex.value; i++) {
    if (!visitedTags.value[i].affix) return true
  }
  return false
})

// 从当前激活标签计算是否可以关闭右侧（用于下拉菜单）
const canCloseRightFromActive = computed(() => {
  if (activeTagIndex.value < 0 || activeTagIndex.value >= visitedTags.value.length - 1) return false
  for (let i = activeTagIndex.value + 1; i < visitedTags.value.length; i++) {
    if (!visitedTags.value[i].affix) return true
  }
  return false
})

// 从当前激活标签计算是否可以关闭其他（用于下拉菜单）
const canCloseOtherFromActive = computed(() => {
  if (activeTagIndex.value < 0) return false
  return visitedTags.value.some((t, i) => i !== activeTagIndex.value && !t.affix)
})

onMounted(() => {
  scrollToActiveTag()
  checkScrollButtons()
  document.addEventListener('mousedown', handleMouseDownOutside)
  if (scrollContainer.value) {
    scrollContainer.value.addEventListener('scroll', checkScrollButtons)
  }
  // 监听窗口大小变化
  window.addEventListener('resize', checkScrollButtons)
})

onUnmounted(() => {
  document.removeEventListener('mousedown', handleMouseDownOutside)
  if (scrollContainer.value) {
    scrollContainer.value.removeEventListener('scroll', checkScrollButtons)
  }
  window.removeEventListener('resize', checkScrollButtons)
})

// 获取标签样式类
const tagsViewClass = computed(() => {
  return [
    'tags-view',
    `tags-view--${layoutPrefs.value.tabsStyle}`
  ]
})
</script>

<template>
  <div :class="tagsViewClass">
    <!-- 背景装饰 -->
    <div class="tags-view__bg"></div>

    <!-- 左侧滚动指示器 -->
    <Transition
      enter-active-class="transition-opacity duration-200"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-200"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <button
        v-if="showScrollLeft"
        class="tags-view__scroll-btn tags-view__scroll-btn--left"
        @click="scrollToLeft"
      >
        <ChevronLeft :size="16" />
      </button>
    </Transition>

    <!-- 标签容器 -->
    <div
      ref="scrollContainer"
      class="tags-view__scroll"
    >
      <div
        v-for="(tag, index) in visitedTags"
        :key="tag.path"
        :class="[
          'tags-view__tag',
          'group',
          { 'tag-active': activeTagPath === tag.path },
          { 'tag-affix': tag.affix },
          { 'dragging': dragIndex === index },
          { 'drag-over': dragOverIndex === index && dragIndex !== index }
        ]"
        :draggable="layoutPrefs.tabsDraggable && !tag.affix"
        @click="handleClick(tag)"
        @dblclick="handleDoubleClick(tag)"
        @contextmenu="handleContextMenu($event, tag)"
        @dragstart="handleDragStart(index)"
        @dragover="handleDragOver($event, index)"
        @drop="handleDrop($event, index)"
        @dragend="handleDragEnd"
      >
        <!-- 拖拽手柄 -->
        <div
          v-if="layoutPrefs.tabsDraggable && !tag.affix"
          class="tags-view__tag-drag"
        >
          <GripVertical :size="12" class="opacity-0 group-hover:opacity-50 transition-opacity" />
        </div>

        <!-- 激活指示器 -->
        <div class="tags-view__tag-indicator"></div>

        <!-- 标签图标 -->
        <span v-if="tag.icon" class="tags-view__tag-icon">
          <component :is="tag.icon" :size="14" />
        </span>

        <!-- 标签标题 -->
        <span class="tags-view__tag-title">{{ getTagTitle(tag) }}</span>

        <!-- 固定图标 -->
        <span v-if="tag.affix" class="tags-view__tag-pin">
          <Pin :size="10" />
        </span>

        <!-- 关闭按钮 -->
        <button
          v-if="!tag.affix"
          class="tags-view__tag-close"
          @click.stop="handleClose(tag, $event)"
        >
          <X :size="12" />
        </button>
      </div>
    </div>

    <!-- 右侧滚动指示器 -->
    <Transition
      enter-active-class="transition-opacity duration-200"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-200"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <button
        v-if="showScrollRight"
        class="tags-view__scroll-btn tags-view__scroll-btn--right"
        @click="scrollToRight"
      >
        <ChevronRight :size="16" />
      </button>
    </Transition>

    <!-- 右侧操作区 -->
    <div class="tags-view__actions">
      <!-- 刷新按钮 -->
      <button
        class="tags-view__action-btn"
        :title="t('tagsView.refresh')"
        @click="handleRefresh({ path: activeTagPath } as TagView)"
      >
        <RefreshCw :size="15" />
      </button>

      <!-- 更多操作下拉按钮 -->
      <div ref="dropdownRef" class="relative">
        <button
          class="tags-view__action-btn"
          :class="{ 'tags-view__action-btn--active': dropdownVisible }"
          @click="openDropdown"
        >
          <ChevronDown :size="15" :class="{ 'rotate-180': dropdownVisible }" class="transition-transform duration-200" />
        </button>
      </div>

      <!-- 最大化按钮 -->
      <button
        class="tags-view__action-btn"
        :title="isMaximized ? t('tagsView.restore') : t('tagsView.maximize')"
        @click="toggleMaximize"
      >
        <Minimize2 v-if="isMaximized" :size="15" />
        <Maximize2 v-else :size="15" />
      </button>
    </div>
  </div>

  <!-- 右键菜单 - Teleport 到 body -->
  <Teleport to="body">
    <Transition
      enter-active-class="transition-all duration-150 ease-out"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition-all duration-100 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-if="contextMenuVisible && contextMenuTag"
        class="tags-context-menu"
        :style="{ left: contextMenuPosition.x + 'px', top: contextMenuPosition.y + 'px' }"
        @click.stop
      >
        <!-- 关闭 -->
        <button
          class="tags-context-menu__item"
          :class="{ 'tags-context-menu__item--disabled': contextMenuTag.affix }"
          :disabled="contextMenuTag.affix"
          @click="handleClose(contextMenuTag)"
        >
          <X :size="16" />
          <span>{{ t('tagsView.close') }}</span>
        </button>

        <!-- 固定/取消固定 -->
        <button
          class="tags-context-menu__item"
          @click="handleTogglePin(contextMenuTag)"
        >
          <Pin :size="16" />
          <span>{{ contextMenuTag.affix ? t('tagsView.unpin') : t('tagsView.pin') }}</span>
        </button>

        <!-- 最大化/还原 -->
        <button
          class="tags-context-menu__item"
          @click="handleMaximize(contextMenuTag)"
        >
          <Minimize2 v-if="isMaximized" :size="16" />
          <Maximize2 v-else :size="16" />
          <span>{{ isMaximized ? t('tagsView.restore') : t('tagsView.maximize') }}</span>
        </button>

        <!-- 重新加载 -->
        <button
          class="tags-context-menu__item"
          @click="handleRefresh(contextMenuTag)"
        >
          <RefreshCw :size="16" />
          <span>{{ t('tagsView.refresh') }}</span>
        </button>

        <!-- 在新窗口打开 -->
        <button
          class="tags-context-menu__item"
          @click="handleOpenInNewWindow(contextMenuTag)"
        >
          <ExternalLink :size="16" />
          <span>{{ t('tagsView.openInNewWindow') }}</span>
        </button>

        <div class="tags-context-menu__divider"></div>

        <!-- 关闭左侧标签页 -->
        <button
          class="tags-context-menu__item"
          :class="{ 'tags-context-menu__item--disabled': !canCloseLeft }"
          :disabled="!canCloseLeft"
          @click="handleCloseLeft(contextMenuTag)"
        >
          <ArrowLeftToLine :size="16" />
          <span>{{ t('tagsView.closeLeft') }}</span>
        </button>

        <!-- 关闭右侧标签页 -->
        <button
          class="tags-context-menu__item"
          :class="{ 'tags-context-menu__item--disabled': !canCloseRight }"
          :disabled="!canCloseRight"
          @click="handleCloseRight(contextMenuTag)"
        >
          <ArrowRightToLine :size="16" />
          <span>{{ t('tagsView.closeRight') }}</span>
        </button>

        <div class="tags-context-menu__divider"></div>

        <!-- 关闭其它标签页 -->
        <button
          class="tags-context-menu__item"
          :class="{ 'tags-context-menu__item--disabled': !canCloseOther }"
          :disabled="!canCloseOther"
          @click="handleCloseOther(contextMenuTag)"
        >
          <FoldHorizontal :size="16" />
          <span>{{ t('tagsView.closeOther') }}</span>
        </button>

        <!-- 关闭全部标签页 -->
        <button
          class="tags-context-menu__item"
          :class="{ 'tags-context-menu__item--disabled': visitedTags.filter(t => !t.affix).length === 0 }"
          :disabled="visitedTags.filter(t => !t.affix).length === 0"
          @click="handleCloseAll"
        >
          <ArrowRightLeft :size="16" />
          <span>{{ t('tagsView.closeAll') }}</span>
        </button>
      </div>
    </Transition>
  </Teleport>

  <!-- 下拉菜单 - Teleport 到 body -->
  <Teleport to="body">
    <Transition
      enter-active-class="transition-all duration-150 ease-out"
      enter-from-class="opacity-0 scale-95 -translate-y-2"
      enter-to-class="opacity-100 scale-100 translate-y-0"
      leave-active-class="transition-all duration-100 ease-in"
      leave-from-class="opacity-100 scale-100 translate-y-0"
      leave-to-class="opacity-0 scale-95 -translate-y-2"
    >
      <div
        v-if="dropdownVisible"
        class="tags-context-menu"
        :style="{ left: dropdownPosition.left + 'px', top: dropdownPosition.top + 'px' }"
        @click.stop
      >
        <!-- 刷新 -->
        <button
          class="tags-context-menu__item"
          @click="handleRefresh({ path: activeTagPath } as TagView)"
        >
          <RefreshCw :size="16" />
          <span>{{ t('tagsView.refresh') }}</span>
        </button>

        <div class="tags-context-menu__divider"></div>

        <!-- 关闭左侧标签页 -->
        <button
          class="tags-context-menu__item"
          :class="{ 'tags-context-menu__item--disabled': !canCloseLeftFromActive }"
          :disabled="!canCloseLeftFromActive"
          @click="handleCloseLeft({ path: activeTagPath } as TagView)"
        >
          <ArrowLeftToLine :size="16" />
          <span>{{ t('tagsView.closeLeft') }}</span>
        </button>

        <!-- 关闭右侧标签页 -->
        <button
          class="tags-context-menu__item"
          :class="{ 'tags-context-menu__item--disabled': !canCloseRightFromActive }"
          :disabled="!canCloseRightFromActive"
          @click="handleCloseRight({ path: activeTagPath } as TagView)"
        >
          <ArrowRightToLine :size="16" />
          <span>{{ t('tagsView.closeRight') }}</span>
        </button>

        <div class="tags-context-menu__divider"></div>

        <!-- 关闭其它标签页 -->
        <button
          class="tags-context-menu__item"
          :class="{ 'tags-context-menu__item--disabled': !canCloseOtherFromActive }"
          :disabled="!canCloseOtherFromActive"
          @click="handleCloseOther({ path: activeTagPath } as TagView)"
        >
          <FoldHorizontal :size="16" />
          <span>{{ t('tagsView.closeOther') }}</span>
        </button>

        <!-- 关闭全部标签页 -->
        <button
          class="tags-context-menu__item"
          :class="{ 'tags-context-menu__item--disabled': visitedTags.filter(t => !t.affix).length === 0 }"
          :disabled="visitedTags.filter(t => !t.affix).length === 0"
          @click="handleCloseAll"
        >
          <ArrowRightLeft :size="16" />
          <span>{{ t('tagsView.closeAll') }}</span>
        </button>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
/* 基础容器样式 */
.tags-view {
  @apply relative flex items-center h-11;
  @apply bg-gradient-to-r from-slate-50 via-white to-slate-50;
  @apply dark:from-slate-900 dark:via-slate-900/95 dark:to-slate-900;
  @apply border-b border-slate-200/80 dark:border-slate-700/50;
  @apply backdrop-blur-xl;
}

.tags-view__bg {
  @apply absolute inset-0 opacity-50;
  background-image: radial-gradient(circle at 1px 1px, rgb(148 163 184 / 0.15) 1px, transparent 0);
  background-size: 24px 24px;
}

.dark .tags-view__bg {
  background-image: radial-gradient(circle at 1px 1px, rgb(100 116 139 / 0.1) 1px, transparent 0);
}

/* 滚动容器 */
.tags-view__scroll {
  @apply relative flex items-center gap-1.5 flex-1 overflow-x-auto px-3 py-1.5;
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.tags-view__scroll::-webkit-scrollbar {
  display: none;
}

/* 滚动按钮 */
.tags-view__scroll-btn {
  @apply absolute z-10 flex items-center justify-center w-7 h-7 rounded-full;
  @apply bg-white/90 dark:bg-slate-800/90;
  @apply border border-slate-200/80 dark:border-slate-700/50;
  @apply shadow-md;
  @apply text-slate-500 dark:text-slate-400;
  @apply hover:bg-slate-50 dark:hover:bg-slate-700;
  @apply hover:text-slate-700 dark:hover:text-slate-200;
  @apply transition-all duration-150 cursor-pointer;
  @apply backdrop-blur-sm;
}

.tags-view__scroll-btn--left {
  @apply left-1;
}

.tags-view__scroll-btn--right {
  @apply right-20;
}

/* 标签样式 */
.tags-view__tag {
  @apply relative flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm;
  @apply cursor-pointer select-none;
  @apply text-slate-600 dark:text-slate-400;
  @apply bg-white/60 dark:bg-slate-800/40;
  @apply border border-slate-200/60 dark:border-slate-700/40;
  @apply transition-all duration-200;
  @apply hover:bg-white dark:hover:bg-slate-800/60;
  @apply hover:border-slate-300 dark:hover:border-slate-600;
  @apply hover:shadow-sm;
}

.tags-view__tag-drag {
  @apply flex items-center justify-center -ml-1;
}

/* 激活指示器 */
.tags-view__tag-indicator {
  @apply absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-4 rounded-full;
  @apply bg-transparent transition-all duration-200;
}

.tags-view__tag.tag-active .tags-view__tag-indicator {
  @apply bg-primary-500;
}

.tags-view__tag-icon {
  @apply flex items-center justify-center text-slate-500 dark:text-slate-400;
}

.tags-view__tag.tag-active .tags-view__tag-icon {
  @apply text-primary-500 dark:text-primary-400;
}

.tags-view__tag-title {
  @apply max-w-[120px] truncate font-medium;
}

.tags-view__tag-pin {
  @apply flex items-center justify-center text-slate-400 dark:text-slate-500;
}

.tags-view__tag-close {
  @apply flex items-center justify-center w-5 h-5 rounded-md;
  @apply opacity-0 transition-all duration-150;
  @apply hover:bg-slate-200 dark:hover:bg-slate-700;
  @apply text-slate-400 hover:text-slate-600;
  @apply dark:hover:text-slate-300;
  cursor: pointer;
}

.tags-view__tag:hover .tags-view__tag-close {
  @apply opacity-100;
}

/* 激活状态 */
.tags-view__tag.tag-active {
  @apply bg-gradient-to-br from-primary-50 to-primary-100/50;
  @apply dark:from-primary-900/30 dark:to-primary-800/20;
  @apply border-primary-200 dark:border-primary-700/50;
  @apply text-primary-700 dark:text-primary-300;
  @apply shadow-sm shadow-primary-500/5;
}

.tags-view__tag.tag-active .tags-view__tag-close {
  @apply opacity-70 hover:bg-primary-200 dark:hover:bg-primary-800/50;
  @apply text-primary-500 dark:text-primary-400;
}

/* 拖拽状态 */
.tags-view__tag.dragging {
  @apply opacity-50 scale-95;
}

.tags-view__tag.drag-over {
  @apply border-primary-400 dark:border-primary-500;
  @apply bg-primary-50/50 dark:bg-primary-900/20;
}

/* 操作区 */
.tags-view__actions {
  @apply relative flex items-center gap-0.5 px-2 py-1.5;
  @apply border-l border-slate-200/60 dark:border-slate-700/40;
}

.tags-view__action-btn {
  @apply flex items-center justify-center w-8 h-8 rounded-lg;
  @apply text-slate-500 dark:text-slate-400;
  @apply hover:bg-slate-100 dark:hover:bg-slate-800;
  @apply hover:text-slate-700 dark:hover:text-slate-200;
  @apply transition-all duration-150 cursor-pointer;
}

.tags-view__action-btn--active {
  @apply bg-slate-100 dark:bg-slate-800 text-primary-600 dark:text-primary-400;
}

/* ========== 右键菜单样式 ========== */
.tags-context-menu {
  @apply fixed py-1.5 min-w-[180px];
  @apply bg-white dark:bg-slate-800 rounded-xl shadow-2xl;
  @apply border border-slate-200/80 dark:border-slate-700/50;
  @apply backdrop-blur-xl z-[9999];
  @apply overflow-hidden;
}

.tags-context-menu__item {
  @apply flex items-center gap-3 w-full px-3 py-2 text-sm text-left;
  @apply text-slate-700 dark:text-slate-300;
  @apply hover:bg-slate-50 dark:hover:bg-slate-700/50;
  @apply transition-colors duration-100 cursor-pointer;
}

.tags-context-menu__item svg {
  @apply text-slate-500 dark:text-slate-400;
}

.tags-context-menu__item:hover svg {
  @apply text-primary-500 dark:text-primary-400;
}

.tags-context-menu__item--disabled {
  @apply opacity-40 cursor-not-allowed pointer-events-none;
}

.tags-context-menu__divider {
  @apply h-px my-1.5 mx-2 bg-slate-200 dark:bg-slate-700;
}

/* ========== 标签页风格 ========== */

/* 谷歌风格 */
.tags-view--chrome {
  @apply bg-slate-100 dark:bg-slate-800/50;
}

.tags-view--chrome .tags-view__scroll {
  @apply gap-0.5 px-2;
}

.tags-view--chrome .tags-view__tag {
  @apply rounded-t-xl rounded-b-none;
  @apply border border-b-0 border-slate-300/80 dark:border-slate-600/50;
  @apply bg-slate-200/60 dark:bg-slate-700/40;
  @apply mb-0;
}

.tags-view--chrome .tags-view__tag.tag-active {
  @apply bg-white dark:bg-slate-800;
  @apply border-slate-300 dark:border-slate-600;
  @apply shadow-none;
}

.tags-view--chrome .tags-view__tag-indicator {
  @apply hidden;
}

/* 朴素风格 */
.tags-view--plain {
  @apply bg-transparent border-b-0 h-10;
}

.tags-view--plain .tags-view__tag {
  @apply rounded-none bg-transparent border-0;
  @apply border-b-2 border-transparent;
  @apply hover:bg-transparent hover:border-slate-300 dark:hover:border-slate-600;
  @apply shadow-none;
}

.tags-view--plain .tags-view__tag-indicator {
  @apply hidden;
}

.tags-view--plain .tags-view__tag.tag-active {
  @apply bg-transparent;
  @apply border-primary-500 dark:border-primary-400;
  @apply text-primary-600 dark:text-primary-400;
}

/* 卡片风格 */
.tags-view--card .tags-view__tag {
  @apply rounded-xl;
  @apply bg-white dark:bg-slate-800/80;
  @apply border border-slate-200 dark:border-slate-700;
  @apply shadow-sm;
}

.tags-view--card .tags-view__tag.tag-active {
  @apply border-primary-400 dark:border-primary-600;
  @apply shadow-md shadow-primary-500/10;
}

.tags-view--card .tags-view__tag-indicator {
  @apply hidden;
}

/* 轻快风格 (Smart) */
.tags-view--smart {
  @apply bg-gradient-to-r from-slate-100 via-slate-50 to-slate-100;
  @apply dark:from-slate-800 dark:via-slate-900 dark:to-slate-800;
}

.tags-view--smart .tags-view__tag {
  @apply rounded-full px-4;
  @apply bg-slate-200/70 dark:bg-slate-700/40;
  @apply border-0;
  @apply backdrop-blur-sm;
}

.tags-view--smart .tags-view__tag-indicator {
  @apply hidden;
}

.tags-view--smart .tags-view__tag.tag-active {
  @apply bg-gradient-to-r from-primary-500 via-primary-500 to-primary-600;
  @apply text-white;
  @apply shadow-lg shadow-primary-500/30;
  @apply border-0;
}

.tags-view--smart .tags-view__tag.tag-active .tags-view__tag-close {
  @apply text-white/70 hover:text-white hover:bg-white/20;
}
</style>
