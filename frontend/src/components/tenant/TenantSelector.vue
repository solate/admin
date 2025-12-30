<template>
  <!-- 对话框选择模式 -->
  <el-dialog
    v-model="dialogVisible"
    title="切换租户"
    width="420px"
    :close-on-click-modal="true"
    class="tenant-selector-dialog"
    @open="handleDialogOpen"
  >
    <div v-if="loading" class="tenant-loading">
      <el-icon class="is-loading" :size="24"><Loading /></el-icon>
      <span>加载中...</span>
    </div>
    <div v-else class="tenant-selector-content">
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
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Check, Loading } from '@element-plus/icons-vue'
import { authApi } from '../../api'
import { saveTokens, getUserInfo } from '../../utils/token'

// Props
interface Props {
  currentTenant?: {
    tenant_id?: string
    name?: string
    tenant_code?: string
  } | null
}

const props = withDefaults(defineProps<Props>(), {
  currentTenant: null
})

// Emit
const emit = defineEmits<{
  'tenant-changed': [tenant: { tenant_id: string; name: string; tenant_code: string }]
}>()

const userInfo = getUserInfo()
const dialogVisible = ref(false)
const loading = ref(false)
const tenantList = ref<Array<{
  tenant_id: string
  name: string
  code: string
  status: number
}>>([])

// 加载租户列表
async function loadTenantList() {
  loading.value = true
  try {
    // TODO: 临时假数据，实际应该调用 API
    // const response = await authApi.getUserTenants(userInfo.user_id)
    // tenantList.value = response.data || []

    // 假数据
    tenantList.value = [
      { tenant_id: '1', name: '默认租户', code: 'default', status: 1 },
      { tenant_id: '2', name: '测试租户A', code: 'tenant-a', status: 1 },
      { tenant_id: '3', name: '生产租户B', code: 'tenant-b', status: 1 },
      { tenant_id: '4', name: '开发租户C', code: 'tenant-c', status: 2 }
    ]
  } catch (error) {
    console.error('获取租户列表失败:', error)
    ElMessage.error('获取租户列表失败')
  } finally {
    loading.value = false
  }
}

// 暴露方法供父组件调用（对话框模式）
function open() {
  dialogVisible.value = true
}

function close() {
  dialogVisible.value = false
}

function handleDialogOpen() {
  if (tenantList.value.length === 0) {
    loadTenantList()
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

  loading.value = true
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
  } finally {
    loading.value = false
  }
}

// 暴露方法
defineExpose({
  open,
  close
})
</script>

<style scoped lang="scss">
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

  .tenant-loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 40px 20px;
    color: var(--text-secondary);
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
