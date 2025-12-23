<template>
  <div class="roles-page">
    <div class="page-header">
      <h1 class="page-title">角色管理</h1>
      <el-button type="primary" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        创建角色
      </el-button>
    </div>

    <el-card>
      <el-table :data="roles" v-loading="loading" stripe>
        <el-table-column prop="name" label="角色名称" width="200" />
        <el-table-column prop="code" label="角色代码" width="200" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="userCount" label="用户数" width="100" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'">
              {{ row.status === 'active' ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" text size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="warning" text size="small" @click="handlePermissions(row)">权限配置</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

const loading = ref(false)
const roles = ref([
  { id: 1, name: '超级管理员', code: 'super_admin', description: '系统超级管理员', userCount: 1, status: 'active' },
  { id: 2, name: '管理员', code: 'admin', description: '系统管理员', userCount: 5, status: 'active' },
  { id: 3, name: '运营', code: 'operator', description: '运营人员', userCount: 10, status: 'active' },
  { id: 4, name: '普通用户', code: 'user', description: '普通用户', userCount: 50, status: 'active' }
])

const handleCreate = () => {
  ElMessage.info('创建角色功能开发中...')
}

const handleEdit = (role: any) => {
  ElMessage.info(`编辑角色：${role.name}`)
}

const handlePermissions = (role: any) => {
  ElMessage.info(`配置权限：${role.name}`)
}

onMounted(() => {
  // 初始化数据
})
</script>

<style scoped lang="scss">
.roles-page {
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