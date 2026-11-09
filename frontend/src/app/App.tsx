import { QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from '@tanstack/react-router'
import { useEffect, useState } from 'react'

import { setTokenProvider } from '@/shared/api/http-client'
import { AuthProvider, type AuthService } from '@/shared/platform/auth'

import { createQueryClient } from './query-client'
import { createAppRouter } from './router'

export function App({ auth }: { auth: AuthService }) {
  const [queryClient] = useState(createQueryClient)
  const [router] = useState(() => createAppRouter(queryClient, auth))

  // The API client gets its bearer token from the auth adapter.
  useEffect(() => {
    setTokenProvider((forceRefresh) => auth.getIdToken(forceRefresh))
    return () => setTokenProvider(null)
  }, [auth])

  // Never show one person's cached data to the next one: drop the cache when
  // someone signs in, out, or the account changes.
  useEffect(() => {
    let uid = auth.getSnapshot().user?.uid ?? null
    return auth.subscribe(() => {
      const next = auth.getSnapshot().user?.uid ?? null
      if (next !== uid) {
        uid = next
        queryClient.clear()
      }
    })
  }, [auth, queryClient])

  return (
    <AuthProvider service={auth}>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </AuthProvider>
  )
}
