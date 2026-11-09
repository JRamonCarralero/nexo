export interface FirebaseConfig {
  apiKey: string
  projectId: string
  authDomain?: string
  /** Auth emulator URL (local development and tests only). */
  emulatorUrl?: string
}

interface Env {
  DEV?: boolean
  PROD?: boolean
  VITE_FIREBASE_API_KEY?: string
  VITE_FIREBASE_PROJECT_ID?: string
  VITE_FIREBASE_AUTH_DOMAIN?: string
  VITE_FIREBASE_AUTH_EMULATOR_URL?: string
}

/**
 * Reads the Firebase web configuration from VITE_* variables.
 *
 * In development everything defaults to the local emulator (project
 * demo-nexo), so `make up` is all it takes. A production build must provide
 * the real values and must not point to the emulator, which accepts unsigned
 * tokens.
 */
export function readFirebaseConfig(env: Env = import.meta.env): FirebaseConfig {
  const emulatorUrl =
    env.VITE_FIREBASE_AUTH_EMULATOR_URL || (env.DEV ? 'http://localhost:9099' : undefined)

  if (env.PROD) {
    if (!env.VITE_FIREBASE_API_KEY || !env.VITE_FIREBASE_PROJECT_ID) {
      throw new Error(
        'VITE_FIREBASE_API_KEY and VITE_FIREBASE_PROJECT_ID are required in production builds',
      )
    }
    if (emulatorUrl) {
      throw new Error('VITE_FIREBASE_AUTH_EMULATOR_URL must not be set in production builds')
    }
  }

  return {
    apiKey: env.VITE_FIREBASE_API_KEY || 'fake-api-key',
    projectId: env.VITE_FIREBASE_PROJECT_ID || 'demo-nexo',
    authDomain: env.VITE_FIREBASE_AUTH_DOMAIN || undefined,
    emulatorUrl,
  }
}
