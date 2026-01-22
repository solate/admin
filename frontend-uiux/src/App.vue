<script setup>
import { useRoute } from 'vue-router'
import { ElConfigProvider } from 'element-plus'
import { useUiStore } from './stores/ui'

const route = useRoute()
const uiStore = useUiStore()
</script>

<template>
  <el-config-provider :locale="uiStore.elementLocale">
    <router-view v-slot="{ Component }">
      <transition :name="route.meta.transition || 'fade'" mode="out-in">
        <component :is="Component" :key="route.path" />
      </transition>
    </router-view>
  </el-config-provider>
</template>

<style>
/* Fade Transition */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 200ms ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Slide Transition */
.slide-enter-active,
.slide-leave-active {
  transition: transform 200ms ease, opacity 200ms ease;
}

.slide-enter-from {
  transform: translateX(20px);
  opacity: 0;
}

.slide-leave-to {
  transform: translateX(-20px);
  opacity: 0;
}
</style>
