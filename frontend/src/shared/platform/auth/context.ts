import { createContext } from 'react'

import type { AuthService } from './types'

export const AuthContext = createContext<AuthService | null>(null)
