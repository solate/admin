import http from './http'
import type { MenuTreeResponse, MenuInfo } from './menu'

// 获取用户菜单树
export const getUserMenu = (): Promise<MenuTreeResponse> => {
  return http.get('/api/v1/user/menu')
}

// 按钮权限信息
export interface ButtonInfo {
  permission_id: string
  name: string
  action?: string
  resource?: string
}

// 获取按钮权限
export const getUserButtons = (menuId: string): Promise<{ buttons: ButtonInfo[] }> => {
  return http.get('/api/v1/user/buttons', { params: { menu_id: menuId } })
}
