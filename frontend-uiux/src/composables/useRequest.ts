// HTTP request composable with loading and error states

import { ref, type Ref } from 'vue'
import { api } from '@/utils/request'
import type { AxiosRequestConfig } from 'axios'

export interface UseRequestOptions<T> {
  onSuccess?: (data: T) => void
  onError?: (error: Error) => void
  immediate?: boolean
}

export interface UseRequestReturn<T> {
  data: Ref<T | null>
  error: Ref<Error | null>
  isLoading: Ref<boolean>
  execute: (config: AxiosRequestConfig) => Promise<T>
  reset: () => void
}

export function useRequest<T = any>(
  options: UseRequestOptions<T> = {}
): UseRequestReturn<T> {
  const data = ref<T | null>(null) as Ref<T | null>
  const error = ref<Error | null>(null)
  const isLoading = ref(false)

  const execute = async (config: AxiosRequestConfig): Promise<T> => {
    isLoading.value = true
    error.value = null

    try {
      const response = await api(config)
      const result = response.data as T

      data.value = result
      options.onSuccess?.(result)

      return result
    } catch (err) {
      const errorObj = err as Error
      error.value = errorObj
      options.onError?.(errorObj)
      throw errorObj
    } finally {
      isLoading.value = false
    }
  }

  const reset = () => {
    data.value = null
    error.value = null
    isLoading.value = false
  }

  return {
    data,
    error,
    isLoading,
    execute,
    reset
  }
}

// Composable for making GET requests
export function useFetch<T = any>(
  url: string,
  options: UseRequestOptions<T> = {}
): UseRequestReturn<T> {
  const request = useRequest<T>(options)

  const execute = (): Promise<T> => {
    return request.execute({ method: 'GET', url })
  }

  if (options.immediate) {
    execute()
  }

  return {
    ...request,
    execute
  }
}

// Composable for pagination
export interface PaginationParams {
  page: number
  pageSize: number
  total: number
}

export function usePagination(initialPageSize = 10) {
  const page = ref(1)
  const pageSize = ref(initialPageSize)
  const total = ref(0)

  const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

  const setTotal = (count: number) => {
    total.value = count
  }

  const setPage = (newPage: number) => {
    page.value = newPage
  }

  const setPageSize = (newPageSize: number) => {
    pageSize.value = newPageSize
    page.value = 1 // Reset to first page when changing page size
  }

  const reset = () => {
    page.value = 1
    total.value = 0
  }

  const getParams = (): PaginationParams => ({
    page: page.value,
    pageSize: pageSize.value,
    total: total.value
  })

  return {
    page,
    pageSize,
    total,
    totalPages,
    setTotal,
    setPage,
    setPageSize,
    reset,
    getParams
  }
}
