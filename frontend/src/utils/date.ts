/**
 * 时间格式化工具函数（基于 dayjs）
 */
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

// 扩展中文语言包
dayjs.locale('zh-cn')
// 扩展相对时间插件
dayjs.extend(relativeTime)

/**
 * 格式化毫秒时间戳为本地时间字符串
 * @param timestamp 毫秒时间戳
 * @param format 格式化模板，默认 'YYYY-MM-DD HH:mm:ss'
 * @returns 格式化后的时间字符串
 */
export function formatTime(
  timestamp: number | string,
  format: string = 'YYYY-MM-DD HH:mm:ss'
): string {
  if (!timestamp) return '-'

  const date = dayjs(timestamp)
  if (!date.isValid()) return '-'

  return date.format(format)
}

/**
 * 格式化日期为相对时间（如：刚刚、5分钟前、1小时前）
 * @param timestamp 毫秒时间戳
 * @returns 相对时间字符串
 */
export function formatRelativeTime(timestamp: number | string): string {
  if (!timestamp) return '-'

  const date = dayjs(timestamp)
  if (!date.isValid()) return '-'

  return date.fromNow()
}

/**
 * 获取当前时间戳（毫秒）
 */
export function now(): number {
  return Date.now()
}

/**
 * 默认导出 dayjs，方便直接使用
 */
export default dayjs
