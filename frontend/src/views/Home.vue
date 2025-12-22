<template>
  <div class="home">
    <el-container style="height: 100vh">
      <el-aside width="200px">
        <el-menu :default-active="active" class="el-menu-vertical-demo">
          <el-menu-item index="1">首页</el-menu-item>
        </el-menu>
      </el-aside>
      <el-container>
        <el-header>
          <div class="logo">后台管理系统</div>
          <div class="actions">
            <el-button text @click="onLogout">退出登录</el-button>
          </div>
        </el-header>
        <el-main>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item>首页</el-breadcrumb-item>
          </el-breadcrumb>
          <div class="content">欢迎使用！</div>
        </el-main>
      </el-container>
    </el-container>
  </div>
  
</template>

<script setup lang="ts">
import { ref } from 'vue'

const active = ref('1')

function onLogout() {
  localStorage.removeItem('token')
  location.href = '/login'
}
</script>

<style scoped>
.logo { font-weight: 600; }
.actions { float: right; }
.content { margin-top: 16px; }
</style>


