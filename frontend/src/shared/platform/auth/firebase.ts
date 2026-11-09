import { FirebaseError, initializeApp } from 'firebase/app'
import {
  GoogleAuthProvider,
  connectAuthEmulator,
  createUserWithEmailAndPassword,
  getAuth,
  onAuthStateChanged,
  sendEmailVerification,
  sendPasswordResetEmail,
  signInWithEmailAndPassword,
  signInWithPopup,
  signOut as firebaseSignOut,
  updateProfile,
  type User,
} from 'firebase/auth'

import type { FirebaseConfig } from './config'
import {
  AuthError,
  type AuthErrorCode,
  type AuthService,
  type AuthSnapshot,
  type AuthUser,
} from './types'

/** Translates Firebase error codes to the app's own (see AuthErrorCode). */
export function toAuthError(error: unknown): AuthError {
  if (error instanceof AuthError) return error
  const code = error instanceof FirebaseError ? error.code : ''
  const mapped: Record<string, AuthErrorCode> = {
    'auth/invalid-credential': 'INVALID_CREDENTIALS',
    'auth/invalid-login-credentials': 'INVALID_CREDENTIALS',
    'auth/wrong-password': 'INVALID_CREDENTIALS',
    'auth/user-not-found': 'INVALID_CREDENTIALS',
    'auth/email-already-in-use': 'EMAIL_IN_USE',
    'auth/weak-password': 'WEAK_PASSWORD',
    'auth/invalid-email': 'INVALID_EMAIL',
    'auth/too-many-requests': 'TOO_MANY_REQUESTS',
    'auth/network-request-failed': 'NETWORK',
    'auth/user-disabled': 'USER_DISABLED',
    'auth/account-exists-with-different-credential': 'ACCOUNT_EXISTS_DIFFERENT_CREDENTIAL',
    'auth/popup-closed-by-user': 'CANCELLED',
    'auth/cancelled-popup-request': 'CANCELLED',
  }
  return new AuthError(mapped[code] ?? 'UNKNOWN', error)
}

function toAuthUser(user: User): AuthUser {
  return {
    uid: user.uid,
    email: user.email,
    emailVerified: user.emailVerified,
    displayName: user.displayName,
  }
}

/** Firebase Authentication behind the AuthService interface. */
export function createFirebaseAuth(config: FirebaseConfig): AuthService {
  const app = initializeApp({
    apiKey: config.apiKey,
    projectId: config.projectId,
    authDomain: config.authDomain,
  })
  const auth = getAuth(app)
  if (config.emulatorUrl) {
    connectAuthEmulator(auth, config.emulatorUrl, { disableWarnings: true })
  }

  let snapshot: AuthSnapshot = { status: 'loading', user: null }
  const listeners = new Set<() => void>()
  const set = (next: AuthSnapshot) => {
    snapshot = next
    listeners.forEach((listener) => listener())
  }

  let markReady: () => void = () => undefined
  const readyPromise = new Promise<void>((resolve) => {
    markReady = resolve
  })

  onAuthStateChanged(auth, (user) => {
    set(user ? { status: 'signedIn', user: toAuthUser(user) } : { status: 'signedOut', user: null })
    markReady()
  })

  // Firebase notifies listeners asynchronously: after a sign-in, publish the
  // new state right away so route guards that run next already see it.
  const publishCurrentUser = () => {
    const user = auth.currentUser
    if (user) set({ status: 'signedIn', user: toAuthUser(user) })
  }

  const attempt = async <T>(action: () => Promise<T>): Promise<T> => {
    try {
      return await action()
    } catch (error) {
      throw toAuthError(error)
    }
  }

  return {
    getSnapshot: () => snapshot,
    subscribe(listener) {
      listeners.add(listener)
      return () => listeners.delete(listener)
    },
    ready: () => readyPromise,

    async getIdToken(forceRefresh = false) {
      return (await auth.currentUser?.getIdToken(forceRefresh)) ?? null
    },

    signInWithPassword: (email, password) =>
      attempt(async () => {
        await signInWithEmailAndPassword(auth, email, password)
        publishCurrentUser()
      }),

    signUpWithPassword: (email, password, name) =>
      attempt(async () => {
        const { user } = await createUserWithEmailAndPassword(auth, email, password)
        await updateProfile(user, { displayName: name })
        // The token issued at sign-up has no name yet: get one that does, so
        // the API creates the profile with it.
        await user.getIdToken(true)
        publishCurrentUser()
        await sendEmailVerification(user).catch(() => undefined) // best effort; can be resent
      }),

    signInWithGoogle: () =>
      attempt(async () => {
        await signInWithPopup(auth, new GoogleAuthProvider())
        publishCurrentUser()
      }),

    signOut: () => attempt(() => firebaseSignOut(auth)),

    sendVerificationEmail: () =>
      attempt(async () => {
        if (auth.currentUser) await sendEmailVerification(auth.currentUser)
      }),

    refreshUser: () =>
      attempt(async () => {
        const user = auth.currentUser
        if (!user) return
        await user.reload()
        await user.getIdToken(true) // the API reads email_verified from the token
        publishCurrentUser()
      }),

    sendPasswordReset: (email) => attempt(() => sendPasswordResetEmail(auth, email)),
  }
}
