import { defineStore } from 'pinia'
import { ref } from 'vue'
import { userMenuApi, type MenuTreeNode } from '@/api/menu'

export const useMenuStore = defineStore('menu', () => {
  const menus = ref<MenuTreeNode[]>([])
  const menuLoaded = ref(false)
  const loading = ref(false)

  // 加载用户菜单
  const loadUserMenus = async () => {
    if (loading.value) return

    loading.value = true
    try {
      const res = await userMenuApi.getUserMenus()
      menus.value = res.list || []
      menuLoaded.value = true
    } catch (error) {
      console.error('加载用户菜单失败:', error)
      menus.value = []
      menuLoaded.value = false
    } finally {
      loading.value = false
    }
  }

  // 清空菜单（用于登出）
  const clearMenus = () => {
    menus.value = []
    menuLoaded.value = false
  }

  // 根据 menu_id 查找菜单
  const findMenu = (menuId: string, menuList: MenuTreeNode[] = menus.value): MenuTreeNode | null => {
    for (const menu of menuList) {
      if (menu.menu_id === menuId) {
        return menu
      }
      if (menu.children && menu.children.length > 0) {
        const found = findMenu(menuId, menu.children)
        if (found) return found
      }
    }
    return null
  }

  // 获取平铺的菜单列表（用于权限检查）
  const flattenMenus = (menuList: MenuTreeNode[] = menus.value): MenuTreeNode[] => {
    const result: MenuTreeNode[] = []
    const flatten = (list: MenuTreeNode[]) => {
      for (const menu of list) {
        result.push(menu)
        if (menu.children && menu.children.length > 0) {
          flatten(menu.children)
        }
      }
    }
    flatten(menuList)
    return result
  }

  // 检查是否有某个菜单的权限
  const hasMenuPermission = (menuId: string): boolean => {
    const flatMenus = flattenMenus()
    return flatMenus.some(m => m.menu_id === menuId)
  }

  // 获取可访问的路由列表
  const getAccessibleRoutes = () => {
    const flatMenus = flattenMenus()
    return flatMenus
      .filter(m => m.path && m.status === 1) // 只返回有路径且状态为显示的菜单
      .map(m => ({
        path: m.path!,
        name: m.name,
        meta: {
          title: m.name,
          icon: m.icon
        }
      }))
  }

  return {
    menus,
    menuLoaded,
    loading,
    loadUserMenus,
    clearMenus,
    findMenu,
    flattenMenus,
    hasMenuPermission,
    getAccessibleRoutes
  }
})
