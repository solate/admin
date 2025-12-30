<template>
  <el-container class="admin-layout">
    <!-- 侧边栏 -->
    <el-aside
      :width="isCollapse ? `${sidebarCollapsedWidth}px` : `${sidebarWidth}px`"
      class="sidebar"
    >
      <!-- 租户选择区域 -->
      <div class="sidebar-header">
        <el-dropdown
          trigger="click"
          placement="bottom-start"
          class="workspace-dropdown"
          @command="handleTenantCommand"
          @visible-change="handleTenantVisibleChange"
        >
          <div class="tenant-trigger" :class="{ 'is-collapsed': isCollapse }">
            <div class="tenant-avatar">
              <el-icon size="20">
                <Promotion />
              </el-icon>
            </div>
            <transition name="fade">
              <div v-show="!isCollapse" class="tenant-meta">
                <div class="tenant-title">后端管理系统</div>
                <div class="tenant-subtitle">{{ currentTenantName }}</div>
              </div>
            </transition>
            <el-icon class="tenant-caret" :class="{ open: tenantDropdownOpen }" :size="14">
              <CaretBottom />
            </el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu class="workspace-menu">
              <div class="workspace-menu-header">
                <div class="workspace-menu-title">选择租户</div>
              </div>

              <el-scrollbar max-height="320px" class="workspace-scroll">
                <div class="workspace-list">
                  <el-dropdown-item
                    v-for="tenant in tenantList"
                    :key="tenant.tenant_id"
                    :command="tenant.tenant_id"
                    class="workspace-dropdown-item"
                  >
                    <div class="workspace-item" :class="{ active: tenant.tenant_id === currentTenantId }">
                      <div class="workspace-avatar">
                        <span class="avatar-text">{{ tenant.name?.charAt(0)?.toUpperCase() || 'T' }}</span>
                      </div>
                      <div class="workspace-details">
                        <div class="workspace-name">{{ tenant.name }}</div>
                        <div class="workspace-code">{{ tenant.code }}</div>
                      </div>
                      <div class="workspace-right">
                        <span class="workspace-badge" :class="{ danger: tenant.status === 2 }">
                          {{ tenant.status === 2 ? '禁用' : '正常' }}
                        </span>
                        <div class="workspace-check">
                          <el-icon class="check-icon"><Check /></el-icon>
                        </div>
                      </div>
                    </div>
                  </el-dropdown-item>
                </div>
              </el-scrollbar>

              <el-dropdown-item divided command="__manage_tenants__" class="workspace-dropdown-item">
                <div class="workspace-add">
                  <div class="workspace-add-icon">
                    <el-icon><Plus /></el-icon>
                  </div>
                  <div class="workspace-add-text">添加新租户</div>
                </div>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
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


          <!-- 用户信息下拉菜单 -->
          <el-dropdown trigger="click" @command="handleUserCommand" class="user-dropdown">
            <div class="user-info" :class="{ 'is-active': userDropdownOpen }">
              <el-avatar :size="32" class="user-avatar">
                {{ userInfo?.user_name?.charAt(0).toUpperCase() || 'A' }}
              </el-avatar>
              <div v-if="!isMobile" class="user-details">
                <div class="user-name">{{ userInfo?.user_name || '管理员' }}</div>
                <div class="user-role">{{ currentTenantName }}</div>
              </div>
              <el-icon class="dropdown-icon" :size="14" :class="{ 'is-rotated': userDropdownOpen }">
                <ArrowDown />
              </el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu class="user-dropdown-menu">
                <!-- 用户信息卡片 -->
                <div class="user-profile-card">
                  <div class="user-profile-header">
                    <el-avatar :size="48" class="user-profile-avatar">
                      {{ userInfo?.user_name?.charAt(0).toUpperCase() || 'A' }}
                    </el-avatar>
                    <div class="user-profile-info">
                      <div class="user-profile-name">{{ userInfo?.user_name || '管理员' }}</div>
                      <div class="user-profile-email">{{ userInfo?.email || userInfo?.phone || '未设置联系方式' }}</div>
                    </div>
                  </div>
                </div>

                <!-- 租户切换卡片 -->
                <div class="tenant-section-compact">
                  <div class="tenant-switcher-compact" @click.stop="handleTenantSwitch">
                    <div class="tenant-current-compact">
                      <div class="tenant-current-icon">
                        <span class="tenant-icon-text">{{ currentTenantName?.charAt(0) || 'T' }}</span>
                      </div>
                      <div class="tenant-current-info">
                        <div class="tenant-current-label">当前租户</div>
                        <div class="tenant-current-name">{{ currentTenantName }}</div>
                      </div>
                    </div>
                    <el-icon class="tenant-switch-icon"><ArrowRight /></el-icon>
                  </div>
                </div>

                <el-dropdown-item divided class="menu-item-profile" command="profile">
                  <div class="menu-item-content">
                    <el-icon class="menu-item-icon"><User /></el-icon>
                    <span class="menu-item-text">个人中心</span>
                  </div>
                </el-dropdown-item>
                <el-dropdown-item class="menu-item-settings" command="settings">
                  <div class="menu-item-content">
                    <el-icon class="menu-item-icon"><Setting /></el-icon>
                    <span class="menu-item-text">账号设置</span>
                  </div>
                </el-dropdown-item>

                <el-dropdown-item divided class="menu-item-logout" command="logout">
                  <div class="menu-item-content">
                    <el-icon class="menu-item-icon"><SwitchButton /></el-icon>
                    <span class="menu-item-text">退出登录</span>
                  </div>
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

    <!-- 租户选择对话框 -->
    <el-dialog
      v-model="tenantSelectorVisible"
      title="切换租户"
      width="420px"
      :close-on-click-modal="true"
      class="tenant-selector-dialog"
    >
      <div class="tenant-selector-content">
        <div class="tenant-selector-list">
          <div
            v-for="tenant in tenantList"
            :key="tenant.tenant_id"
            class="tenant-selector-item"
            :class="{ active: tenant.tenant_id === currentTenantId }"
            @click="handleTenantSelect(tenant.tenant_id)"
          >
            <div class="tenant-selector-avatar">
              <span class="tenant-selector-avatar-text">{{ tenant.name?.charAt(0)?.toUpperCase() || 'T' }}</span>
            </div>
            <div class="tenant-selector-info">
              <div class="tenant-selector-name">{{ tenant.name }}</div>
              <div class="tenant-selector-code">{{ tenant.code }}</div>
            </div>
            <div class="tenant-selector-status">
              <span class="tenant-selector-badge" :class="{ danger: tenant.status === 2 }">
                {{ tenant.status === 2 ? '禁用' : '正常' }}
              </span>
              <el-icon v-if="tenant.tenant_id === currentTenantId" class="tenant-selector-check">
                <Check />
              </el-icon>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowRight, CaretBottom } from '@element-plus/icons-vue'
