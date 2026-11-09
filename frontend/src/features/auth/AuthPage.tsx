import { zodResolver } from '@hookform/resolvers/zod'
import { useRouter } from '@tanstack/react-router'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { z } from 'zod'

import { useFieldError } from '@/shared/forms'
import { AuthError, useAuthService } from '@/shared/platform/auth'
import { Alert, Button, Card, Field, Input } from '@/shared/ui'

const email = z.string().trim().min(1, 'validation.required').pipe(z.email('validation.email'))
const password = z.string().min(8, 'validation.passwordMin')

const signInSchema = z.object({ email, password: z.string().min(1, 'validation.required') })
const signUpSchema = z.object({
  name: z.string().trim().min(1, 'validation.required').max(100, 'validation.nameTooLong'),
  email,
  password,
})
const resetSchema = z.object({ email })

type Mode = 'signIn' | 'signUp' | 'reset'

/** Sign in, create an account or recover a password. `redirectTo` is where to go afterwards. */
export function AuthPage({ redirectTo }: { redirectTo?: string }) {
  const { t } = useTranslation()
  const [mode, setMode] = useState<Mode>('signIn')

  return (
    <Card className="w-full max-w-md">
      <h1 className="text-2xl font-semibold tracking-tight">{t(`auth.${mode}.title`)}</h1>
      <p className="mt-1 text-sm text-slate-600">{t(`auth.${mode}.subtitle`)}</p>

      <div className="mt-6 flex flex-col gap-4">
        {mode === 'signIn' && (
          <SignInForm redirectTo={redirectTo} onForgot={() => setMode('reset')} />
        )}
        {mode === 'signUp' && <SignUpForm redirectTo={redirectTo} />}
        {mode === 'reset' && <ResetForm />}

        {mode !== 'reset' && <GoogleButton redirectTo={redirectTo} />}

        <p className="text-sm text-slate-600">
          {mode === 'signIn' ? (
            <>
              {t('auth.switch.toSignUpText')}{' '}
              <LinkButton onClick={() => setMode('signUp')}>{t('auth.switch.toSignUp')}</LinkButton>
            </>
          ) : (
            <>
              {mode === 'signUp' ? t('auth.switch.toSignInText') : null}{' '}
              <LinkButton onClick={() => setMode('signIn')}>{t('auth.switch.toSignIn')}</LinkButton>
            </>
          )}
        </p>
      </div>
    </Card>
  )
}

function LinkButton({ onClick, children }: { onClick: () => void; children: string }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="font-medium text-slate-900 underline underline-offset-2 hover:text-slate-700"
    >
      {children}
    </button>
  )
}

/** Where to go once signed in. */
function useAfterAuth(redirectTo?: string) {
  const router = useRouter()
  return () => router.history.push(redirectTo ?? '/')
}

function useAuthErrorMessage() {
  const { t } = useTranslation()
  return (error: unknown) =>
    error instanceof AuthError ? t(`auth.errors.${error.code}`) : t('auth.errors.UNKNOWN')
}

function SignInForm({ redirectTo, onForgot }: { redirectTo?: string; onForgot: () => void }) {
  const { t } = useTranslation()
  const auth = useAuthService()
  const afterAuth = useAfterAuth(redirectTo)
  const fieldError = useFieldError()
  const authError = useAuthErrorMessage()
  const [error, setError] = useState<string>()
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm({ resolver: zodResolver(signInSchema) })

  const onSubmit = handleSubmit(async (values) => {
    setError(undefined)
    try {
      await auth.signInWithPassword(values.email, values.password)
      afterAuth()
    } catch (e) {
      setError(authError(e))
    }
  })

  return (
    <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
      {error && <Alert tone="error">{error}</Alert>}
      <Field label={t('auth.fields.email')} error={fieldError(errors.email)}>
        {(control) => (
          <Input type="email" autoComplete="email" {...control} {...register('email')} />
        )}
      </Field>
      <Field label={t('auth.fields.password')} error={fieldError(errors.password)}>
        {(control) => (
          <Input
            type="password"
            autoComplete="current-password"
            {...control}
            {...register('password')}
          />
        )}
      </Field>
      <button
        type="button"
        onClick={onForgot}
        className="self-start text-sm text-slate-600 underline underline-offset-2 hover:text-slate-900"
      >
        {t('auth.actions.forgotPassword')}
      </button>
      <Button type="submit" loading={isSubmitting}>
        {t('auth.actions.signIn')}
      </Button>
    </form>
  )
}

