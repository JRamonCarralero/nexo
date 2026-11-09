import type { QueryClient } from '@tanstack/react-query'
import { createRouter, type RouterHistory } from '@tanstack/react-router'

import type { AuthService } from '@/shared/platform/auth'

import { routeTree } from './routeTree.gen'

export interface RouterContext {
  queryClient: QueryClient
  auth: AuthService
}

export function createAppRouter(
  queryClient: QueryClient,
  auth: AuthService,
  options: { history?: RouterHistory } = {},
) {
  return createRouter({
    routeTree,
    context: { queryClient, auth },
    history: options.history,
    defaultPreload: 'intent',
    scrollRestoration: true,
  })
}

declare module '@tanstack/react-router' {
  interface Register {
    router: ReturnType<typeof createAppRouter>
  }
}
