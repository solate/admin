<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    :fill="fill"
    :viewBox="viewBox"
    stroke="none"
    v-bind="$attrs"
  >
    <slot />
  </svg>
</template>

<script setup>
defineProps({
  fill: {
    type: String,
    default: 'currentColor'
  },
  viewBox: {
    type: String,
    default: '0 0 24 24'
  }
})
</script>
