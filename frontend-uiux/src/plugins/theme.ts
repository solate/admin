/**
 * 主题初始化插件
 */

import type { App } from 'vue'
import { usePreferencesStore } from '@/stores/modules/preferences'

export default {
  install: (app: App) => {
    // 使用 mixin 在根组件挂载时初始化主题
    app.mixin({
      mounted() {
        // 只在根组件上执行一次
        if (this.$root !== this) return

        const preferencesStore = usePreferencesStore()
        preferencesStore.initialize()
      }
    })
  }
}
