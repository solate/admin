<template>
  <el-config-provider :locale="elementLocale">
    <router-view />
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useLocaleStore } from '@/stores/locale'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'

const localeStore = useLocaleStore()

// Element Plus 语言包映射
const elementLocaleMap: Record<string, any> = {
  'zh-CN': zhCn,
  'en-US': en
}

// 当前 Element Plus 语言
const elementLocale = computed(() => elementLocaleMap[localeStore.currentLocale.value])
</script>

<style>
html, body, #app {
  height: 100%;
}
body {
  margin: 0;
}
</style>
