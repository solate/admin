import { createRouter, createWebHistory } from 'vue-router'
import { ensureValidToken, clearTokens } from '../utils/token'
import { useMenuStore } from '@/stores/menu'

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

// 组件映射表（用于动态路由）
const componentMap: Record<string, () => Promise<any>> = {
  Dashboard,
  TenantList,
  TenantPackages,
  TenantSubscription,
  TenantBilling,
  OrganizationDepartments,
  OrganizationPositions,
  AccessUsers,
  AccessRoles,
  AccessMenus,
  AccessDataPermissions,
  BusinessFactories,
  BusinessProducts,
  BusinessOrders,
  BusinessStatistics,
  AuditLogin,
  AuditOperation,
  AuditData,
  SettingsDictionary,
  SettingsParameters,
  SettingsNotifications,
  SettingsStorage,
  SettingsMonitor
}

// 基础路由（不需要权限）
const constantRoutes = [
  { path: '/login/:tenantCode', name: 'login', component: Login, meta: { public: true, title: '登录' } },
  { path: '/login', redirect: '/login/default' },
  {
    path: '/',
    name: 'layout', // 添加 name 以便动态路由可以引用
    component: Layout,
    redirect: '/dashboard',
    children: [
      { path: 'dashboard', name: 'dashboard', component: Dashboard, meta: { title: '工作台', icon: 'Dashboard' } }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes: constantRoutes
})

// 动态添加路由
export function addDynamicRoutes(menus: any[]) {
  menus.forEach(menu => {
    // 跳过没有 path 的菜单（通常是父级菜单）
    if (!menu.path) return

    // 根据 path 查找对应的组件
    let component = Dashboard
    const path = menu.path?.replace(/^\//, '') || ''

    // 尝试从 componentMap 中查找组件
    const componentName = pathToComponentName(path)
    if (componentName && componentMap[componentName]) {
      component = componentMap[componentName]
    }

    const route = {
      path: path,
      name: menu.menu_id || `route-${menu.name}`,
      component,
      meta: {
        title: menu.name,
        icon: menu.icon,
        menuId: menu.menu_id
      }
    }

    // 添加路由到 Layout 的 children（使用父路由名称 'layout'）
    router.addRoute('layout', route)
  })

  // 添加 404 路由（放在最后）
  router.addRoute('layout', {
    path: ':pathMatch(.*)*',
    redirect: '/dashboard'
  })
}

// 路径转组件名
function pathToComponentName(path: string): string {
  const mapping: Record<string, string> = {
    'tenant/list': 'TenantList',
    'tenant/packages': 'TenantPackages',
    'tenant/subscription': 'TenantSubscription',
    'tenant/billing': 'TenantBilling',
    'organization/departments': 'OrganizationDepartments',
    'organization/positions': 'OrganizationPositions',
    'access/users': 'AccessUsers',
    'access/roles': 'AccessRoles',
    'access/menus': 'AccessMenus',
    'access/data-permissions': 'AccessDataPermissions',
    'business/factories': 'BusinessFactories',
    'business/products': 'BusinessProducts',
    'business/orders': 'BusinessOrders',
    'business/statistics': 'BusinessStatistics',
    'audit/login': 'AuditLogin',
    'audit/operation': 'AuditOperation',
    'audit/data': 'AuditData',
    'settings/dictionary': 'SettingsDictionary',
    'settings/parameters': 'SettingsParameters',
    'settings/notifications': 'SettingsNotifications',
    'settings/storage': 'SettingsStorage',
    'settings/monitor': 'SettingsMonitor'
  }

  return mapping[path] || ''
}

// 重置路由（用于登出）
export function resetRouter() {
  const newRouter = createRouter({
    history: createWebHistory(),
    routes: constantRoutes
  })
  ;(router as any).matcher = (newRouter as any).matcher
}

// 路由守卫
router.beforeEach(async (to, _from, next) => {
  // 公开路由直接放行
  if (to.meta.public) {
    return next()
  }

  // 检查 token 是否存在
  const token = localStorage.getItem('access_token')
  if (!token) {
    return next({ path: '/login/default', query: { redirect: to.fullPath } })
  }

  // 确保 token 有效（自动刷新过期或即将过期的 token）
  const isValid = await ensureValidToken()
  if (!isValid) {
    clearTokens()
    return next({ path: '/login/default', query: { redirect: to.fullPath } })
  }

  // 加载用户菜单（只加载一次）
  const menuStore = useMenuStore()
  if (!menuStore.menuLoaded) {
    try {
      await menuStore.loadUserMenus()
      // 动态添加路由
      const flatMenus = menuStore.flattenMenus()
      addDynamicRoutes(flatMenus)
      // 重新进入当前路由
      return next({ ...to, replace: true })
    } catch (error) {
      console.error('加载菜单失败:', error)
      // 菜单加载失败，仍然允许访问基础页面
    }
  }

  next()
})

export default router
