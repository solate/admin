// LocalStorage composable

import { watch, type Ref } from 'vue'
import { storage } from '@/utils/storage'

export function useLocalStorage<T>(key: string, defaultValue: T) {
  const storedValue = storage.get<T>(key)
  const state = ref<T>(storedValue !== null ? storedValue : defaultValue) as Ref<T>

  // Watch for changes and update localStorage
  watch(
    state,
    (newValue) => {
      storage.set(key, newValue)
    },
    { deep: true }
  )

  // Update state and localStorage
  const setValue = (value: T) => {
    state.value = value
  }

  // Clear state and localStorage
  const clearValue = () => {
    state.value = defaultValue
    storage.remove(key)
  }

  return {
    state,
    setValue,
    clearValue
  }
}
