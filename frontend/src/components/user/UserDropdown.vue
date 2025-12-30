<template>
  <el-dropdown trigger="click" @command="handleCommand" @visible-change="handleVisibleChange" class="user-dropdown">
    <div class="user-info" :class="{ 'is-active': isOpen }">
      <el-avatar :size="36" class="user-avatar">
        {{ userInfo?.user_name?.charAt(0).toUpperCase() || 'A' }}
      </el-avatar>
      <el-icon class="dropdown-icon" :size="14" :class="{ 'is-rotated': isOpen }">
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
              <div class="user-profile-name">{{ userInfo?.nickname || userInfo?.user_name || '管理员' }}</div>
              <!-- 角色标签 -->
              <div v-if="roles.length > 0" class="user-profile-roles">
                <span v-for="role in roles" :key="role.role_id" class="role-tag-inline">
                  {{ role.name }}
                </span>
              </div>
              <div class="user-profile-email">{{ userInfo?.email || userInfo?.phone || '未设置联系方式' }}</div>
            </div>
          </div>
        </div>

        <!-- 租户信息卡片 -->
        <div class="tenant-section-compact">
          <div class="tenant-switcher-compact">
            <div class="tenant-current-compact">
              <div class="tenant-current-icon">
                <span class="tenant-icon-text">{{ tenant?.name?.charAt(0) || 'T' }}</span>
              </div>
              <div class="tenant-current-info">
                <div class="tenant-current-name">{{ tenant?.name || '默认租户' }}</div>
                <div class="tenant-current-code">{{ tenant?.tenant_code || 'default' }}</div>
              </div>
            </div>
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
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import { authApi } from '../../api'
import { clearTokens } from '../../utils/token'

// 定义 emit
const emit = defineEmits<{
  command: [command: string]
}>()

// Props
interface Props {
  userInfo?: {
    user_name?: string
    nickname?: string
    email?: string
    phone?: string
  } | null
  tenant?: {
    name?: string
    tenant_code?: string
  } | null
  roles?: Array<{
    role_id: string
    name: string
  }>
}

const props = withDefaults(defineProps<Props>(), {
  userInfo: null,
  tenant: null,
  roles: () => []
})

const router = useRouter()
const isOpen = ref(false)

function handleVisibleChange(visible: boolean) {
  isOpen.value = visible
}

async function handleCommand(command: string) {
  isOpen.value = false
  emit('command', command)

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
</script>

<style scoped lang="scss">
.user-dropdown {
  .user-info {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 8px;
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
          margin-bottom: 6px;
        }

        .user-profile-roles {
          display: flex;
          flex-wrap: wrap;
          gap: 6px;
          margin-bottom: 6px;

          .role-tag-inline {
            display: inline-flex;
            align-items: center;
            padding: 3px 10px;
            background: linear-gradient(135deg, rgba(66, 133, 244, 0.12) 0%, rgba(66, 133, 244, 0.06) 100%);
            border: 1px solid rgba(66, 133, 244, 0.2);
            border-radius: 6px;
            font-size: 12px;
            font-weight: 600;
            color: var(--primary-color);
            line-height: 1.2;
          }
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

  // 租户区域
  .tenant-section-compact {
    margin: 4px 4px 8px;
    padding: 0;

    .tenant-switcher-compact {
      display: flex;
      align-items: center;
      justify-content: flex-start;
      padding: 12px 14px;
      background: linear-gradient(135deg, rgba(66, 133, 244, 0.06) 0%, rgba(66, 133, 244, 0.02) 100%);
      border: 1px solid rgba(66, 133, 244, 0.15);
      border-radius: 12px;

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

          .tenant-current-name {
            font-size: 14px;
            font-weight: 600;
            color: var(--text-primary);
            line-height: 1.2;
          }

          .tenant-current-code {
            font-size: 12px;
            color: var(--text-secondary);
            font-weight: 500;
          }
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
}
</style>
