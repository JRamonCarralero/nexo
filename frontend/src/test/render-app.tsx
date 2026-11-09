import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createMemoryHistory, RouterProvider } from '@tanstack/react-router'
import { render } from '@testing-library/react'

import { createAppRouter } from '@/app/router'
import { AuthProvider } from '@/shared/platform/auth'

import { createFakeAuth, type FakeAuth } from './fake-auth'

/**
 * Renders the whole app (real routes and guards) at `path` with a fake
 * authentication. Pair it with mockApi() to fake the server.
 */
export function renderApp({
  path = '/',
  auth = createFakeAuth(),
}: { path?: string; auth?: FakeAuth } = {}) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  const history = createMemoryHistory({ initialEntries: [path] })
  const router = createAppRouter(queryClient, auth, { history })

  render(
    <AuthProvider service={auth}>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </AuthProvider>,
  )

  return { auth, history, router, queryClient }
}
