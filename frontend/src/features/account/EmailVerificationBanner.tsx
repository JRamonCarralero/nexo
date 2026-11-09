import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { useAuth, useAuthService } from '@/shared/platform/auth'
import { Alert, Button } from '@/shared/ui'

/**
 * Asks people who signed up with email and password to verify their address.
 * Some things need it (accepting an invitation sent to that email). Google
 * accounts arrive verified, so they never see it.
 */
export function EmailVerificationBanner() {
  const { t } = useTranslation()
  const { user } = useAuth()
  const auth = useAuthService()
  const [state, setState] = useState<'idle' | 'resent' | 'stillUnverified' | 'failed'>('idle')
  const [busy, setBusy] = useState<'resend' | 'check' | null>(null)

  if (!user || user.emailVerified) return null

  const run = async (kind: 'resend' | 'check') => {
    setBusy(kind)
    try {
      if (kind === 'resend') {
        await auth.sendVerificationEmail()
        setState('resent')
      } else {
        await auth.refreshUser()
        // On success the banner disappears by itself (the user is now verified).
        setState('stillUnverified')
      }
    } catch {
      setState('failed')
    } finally {
      setBusy(null)
    }
  }

  return (
    <div className="border-b border-amber-200 bg-amber-50">
      <Alert tone="warning" className="mx-auto max-w-5xl rounded-none border-0 bg-transparent">
        <p>{t('account.verify.message', { email: user.email })}</p>
        <div className="mt-2 flex flex-wrap items-center gap-2">
          <Button
            size="sm"
            variant="secondary"
            loading={busy === 'resend'}
            onClick={() => run('resend')}
          >
            {t('account.verify.resend')}
          </Button>
          <Button
            size="sm"
            variant="secondary"
            loading={busy === 'check'}
            onClick={() => run('check')}
          >
            {t('account.verify.check')}
          </Button>
          {state === 'resent' && <span>{t('account.verify.resent')}</span>}
          {state === 'stillUnverified' && <span>{t('account.verify.stillUnverified')}</span>}
          {state === 'failed' && <span>{t('account.verify.failed')}</span>}
        </div>
      </Alert>
    </div>
  )
}
