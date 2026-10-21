import { QueryClient } from '@tanstack/react-query'

import { ApiError } from '@/shared/api/http-client'

export function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        // Client errors (4xx) will not fix themselves: retry only network and server errors.
        retry: (failureCount, error) => {
          if (error instanceof ApiError && error.status >= 400 && error.status < 500) return false
          return failureCount < 2
        },
      },
    },
  })
}
