<script setup lang="ts">
/**
 * 应用页脚组件
 * 支持固定在底部、版权信息显示
 */
import { computed } from 'vue'
import { usePreferencesStore } from '@/stores/modules/preferences'
import { useI18n } from '@/locales/composables'

const props = withDefaults(defineProps<{
  /** 是否固定在底部 */
  fixed?: boolean
  /** 是否显示版权信息 */
  showCopyright?: boolean
}>(), {
  fixed: false,
  showCopyright: true
})

const { t } = useI18n()
const preferencesStore = usePreferencesStore()

// ==================== 计算属性 ====================

const footerClass = computed(() => [
  'app-footer',
  {
    'app-footer--fixed': props.fixed
  }
])

const copyrightInfo = computed(() => preferencesStore.copyright)

const currentYear = computed(() => new Date().getFullYear())
</script>

<template>
  <footer :class="footerClass">
    <div class="footer-content">
      <!-- 默认插槽 -->
      <slot>
        <div class="flex flex-col sm:flex-row items-center justify-between gap-2 text-sm text-slate-500 dark:text-slate-400">
          <!-- 版权信息 -->
          <p>
            &copy; {{ copyrightInfo.date || currentYear }}
            <a
              v-if="copyrightInfo.companySiteLink"
              :href="copyrightInfo.companySiteLink"
              target="_blank"
              rel="noopener noreferrer"
              class="hover:text-primary-600 dark:hover:text-primary-400 transition-colors"
            >
              {{ copyrightInfo.companyName }}
            </a>
            <span v-else>{{ copyrightInfo.companyName }}</span>
            . {{ t('common.allRightsReserved', 'All rights reserved.') }}
          </p>

          <!-- 技术栈信息 -->
          <p v-if="showCopyright" class="text-xs">
            Built with Vue 3 + Vite + Tailwind CSS
          </p>
        </div>

        <!-- ICP 备案 -->
        <div v-if="copyrightInfo.icp" class="mt-2 text-center">
          <a
            :href="copyrightInfo.icpLink || 'https://beian.miit.gov.cn/'"
            target="_blank"
            rel="noopener noreferrer"
            class="text-xs text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-400 transition-colors"
          >
            {{ copyrightInfo.icp }}
          </a>
        </div>
      </slot>
    </div>
  </footer>
</template>

<style scoped>
.app-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid rgb(226 232 240 / 0.6);
  background: linear-gradient(to bottom, rgb(255 255 255 / 0.5), rgb(255 255 255 / 0.8));
}

.dark .app-footer {
  border-top-color: rgb(51 65 85 / 0.6);
  background: linear-gradient(to bottom, rgb(15 23 42 / 0.5), rgb(15 23 42 / 0.8));
}

.app-footer--fixed {
  position: sticky;
  bottom: 0;
  z-index: 10;
  backdrop-filter: blur(8px);
}

.footer-content {
  max-width: 100%;
}
</style>
