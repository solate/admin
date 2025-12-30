<template>
  <!-- 下拉选择模式 -->
  <el-dropdown
    v-if="mode === 'dropdown'"
    trigger="click"
    @command="handleCommand"
    @visible-change="handleVisibleChange"
    class="tenant-selector-dropdown"
  >
    <div class="tenant-trigger" :class="{ 'is-active': isOpen }">
      <div class="tenant-current">
        <div class="tenant-avatar">
          <span class="avatar-text">{{ currentTenant?.name?.charAt(0)?.toUpperCase() || 'T' }}</span>
        </div>
        <div class="tenant-info">
          <div class="tenant-name">{{ currentTenant?.name || '默认租户' }}</div>
          <div class="tenant-code">{{ currentTenant?.tenant_code || 'default' }}</div>
        </div>
      </div>
      <el-icon class="dropdown-icon" :size="14" :class="{ 'is-rotated': isOpen }">
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
              <div class="workspace-item" :class="{ active: tenant.tenant_id === currentTenant?.tenant_id }">
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

  <!-- 对话框选择模式 -->
  <el-dialog
    v-else-if="mode === 'dialog'"
    v-model="dialogVisible"
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
          :class="{ active: tenant.tenant_id === currentTenant?.tenant_id }"
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
            <el-icon v-if="tenant.tenant_id === currentTenant?.tenant_id" class="tenant-selector-check">
              <Check />
            </el-icon>
          </div>
        </div>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { CaretBottom, Check, Plus } from '@element-plus/icons-vue'
import { authApi } from '../../api'
import { saveTokens, getUserInfo } from '../../utils/token'

// Props
interface Props {
  mode?: 'dropdown' | 'dialog'  // 选择模式：下拉选择或对话框
  currentTenant?: {
    tenant_id?: string
    name?: string
    tenant_code?: string
  } | null
  tenantList?: Array<{
    tenant_id: string
    name: string
    code: string
    status: number
  }>
}

const props = withDefaults(defineProps<Props>(), {
  mode: 'dropdown',
  currentTenant: null,
  tenantList: () => []
})

// Emit
const emit = defineEmits<{
  'tenant-changed': [tenant: { tenant_id: string; name: string; tenant_code: string }]
}>()

const router = useRouter()
const isOpen = ref(false)
const dialogVisible = ref(false)
const userInfo = getUserInfo()

// 暴露方法供父组件调用（对话框模式）
function open() {
  dialogVisible.value = true
}

function close() {
  dialogVisible.value = false
}

function handleVisibleChange(visible: boolean) {
  isOpen.value = visible
}

async function handleCommand(command: string) {
  if (command === '__manage_tenants__') {
    router.push('/system/tenants')
    return
  }

  if (command === props.currentTenant?.tenant_id) return

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
      tenant: response.tenant,
      roles: response.roles
    })

    emit('tenant-changed', response.tenant)
    ElMessage.success('租户切换成功')
    location.reload()
  } catch (error: any) {
    ElMessage.error(error?.message || '租户切换失败')
  }
}

async function handleTenantSelect(tenantId: string) {
  if (!userInfo?.user_id) {
    ElMessage.error('用户信息缺失，请重新登录')
    return
  }

  if (tenantId === props.currentTenant?.tenant_id) {
    close()
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
      tenant: response.tenant,
      roles: response.roles
    })

    emit('tenant-changed', response.tenant)
    close()
    ElMessage.success('租户切换成功')
    location.reload()
  } catch (error: any) {
    ElMessage.error(error?.message || '租户切换失败')
  }
}

// 暴露方法
defineExpose({
  open,
  close
})
</script>

