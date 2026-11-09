import { FirebaseError } from 'firebase/app'

import { readFirebaseConfig } from './config'
import { toAuthError } from './firebase'
import { AuthError } from './types'

describe('readFirebaseConfig', () => {
  it('defaults to the local emulator in development', () => {
    expect(readFirebaseConfig({ DEV: true })).toEqual({
      apiKey: 'fake-api-key',
      projectId: 'demo-nexo',
      authDomain: undefined,
      emulatorUrl: 'http://localhost:9099',
    })
  })

  it('uses the variables that are set', () => {
    const config = readFirebaseConfig({
      DEV: true,
      VITE_FIREBASE_API_KEY: 'key',
      VITE_FIREBASE_PROJECT_ID: 'my-project',
      VITE_FIREBASE_AUTH_DOMAIN: 'my-project.firebaseapp.com',
      VITE_FIREBASE_AUTH_EMULATOR_URL: 'http://localhost:9199',
    })

    expect(config).toEqual({
      apiKey: 'key',
      projectId: 'my-project',
      authDomain: 'my-project.firebaseapp.com',
      emulatorUrl: 'http://localhost:9199',
    })
  })

  it('accepts a complete production configuration without emulator', () => {
    const config = readFirebaseConfig({
      PROD: true,
      VITE_FIREBASE_API_KEY: 'key',
      VITE_FIREBASE_PROJECT_ID: 'my-project',
    })

    expect(config.emulatorUrl).toBeUndefined()
    expect(config.projectId).toBe('my-project')
  })

  it('refuses incomplete production builds', () => {
    expect(() => readFirebaseConfig({ PROD: true })).toThrow(/required in production/)
    expect(() => readFirebaseConfig({ PROD: true, VITE_FIREBASE_API_KEY: 'key' })).toThrow(
      /required in production/,
    )
  })

  it('refuses the emulator in production builds: it accepts unsigned tokens', () => {
    expect(() =>
      readFirebaseConfig({
        PROD: true,
        VITE_FIREBASE_API_KEY: 'key',
        VITE_FIREBASE_PROJECT_ID: 'my-project',
        VITE_FIREBASE_AUTH_EMULATOR_URL: 'http://localhost:9099',
      }),
    ).toThrow(/must not be set in production/)
  })
})

describe('toAuthError', () => {
  const firebaseError = (code: string) => new FirebaseError(code, 'message')

  it.each([
    ['auth/invalid-credential', 'INVALID_CREDENTIALS'],
    ['auth/wrong-password', 'INVALID_CREDENTIALS'],
    ['auth/user-not-found', 'INVALID_CREDENTIALS'],
    ['auth/email-already-in-use', 'EMAIL_IN_USE'],
    ['auth/weak-password', 'WEAK_PASSWORD'],
    ['auth/invalid-email', 'INVALID_EMAIL'],
    ['auth/too-many-requests', 'TOO_MANY_REQUESTS'],
    ['auth/network-request-failed', 'NETWORK'],
    ['auth/user-disabled', 'USER_DISABLED'],
    ['auth/account-exists-with-different-credential', 'ACCOUNT_EXISTS_DIFFERENT_CREDENTIAL'],
    ['auth/popup-closed-by-user', 'CANCELLED'],
    ['auth/something-new', 'UNKNOWN'],
  ])('maps %s to %s', (code, expected) => {
    const error = toAuthError(firebaseError(code))

    expect(error).toBeInstanceOf(AuthError)
    expect(error.code).toBe(expected)
  })

  it('maps anything that is not a Firebase error to UNKNOWN and keeps the cause', () => {
    const cause = new TypeError('boom')

    const error = toAuthError(cause)

    expect(error.code).toBe('UNKNOWN')
    expect(error.cause).toBe(cause)
  })

  it('leaves our own errors untouched', () => {
    const own = new AuthError('NETWORK')

    expect(toAuthError(own)).toBe(own)
  })
})
