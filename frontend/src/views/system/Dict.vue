<template>
  <div class="dict-page">
    <div class="page-header">
      <h1 class="page-title">字典管理</h1>
      <el-button type="primary">
        <el-icon><Plus /></el-icon>
        新增字典
      </el-button>
    </div>

    <el-card>
      <el-empty description="字典管理功能开发中..." />
    </el-card>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped lang="scss">
.dict-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;

    .page-title {
      font-size: 28px;
      font-weight: 700;
      margin: 0;
    }
  }
}
</style>