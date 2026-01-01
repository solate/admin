import { createRouter, createWebHistory } from 'vue-router'
import { ensureValidToken, clearTokens } from '../utils/token'

// 动态导入组件
const Login = () => import('../views/Login.vue')
const Layout = () => import('../views/Layout.vue')
const Dashboard = () => import('../views/Dashboard.vue')

// 动态导入系统管理相关页面
const SystemUsers = () => import('../views/system/Users.vue')
const SystemRoles = () => import('../views/system/Roles.vue')
const SystemMenus = () => import('../views/system/Menus.vue')
const SystemTenants = () => import('../views/system/Tenants.vue')
const SystemTenantMembers = () => import('../views/system/TenantMembers.vue')
const SystemLoginLogs = () => import('../views/system/LoginLogs.vue')
const SystemOperationLogs = () => import('../views/system/OperationLogs.vue')

const routes = [
  { path: '/login/:tenantCode', name: 'login', component: Login, meta: { public: true, title: '登录' } },
  { path: '/login', redirect: '/login/default' },  // 默认重定向到 /login/default
  {
    path: '/',
    component: Layout,
    children: [
      { path: '', name: 'dashboard', component: Dashboard, meta: { title: '仪表盘' } },
      { path: 'login-logs', name: 'login-logs', component: SystemLoginLogs, meta: { title: '登录日志' } },
      { path: 'operation-logs', name: 'operation-logs', component: SystemOperationLogs, meta: { title: '操作日志' } },
      { path: 'system/users', name: 'users', component: SystemUsers, meta: { title: '用户管理' } },
      { path: 'system/roles', name: 'roles', component: SystemRoles, meta: { title: '角色管理' } },
      { path: 'system/menus', name: 'menus', component: SystemMenus, meta: { title: '菜单管理' } },
      { path: 'system/tenants', name: 'tenants', component: SystemTenants, meta: { title: '租户管理' } },
      { path: 'system/tenant-members', name: 'tenant-members', component: SystemTenantMembers, meta: { title: '租户成员管理' } }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to, _from, next) => {
  // 公开路由直接放行
  if (to.meta.public) return next()

  // 检查 token 是否存在
  const token = localStorage.getItem('access_token')
  if (!token) {
    return next({ path: '/login/default', query: { redirect: to.fullPath } })
  }

  // 确保 token 有效（自动刷新过期或即将过期的 token）
  const isValid = await ensureValidToken()
  if (!isValid) {
    // token 刷新失败，清除并跳转登录
    clearTokens()
    return next({ path: '/login/default', query: { redirect: to.fullPath } })
  }

  next()
})

export default router
