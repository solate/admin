/**
 * 布局逻辑 Composable
 * 封装所有布局相关的计算逻辑，提供统一的布局状态访问接口
 */

import { computed, ref, onMounted, onUnmounted } from 'vue'
import { usePreferencesStore } from '@/stores/modules/preferences'
import { useUiStore } from '@/stores/modules/ui'
import type { LayoutMode, HeaderMode, ContentWidthMode, NavStyle, TabsStyle, BreadcrumbStyle, WidgetsPosition } from '@/types/preferences'

// 全局设置抽屉状态 - 模块级别变量，布局切换时不会被重置
const globalShowSettingsDrawer = ref(false)
const globalSettingsActiveTab = ref<'appearance' | 'layout' | 'shortcuts' | 'general' | 'advanced'>('appearance')

/**
 * 布局 Composable
 */
export function useLayout() {
  const preferencesStore = usePreferencesStore()
  const uiStore = useUiStore()

  // ==================== 布局模式 ====================

  /** 当前布局模式 */
  const layoutMode = computed<LayoutMode>(() => preferencesStore.layout.layoutMode)

  /** 是否为侧边栏模式 */
  const isSidebarMode = computed(() => layoutMode.value === 'sidebar')

  /** 是否为顶部导航模式 */
  const isTopbarMode = computed(() => layoutMode.value === 'topbar')

  /** 是否为混合模式 */
  const isMixedMode = computed(() => layoutMode.value === 'mixed')

  /** 是否为水平模式 */
  const isHorizontalMode = computed(() => layoutMode.value === 'horizontal')

  /** 是否为双列菜单模式 */
  const isDoubleSidebarMode = computed(() => layoutMode.value === 'double-sidebar')

  // ==================== 顶栏相关 ====================

  /** 顶栏模式 */
  const headerMode = computed<HeaderMode>(() => preferencesStore.layout.headerMode)

  /** 顶栏高度（像素值） */
  const headerHeight = computed(() => preferencesStore.layout.headerHeight)

  /** 顶栏高度（带单位的字符串） */
  const headerHeightPx = computed(() => `${headerHeight.value}px`)

  /** 顶栏是否固定 */
  const isHeaderFixed = computed(() => headerMode.value === 'fixed')

  /** 顶栏是否自动隐藏 */
  const isHeaderAutoHide = computed(() => headerMode.value === 'auto-hide')

  // ==================== 侧边栏相关 ====================

  /** 侧边栏宽度 */
  const sidebarWidth = computed(() => preferencesStore.layout.sidebarWidth)

  /** 侧边栏折叠宽度 */
  const sidebarCollapsedWidth = computed(() => preferencesStore.layout.sidebarCollapsedWidth)

  /** 侧边栏是否可折叠 */
  const sidebarCollapsible = computed(() => preferencesStore.layout.sidebarCollapsible)

  /** 侧边栏默认折叠状态 */
  const sidebarCollapsed = computed(() => preferencesStore.layout.sidebarCollapsed)

  /** 侧边栏当前宽度（像素值） */
  const sidebarWidthPx = computed(() => {
    return uiStore.sidebarOpen ? sidebarWidth.value : sidebarCollapsedWidth.value
  })

  /** 侧边栏当前宽度（带单位的字符串） */
  const sidebarWidthPxStr = computed(() => `${sidebarWidthPx.value}px`)

  // ==================== 导航相关 ====================

  /** 导航样式 */
  const navStyle = computed<NavStyle>(() => preferencesStore.layout.navStyle)

  /** 导航手风琴模式 */
  const navAccordion = computed(() => preferencesStore.layout.navAccordion)

  /** 是否仅显示图标 */
  const iconOnlyNav = computed(() => navStyle.value === 'icon-only')

  // ==================== 内容区域 ====================

  /** 内容宽度模式 */
  const contentWidthMode = computed<ContentWidthMode>(() => preferencesStore.layout.contentWidthMode)

  /** 内容定宽值 */
  const contentFixedWidth = computed(() => preferencesStore.layout.contentFixedWidth || 1200)

  /** 内容区域是否定宽 */
  const isContentFixed = computed(() => contentWidthMode.value === 'fixed')

  /** 内容区域样式 */
  const contentStyle = computed(() => {
    if (contentWidthMode.value === 'fixed') {
      return {
        maxWidth: `${contentFixedWidth.value}px`,
        margin: '0 auto'
      }
    }
    return {}
  })

  // ==================== 面包屑 ====================

  /** 显示面包屑 */
  const showBreadcrumbs = computed(() => preferencesStore.layout.showBreadcrumbs)

  /** 仅有一个面包屑时隐藏 */
  const hideSingleBreadcrumb = computed(() => preferencesStore.layout.hideSingleBreadcrumb)

  /** 面包屑样式 */
  const breadcrumbStyle = computed<BreadcrumbStyle>(() => preferencesStore.layout.breadcrumbStyle)

  /** 面包屑显示首页 */
  const breadcrumbShowHome = computed(() => preferencesStore.layout.breadcrumbShowHome)

  // ==================== 标签页 ====================

  /** 显示标签页 */
  const showTabs = computed(() => preferencesStore.layout.showTabs)

  /** 标签页持久化 */
  const tabsPersistent = computed(() => preferencesStore.layout.tabsPersistent)

  /** 标签页可拖拽 */
  const tabsDraggable = computed(() => preferencesStore.layout.tabsDraggable)

  /** 标签页样式 */
  const tabsStyle = computed<TabsStyle>(() => preferencesStore.layout.tabsStyle)

  // ==================== 页脚 ====================

  /** 显示页脚 */
  const showFooter = computed(() => preferencesStore.layout.showFooter)

  /** 页脚固定 */
  const footerFixed = computed(() => preferencesStore.layout.footerFixed)

  /** 显示版权信息 */
  const showCopyright = computed(() => preferencesStore.layout.showCopyright)

  // ==================== 小部件 ====================

  /** 显示小部件 */
  const showWidgets = computed(() => preferencesStore.layout.showWidgets)

  /** 小部件位置 */
  const widgetsPosition = computed<WidgetsPosition>(() => preferencesStore.layout.widgetsPosition)

  // ==================== 通用设置 ====================

  /** 通用设置 */
  const generalPrefs = computed(() => preferencesStore.general)

  // ==================== 组合样式 ====================

  /** 窗口宽度（用于响应式判断） */
  const windowWidth = ref(typeof window !== 'undefined' ? window.innerWidth : 1024)

  /** 是否为桌面端 */
  const isDesktop = computed(() => windowWidth.value >= 1024)

  /** 是否为移动端 */
  const isMobile = computed(() => !isDesktop.value)

  /** 主内容区样式（根据布局模式动态计算） */
  const mainContentStyle = computed(() => {
    const styles: Record<string, string> = {}

    // 侧边栏模式和混合模式下，桌面端需要左边距
    if ((isSidebarMode.value || isMixedMode.value) && isDesktop.value) {
      styles.marginLeft = sidebarWidthPxStr.value
    }

    // 固定顶栏模式下，内容区需要上边距
    if (isHeaderFixed.value || headerMode.value === 'static') {
      // static 模式下也需要考虑顶栏高度对布局的影响
    }

    return styles
  })

  /** 顶栏定位类 */
  const headerPositionClass = computed(() => {
    // 侧边栏模式、混合模式和双列菜单模式下，顶栏不应固定到视口顶部，而是跟随主内容区
    // 这样可以避免顶栏覆盖侧边栏
    if (isSidebarMode.value || isMixedMode.value || isDoubleSidebarMode.value) {
      switch (headerMode.value) {
        case 'static':
          return 'relative'
        case 'auto-hide':
          // 自动隐藏模式在侧边栏布局下使用 sticky
          return 'sticky top-0 z-20'
        default:
          // fixed 和默认情况都使用 sticky
          return 'sticky top-0 z-20'
      }
    }

    // 顶部导航模式和水平模式下，顶栏可以固定到视口顶部
    switch (headerMode.value) {
      case 'fixed':
        return 'fixed top-0 left-0 right-0 z-30'
      case 'static':
        return 'relative'
      case 'auto-hide':
        return 'fixed top-0 left-0 right-0 z-30'
      default:
        return 'sticky top-0 z-20'
    }
  })

  /** 顶栏内联样式 */
  const headerStyle = computed(() => ({
    height: headerHeightPx.value
  }))

  /** 侧边栏内联样式 */
  const sidebarStyle = computed(() => ({
    width: sidebarWidthPxStr.value
  }))

  // ==================== 窗口尺寸监听 ====================

  const handleResize = () => {
    windowWidth.value = window.innerWidth
  }

  onMounted(() => {
    window.addEventListener('resize', handleResize)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', handleResize)
  })

  // ==================== 返回 ====================

  return {
    // 设置抽屉状态（全局共享）
    showSettingsDrawer: globalShowSettingsDrawer,
    settingsActiveTab: globalSettingsActiveTab,

    // 布局模式
    layoutMode,
    isSidebarMode,
    isTopbarMode,
    isMixedMode,
    isHorizontalMode,
    isDoubleSidebarMode,

    // 顶栏
    headerMode,
    headerHeight,
    headerHeightPx,
    isHeaderFixed,
    isHeaderAutoHide,
    headerPositionClass,
    headerStyle,

    // 侧边栏
    sidebarWidth,
    sidebarCollapsedWidth,
    sidebarCollapsible,
    sidebarCollapsed,
    sidebarWidthPx,
    sidebarWidthPxStr,
    sidebarStyle,

    // 导航
    navStyle,
    navAccordion,
    iconOnlyNav,

    // 内容区域
    contentWidthMode,
    contentFixedWidth,
    isContentFixed,
    contentStyle,
    mainContentStyle,

    // 面包屑
    showBreadcrumbs,
    hideSingleBreadcrumb,
    breadcrumbStyle,
    breadcrumbShowHome,

    // 标签页
    showTabs,
    tabsPersistent,
    tabsDraggable,
    tabsStyle,

    // 页脚
    showFooter,
    footerFixed,
    showCopyright,

    // 小部件
    showWidgets,
    widgetsPosition,

    // 通用设置
    generalPrefs,

    // 响应式
    windowWidth,
    isDesktop,
    isMobile
  }
}

/**
 * 顶栏自动隐藏逻辑
 */
export function useHeaderAutoHide(enabled: () => boolean, threshold = 60) {
  const isVisible = ref(true)
  const lastScrollY = ref(0)

  const handleScroll = () => {
    if (!enabled()) {
      isVisible.value = true
      return
    }

    const currentScrollY = window.scrollY

    if (currentScrollY > lastScrollY.value && currentScrollY > threshold) {
      // 向下滚动超过阈值，隐藏顶栏
      isVisible.value = false
    } else {
      // 向上滚动或在阈值内，显示顶栏
      isVisible.value = true
    }

    lastScrollY.value = currentScrollY
  }

  onMounted(() => {
    window.addEventListener('scroll', handleScroll, { passive: true })
  })

  onUnmounted(() => {
    window.removeEventListener('scroll', handleScroll)
  })

  return {
    isVisible,
    lastScrollY
  }
}
