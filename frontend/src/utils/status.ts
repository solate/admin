/**
 * 状态值常量
 */
export const STATUS = {
  ACTIVE: 1,
  INACTIVE: 2,
  ENABLED: 1,
  DISABLED: 2,
  NORMAL: 1,
  DISABLE: 2
} as const

/**
 * 状态工具类
 */
export class StatusUtils {
  /**
   * 判断状态是否为激活/正常/启用状态
   * @param status 状态值 (1 | 2 | number | undefined | null)
   * @returns 是否为激活状态
   */
  static isActive(status: number | undefined | null): boolean {
    return Number(status) === 1
  }

  /**
   * 判断状态是否为停用/禁用状态
   * @param status 状态值 (1 | 2 | number | undefined | null)
   * @returns 是否为停用状态
   */
  static isInactive(status: number | undefined | null): boolean {
    return Number(status) === 2
  }

  /**
   * 获取状态对应的 Element Plus Tag 类型
   * @param status 状态值
   * @param activeType 激活状态的类型，默认 'success'
   * @param inactiveType 停用状态的类型，默认 'info'
   * @returns Element Plus Tag type
   */
  static getTagType(
    status: number | undefined | null,
    activeType: 'success' | 'primary' | 'warning' | 'danger' | 'info' = 'success',
    inactiveType: 'success' | 'primary' | 'warning' | 'danger' | 'info' = 'info'
  ): 'success' | 'primary' | 'warning' | 'danger' | 'info' {
    return this.isActive(status) ? activeType : inactiveType
  }

  /**
   * 获取状态对应的 Element Plus Button 类型
   * @param status 状态值
   * @param activeType 激活状态的类型
   * @param inactiveType 停用状态的类型
   * @returns Element Plus Button type
   */
  static getButtonType(
    status: number | undefined | null,
    activeType: 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'text' = 'primary',
    inactiveType: 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'text' = 'success'
  ): 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'text' {
    return this.isActive(status) ? activeType : inactiveType
  }

  /**
   * 获取状态文本
   * @param status 状态值
   * @param activeText 激活状态文本，默认 '正常'
   * @param inactiveText 停用状态文本，默认 '禁用'
   * @returns 状态文本
   */
  static getStatusText(
    status: number | undefined | null,
    activeText = '正常',
    inactiveText = '禁用'
  ): string {
    return this.isActive(status) ? activeText : inactiveText
  }

  /**
   * 切换状态值 (1 -> 2, 2 -> 1)
   * @param status 当前状态值
   * @returns 切换后的状态值
   */
  static toggleStatus(status: number | undefined | null): 1 | 2 {
    return this.isActive(status) ? 2 : 1
  }

  /**
   * 获取切换后的状态文本
   * @param status 当前状态值
   * @param activeText 激活操作文本，默认 '禁用'
   * @param inactiveText 停用操作文本，默认 '启用'
   * @returns 操作文本
   */
  static getToggleActionText(
    status: number | undefined | null,
    activeText = '禁用',
    inactiveText = '启用'
  ): string {
    return this.isActive(status) ? activeText : inactiveText
  }

  /**
   * 将状态值转换为数字（处理 null/undefined）
   * @param status 状态值
   * @returns 数字状态值
   */
  static toNumber(status: number | undefined | null): 1 | 2 {
    const num = Number(status)
    return num === 1 ? 1 : 2
  }
}

/**
 * 状态类型的 Vue 组合式函数
 * @returns 状态工具方法
 */
export function useStatus() {
  return {
    isActive: StatusUtils.isActive.bind(StatusUtils),
    isInactive: StatusUtils.isInactive.bind(StatusUtils),
    getTagType: StatusUtils.getTagType.bind(StatusUtils),
    getButtonType: StatusUtils.getButtonType.bind(StatusUtils),
    getStatusText: StatusUtils.getStatusText.bind(StatusUtils),
    toggleStatus: StatusUtils.toggleStatus.bind(StatusUtils),
    getToggleActionText: StatusUtils.getToggleActionText.bind(StatusUtils),
    toNumber: StatusUtils.toNumber.bind(StatusUtils)
  }
}
