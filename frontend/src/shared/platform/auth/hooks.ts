import { useContext, useSyncExternalStore } from 'react'

import { AuthContext } from './context'
import type { AuthService, AuthSnapshot } from './types'

/** The auth service, for actions (sign in, sign out…). */
export function useAuthService(): AuthService {
  const service = useContext(AuthContext)
  if (!service) throw new Error('useAuthService must be used inside <AuthProvider>')
  return service
}

/** The current authentication state; re-renders when it changes. */
export function useAuth(): AuthSnapshot {
  const service = useAuthService()
  return useSyncExternalStore(service.subscribe, service.getSnapshot)
}
