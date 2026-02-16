<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

onMounted(() => {
  const { path } = route.params
  router.replace('/' + (Array.isArray(path) ? path.join('/') : path))
})
</script>

<template>
  <div />
</template>
