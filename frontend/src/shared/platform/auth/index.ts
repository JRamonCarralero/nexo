// Public API of the auth adapter. The Firebase implementation is imported only
// where the app is assembled (see src/main.tsx).
export { readFirebaseConfig, type FirebaseConfig } from './config'
export { createFirebaseAuth } from './firebase'
export { AuthProvider } from './AuthProvider'
export { useAuth, useAuthService } from './hooks'
export {
  AUTH_ERROR_CODES,
  AuthError,
  type AuthErrorCode,
  type AuthService,
  type AuthSnapshot,
  type AuthStatus,
  type AuthUser,
} from './types'
