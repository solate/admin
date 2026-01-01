import { createRouter, createWebHistory } from 'vue-router'
import { ensureValidToken, clearTokens } from '../utils/token'

// 动态导入组件
const Login = () => import('../views/Login.vue')
const Layout = () => import('../views/Layout.vue')
const Dashboard = () => import('../views/Dashboard.vue')

// 租户管理
const TenantList = () => import('../views/tenant/TenantList.vue')
const TenantPackages = () => import('../views/tenant/TenantPackages.vue')
const TenantSubscription = () => import('../views/tenant/TenantSubscription.vue')
const TenantBilling = () => import('../views/tenant/TenantBilling.vue')

// 组织架构
const OrganizationDepartments = () => import('../views/organization/Departments.vue')
const OrganizationPositions = () => import('../views/organization/Positions.vue')

// 用户与权限
const AccessUsers = () => import('../views/access/Users.vue')
const AccessRoles = () => import('../views/access/Roles.vue')
const AccessMenus = () => import('../views/access/Menus.vue')
const AccessDataPermissions = () => import('../views/access/DataPermissions.vue')

// 业务管理
const BusinessFactories = () => import('../views/business/Factories.vue')
const BusinessProducts = () => import('../views/business/Products.vue')
const BusinessOrders = () => import('../views/business/Orders.vue')
const BusinessStatistics = () => import('../views/business/Statistics.vue')

// 审计日志
const AuditLogin = () => import('../views/audit/LoginLogs.vue')
const AuditOperation = () => import('../views/audit/OperationLogs.vue')
const AuditData = () => import('../views/audit/DataLogs.vue')

// 系统设置
const SettingsDictionary = () => import('../views/settings/Dictionary.vue')
const SettingsParameters = () => import('../views/settings/Parameters.vue')
const SettingsNotifications = () => import('../views/settings/Notifications.vue')
const SettingsStorage = () => import('../views/settings/Storage.vue')
const SettingsMonitor = () => import('../views/settings/Monitor.vue')

const routes = [
  { path: '/login/:tenantCode', name: 'login', component: Login, meta: { public: true, title: '登录' } },
  { path: '/login', redirect: '/login/default' },  // 默认重定向到 /login/default
  {
    path: '/',
    component: Layout,
    children: [
      // 工作台
      { path: '', name: 'dashboard', component: Dashboard, meta: { title: '工作台' } },

      // 租户管理
      { path: 'tenant/list', name: 'tenant-list', component: TenantList, meta: { title: '租户列表' } },
      { path: 'tenant/packages', name: 'tenant-packages', component: TenantPackages, meta: { title: '套餐管理' } },
      { path: 'tenant/subscription', name: 'tenant-subscription', component: TenantSubscription, meta: { title: '订阅管理' } },
      { path: 'tenant/billing', name: 'tenant-billing', component: TenantBilling, meta: { title: '账单管理' } },

      // 组织架构
      { path: 'organization/departments', name: 'organization-departments', component: OrganizationDepartments, meta: { title: '部门管理' } },
      { path: 'organization/positions', name: 'organization-positions', component: OrganizationPositions, meta: { title: '岗位管理' } },

      // 用户与权限
      { path: 'access/users', name: 'access-users', component: AccessUsers, meta: { title: '用户管理' } },
      { path: 'access/roles', name: 'access-roles', component: AccessRoles, meta: { title: '角色管理' } },
      { path: 'access/menus', name: 'access-menus', component: AccessMenus, meta: { title: '菜单权限' } },
      { path: 'access/data-permissions', name: 'access-data-permissions', component: AccessDataPermissions, meta: { title: '数据权限' } },

      // 业务管理
      { path: 'business/factories', name: 'business-factories', component: BusinessFactories, meta: { title: '工厂管理' } },
      { path: 'business/products', name: 'business-products', component: BusinessProducts, meta: { title: '商品管理' } },
      { path: 'business/orders', name: 'business-orders', component: BusinessOrders, meta: { title: '订单管理' } },
      { path: 'business/statistics', name: 'business-statistics', component: BusinessStatistics, meta: { title: '数据统计' } },

      // 审计日志
      { path: 'audit/login', name: 'audit-login', component: AuditLogin, meta: { title: '登录日志' } },
      { path: 'audit/operation', name: 'audit-operation', component: AuditOperation, meta: { title: '操作日志' } },
      { path: 'audit/data', name: 'audit-data', component: AuditData, meta: { title: '数据变更' } },

      // 系统设置
      { path: 'settings/dictionary', name: 'settings-dictionary', component: SettingsDictionary, meta: { title: '字典管理' } },
      { path: 'settings/parameters', name: 'settings-parameters', component: SettingsParameters, meta: { title: '系统参数' } },
      { path: 'settings/notifications', name: 'settings-notifications', component: SettingsNotifications, meta: { title: '通知配置' } },
      { path: 'settings/storage', name: 'settings-storage', component: SettingsStorage, meta: { title: '存储配置' } },
      { path: 'settings/monitor', name: 'settings-monitor', component: SettingsMonitor, meta: { title: '系统监控' } },

      // 兼容旧路由（重定向到新路由）
      { path: 'login-logs', redirect: '/audit/login' },
      { path: 'operation-logs', redirect: '/audit/operation' },
      { path: 'system/users', redirect: '/access/users' },
      { path: 'system/roles', redirect: '/access/roles' },
      { path: 'system/menus', redirect: '/access/menus' },
      { path: 'system/tenants', redirect: '/tenant/list' },
      { path: 'system/tenant-members', redirect: '/tenant/list' }
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
