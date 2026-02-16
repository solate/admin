// Mock module entry point

import { env } from '@/config/env'

// Check if mock mode is enabled
export const isMockEnabled = (): boolean => {
  return import.meta.env.VITE_MOCK_ENABLED === 'true' || env.isDev
}

// Re-export handlers
export * from './handlers'
export * from './data'
