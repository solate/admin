<template>
  <div class="notifications-page">
    <el-card>
      <template #header>
        <span>通知配置</span>
      </template>
      <el-empty description="功能开发中..." />
    </el-card>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped lang="scss">
.notifications-page {
  :deep(.el-card__body) {
    min-height: 400px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}
</style>
