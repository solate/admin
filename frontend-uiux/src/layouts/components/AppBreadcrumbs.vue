<script setup lang="ts">
/**
 * 应用面包屑组件
 * 支持显示/隐藏控制、样式切换、单层隐藏、首页显示
 */
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { usePreferencesStore } from '@/stores/modules/preferences'
import { useI18n } from '@/locales/composables'
import { Home, ChevronRight } from 'lucide-vue-next'
import type { BreadcrumbStyle } from '@/types/preferences'

const route = useRoute()
const { t } = useI18n()
const preferencesStore = usePreferencesStore()

const layoutPrefs = computed(() => preferencesStore.layout)

// ==================== 面包屑数据 ====================

interface BreadcrumbItem {
  label: string
  path: string
  isHome?: boolean
}

const breadcrumbs = computed((): BreadcrumbItem[] => {
  const pathSegments = route.path.split('/').filter(Boolean)
  const crumbs: BreadcrumbItem[] = []

  // 路由标签映射
  const labels: Record<string, string> = {
    dashboard: t('nav.dashboard'),
    overview: t('nav.overview'),
    tenants: t('nav.tenants'),
    services: t('nav.services'),
    users: t('nav.users'),
    analytics: t('nav.analytics'),
    settings: t('nav.settings'),
    profile: t('nav.profile'),
    notifications: t('nav.notifications')
  }

  // 如果配置显示首页，添加首页
  if (layoutPrefs.value.breadcrumbShowHome) {
    crumbs.push({
      label: t('nav.overview'),
      path: '/dashboard/overview',
      isHome: true
    })
  }

  // 添加路由面包屑
  if (pathSegments[0] === 'dashboard') {
    let currentPath = ''
    for (let i = 0; i < pathSegments.length; i++) {
      currentPath += `/${pathSegments[i]}`
      // 如果已经添加了首页，且当前是 overview，跳过
      if (layoutPrefs.value.breadcrumbShowHome && pathSegments[i] === 'overview') {
        continue
      }
      crumbs.push({
        label: labels[pathSegments[i]] || pathSegments[i],
        path: currentPath
      })
    }
  }

  return crumbs
})

// ==================== 显示控制 ====================

/** 是否应该显示面包屑 */
const shouldShowBreadcrumbs = computed(() => {
  // 如果配置不显示面包屑
  if (!layoutPrefs.value.showBreadcrumbs) return false

  // 如果配置单层隐藏且只有一个面包屑
  if (layoutPrefs.value.hideSingleBreadcrumb && breadcrumbs.value.length <= 1) {
    return false
  }

  return breadcrumbs.value.length > 0
})

// ==================== 样式 ====================

const breadcrumbClass = computed(() => [
  'app-breadcrumbs',
  `app-breadcrumbs--${layoutPrefs.value.breadcrumbStyle}`
])

/** 是否为最后一项 */
const isLast = (index: number) => index === breadcrumbs.value.length - 1
</script>

<template>
  <nav v-if="shouldShowBreadcrumbs" :class="breadcrumbClass" aria-label="Breadcrumb">
    <template v-for="(crumb, index) in breadcrumbs" :key="crumb.path">
      <!-- 面包屑项 -->
      <router-link
        :to="crumb.path"
        class="breadcrumb-item"
        :class="{
          'breadcrumb-item--active': isLast(index),
          'breadcrumb-item--home': crumb.isHome
        }"
        :aria-current="isLast(index) ? 'page' : undefined"
      >
        <Home v-if="crumb.isHome" :size="14" class="breadcrumb-icon" />
        <span class="breadcrumb-label">{{ crumb.label }}</span>
      </router-link>

      <!-- 分隔符 -->
      <ChevronRight
        v-if="!isLast(index)"
        :size="14"
        class="breadcrumb-separator"
      />
    </template>
  </nav>
</template>

<style scoped>
.app-breadcrumbs {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: nowrap;
  overflow: hidden;
}

/* 常规样式 */
.app-breadcrumbs--normal {
  /* 使用默认样式 */
}

/* 背景样式 */
.app-breadcrumbs--background {
  padding: 0.5rem 1rem;
  background-color: rgb(241 245 249 / 1);
  border-radius: var(--border-radius);
}

.dark .app-breadcrumbs--background {
  background-color: rgb(30 41 59 / 0.5);
}

/* 面包屑项 */
.breadcrumb-item {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: rgb(100 116 139 / 1);
  text-decoration: none;
  transition: color 0.2s;
  cursor: pointer;
}

.breadcrumb-item:hover {
  color: rgb(51 65 85 / 1);
}

.dark .breadcrumb-item {
  color: rgb(148 163 184 / 1);
}

.dark .breadcrumb-item:hover {
  color: rgb(203 213 225 / 1);
}

/* 激活状态 */
.breadcrumb-item--active {
  color: rgb(15 23 42 / 1);
  font-weight: 600;
  cursor: default;
}

.dark .breadcrumb-item--active {
  color: rgb(241 245 249 / 1);
}

/* 首页图标 */
.breadcrumb-icon {
  flex-shrink: 0;
}

.breadcrumb-label {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 150px;
}

/* 分隔符 */
.breadcrumb-separator {
  flex-shrink: 0;
  color: rgb(148 163 184 / 1);
}

.dark .breadcrumb-separator {
  color: rgb(71 85 105 / 1);
}
</style>
