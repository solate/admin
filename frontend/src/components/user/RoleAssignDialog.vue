<template>
  <el-dialog
    v-model="visible"
    title="分配角色"
    width="600px"
    @close="handleClose"
  >
    <div class="role-assign-dialog">
      <div class="user-info">
        <el-avatar :size="50">{{ currentUser?.user_name?.charAt(0)?.toUpperCase() || '?' }}</el-avatar>
        <div class="info">
          <div class="username">{{ currentUser?.user_name }}</div>
          <div class="email">{{ currentUser?.email || '-' }}</div>
        </div>
      </div>

      <el-divider />

      <div class="role-selection">
        <div class="section-title">
          <span>选择角色</span>
          <el-tag type="info" size="small">可多选</el-tag>
        </div>

        <el-checkbox-group v-model="selectedRoleCodes" class="role-checkbox-group">
          <el-checkbox
            v-for="role in roleOptions"
            :key="role.role_code"
            :label="role.role_code"
            class="role-checkbox"
          >
            <div class="role-item">
              <div class="role-name">{{ role.name }}</div>
              <div class="role-code">{{ role.role_code }}</div>
              <div v-if="role.description" class="role-description">{{ role.description }}</div>
            </div>
          </el-checkbox>
        </el-checkbox-group>
      </div>

      <div v-if="selectedRoles.length > 0" class="selected-roles">
        <div class="section-title">已选择角色 ({{ selectedRoles.length }})</div>
        <div class="selected-tags">
          <el-tag
            v-for="role in selectedRoles"
            :key="role.role_code"
            closable
            @close="removeRole(role.role_code)"
          >
            {{ role.name }}
          </el-tag>
        </div>
      </div>
    </div>

    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleConfirm">
        确定 (已选 {{ selectedRoles.length }} 个角色)
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { userApi, type AssignRolesRequest, type UserRoleInfo } from '@/api/user'
import { roleApi, type RoleInfo } from '@/api/role'

interface User {
  user_id: string
  user_name: string
  email?: string
}

interface Props {
  modelValue: boolean
  user: User | null
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void
  (e: 'success'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const currentUser = computed(() => props.user)

const loading = ref(false)
const roleOptions = ref<RoleInfo[]>([])
const selectedRoleCodes = ref<string[]>([])
const userRoles = ref<UserRoleInfo[]>([])

// 计算已选择的角色对象
const selectedRoles = computed(() => {
  return roleOptions.value.filter(role =>
    selectedRoleCodes.value.includes(role.role_code)
  )
})

// 获取所有角色列表
const fetchAllRoles = async () => {
  try {
    const response = await roleApi.getAllRoles()
    roleOptions.value = response.list
  } catch (error: any) {
    ElMessage.error(error?.message || '获取角色列表失败')
  }
}

// 获取用户的当前角色
const fetchUserRoles = async () => {
  if (!currentUser.value?.user_id) return

  try {
    const response = await userApi.getUserRoles(currentUser.value.user_id)
    userRoles.value = response.roles
    selectedRoleCodes.value = response.roles.map(role => role.role_code)
  } catch (error: any) {
    ElMessage.error(error?.message || '获取用户角色失败')
  }
}

// 移除角色
const removeRole = (roleCode: string) => {
  const index = selectedRoleCodes.value.indexOf(roleCode)
  if (index > -1) {
    selectedRoleCodes.value.splice(index, 1)
  }
}

// 确认分配
const handleConfirm = async () => {
  if (!currentUser.value?.user_id) return

  loading.value = true
  try {
    const data: AssignRolesRequest = {
      role_codes: selectedRoleCodes.value
    }

    await userApi.assignRoles(currentUser.value.user_id, data)
    ElMessage.success('角色分配成功')
    emit('success')
    handleClose()
  } catch (error: any) {
    ElMessage.error(error?.message || '角色分配失败')
  } finally {
    loading.value = false
  }
}

// 关闭对话框
const handleClose = () => {
  visible.value = false
  selectedRoleCodes.value = []
  userRoles.value = []
}

// 监听对话框打开
watch(() => props.modelValue, (newVal) => {
  if (newVal) {
    fetchAllRoles()
    fetchUserRoles()
  }
})
</script>

<style scoped lang="scss">
.role-assign-dialog {
  .user-info {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 16px;
    background: var(--bg-color-page);
    border-radius: 8px;

    .info {
      flex: 1;

      .username {
        font-size: 16px;
        font-weight: 600;
        color: var(--text-primary);
        margin-bottom: 4px;
      }

      .email {
        font-size: 14px;
        color: var(--text-secondary);
      }
    }
  }

  .role-selection {
    margin-top: 20px;

    .section-title {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 12px;
      font-size: 14px;
      font-weight: 600;
      color: var(--text-primary);
    }

    .role-checkbox-group {
      display: flex;
      flex-direction: column;
      gap: 12px;

      .role-checkbox {
        width: 100%;
        margin: 0;
        padding: 12px;
        border: 1px solid var(--border-lighter);
        border-radius: 8px;
        transition: all 0.3s ease;

        &:hover {
          border-color: var(--color-primary);
          background-color: var(--color-primary-light-9);
        }

        :deep(.el-checkbox__label) {
          flex: 1;
          padding-left: 12px;
        }

        .role-item {
          flex: 1;

          .role-name {
            font-size: 14px;
            font-weight: 600;
            color: var(--text-primary);
            margin-bottom: 4px;
          }

          .role-code {
            font-size: 12px;
            color: var(--text-secondary);
            margin-bottom: 4px;
          }

          .role-description {
            font-size: 12px;
            color: var(--text-regular);
            line-height: 1.5;
          }
        }
      }
    }
  }

  .selected-roles {
    margin-top: 24px;
    padding: 16px;
    background: var(--bg-color-page);
    border-radius: 8px;

    .section-title {
      font-size: 14px;
      font-weight: 600;
      color: var(--text-primary);
      margin-bottom: 12px;
    }

    .selected-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;

      .el-tag {
        margin: 0;
      }
    }
  }
}
</style>
