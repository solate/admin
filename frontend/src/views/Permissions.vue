<template>
  <div class="permissions-page">
    <div class="page-header">
      <h1 class="page-title">{{ pageTitle }}</h1>
      <el-button type="primary" @click="handleSave">
        <el-icon><Check /></el-icon>
        保存配置
      </el-button>
    </div>

    <el-card>
      <el-empty description="权限管理功能开发中..." />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const pageTitle = computed(() => {
  const titles: Record<string, string> = {
    'permissions-menu': '菜单权限',
    'permissions-api': '接口权限',
    'permissions-data': '数据权限'
  }
  return titles[route.name as string] || '权限管理'
})

const handleSave = () => {
  console.log('保存权限配置')
}
</script>

<style scoped lang="scss">
.permissions-page {
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