function SignUpForm({ redirectTo }: { redirectTo?: string }) {
  const { t } = useTranslation()
  const auth = useAuthService()
  const afterAuth = useAfterAuth(redirectTo)
  const fieldError = useFieldError()
  const authError = useAuthErrorMessage()
  const [error, setError] = useState<string>()
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm({ resolver: zodResolver(signUpSchema) })

  const onSubmit = handleSubmit(async (values) => {
    setError(undefined)
    try {
      await auth.signUpWithPassword(values.email, values.password, values.name)
      afterAuth()
    } catch (e) {
      setError(authError(e))
    }
  })

  return (
    <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
      {error && <Alert tone="error">{error}</Alert>}
      <Field label={t('auth.fields.name')} error={fieldError(errors.name)}>
        {(control) => <Input autoComplete="name" {...control} {...register('name')} />}
      </Field>
      <Field label={t('auth.fields.email')} error={fieldError(errors.email)}>
        {(control) => (
          <Input type="email" autoComplete="email" {...control} {...register('email')} />
        )}
      </Field>
      <Field
        label={t('auth.fields.password')}
        hint={t('auth.fields.passwordHint')}
        error={fieldError(errors.password)}
      >
        {(control) => (
          <Input
            type="password"
            autoComplete="new-password"
            {...control}
            {...register('password')}
          />
        )}
      </Field>
      <Button type="submit" loading={isSubmitting}>
        {t('auth.actions.signUp')}
      </Button>
    </form>
  )
}

function ResetForm() {
  const { t } = useTranslation()
  const auth = useAuthService()
  const fieldError = useFieldError()
  const authError = useAuthErrorMessage()
  const [error, setError] = useState<string>()
  const [sent, setSent] = useState(false)
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm({ resolver: zodResolver(resetSchema) })

  const onSubmit = handleSubmit(async (values) => {
    setError(undefined)
    try {
      await auth.sendPasswordReset(values.email)
      setSent(true)
    } catch (e) {
      setError(authError(e))
    }
  })

  if (sent) return <Alert tone="success">{t('auth.reset.sent')}</Alert>

  return (
    <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
      {error && <Alert tone="error">{error}</Alert>}
      <Field label={t('auth.fields.email')} error={fieldError(errors.email)}>
        {(control) => (
          <Input type="email" autoComplete="email" {...control} {...register('email')} />
        )}
      </Field>
      <Button type="submit" loading={isSubmitting}>
        {t('auth.actions.sendReset')}
      </Button>
    </form>
  )
}

function GoogleButton({ redirectTo }: { redirectTo?: string }) {
  const { t } = useTranslation()
  const auth = useAuthService()
  const afterAuth = useAfterAuth(redirectTo)
  const authError = useAuthErrorMessage()
  const [error, setError] = useState<string>()
  const [busy, setBusy] = useState(false)

  const onClick = async () => {
    setError(undefined)
    setBusy(true)
    try {
      await auth.signInWithGoogle()
      afterAuth()
    } catch (e) {
      // Closing the popup is not an error worth showing.
      if (!(e instanceof AuthError && e.code === 'CANCELLED')) setError(authError(e))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center gap-3 text-xs text-slate-500" aria-hidden="true">
        <span className="h-px flex-1 bg-slate-200" />
        {t('auth.or')}
        <span className="h-px flex-1 bg-slate-200" />
      </div>
      {error && <Alert tone="error">{error}</Alert>}
      <Button variant="secondary" onClick={onClick} loading={busy}>
        {t('auth.actions.google')}
      </Button>
    </div>
  )
}
