<template>
  <el-container class="admin-layout">
    <!-- 侧边栏 -->
    <el-aside
      :width="isCollapse ? `${sidebarCollapsedWidth}px` : `${sidebarWidth}px`"
      class="sidebar"
    >
      <!-- Logo区域 -->
      <div class="sidebar-header">
        <div class="logo-container" @click="goHome">
          <div class="logo-icon">
            <el-icon size="24">
              <Promotion />
            </el-icon>
          </div>
          <transition name="fade">
            <span v-show="!isCollapse" class="logo-text">管理系统</span>
          </transition>
        </div>
      </div>


      <!-- 菜单 -->
      <el-scrollbar class="sidebar-scrollbar">
        <el-menu
          :default-active="activeMenu"
          :collapse="isCollapse"
          :unique-opened="true"
          router
          class="sidebar-menu"
          :collapse-transition="false"
        >
          <el-menu-item index="/">
            <el-icon><House /></el-icon>
            <template #title>仪表盘</template>
          </el-menu-item>

          <el-sub-menu index="/system">
            <template #title>
              <el-icon><Setting /></el-icon>
              <span>系统管理</span>
            </template>
            <el-menu-item index="/system/users">用户管理</el-menu-item>
            <el-menu-item index="/system/roles">角色管理</el-menu-item>
            <el-menu-item index="/system/tenants">租户管理</el-menu-item>
            <el-menu-item index="/system/permissions/menu">菜单权限</el-menu-item>
            <el-menu-item index="/system/permissions/api">接口权限</el-menu-item>
            <el-menu-item index="/system/permissions/data">数据权限</el-menu-item>
            <el-menu-item index="/system/dict">字典管理</el-menu-item>
            <el-menu-item index="/system/logs">操作日志</el-menu-item>
            <el-menu-item index="/system/monitor">系统监控</el-menu-item>
          </el-sub-menu>

          <el-menu-item index="/factories">
            <el-icon><OfficeBuilding /></el-icon>
            <template #title>工厂管理</template>
          </el-menu-item>

          <el-menu-item index="/products">
            <el-icon><Box /></el-icon>
            <template #title>商品管理</template>
          </el-menu-item>

          <el-menu-item index="/statistics">
            <el-icon><TrendCharts /></el-icon>
            <template #title>数据统计</template>
          </el-menu-item>
        </el-menu>
      </el-scrollbar>
    </el-aside>

    <!-- 主内容区域 -->
    <el-container class="main-container">
      <!-- 顶部导航 -->
      <el-header class="header">
        <div class="header-left">
          <!-- 折叠按钮 -->
          <el-button text @click="toggleCollapse" class="collapse-btn">
            <el-icon :size="18">
              <Fold v-if="!isCollapse" />
              <Expand v-else />
            </el-icon>
          </el-button>

          <!-- 面包屑导航 -->
          <el-breadcrumb separator="/" class="breadcrumb">
            <el-breadcrumb-item
              v-for="item in breadcrumbList"
              :key="item.path"
              :to="item.path"
            >
              {{ item.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <div class="header-right">
          <!-- 全局搜索 -->
          <div class="search-box">
            <el-input
              v-model="searchQuery"
              placeholder="搜索资源..."
              prefix-icon="Search"
              class="search-input"
            />
          </div>

          <!-- 全屏按钮 -->
          <el-tooltip content="全屏" placement="bottom">
            <el-button text @click="toggleFullscreen" class="header-action">
              <el-icon :size="18">
                <FullScreen v-if="!isFullscreen" />
                <Aim v-else />
              </el-icon>
            </el-button>
          </el-tooltip>

          <!-- 主题切换 -->
          <el-tooltip :content="themeStore.theme === 'light' ? '切换到暗色模式' : '切换到亮色模式'" placement="bottom">
            <el-button text @click="themeStore.toggleTheme" class="header-action">
              <el-icon :size="18">
                <Moon v-if="themeStore.theme === 'light'" />
                <Sunny v-else />
              </el-icon>
            </el-button>
          </el-tooltip>

          <!-- 通知 -->
          <el-tooltip content="通知" placement="bottom">
            <el-badge :value="unreadCount" :max="99" class="header-badge">
              <el-button text @click="showNotifications" class="header-action">
                <el-icon :size="18">
                  <Bell />
                </el-icon>
              </el-button>
            </el-badge>
          </el-tooltip>

          <!-- 租户切换器 -->
          <div class="tenant-selector-header">
            <el-popover
              :width="280"
              trigger="click"
              placement="bottom-end"
            >
              <template #reference>
                <div class="tenant-current">
                  <div class="tenant-avatar">
                    T
                  </div>
                  <span class="tenant-name">{{ currentTenant }}</span>
                  <el-icon class="dropdown-arrow" :size="14">
                    <ArrowDown />
                  </el-icon>
                </div>
              </template>
              <div class="tenant-popover">
                <div class="popover-header">
                  <span class="popover-title">切换租户</span>
                </div>
                <div class="popover-list">
                  <div
                    v-for="tenant in tenants"
                    :key="tenant.id"
                    class="popover-item"
                    :class="{ active: tenant.id === currentTenantId }"
                    @click="switchTenant(tenant)"
                  >
                    <div class="item-info">
                      <div class="item-name">{{ tenant.name }}</div>
                      <div class="item-desc">{{ tenant.description }}</div>
                    </div>
                    <span class="item-status" :class="tenant.status">
                      {{ tenant.statusText }}
                    </span>
                  </div>
                </div>
              </div>
            </el-popover>
          </div>

          <!-- 用户信息下拉菜单 -->
          <el-dropdown @command="handleUserCommand" class="user-dropdown">
            <div class="user-info">
              <el-avatar :size="32">
                {{ userInfo?.user_name?.charAt(0).toUpperCase() || 'A' }}
              </el-avatar>
              <div v-if="!isMobile" class="user-details">
                <div class="user-name">{{ userInfo?.user_name || '管理员' }}</div>
                <div class="user-role">超级管理员</div>
              </div>
              <el-icon class="dropdown-icon" :size="14">
                <ArrowDown />
              </el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon>
                  个人中心
                </el-dropdown-item>
                <el-dropdown-item command="settings">
                  <el-icon><Setting /></el-icon>
                  账号设置
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 内容区域 -->
      <el-main class="main-content">
        <router-view v-slot="{ Component, route }">
          <transition name="slide-fade" mode="out-in">
            <component :is="Component" :key="route.path" />
          </transition>
        </router-view>
      </el-main>

      <!-- 页脚 -->
      <el-footer class="footer">
        <div class="footer-content">
          <span>&copy; 2025 多租户管理系统 v1.0.0</span>
        </div>
      </el-footer>
    </el-container>

    <!-- 通知抽屉 -->
    <el-drawer v-model="notificationVisible" title="系统通知" size="400px">
      <div class="notification-list">
        <div
          v-for="item in notifications"
          :key="item.id"
          class="notification-item"
          :class="{ 'is-read': item.read }"
          @click="markAsRead(item)"
        >
          <div class="notification-content">
            <div class="notification-title">{{ item.title }}</div>
            <div class="notification-desc">{{ item.description }}</div>
            <div class="notification-time">{{ item.time }}</div>
          </div>
        </div>
      </div>
    </el-drawer>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useThemeStore } from '../stores/theme'
import { authApi } from '../api'
import { clearTokens, getUserInfo } from '../utils/token'
import NProgress from 'nprogress'

const route = useRoute()
const router = useRouter()
const themeStore = useThemeStore()

const isCollapse = ref(false)
const isFullscreen = ref(false)
const notificationVisible = ref(false)
const searchQuery = ref('')
const userInfo = getUserInfo()
const isMobile = ref(window.innerWidth < 768)

const sidebarWidth = 260
const sidebarCollapsedWidth = 72

const currentTenantId = ref(1)
const currentTenant = ref('默认租户')
const tenants = ref([
  { id: 1, name: '默认租户', description: '主租户账号', status: 'online', statusText: '运行中' },
  { id: 2, name: '测试租户', description: '测试环境', status: 'warning', statusText: '维护中' },
  { id: 3, name: '开发租户', description: '开发环境', status: 'offline', statusText: '已停用' }
])

const unreadCount = computed(() => notifications.value.filter(n => !n.read).length)

const breadcrumbConfig: Record<string, { title: string }> = {
  '/': { title: '首页' },
  '/system': { title: '系统管理' },
  '/system/users': { title: '用户管理' },
  '/system/roles': { title: '角色管理' },
  '/system/tenants': { title: '租户管理' },
  '/system/permissions': { title: '权限管理' },
  '/system/permissions/menu': { title: '菜单权限' },
  '/system/permissions/api': { title: '接口权限' },
  '/system/permissions/data': { title: '数据权限' },
  '/system/dict': { title: '字典管理' },
  '/system/logs': { title: '操作日志' },
  '/system/monitor': { title: '系统监控' },
  '/factories': { title: '工厂管理' },
  '/products': { title: '商品管理' },
  '/statistics': { title: '数据统计' }
}

const activeMenu = computed(() => route.path)

const breadcrumbList = computed(() => {
  const pathArray = route.path.split('/').filter(Boolean)
  const breadcrumbs = [{ path: '/', title: '首页' }]
  let currentPath = ''
  pathArray.forEach((path) => {
    currentPath += `/${path}`
    const config = breadcrumbConfig[currentPath]
    if (config) {
      breadcrumbs.push({ path: currentPath, title: config.title })
    }
  })
  return breadcrumbs
})

const notifications = ref([
  { id: 1, title: '系统更新', description: '系统已更新至 v1.0.1 版本', time: '10分钟前', read: false },
  { id: 2, title: '存储空间告警', description: '当前存储空间使用率已达到 80%', time: '1小时前', read: false },
  { id: 3, title: '数据备份完成', description: '系统数据已成功备份至云端服务器', time: '2小时前', read: true }
])

watch(() => route.path, () => {
  NProgress.start()
  setTimeout(() => NProgress.done(), 300)
})

function toggleCollapse() {
  isCollapse.value = !isCollapse.value
}

function goHome() {
  router.push('/')
}

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen()
    isFullscreen.value = true
  } else {
    document.exitFullscreen()
    isFullscreen.value = false
  }
}

