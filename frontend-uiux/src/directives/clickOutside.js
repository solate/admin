/**
 * Click outside directive
 * Usage: v-click-outside="handler"
 */
export default {
  mounted(el, binding) {
    // 使用 nextTick 避免竞态条件：
    // 当元素刚挂载时，如果同一事件正在冒泡，
    // 立即添加监听器会导致该事件立即触发关闭
    // 使用 setTimeout 将监听器添加推迟到下一个事件循环
    el._clickOutside = (event) => {
      if (!(el === event.target || el.contains(event.target))) {
        binding.value(event)
      }
    }
    // 延迟添加监听器，避免当前事件触发
    setTimeout(() => {
      document.addEventListener('click', el._clickOutside)
    }, 0)
  },
  unmounted(el) {
    document.removeEventListener('click', el._clickOutside)
  }
}
