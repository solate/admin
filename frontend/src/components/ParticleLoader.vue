<template>
  <div class="particle-loader-wrapper">
    <div class="particle-loader">
      <div
        v-for="i in particleCount"
        :key="i"
        class="particle"
        :style="getParticleStyle(i)"
      ></div>
    </div>
    <div v-if="text" class="loader-text">{{ text }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  text?: string
  particleCount?: number
  color?: string
}

const props = withDefaults(defineProps<Props>(), {
  text: '',
  particleCount: 20,
  color: ''
})

const particlePositions = computed(() => {
  const positions = []
  for (let i = 0; i < props.particleCount; i++) {
    positions.push({
      left: Math.random() * 100,
      animationDelay: Math.random() * 2,
      animationDuration: 1.5 + Math.random() * 1,
      tx: (Math.random() - 0.5) * 40
    })
  }
  return positions
})

function getParticleStyle(index: number) {
  const pos = particlePositions.value[index - 1]
  return {
    left: `${pos.left}%`,
    animationDelay: `${pos.animationDelay}s`,
    animationDuration: `${pos.animationDuration}s`,
    '--tx': `${pos.tx}px`,
    background: props.color || 'var(--gradient-aurora)'
  }
}
</script>

<style scoped lang="scss">
.particle-loader-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-large);
}

.particle-loader {
  position: relative;
  width: 100%;
  height: 200px;
  overflow: hidden;
  background: var(--bg-glass);
  border-radius: var(--border-radius-large);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
}

@keyframes particle-float {
  0% {
    transform: translateY(0) translateX(0) scale(1);
    opacity: 0;
  }
  10% {
    opacity: 1;
  }
  90% {
    opacity: 1;
  }
  100% {
    transform: translateY(-150px) translateX(var(--tx)) scale(0);
    opacity: 0;
  }
}

.particle {
  position: absolute;
  bottom: 10px;
  width: 4px;
  height: 4px;
  border-radius: 50%;
  animation: particle-float 2s ease-in-out infinite;
  box-shadow: 0 0 6px currentColor;
}

.loader-text {
  margin-top: var(--spacing-base);
  font-size: var(--font-size-small);
  color: var(--text-secondary);
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 0.6;
  }
  50% {
    opacity: 1;
  }
}
</style>
