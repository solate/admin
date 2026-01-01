<template>
  <el-container class="admin-layout">
    <!-- 侧边栏 -->
    <el-aside
      :width="isCollapse ? `${sidebarCollapsedWidth}px` : `${sidebarWidth}px`"
      class="sidebar"
    >

      <!-- 侧边栏标题区域 -->
      <div class="sidebar-header">
        <div class="tenant-trigger-static" :class="{ 'is-collapsed': isCollapse }">
          <div class="tenant-avatar">
            <el-icon size="20">
              <Promotion />
            </el-icon>
          </div>
          <transition name="fade">
            <div v-show="!isCollapse" class="tenant-meta">
              <div class="tenant-title">后端管理系统</div>
            </div>
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
          <!-- 工作台 -->
          <el-menu-item index="/">
            <el-icon><DataBoard /></el-icon>
            <template #title>工作台</template>
          </el-menu-item>

          <!-- 租户管理 -->
          <el-sub-menu index="/tenant">
            <template #title>
              <el-icon><OfficeBuilding /></el-icon>
              <span>租户管理</span>
            </template>
            <el-menu-item index="/tenant/list">租户列表</el-menu-item>
            <el-menu-item index="/tenant/packages">套餐管理</el-menu-item>
            <el-menu-item index="/tenant/subscription">订阅管理</el-menu-item>
            <el-menu-item index="/tenant/billing">账单管理</el-menu-item>
          </el-sub-menu>

          <!-- 组织架构 -->
          <el-sub-menu index="/organization">
            <template #title>
              <el-icon><Share /></el-icon>
              <span>组织架构</span>
            </template>
            <el-menu-item index="/organization/departments">部门管理</el-menu-item>
            <el-menu-item index="/organization/positions">岗位管理</el-menu-item>
          </el-sub-menu>

          <!-- 用户与权限 -->
          <el-sub-menu index="/access">
            <template #title>
              <el-icon><Lock /></el-icon>
              <span>用户与权限</span>
            </template>
            <el-menu-item index="/access/users">用户管理</el-menu-item>
            <el-menu-item index="/access/roles">角色管理</el-menu-item>
            <el-menu-item index="/access/menus">菜单权限</el-menu-item>
            <el-menu-item index="/access/data-permissions">数据权限</el-menu-item>
          </el-sub-menu>

          <!-- 业务管理（预留） -->
          <el-sub-menu index="/business">
            <template #title>
              <el-icon><Briefcase /></el-icon>
              <span>业务管理</span>
            </template>
            <el-menu-item index="/business/factories">工厂管理</el-menu-item>
            <el-menu-item index="/business/products">商品管理</el-menu-item>
            <el-menu-item index="/business/orders">订单管理</el-menu-item>
            <el-menu-item index="/business/statistics">数据统计</el-menu-item>
          </el-sub-menu>

          <!-- 审计日志 -->
          <el-sub-menu index="/audit">
            <template #title>
              <el-icon><Document /></el-icon>
              <span>审计日志</span>
            </template>
            <el-menu-item index="/audit/login">登录日志</el-menu-item>
            <el-menu-item index="/audit/operation">操作日志</el-menu-item>
            <el-menu-item index="/audit/data">数据变更</el-menu-item>
          </el-sub-menu>

          <!-- 系统设置 -->
          <el-sub-menu index="/settings">
            <template #title>
              <el-icon><Setting /></el-icon>
              <span>系统设置</span>
            </template>
            <el-menu-item index="/settings/dictionary">字典管理</el-menu-item>
            <el-menu-item index="/settings/parameters">系统参数</el-menu-item>
            <el-menu-item index="/settings/notifications">通知配置</el-menu-item>
            <el-menu-item index="/settings/storage">存储配置</el-menu-item>
            <el-menu-item index="/settings/monitor">系统监控</el-menu-item>
          </el-sub-menu>
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

          <!-- 用户信息下拉菜单 -->
          <UserDropdown
            :user-info="userInfo"
            :tenant="userInfo?.tenant || null"
            :roles="userRoles"
            @command="handleUserCommand"
            @tenant-changed="handleTenantChanged"
          />
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
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowRight } from '@element-plus/icons-vue'
import { useThemeStore } from '../stores/theme'
import { getUserInfo } from '../utils/token'
import NProgress from 'nprogress'
import UserDropdown from '../components/user/UserDropdown.vue'

