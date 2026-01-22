// LocalStorage utilities

const STORAGE_PREFIX = 'app_'

export const storage = {
  // Get item from localStorage
  get<T = any>(key: string): T | null {
    try {
      const item = localStorage.getItem(STORAGE_PREFIX + key)
      return item ? JSON.parse(item) : null
    } catch {
      return null
    }
  },

  // Set item to localStorage
  set(key: string, value: any): void {
    try {
      localStorage.setItem(STORAGE_PREFIX + key, JSON.stringify(value))
    } catch (error) {
      console.error('Error saving to localStorage:', error)
    }
  },

  // Remove item from localStorage
  remove(key: string): void {
    localStorage.removeItem(STORAGE_PREFIX + key)
  },

  // Clear all app items from localStorage
  clear(): void {
    const keys = Object.keys(localStorage)
    keys.forEach((key) => {
      if (key.startsWith(STORAGE_PREFIX)) {
        localStorage.removeItem(key)
      }
    })
  }
}

// Specific storage helpers
export const authStorage = {
  getToken: (): string | null => storage.get<string>('token'),
  setToken: (token: string) => storage.set('token', token),
  removeToken: () => storage.remove('token')
}

export const tenantStorage = {
  getCurrentTenantId: (): string | null => storage.get<string>('currentTenantId'),
  setCurrentTenantId: (id: string) => storage.set('currentTenantId', id),
  removeCurrentTenantId: () => storage.remove('currentTenantId')
}

export const localeStorage = {
  getLocale: (): string | null => storage.get<string>('locale'),
  setLocale: (locale: string) => storage.set('locale', locale)
}