<style scoped lang="scss">
.tenant-selector-dropdown {
  .tenant-trigger {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    margin: 0 12px;
    border-radius: 12px;
    cursor: pointer;
    transition: var(--transition-base);
    background: transparent;

    &:hover {
      background: rgba(0, 0, 0, 0.03);
    }

    &.is-active {
      background: rgba(0, 0, 0, 0.04);
    }

    .tenant-current {
      display: flex;
      align-items: center;
      gap: 12px;
      flex: 1;
      min-width: 0;

      .tenant-avatar {
        width: 36px;
        height: 36px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--gradient-primary);
        border-radius: 10px;
        color: white;
        box-shadow: 0 4px 12px rgba(66, 133, 244, 0.2);
        flex-shrink: 0;

        .avatar-text {
          font-size: 14px;
          font-weight: 700;
        }
      }

      .tenant-info {
        display: flex;
        flex-direction: column;
        gap: 2px;
        min-width: 0;

        .tenant-name {
          font-size: 14px;
          font-weight: 600;
          color: var(--text-primary);
          line-height: 1.2;
        }

        .tenant-code {
          font-size: 12px;
          color: var(--text-secondary);
          line-height: 1.2;
        }
      }
    }

    .dropdown-icon {
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

// 租户下拉菜单样式
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
    }
  }

  .workspace-scroll {
    :deep(.el-scrollbar__wrap) {
      max-height: 320px;
    }
  }

  .workspace-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-bottom: 8px;

    .workspace-dropdown-item {
      padding: 0;
      margin: 0;

      :deep(.el-dropdown-menu__item) {
        padding: 0;
        border-radius: 10px;
        margin: 0;
        line-height: normal;
      }
    }
  }

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

      .workspace-check {
        opacity: 1;
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
      gap: 8px;

      .workspace-badge {
        padding: 3px 8px;
        border-radius: 6px;
        font-size: 11px;
        font-weight: 600;
        background: rgba(103, 194, 58, 0.1);
        color: var(--success-color);
        border: 1px solid rgba(103, 194, 58, 0.2);

        &.danger {
          background: rgba(239, 68, 68, 0.1);
          color: var(--danger-color);
          border-color: rgba(239, 68, 68, 0.2);
        }
      }

      .workspace-check {
        width: 20px;
        height: 20px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--gradient-primary);
        border-radius: 50%;
        color: white;
        opacity: 0;
        transition: var(--transition-base);

        .check-icon {
          font-size: 14px;
        }
      }
    }
  }

  .workspace-add {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 14px;
    border-radius: 12px;
    cursor: pointer;
    transition: var(--transition-base);
    border: 1.5px dashed var(--border-base);
    width: 100%;

    &:hover {
      border-color: var(--primary-color);
      background: rgba(66, 133, 244, 0.04);

      .workspace-add-icon {
        background: var(--gradient-primary);
        color: white;
      }

      .workspace-add-text {
        color: var(--primary-color);
      }
    }

    .workspace-add-icon {
      width: 36px;
      height: 36px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: var(--bg-light);
      border-radius: 10px;
      color: var(--text-secondary);
      font-size: 16px;
      transition: var(--transition-base);
    }

    .workspace-add-text {
      font-size: 14px;
      font-weight: 600;
      color: var(--text-secondary);
      transition: var(--transition-base);
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
        gap: 14px;
        padding: 14px;
        background: var(--bg-white);
        border: 1.5px solid var(--border-base);
        border-radius: 14px;
        cursor: pointer;
        transition: var(--transition-base);

        &:hover {
          border-color: rgba(66, 133, 244, 0.25);
          box-shadow: 0 14px 30px rgba(66, 133, 244, 0.12);
          transform: translateY(-1px);
        }

        &.active {
          background: rgba(66, 133, 244, 0.06);
          border-color: var(--primary-color);

          .tenant-selector-avatar {
            background: var(--gradient-primary);
            color: white;
            box-shadow: 0 14px 28px rgba(66, 133, 244, 0.22);
          }

          .tenant-selector-name {
            color: var(--primary-color);
          }
        }

        .tenant-selector-avatar {
          width: 44px;
          height: 44px;
          display: flex;
          align-items: center;
          justify-content: center;
          background: var(--bg-light);
          border-radius: 14px;
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
            font-size: 15px;
            font-weight: 700;
            color: var(--text-primary);
            line-height: 1.3;
            margin-bottom: 3px;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
          }

          .tenant-selector-code {
            font-size: 13px;
            color: var(--text-secondary);
            font-weight: 500;
            text-transform: lowercase;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
          }
        }

        .tenant-selector-status {
          display: flex;
          align-items: center;
          gap: 10px;
          flex-shrink: 0;

          .tenant-selector-badge {
            padding: 4px 10px;
            border-radius: 6px;
            font-size: 12px;
            font-weight: 600;
            background: rgba(103, 194, 58, 0.1);
            color: var(--success-color);
            border: 1px solid rgba(103, 194, 58, 0.2);

            &.danger {
              background: rgba(239, 68, 68, 0.1);
              color: var(--danger-color);
              border-color: rgba(239, 68, 68, 0.2);
            }
          }

          .tenant-selector-check {
            width: 22px;
            height: 22px;
            display: flex;
            align-items: center;
            justify-content: center;
            background: var(--gradient-primary);
            border-radius: 50%;
            color: white;
            font-size: 14px;
          }
        }
      }
    }
  }
}
</style>
