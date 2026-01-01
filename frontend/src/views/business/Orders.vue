<template>
  <div class="orders-page">
    <el-card>
      <template #header>
        <span>订单管理</span>
      </template>
      <el-empty description="功能开发中..." />
    </el-card>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped lang="scss">
.orders-page {
  :deep(.el-card__body) {
    min-height: 400px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}
</style>
