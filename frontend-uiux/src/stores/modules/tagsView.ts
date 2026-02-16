// 标签页状态管理

import { defineStore } from 'pinia'
import { ref, watch, computed } from 'vue'
import type { RouteLocationNormalized } from 'vue-router'
import type { TagView, TagsViewState } from '@/types/tagsView'
import { TAGS_VIEW_STORAGE_KEY } from '@/types/tagsView'
import { usePreferencesStore } from './preferences'

export const useTagsViewStore = defineStore('tagsView', () => {
  // 从 localStorage 读取初始状态
  const getStoredState = (): TagsViewState | null => {
    try {
      const stored = localStorage.getItem(TAGS_VIEW_STORAGE_KEY)
      return stored ? JSON.parse(stored) : null
    } catch {
      return null
    }
  }

  const storedState = getStoredState()

  // State
  const visitedTags = ref<TagView[]>(storedState?.visitedTags || [])
  const cachedTags = ref<string[]>(storedState?.cachedTags || [])
  const activeTagPath = ref<string>(storedState?.activeTagPath || '')
  const isMaximized = ref(false)

  /**
   * 切换最大化状态
   */
  function toggleMaximized() {
    isMaximized.value = !isMaximized.value
  }

  /**
   * 设置最大化状态
   */
  function setMaximized(value: boolean) {
    isMaximized.value = value
  }

  /**
   * 初始化标签页
   */
  function initialize(routes: RouteLocationNormalized[]) {
    const preferencesStore = usePreferencesStore()

    // 如果开启了持久化且有存储的标签，则恢复
    if (preferencesStore.layout.tabsPersistent && storedState?.visitedTags.length) {
      return
    }

    // 查找需要固定的路由（affix: true）
    const affixRoutes = routes.filter(route => route.meta?.affix)
    affixRoutes.forEach(route => {
      addTag(route)
    })
  }

  /**
   * 从路由生成标签
   */
  function generateTagFromRoute(route: RouteLocationNormalized): TagView {
    const title = (route.meta?.title as string) || route.name?.toString() || route.path
    const icon = route.meta?.icon as string | undefined

    return {
      path: route.path,
      name: route.name?.toString() || '',
      title,
      icon,
      query: route.query as Record<string, string>,
      fullPath: route.fullPath,
      affix: route.meta?.affix as boolean | undefined
    }
  }

  /**
   * 添加标签
   */
  function addTag(route: RouteLocationNormalized) {
    const tag = generateTagFromRoute(route)

    // 检查是否已存在
    const exists = visitedTags.value.some(t => t.path === tag.path)
    if (!exists) {
      visitedTags.value.push(tag)
    }

    // 更新激活标签
    activeTagPath.value = tag.path

    // 添加到缓存（用于 keep-alive）
    if (route.name && route.meta?.keepAlive !== false) {
      const name = route.name.toString()
      if (!cachedTags.value.includes(name)) {
        cachedTags.value.push(name)
      }
    }
  }

  /**
   * 关闭标签
   */
  function removeTag(path: string): TagView | null {
    // 查找要关闭的标签
    const index = visitedTags.value.findIndex(t => t.path === path)
    if (index === -1) return null

    const removedTag = visitedTags.value[index]

    // 固定标签不能关闭
    if (removedTag.affix) return null

    // 移除标签
    visitedTags.value.splice(index, 1)

    // 从缓存中移除
    if (removedTag.name) {
      const cacheIndex = cachedTags.value.indexOf(removedTag.name)
      if (cacheIndex > -1) {
        cachedTags.value.splice(cacheIndex, 1)
      }
    }

    // 如果关闭的是当前激活的标签，需要切换到其他标签
    if (activeTagPath.value === path) {
      // 优先切换到右边的标签，否则切换到左边
      const nextTag = visitedTags.value[index] || visitedTags.value[index - 1]
      return nextTag || null
    }

    return null
  }

  /**
   * 关闭其他标签
   */
  function closeOtherTags(currentPath?: string) {
    const path = currentPath || activeTagPath.value
    visitedTags.value = visitedTags.value.filter(t => t.affix || t.path === path)

    // 更新缓存
    const currentTag = visitedTags.value.find(t => t.path === path)
    if (currentTag?.name) {
      cachedTags.value = [currentTag.name]
    } else {
      cachedTags.value = []
    }
  }

  /**
   * 关闭左侧标签
   */
  function closeLeftTags(path: string): TagView | null {
    const currentIndex = visitedTags.value.findIndex(t => t.path === path)
    if (currentIndex === -1) return null

    // 保留当前标签及其右侧的标签，以及固定标签
    visitedTags.value = visitedTags.value.filter((t, index) => t.affix || index >= currentIndex)

    // 更新缓存
    cachedTags.value = visitedTags.value
      .filter(t => t.name)
      .map(t => t.name!)

    // 如果当前激活的标签被关闭，返回新的目标标签
    if (!visitedTags.value.find(t => t.path === activeTagPath.value)) {
      return visitedTags.value[visitedTags.value.length - 1] || null
    }

    return null
  }

  /**
   * 关闭右侧标签
   */
  function closeRightTags(path: string): TagView | null {
    const currentIndex = visitedTags.value.findIndex(t => t.path === path)
    if (currentIndex === -1) return null

    // 保留当前标签及其左侧的标签，以及固定标签
    visitedTags.value = visitedTags.value.filter((t, index) => t.affix || index <= currentIndex)

    // 更新缓存
    cachedTags.value = visitedTags.value
      .filter(t => t.name)
      .map(t => t.name!)

    // 如果当前激活的标签被关闭，返回新的目标标签
    if (!visitedTags.value.find(t => t.path === activeTagPath.value)) {
      return visitedTags.value[currentIndex] || null
    }

    return null
  }

  /**
   * 关闭所有标签
   */
  function closeAllTags(): TagView | null {
    // 只保留固定标签
    visitedTags.value = visitedTags.value.filter(t => t.affix)
    cachedTags.value = visitedTags.value
      .filter(t => t.name)
      .map(t => t.name!)

    // 返回最后一个固定标签（通常是首页）
    return visitedTags.value[visitedTags.value.length - 1] || null
  }

  /**
   * 更新标签顺序（拖拽排序）
   */
  function updateTagOrder(oldIndex: number, newIndex: number) {
    const preferencesStore = usePreferencesStore()
    if (!preferencesStore.layout.tabsDraggable) return

    const [removed] = visitedTags.value.splice(oldIndex, 1)
    visitedTags.value.splice(newIndex, 0, removed)
  }

  /**
   * 更新标签标题
   */
  function updateTagTitle(path: string, title: string) {
    const tag = visitedTags.value.find(t => t.path === path)
    if (tag) {
      tag.title = title
    }
  }

  /**
   * 切换标签固定状态
   */
  function toggleAffix(path: string) {
    const tag = visitedTags.value.find(t => t.path === path)
    if (tag) {
      tag.affix = !tag.affix
    }
  }

  /**
   * 设置激活标签
   */
  function setActiveTag(path: string) {
    activeTagPath.value = path
  }

  /**
   * 刷新当前标签（移除缓存后重新加载）
   */
  function refreshTag(path: string): string | null {
    const tag = visitedTags.value.find(t => t.path === path)
    if (!tag?.name) return null

    // 从缓存中移除
    const cacheIndex = cachedTags.value.indexOf(tag.name)
    if (cacheIndex > -1) {
      cachedTags.value.splice(cacheIndex, 1)
    }

    return tag.name
  }

  // 持久化状态
  const preferencesStore = usePreferencesStore()

  watch(
    [visitedTags, cachedTags, activeTagPath],
    () => {
      if (preferencesStore.layout.tabsPersistent) {
        const state: TagsViewState = {
          visitedTags: visitedTags.value,
          cachedTags: cachedTags.value,
          activeTagPath: activeTagPath.value
        }
        localStorage.setItem(TAGS_VIEW_STORAGE_KEY, JSON.stringify(state))
      }
    },
    { deep: true }
  )

  return {
    // State
    visitedTags,
    cachedTags,
    activeTagPath,
    isMaximized,

    // Actions
    initialize,
    addTag,
    removeTag,
    closeOtherTags,
    closeLeftTags,
    closeRightTags,
    closeAllTags,
    updateTagOrder,
    updateTagTitle,
    toggleAffix,
    setActiveTag,
    refreshTag,
    toggleMaximized,
    setMaximized
  }
})