function handleFullscreenChange() {
  isFullscreen.value = !!document.fullscreenElement
}

function showNotifications() {
  notificationVisible.value = true
}

function markAsRead(item: any) {
  item.read = true
}

function switchTenant(tenant: any) {
  currentTenantId.value = tenant.id
  currentTenant.value = tenant.name
  ElMessage.success(`已切换到租户：${tenant.name}`)
}

async function handleUserCommand(command: string) {
  switch (command) {
    case 'profile':
    case 'settings':
      ElMessage.info('功能开发中...')
      break
    case 'logout':
      try {
        await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        try {
          await authApi.logout()
        } catch {}
        clearTokens()
        ElMessage.success('已安全退出登录')
        router.push('/login')
      } catch {}
      break
  }
}

function handleResize() {
  isMobile.value = window.innerWidth < 768
  if (window.innerWidth < 768) {
    isCollapse.value = true
  }
}

onMounted(() => {
  document.addEventListener('fullscreenchange', handleFullscreenChange)
  window.addEventListener('resize', handleResize)
  NProgress.configure({ showSpinner: false, minimum: 0.2, easing: 'ease', speed: 500 })
  handleResize()
})

onUnmounted(() => {
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped lang="scss">
.admin-layout {
  height: 100vh;
  background: var(--bg-page);
  overflow: hidden;

  .sidebar {
    background: var(--bg-white);
    border-right: 1px solid var(--border-base);
    display: flex;
    flex-direction: column;
    box-shadow: var(--box-shadow-light);

    .sidebar-header {
      height: var(--header-height);
      display: flex;
      align-items: center;
      padding: 0 20px;
      border-bottom: 1px solid var(--border-base);

      .logo-container {
        display: flex;
        align-items: center;
        gap: 12px;
        cursor: pointer;
        width: 100%;

        .logo-icon {
          width: 36px;
          height: 36px;
          display: flex;
          align-items: center;
          justify-content: center;
          background: var(--gradient-primary);
          border-radius: var(--border-radius);
          color: white;
          flex-shrink: 0;
        }

        .logo-text {
          flex: 1;
          font-size: 16px;
          font-weight: 600;
          color: var(--text-primary);
          white-space: nowrap;
        }
      }
    }

    .sidebar-scrollbar {
      flex: 1;

      :deep(.el-scrollbar__view) {
        height: 100%;
      }
    }

    .sidebar-menu {
      border: none;
      padding: 8px 0;
    }
  }

  .main-container {
    display: flex;
    flex-direction: column;
    overflow: hidden;

    .header {
      background: var(--bg-white);
      border-bottom: 1px solid var(--border-base);
      padding: 0 24px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      box-shadow: var(--box-shadow-light);
      height: var(--header-height);

      .header-left {
        display: flex;
        align-items: center;
        gap: 16px;

        .collapse-btn {
          color: var(--text-regular);

          &:hover {
            color: var(--primary-color);
          }
        }

        .breadcrumb :deep(.el-breadcrumb__item) {
          .el-breadcrumb__inner {
            color: var(--text-regular);
            font-weight: 500;

            &:hover {
              color: var(--primary-color);
            }
          }

          &:last-child .el-breadcrumb__inner {
            color: var(--text-primary);
            font-weight: 600;
          }
        }
      }

      .header-right {
        display: flex;
        align-items: center;
        gap: 12px;

        .tenant-selector-header {
          .tenant-current {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 4px 8px;
            border-radius: var(--border-radius);
            cursor: pointer;
            transition: var(--transition-base);

            .tenant-avatar {
              width: 28px;
              height: 28px;
              display: flex;
              align-items: center;
              justify-content: center;
              background: var(--gradient-primary);
              border-radius: 50%;
              color: white;
              font-size: 13px;
              font-weight: 600;
            }

            .tenant-name {
              font-size: 14px;
              font-weight: 600;
              color: var(--text-primary);
              line-height: 1;
            }

            .dropdown-arrow {
              font-size: 14px;
              color: var(--text-secondary);
              transition: transform 0.3s;
            }

            &:hover {
              background: var(--bg-light);

              .dropdown-arrow {
                transform: rotate(180deg);
              }
            }
          }
        }

        .search-box .search-input {
          width: 200px;
        }

        .header-action {
          color: var(--text-regular);
          width: 36px;
          height: 36px;
          padding: 0;

          &:hover {
            color: var(--primary-color);
          }
        }

        .header-badge {
          :deep(.el-badge__content) {
            font-size: 11px;
            height: 16px;
            line-height: 16px;
            padding: 0 5px;
            border: 1.5px solid var(--bg-white);
          }
        }

        .user-dropdown {
          .user-info {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 4px 8px;
            border-radius: var(--border-radius);
            cursor: pointer;

            &:hover {
              background: var(--bg-light);
            }

            .el-avatar {
              background: var(--gradient-primary);
            }

            .user-details {
              display: flex;
              flex-direction: column;
              gap: 2px;

              .user-name {
                font-size: 14px;
                font-weight: 600;
                color: var(--text-primary);
                line-height: 1;
              }

              .user-role {
                font-size: 12px;
                color: var(--text-secondary);
                line-height: 1;
              }
            }

            .dropdown-icon {
              font-size: 14px;
              color: var(--text-secondary);
              transition: transform 0.3s;
            }

            &:hover .dropdown-icon {
              transform: rotate(180deg);
            }
          }
        }
      }
    }

    .main-content {
      flex: 1;
      padding: 24px;
      overflow-y: auto;
      background: var(--bg-page);
    }

    .footer {
      height: auto;
      padding: 16px 24px;
      background: var(--bg-white);
      border-top: 1px solid var(--border-base);

      .footer-content {
        text-align: center;
        color: var(--text-secondary);
        font-size: 13px;
      }
    }
  }
}

// 租户选择弹出框样式
.tenant-popover {
  .popover-header {
    padding: 12px 0;
    border-bottom: 1px solid var(--border-base);
    margin-bottom: 8px;

    .popover-title {
      font-size: 14px;
      font-weight: 600;
      color: var(--text-primary);
    }
  }

  .popover-list {
    display: flex;
    flex-direction: column;
    gap: 4px;

    .popover-item {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 10px 12px;
      border-radius: 6px;
      cursor: pointer;
      transition: var(--transition-base);

      .item-info {
        flex: 1;

        .item-name {
          font-size: 14px;
          font-weight: 500;
          color: var(--text-primary);
          line-height: 1.4;
        }

        .item-desc {
          font-size: 12px;
          color: var(--text-secondary);
          margin-top: 2px;
        }
      }

      .item-status {
        font-size: 12px;
        padding: 2px 8px;
        border-radius: 4px;

        &.online {
          color: #67c23a;
          background: rgba(103, 194, 58, 0.1);
        }

        &.warning {
          color: #e6a23c;
          background: rgba(230, 162, 60, 0.1);
        }

        &.offline {
          color: #909399;
          background: rgba(144, 147, 153, 0.1);
        }
      }

      &:hover {
        background: var(--bg-light);
      }

      &.active {
        background: var(--bg-light);
        border-radius: 6px;

        .item-name {
          color: var(--primary-color);
          font-weight: 600;
        }
      }
    }
  }
}

.notification-list {
  display: flex;
  flex-direction: column;
  gap: 12px;

  .notification-item {
    padding: 16px;
    border-radius: var(--border-radius);
    border: 1px solid var(--border-base);
    cursor: pointer;
    transition: var(--transition-base);

    &:hover {
      background: var(--bg-light);
    }

    &.is-read {
      opacity: 0.6;
    }

    .notification-content {
      .notification-title {
        font-size: 14px;
        font-weight: 600;
        color: var(--text-primary);
        margin-bottom: 4px;
      }

      .notification-desc {
        font-size: 13px;
        color: var(--text-regular);
        margin-bottom: 4px;
      }

      .notification-time {
        font-size: 12px;
        color: var(--text-secondary);
      }
    }
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

@media (max-width: 768px) {
  .admin-layout {
    .sidebar {
      position: fixed;
      left: 0;
      top: 0;
      height: 100vh;
      z-index: 1001;
      transform: translateX(-100%);
      transition: transform 0.3s;
    }

    .main-container {
      .header {
        padding: 0 16px;

        .header-left .breadcrumb {
          display: none;
        }

        .header-right {
          gap: 8px;

          .tenant-selector-header {
            .tenant-current {
              padding: 4px 8px;

              .tenant-name {
                font-size: 12px;
              }
            }
          }

          .search-box {
            display: none;
          }

          .user-details {
            display: none;
          }
        }
      }

      .main-content {
        padding: 16px;
      }
    }
  }
}
</style>
