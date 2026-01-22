/**
 * Click outside directive
 * Usage: v-click-outside="handler"
 */
import type { Directive, DirectiveBinding } from 'vue'

const clickOutside: Directive = {
  mounted(el: HTMLElement, binding: DirectiveBinding) {
    // Use nextTick to avoid race conditions:
    // When an element is first mounted, if an event is currently bubbling,
    // adding a listener immediately will cause that event to trigger the close handler.
    // Using setTimeout defers listener addition to the next event loop.
    el._clickOutside = (event: Event) => {
      if (!(el === event.target || el.contains(event.target as Node))) {
        binding.value(event)
      }
    }
    // Delay adding listener to avoid triggering on current event
    setTimeout(() => {
      document.addEventListener('click', el._clickOutside as EventListener)
    }, 0)
  },
  unmounted(el: HTMLElement) {
    document.removeEventListener('click', el._clickOutside as EventListener)
  }
}

export default clickOutside
