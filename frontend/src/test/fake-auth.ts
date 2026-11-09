import { vi } from 'vitest'

import type { AuthService, AuthSnapshot, AuthUser } from '@/shared/platform/auth'

export const ana: AuthUser = {
  uid: 'u1',
  email: 'ana@example.com',
  emailVerified: true,
  displayName: 'Ana',
}

/** An AuthService for tests: starts signed in (or out) and records every action. */
export function createFakeAuth(initialUser: AuthUser | null = ana) {
  let snapshot: AuthSnapshot = initialUser
    ? { status: 'signedIn', user: initialUser }
    : { status: 'signedOut', user: null }
  const listeners = new Set<() => void>()
  const set = (next: AuthSnapshot) => {
    snapshot = next
    listeners.forEach((listener) => listener())
  }
  const signedInAs = (email: string): AuthSnapshot => ({
    status: 'signedIn',
    user: { uid: 'u-new', email, emailVerified: false, displayName: null },
  })

  return {
    getSnapshot: () => snapshot,
    subscribe(listener: () => void) {
      listeners.add(listener)
      return () => listeners.delete(listener)
    },
    ready: () => Promise.resolve(),
    getIdToken: vi.fn(async () => (snapshot.user ? 'fake-token' : null)),
    signInWithPassword: vi.fn(async (email: string) => set(signedInAs(email))),
    signUpWithPassword: vi.fn(async (email: string, _password: string, name: string) =>
      set({
        status: 'signedIn',
        user: { uid: 'u-new', email, emailVerified: false, displayName: name },
      }),
    ),
    signInWithGoogle: vi.fn(async () => set(signedInAs('google@example.com'))),
    signOut: vi.fn(async () => set({ status: 'signedOut', user: null })),
    sendVerificationEmail: vi.fn(() => Promise.resolve()),
    refreshUser: vi.fn(() => Promise.resolve()),
    sendPasswordReset: vi.fn(() => Promise.resolve()),
    /** Test helper: change who is signed in. */
    setUser: (user: AuthUser | null) =>
      set(user ? { status: 'signedIn', user } : { status: 'signedOut', user: null }),
  } satisfies AuthService & { setUser: (user: AuthUser | null) => void }
}

export type FakeAuth = ReturnType<typeof createFakeAuth>