const route = useRoute()
const themeStore = useThemeStore()

const isCollapse = ref(false)
const isFullscreen = ref(false)
const notificationVisible = ref(false)
const searchQuery = ref('')
const userInfo = getUserInfo()
const isMobile = ref(window.innerWidth < 768)

const sidebarWidth = 260
const sidebarCollapsedWidth = 72

// 从 token 中获取租户信息（用于侧边栏显示）
const currentTenantCode = ref(userInfo?.tenant?.tenant_code || 'default')
const currentTenantName = ref(userInfo?.tenant?.name || '默认租户')
// 角色列表
const userRoles = ref(userInfo?.roles || [])

const unreadCount = computed(() => notifications.value.filter(n => !n.read).length)

const breadcrumbConfig: Record<string, { title: string }> = {
  '/': { title: '工作台' },
  '/tenant': { title: '租户管理' },
  '/tenant/list': { title: '租户列表' },
  '/tenant/packages': { title: '套餐管理' },
  '/tenant/subscription': { title: '订阅管理' },
  '/tenant/billing': { title: '账单管理' },
  '/organization': { title: '组织架构' },
  '/organization/departments': { title: '部门管理' },
  '/organization/positions': { title: '岗位管理' },
  '/access': { title: '用户与权限' },
  '/access/users': { title: '用户管理' },
  '/access/roles': { title: '角色管理' },
  '/access/menus': { title: '菜单权限' },
  '/access/data-permissions': { title: '数据权限' },
  '/business': { title: '业务管理' },
  '/business/factories': { title: '工厂管理' },
  '/business/products': { title: '商品管理' },
  '/business/orders': { title: '订单管理' },
  '/business/statistics': { title: '数据统计' },
  '/audit': { title: '审计日志' },
  '/audit/login': { title: '登录日志' },
  '/audit/operation': { title: '操作日志' },
  '/audit/data': { title: '数据变更' },
  '/settings': { title: '系统设置' },
  '/settings/dictionary': { title: '字典管理' },
  '/settings/parameters': { title: '系统参数' },
  '/settings/notifications': { title: '通知配置' },
  '/settings/storage': { title: '存储配置' },
  '/settings/monitor': { title: '系统监控' }
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

function handleUserCommand(command: string) {
  // UserDropdown 组件内部已处理 logout 逻辑
  // 这里只处理其他命令
  switch (command) {
    case 'profile':
    case 'settings':
      ElMessage.info('功能开发中...')
      break
  }
}

function handleTenantChanged(tenant: { tenant_id: string; name: string; tenant_code: string }) {
  // 更新当前租户信息
  currentTenantCode.value = tenant.tenant_code
  currentTenantName.value = tenant.name
  // 更新用户信息中的租户信息
  if (userInfo.value) {
    userInfo.value.tenant = tenant
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
      padding: 0 14px;
      border-bottom: 1px solid var(--border-base);
      background: var(--bg-white);

      .workspace-dropdown {
        width: 100%;
        height: 100%;
      }

      // 静态版本 (移除租户切换功能后使用)
      .tenant-trigger-static {
        display: flex;
        align-items: center;
        gap: 12px;
        width: 100%;
        height: 100%;
        padding: 10px 10px;
        border-radius: 12px;
        border: 1px solid transparent;
        transition: var(--transition-base);

        &.is-collapsed {
          justify-content: center;
          padding: 10px 0;
        }

        .tenant-avatar {
          width: 38px;
          height: 38px;
          display: flex;
          align-items: center;
          justify-content: center;
          background: var(--gradient-primary);
          border-radius: 10px;
          color: white;
          flex-shrink: 0;
          box-shadow: 0 4px 12px rgba(66, 133, 244, 0.2);
        }

        .tenant-meta {
          display: flex;
          flex-direction: column;
          gap: 4px;
          flex: 1;
          min-width: 0;

          .tenant-title {
            font-size: 15px;
            font-weight: 600;
            color: var(--text-primary);
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            line-height: 1.3;
            letter-spacing: -0.3px;
          }

          .tenant-subtitle {
            font-size: 12px;
            font-weight: 500;
            color: var(--text-secondary);
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            opacity: 0.9;
          }
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

          .search-box {
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
