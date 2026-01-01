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
            <div class="tenant-selector-code">{{ tenant.tenant_code }}</div>
          </div>
          <div class="tenant-selector-status">
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
import { authApi, type TenantInfo } from '../../api'
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
const tenantList = ref<TenantInfo[]>([])

// 加载租户列表
async function loadTenantList() {
  loading.value = true
  try {
    const response = await authApi.getAvailableTenants()
    tenantList.value = response.tenants || []
  } catch (error: any) {
    console.error('获取租户列表失败:', error)
    ElMessage.error(error.message || '获取租户列表失败')
    tenantList.value = []
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
    // 1. 调用切换租户接口获取新 token
    const switchResponse = await authApi.switchTenant({ tenant_id: tenantId })

    // 2. 保存新 token
    localStorage.setItem('access_token', switchResponse.access_token)
    localStorage.setItem('refresh_token', switchResponse.refresh_token)

    // 3. 调用 profile 接口获取完整的用户信息（包括新租户信息）
    const profileResponse = await authApi.getProfile()

    // 4. 保存完整的用户信息
    saveTokens({
      access_token: switchResponse.access_token,
      refresh_token: switchResponse.refresh_token,
      user_id: profileResponse.user.user_id,
      username: profileResponse.user.username,
      email: profileResponse.user.email,
      phone: profileResponse.user.phone,
      tenant: profileResponse.tenant,
      roles: profileResponse.roles
    })

    emit('tenant-changed', profileResponse.tenant)
    close()
    ElMessage.success(`已切换到 ${profileResponse.tenant.name}`)
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
