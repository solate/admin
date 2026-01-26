/**
 * 页面标题守卫
 * 自动更新页面标题
 */

import type { Router } from 'vue-router'
import { appConfig } from '@/config/app'

/**
 * 设置标题守卫
 */
export function setupTitleGuard(router: Router) {
  router.afterEach((to) => {
    // 从路由元信息获取标题
    const title = to.meta.title as string | undefined

    // 构建完整标题
    const fullTitle = title
      ? `${title} - ${appConfig.name}`
      : appConfig.name

    // 更新文档标题
    document.title = fullTitle
  })
}
