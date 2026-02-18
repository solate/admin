<script setup lang="ts">
import { useRoute } from 'vue-router'
import { ElConfigProvider } from 'element-plus'
import { useUiStore } from '@/stores/modules/ui'

const route = useRoute()
const uiStore = useUiStore()
</script>

<template>
  <!-- SVG 滤镜定义 - 用于色盲/色弱模式 -->
  <svg class="svg-filters" style="position: absolute; width: 0; height: 0; overflow: hidden;" aria-hidden="true">
    <defs>
      <!-- 红色盲 (Protanopia) - 完全无法感知红色 -->
      <filter id="protanopia-filter">
        <feColorMatrix
          type="matrix"
          values="0.567, 0.433, 0,     0, 0
                  0.558, 0.442, 0,     0, 0
                  0,     0.242, 0.758, 0, 0
                  0,     0,     0,     1, 0"
        />
      </filter>

      <!-- 绿色盲 (Deuteranopia) - 完全无法感知绿色 -->
      <filter id="deuteranopia-filter">
        <feColorMatrix
          type="matrix"
          values="0.625, 0.375, 0,   0, 0
                  0.7,   0.3,   0,   0, 0
                  0,     0.3,   0.7, 0, 0
                  0,     0,     0,   1, 0"
        />
      </filter>

      <!-- 蓝色盲 (Tritanopia) - 完全无法感知蓝色 -->
      <filter id="tritanopia-filter">
        <feColorMatrix
          type="matrix"
          values="0.95, 0.05,  0,     0, 0
                  0,    0.433, 0.567, 0, 0
                  0,    0.475, 0.525, 0, 0
                  0,    0,     0,     1, 0"
        />
      </filter>

      <!-- 红色弱 (Protanomaly) - 红色敏感度降低 -->
      <filter id="protanomaly-filter">
        <feColorMatrix
          type="matrix"
          values="0.817, 0.183, 0,     0, 0
                  0.333, 0.667, 0,     0, 0
                  0,     0.125, 0.875, 0, 0
                  0,     0,     0,     1, 0"
        />
      </filter>

      <!-- 绿色弱 (Deuteranomaly) - 绿色敏感度降低（最常见） -->
      <filter id="deuteranomaly-filter">
        <feColorMatrix
          type="matrix"
          values="0.8,   0.2,   0,     0, 0
                  0.258, 0.742, 0,     0, 0
                  0,     0.142, 0.858, 0, 0
                  0,     0,     0,     1, 0"
        />
      </filter>

      <!-- 蓝色弱 (Tritanomaly) - 蓝色敏感度降低 -->
      <filter id="tritanomaly-filter">
        <feColorMatrix
          type="matrix"
          values="0.967, 0.033, 0,     0, 0
                  0,     0.733, 0.267, 0, 0
                  0,     0.183, 0.817, 0, 0
                  0,     0,     0,     1, 0"
        />
      </filter>
    </defs>
  </svg>

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
