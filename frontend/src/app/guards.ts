import { redirect } from '@tanstack/react-router'

import type { AuthService } from '@/shared/platform/auth'

import { safeRedirect } from './safe-redirect'

/**
 * Route guard for pages that need a signed-in person. Waits for the initial
 * authentication state, then sends signed-out visitors to the login page,
 * remembering where they were going.
 */
export async function requireAuth(auth: AuthService, location: { href: string }) {
  await auth.ready()
  if (auth.getSnapshot().status !== 'signedIn') {
    throw redirect({ to: '/login', search: { redirect: location.href } })
  }
}

/** Route guard for the login page: people already signed in go on to where they were heading. */
export async function redirectIfSignedIn(auth: AuthService, target: string | undefined) {
  await auth.ready()
  if (auth.getSnapshot().status === 'signedIn') {
    throw redirect({ to: safeRedirect(target) ?? '/' })
  }
}
