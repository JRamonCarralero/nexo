/** The signed-in person, as the identity provider knows them. */
export interface AuthUser {
  uid: string
  email: string | null
  emailVerified: boolean
  displayName: string | null
}

export type AuthStatus = 'loading' | 'signedIn' | 'signedOut'

export interface AuthSnapshot {
  status: AuthStatus
  user: AuthUser | null
}

/** Why an authentication action failed. The UI translates `auth.errors.<code>`. */
export const AUTH_ERROR_CODES = [
  'INVALID_CREDENTIALS',
  'EMAIL_IN_USE',
  'WEAK_PASSWORD',
  'INVALID_EMAIL',
  'TOO_MANY_REQUESTS',
  'NETWORK',
  'USER_DISABLED',
  'ACCOUNT_EXISTS_DIFFERENT_CREDENTIAL',
  'CANCELLED',
  'UNKNOWN',
] as const

export type AuthErrorCode = (typeof AUTH_ERROR_CODES)[number]

export class AuthError extends Error {
  readonly code: AuthErrorCode

  constructor(code: AuthErrorCode, cause?: unknown) {
    super(code, { cause })
    this.name = 'AuthError'
    this.code = code
  }
}

/**
 * What the app needs from an identity provider. Firebase implements it in
 * ./firebase.ts; the rest of the code never sees the SDK.
 */
export interface AuthService {
  /** Current state. Returns the same object until the state changes. */
  getSnapshot(): AuthSnapshot
  /** Calls the listener on every state change; returns the unsubscribe function. */
  subscribe(listener: () => void): () => void
  /** Resolves once the initial state (signed in or out) is known. */
  ready(): Promise<void>
  /** The ID token for the API, or null when signed out. */
  getIdToken(forceRefresh?: boolean): Promise<string | null>
  signInWithPassword(email: string, password: string): Promise<void>
  /** Creates the account, sets the name and sends the verification email. */
  signUpWithPassword(email: string, password: string, name: string): Promise<void>
  signInWithGoogle(): Promise<void>
  signOut(): Promise<void>
  sendVerificationEmail(): Promise<void>
  /** Reloads the account so a just-verified email is reflected (also in the token). */
  refreshUser(): Promise<void>
  sendPasswordReset(email: string): Promise<void>
}