import { useThemeStore } from '../stores/theme'
import { authApi } from '../api'
import { clearTokens, getUserInfo, saveTokens } from '../utils/token'
import NProgress from 'nprogress'

const route = useRoute()
const router = useRouter()
const themeStore = useThemeStore()

const isCollapse = ref(false)
const isFullscreen = ref(false)
const notificationVisible = ref(false)
const userDropdownOpen = ref(false)
const searchQuery = ref('')
const userInfo = getUserInfo()
const isMobile = ref(window.innerWidth < 768)

const sidebarWidth = 260
const sidebarCollapsedWidth = 72

// 从 token 中获取租户信息
const currentTenantCode = ref(userInfo?.current_tenant?.tenant_code || 'default')
const currentTenantName = ref(userInfo?.current_tenant?.tenant_name || '默认租户')
const currentTenantId = ref(userInfo?.tenant_id || '')

const tenantDropdownOpen = ref(false)
const tenantList = ref<Array<{ tenant_id: string; name: string; code: string; status: number }>>([
  { tenant_id: '1', name: '默认租户', code: 'default', status: 1 },
  { tenant_id: '2', name: '测试租户A', code: 'tenant-a', status: 1 },
  { tenant_id: '3', name: '测试租户B', code: 'tenant-b', status: 1 },
  { tenant_id: '4', name: '开发环境', code: 'dev', status: 1 }
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

function handleTenantVisibleChange(visible: boolean) {
  tenantDropdownOpen.value = visible
}

async function handleTenantCommand(command: string) {
  if (command === '__manage_tenants__') {
    router.push('/system/tenants')
    return
  }

  if (command === currentTenantId.value) return

  if (!userInfo?.user_id) {
    ElMessage.error('用户信息缺失，请重新登录')
    return
  }

  try {
    const response = await authApi.selectTenant(userInfo.user_id, { tenant_id: command })
    saveTokens({
      access_token: response.access_token,
      refresh_token: response.refresh_token,
      user_id: userInfo.user_id,
      username: userInfo.user_name || undefined,
      email: userInfo.email || undefined,
      phone: userInfo.phone || undefined,
      current_tenant: response.current_tenant
    })
    currentTenantId.value = response.current_tenant.tenant_id
    currentTenantName.value = response.current_tenant.tenant_name
    currentTenantCode.value = response.current_tenant.tenant_code
    ElMessage.success('租户切换成功')
    location.reload()
  } catch (error: any) {
    ElMessage.error(error?.message || '租户切换失败')
  }
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

async function handleUserCommand(command: string) {
  userDropdownOpen.value = false
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

function handleTenantSwitch() {
  userDropdownOpen.value = false
  // 打开租户选择对话框
  showTenantSelector()
}

// 租户选择器对话框
const tenantSelectorVisible = ref(false)

function showTenantSelector() {
  tenantSelectorVisible.value = true
}

async function handleTenantSelect(tenantId: string) {
  if (!userInfo?.user_id) {
    ElMessage.error('用户信息缺失，请重新登录')
    return
  }

  try {
    const response = await authApi.selectTenant(userInfo.user_id, { tenant_id: tenantId })
    saveTokens({
      access_token: response.access_token,
      refresh_token: response.refresh_token,
      user_id: userInfo.user_id,
      username: userInfo.user_name || undefined,
      email: userInfo.email || undefined,
      phone: userInfo.phone || undefined,
      current_tenant: response.current_tenant
    })
    currentTenantId.value = response.current_tenant.tenant_id
    currentTenantName.value = response.current_tenant.tenant_name
    currentTenantCode.value = response.current_tenant.tenant_code
    ElMessage.success('租户切换成功')
    tenantSelectorVisible.value = false
    setTimeout(() => location.reload(), 300)
  } catch (error: any) {
    ElMessage.error(error?.message || '租户切换失败')
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

      .tenant-trigger {
        display: flex;
        align-items: center;
        gap: 12px;
        cursor: pointer;
        width: 100%;
        height: 100%;
        padding: 10px 10px;
        border-radius: 12px;
        border: 1px solid transparent;
        transition: var(--transition-base);

        &:hover {
          background: rgba(0, 0, 0, 0.03);
        }

        &:active {
          transform: translateY(0);
        }

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

        .tenant-caret {
          color: var(--text-secondary);
          transition: transform 0.22s ease, opacity 0.22s ease;
          opacity: 0.7;
          flex-shrink: 0;
        }

        .tenant-caret.open {
          transform: rotate(180deg);
          opacity: 1;
        }

        &.is-collapsed .tenant-caret {
          display: none;
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

        .user-dropdown {
          .user-info {
            display: flex;
            align-items: center;
            gap: 10px;
            padding: 6px 12px;
            border-radius: 12px;
            cursor: pointer;
            transition: var(--transition-base);
            border: 1px solid transparent;

            &:hover {
              background: rgba(0, 0, 0, 0.03);
            }

            &.is-active {
              background: rgba(0, 0, 0, 0.04);
              border-color: rgba(66, 133, 244, 0.15);
            }

            .user-avatar {
              background: var(--gradient-primary);
              box-shadow: 0 4px 12px rgba(66, 133, 244, 0.2);
            }

            .user-details {
              display: flex;
              flex-direction: column;
              gap: 2px;

              .user-name {
                font-size: 14px;
                font-weight: 600;
                color: var(--text-primary);
                line-height: 1.2;
              }

              .user-role {
                font-size: 12px;
                color: var(--text-secondary);
                line-height: 1.2;
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
                max-width: 140px;
              }
            }

            .dropdown-icon {
              font-size: 14px;
              color: var(--text-secondary);
              transition: transform 0.3s;
              opacity: 0.7;

              &.is-rotated {
                transform: rotate(180deg);
                opacity: 1;
              }
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

// 租户切换下拉菜单样式
.workspace-menu {
  padding: 10px;
  min-width: 290px;
  border: 1px solid var(--border-base);
  box-shadow: 0 22px 64px rgba(15, 23, 42, 0.14);
  border-radius: 14px;
  overflow: hidden;

  :deep(.el-dropdown-menu__item) {
    padding: 0;
    line-height: normal;
  }
  :deep(.el-dropdown-menu__item:hover) {
    background: transparent;
    color: inherit;
  }

  .workspace-menu-header {
    padding: 4px 6px 10px;
    border-bottom: 1px solid var(--border-base);
    margin-bottom: 8px;

    .workspace-menu-title {
      font-size: 13px;
      font-weight: 700;
      color: var(--text-primary);
      line-height: 1.2;
    }

    .workspace-menu-subtitle {
      margin-top: 4px;
      font-size: 12px;
      color: var(--text-placeholder);
      line-height: 1.2;
    }
  }

  .workspace-scroll {
    padding: 2px 2px 8px;

    :deep(.el-scrollbar__view) {
      padding-right: 6px;
    }
  }

  .workspace-list {
    display: flex;
    flex-direction: column;
    gap: 10px;

    .workspace-item {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 12px;
      border-radius: 14px;
      cursor: pointer;
      transition: var(--transition-base);
      background: var(--bg-white);
      border: 1px solid var(--border-base);
      width: 100%;

      &:hover {
        border-color: rgba(66, 133, 244, 0.25);
        box-shadow: 0 14px 30px rgba(66, 133, 244, 0.12);
        transform: translateY(-1px);
      }

      &.active {
        background: rgba(66, 133, 244, 0.06);
        border-color: rgba(66, 133, 244, 0.3);

        .workspace-avatar {
          background: var(--gradient-primary);
          color: white;
          box-shadow: 0 14px 28px rgba(66, 133, 244, 0.22);
        }

        .workspace-name {
          color: var(--primary-color);
        }
      }

      .workspace-avatar {
        width: 40px;
        height: 40px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--bg-light);
        border-radius: 14px;
        color: var(--text-secondary);
        font-size: 15px;
        font-weight: 700;
        transition: var(--transition-base);
        flex-shrink: 0;

        .avatar-text {
          text-transform: uppercase;
        }
      }

      .workspace-details {
        flex: 1;
        min-width: 0;

        .workspace-name {
          font-size: 14px;
          font-weight: 700;
          color: var(--text-primary);
          line-height: 1.3;
          margin-bottom: 2px;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }

        .workspace-code {
          font-size: 12px;
          color: var(--text-secondary);
          font-weight: 500;
          text-transform: lowercase;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }
      }

      .workspace-right {
        display: flex;
        align-items: center;
        gap: 10px;
        flex-shrink: 0;
      }

      .workspace-badge {
        height: 20px;
        padding: 0 8px;
        border-radius: 999px;
        font-size: 11px;
        font-weight: 700;
        line-height: 20px;
        border: 1px solid var(--border-base);
        background: var(--bg-light);
        color: var(--text-secondary);

        &.danger {
          border-color: rgba(239, 68, 68, 0.22);
          background: rgba(239, 68, 68, 0.1);
          color: var(--danger-color);
        }
      }

      .workspace-check {
        width: 22px;
        height: 22px;
        border-radius: 999px;
        display: flex;
        align-items: center;
        justify-content: center;
        border: 1.5px solid var(--border-base);
        transition: var(--transition-base);
      }

      .check-icon {
        font-size: 14px;
        opacity: 0;
        color: white;
        transition: var(--transition-base);
      }

      &.active {
        .workspace-check {
          background: var(--gradient-primary);
          border-color: transparent;
          box-shadow: 0 10px 22px rgba(66, 133, 244, 0.25);
        }

        .check-icon {
          opacity: 1;
        }
      }
    }
  }

  .workspace-add {
    display: flex;
    align-items: center;
    gap: 12px;
    width: 100%;
    padding: 12px;
    border-radius: 14px;
    border: 1px dashed rgba(66, 133, 244, 0.35);
    background: rgba(66, 133, 244, 0.08);
    transition: var(--transition-base);

    &:hover {
      background: rgba(66, 133, 244, 0.12);
      border-color: rgba(66, 133, 244, 0.5);
      box-shadow: 0 14px 30px rgba(66, 133, 244, 0.12);
      transform: translateY(-1px);
    }

    .workspace-add-icon {
      width: 40px;
      height: 40px;
      border-radius: 14px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: var(--bg-white);
      border: 1px solid rgba(66, 133, 244, 0.18);
      color: var(--primary-color);
      flex-shrink: 0;
    }

    .workspace-add-text {
      font-size: 13px;
      font-weight: 700;
      color: var(--primary-color);
      line-height: 1.2;
    }
  }
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

// 用户下拉菜单样式
.user-dropdown-menu {
  min-width: 300px;
  padding: 12px;
  border: 1px solid var(--border-base);
  box-shadow: 0 22px 64px rgba(15, 23, 42, 0.14);
  border-radius: 16px;
  overflow: hidden;

  :deep(.el-dropdown-menu__item) {
    padding: 0;
    border-radius: 10px;
    margin: 0;
    line-height: normal;

    &:not(.menu-item-profile):not(.menu-item-settings):not(.menu-item-logout) {
      display: none;
    }

    &.el-dropdown-menu__item--divided {
      margin-top: 4px;
      padding-top: 0;

      &::before {
        display: none;
      }
    }
  }

  // 用户信息卡片
  .user-profile-card {
    padding: 0 4px 12px;
    border-bottom: 1px solid var(--border-base);
    margin-bottom: 8px;

    .user-profile-header {
      display: flex;
      align-items: center;
      gap: 14px;

      .user-profile-avatar {
        background: var(--gradient-primary);
        box-shadow: 0 6px 18px rgba(66, 133, 244, 0.22);
        flex-shrink: 0;
      }

      .user-profile-info {
        flex: 1;
        min-width: 0;

        .user-profile-name {
          font-size: 15px;
          font-weight: 700;
          color: var(--text-primary);
          line-height: 1.3;
          margin-bottom: 4px;
        }

        .user-profile-email {
          font-size: 13px;
          color: var(--text-secondary);
          line-height: 1.3;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }
      }
    }
  }

  // 菜单项
  .menu-item-content {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 12px;
    width: 100%;
    transition: var(--transition-base);
    border-radius: 10px;

    .menu-item-icon {
      font-size: 18px;
      color: var(--text-secondary);
      transition: var(--transition-base);
    }

    .menu-item-text {
      font-size: 14px;
      font-weight: 500;
      color: var(--text-primary);
    }
  }

  :deep(.el-dropdown-menu__item:hover .menu-item-content) {
    background: var(--bg-light);

    .menu-item-icon {
      color: var(--primary-color);
    }

    .menu-item-text {
      color: var(--primary-color);
    }
  }

  .menu-item-logout {
    :deep(.el-dropdown-menu__item:hover .menu-item-content) {
      background: rgba(239, 68, 68, 0.08);

      .menu-item-icon {
        color: var(--danger-color);
      }

      .menu-item-text {
        color: var(--danger-color);
      }
    }
  }

  // 租户区域
  .tenant-section-compact {
    margin: 4px 4px 8px;
    padding: 0;

    .tenant-switcher-compact {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 12px 14px;
      background: linear-gradient(135deg, rgba(66, 133, 244, 0.06) 0%, rgba(66, 133, 244, 0.02) 100%);
      border: 1px solid rgba(66, 133, 244, 0.15);
      border-radius: 12px;
      cursor: pointer;
      transition: var(--transition-base);

      &:hover {
        background: linear-gradient(135deg, rgba(66, 133, 244, 0.1) 0%, rgba(66, 133, 244, 0.04) 100%);
        border-color: rgba(66, 133, 244, 0.3);
        box-shadow: 0 4px 16px rgba(66, 133, 244, 0.12);
        transform: translateY(-1px);
      }

      .tenant-current-compact {
        display: flex;
        align-items: center;
        gap: 12px;

        .tenant-current-icon {
          width: 38px;
          height: 38px;
          display: flex;
          align-items: center;
          justify-content: center;
          background: var(--gradient-primary);
          border-radius: 10px;
          color: white;
          box-shadow: 0 4px 12px rgba(66, 133, 244, 0.2);
          flex-shrink: 0;

          .tenant-icon-text {
            font-size: 15px;
            font-weight: 700;
          }
        }

        .tenant-current-info {
          display: flex;
          flex-direction: column;
          gap: 2px;

          .tenant-current-label {
            font-size: 11px;
            font-weight: 600;
            color: var(--text-secondary);
            text-transform: uppercase;
            letter-spacing: 0.3px;
          }

          .tenant-current-name {
            font-size: 14px;
            font-weight: 600;
            color: var(--text-primary);
            line-height: 1.2;
          }
        }
      }

      .tenant-switch-icon {
        font-size: 16px;
        color: var(--primary-color);
        opacity: 0.7;
        transition: transform 0.2s, opacity 0.2s;
        flex-shrink: 0;
      }

      &:hover .tenant-switch-icon {
        opacity: 1;
        transform: translateX(3px);
      }
    }
  }
}

// 租户选择器对话框样式
.tenant-selector-dialog {
  :deep(.el-dialog__header) {
    padding: 20px 24px 16px;
    border-bottom: 1px solid var(--border-base);

    .el-dialog__title {
      font-size: 17px;
      font-weight: 700;
      color: var(--text-primary);
    }
  }

  :deep(.el-dialog__body) {
    padding: 20px 24px 24px;
  }

  .tenant-selector-content {
    .tenant-selector-list {
      display: flex;
      flex-direction: column;
      gap: 10px;

      .tenant-selector-item {
        display: flex;
        align-items: center;
        gap: 12px;
        padding: 14px;
        background: var(--bg-white);
        border: 1px solid var(--border-base);
        border-radius: 14px;
        cursor: pointer;
        transition: var(--transition-base);

        &:hover {
          border-color: rgba(66, 133, 244, 0.3);
          box-shadow: 0 6px 20px rgba(66, 133, 244, 0.12);
          transform: translateY(-1px);
        }

        &.active {
          background: rgba(66, 133, 244, 0.06);
          border-color: rgba(66, 133, 244, 0.3);

          .tenant-selector-avatar {
            background: var(--gradient-primary);
            color: white;
            box-shadow: 0 6px 18px rgba(66, 133, 244, 0.22);
          }

          .tenant-selector-name {
            color: var(--primary-color);
          }
        }

        .tenant-selector-avatar {
          width: 42px;
          height: 42px;
          display: flex;
          align-items: center;
          justify-content: center;
          background: var(--bg-light);
          border-radius: 12px;
          color: var(--text-secondary);
          font-size: 16px;
          font-weight: 700;
          transition: var(--transition-base);
          flex-shrink: 0;

          .tenant-selector-avatar-text {
            text-transform: uppercase;
          }
        }

        .tenant-selector-info {
          flex: 1;
          min-width: 0;

          .tenant-selector-name {
            font-size: 14px;
            font-weight: 600;
            color: var(--text-primary);
            line-height: 1.4;
            margin-bottom: 2px;
          }

          .tenant-selector-code {
            font-size: 12px;
            color: var(--text-secondary);
            font-weight: 500;
            text-transform: lowercase;
          }
        }

        .tenant-selector-status {
          display: flex;
          align-items: center;
          gap: 8px;
          flex-shrink: 0;

          .tenant-selector-badge {
            height: 22px;
            padding: 0 8px;
            border-radius: 999px;
            font-size: 11px;
            font-weight: 700;
            line-height: 22px;
            border: 1px solid var(--border-base);
            background: var(--bg-light);
            color: var(--text-secondary);

            &.danger {
              border-color: rgba(239, 68, 68, 0.22);
              background: rgba(239, 68, 68, 0.1);
              color: var(--danger-color);
            }
          }

          .tenant-selector-check {
            width: 24px;
            height: 24px;
            border-radius: 999px;
            background: var(--gradient-primary);
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
            font-size: 14px;
            box-shadow: 0 4px 12px rgba(66, 133, 244, 0.25);
          }
        }
      }
    }
  }
}
</style>
