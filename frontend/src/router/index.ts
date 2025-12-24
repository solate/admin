import { createRouter, createWebHistory } from 'vue-router'

// 动态导入组件
const Login = () => import('../views/Login.vue')
const Layout = () => import('../views/Layout.vue')
const Dashboard = () => import('../views/Dashboard.vue')
const Factories = () => import('../views/Factories.vue')
const Products = () => import('../views/Products.vue')
const Statistics = () => import('../views/Statistics.vue')

// 动态导入注册组件
const Register = () => import('../views/Register.vue')

// 动态导入系统管理相关页面
const SystemUsers = () => import('../views/system/Users.vue')
const SystemRoles = () => import('../views/system/Roles.vue')
const SystemTenants = () => import('../views/system/Tenants.vue')
const SystemTenantMembers = () => import('../views/system/TenantMembers.vue')
const SystemPermissions = () => import('../views/system/Permissions.vue')
const SystemDict = () => import('../views/system/Dict.vue')
const SystemLogs = () => import('../views/system/Logs.vue')
const SystemMonitor = () => import('../views/system/Monitor.vue')

const routes = [
  { path: '/login', name: 'login', component: Login, meta: { public: true, title: '登录' } },
  { path: '/register', name: 'register', component: Register, meta: { public: true, title: '注册' } },
  {
    path: '/',
    component: Layout,
    children: [
      { path: '', name: 'dashboard', component: Dashboard, meta: { title: '仪表盘' } },
      { path: 'system/users', name: 'users', component: SystemUsers, meta: { title: '用户管理' } },
      { path: 'system/roles', name: 'roles', component: SystemRoles, meta: { title: '角色管理' } },
      { path: 'system/tenants', name: 'tenants', component: SystemTenants, meta: { title: '租户管理' } },
      { path: 'system/tenant-members', name: 'tenant-members', component: SystemTenantMembers, meta: { title: '租户成员管理' } },
      { path: 'system/permissions/menu', name: 'permissions-menu', component: SystemPermissions, meta: { title: '菜单权限' } },
      { path: 'system/permissions/api', name: 'permissions-api', component: SystemPermissions, meta: { title: '接口权限' } },
      { path: 'system/permissions/data', name: 'permissions-data', component: SystemPermissions, meta: { title: '数据权限' } },
      { path: 'system/dict', name: 'system-dict', component: SystemDict, meta: { title: '字典管理' } },
      { path: 'system/logs', name: 'system-logs', component: SystemLogs, meta: { title: '操作日志' } },
      { path: 'system/monitor', name: 'system-monitor', component: SystemMonitor, meta: { title: '系统监控' } },
      { path: 'factories', name: 'factories', component: Factories, meta: { title: '工厂管理' } },
      { path: 'products', name: 'products', component: Products, meta: { title: '商品管理' } },
      { path: 'statistics', name: 'statistics', component: Statistics, meta: { title: '数据统计' } }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, _from, next) => {
  if (to.meta.public) return next()
  const token = localStorage.getItem('access_token')  // 修改为 access_token
  if (!token) return next({ path: '/login', query: { redirect: to.fullPath } })
  next()
})

export default router


