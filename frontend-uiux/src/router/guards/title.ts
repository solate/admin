/**
 * 页面标题守卫
 * 自动更新页面标题，支持多语言
 */

import type { Router } from 'vue-router'
import i18n from '@/locales'

/**
 * 翻译标题
 * 将 i18n key 转换为对应语言的文本
 */
function translateTitle(title: string): string {
  // 如果标题看起来像是 i18n key（包含点号），尝试翻译
  if (title && title.includes('.')) {
    // 使用类型断言处理 legacy: false 模式下的类型问题
    const t = i18n.global.t as (key: string) => string
    const translated = t(title)
    // 如果翻译结果和 key 不同，说明翻译成功
    if (translated !== title) {
      return translated
    }
  }
  return title
}

/**
 * 获取应用名称
 */
function getAppName(): string {
  // 产品名称保持英文（品牌一致性）
  // 可选：使用 i18n.global.t('common.appName') 如果需要翻译
  return 'MultiTenant'
}

/**
 * 设置标题守卫
 */
export function setupTitleGuard(router: Router) {
  router.afterEach((to) => {
    // 从路由元信息获取标题
    const titleKey = to.meta.title as string | undefined

    // 翻译标题
    const pageTitle = titleKey ? translateTitle(titleKey) : ''
    const appName = getAppName()

    // 构建完整标题
    // 格式：页面名称 - 产品名称（符合主流 SaaS 设计）
    const fullTitle = pageTitle
      ? `${pageTitle} - ${appName}`
      : appName

    // 更新文档标题
    document.title = fullTitle
  })
}
