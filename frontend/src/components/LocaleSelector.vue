<template>
  <el-dropdown trigger="click" @command="handleLocaleChange">
    <span class="locale-selector">
      <el-icon><Language /></el-icon>
    </span>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item
          v-for="locale in locales"
          :key="locale.key"
          :command="locale.key"
          :class="{ 'is-active': locale.key === currentLocale }"
        >
          {{ locale.label }}
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useLocaleStore } from '@/stores/locale'
import { locales, type LocaleType } from '@/locales'
import { useI18n } from 'vue-i18n'

const localeStore = useLocaleStore()
const { locale } = useI18n()

// 当前语言
const currentLocale = computed(() => localeStore.currentLocale)

// 切换语言
const handleLocaleChange = (newLocale: LocaleType) => {
  localeStore.setLocale(newLocale)
  locale.value = newLocale
}
</script>

<style scoped lang="scss">
.locale-selector {
  display: flex;
  align-items: center;
  cursor: pointer;
  color: var(--el-text-color-regular);
  transition: color 0.3s;

  &:hover {
    color: var(--el-color-primary);
  }
}

:deep(.el-dropdown-menu__item.is-active) {
  color: var(--el-color-primary);
  font-weight: 500;
}
</style>
