import type { ReactNode } from 'react'

import { AuthContext } from './context'
import type { AuthService } from './types'

export function AuthProvider({ service, children }: { service: AuthService; children: ReactNode }) {
  return <AuthContext.Provider value={service}>{children}</AuthContext.Provider>
}
