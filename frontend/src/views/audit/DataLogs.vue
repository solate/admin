<template>
  <div class="data-logs-page">
    <el-card>
      <template #header>
        <span>数据变更</span>
      </template>
      <el-empty description="功能开发中..." />
    </el-card>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped lang="scss">
.data-logs-page {
  :deep(.el-card__body) {
    min-height: 400px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}
</style>
