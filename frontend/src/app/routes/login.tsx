import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { CenteredLayout } from '@/app/CenteredLayout'
import { redirectIfSignedIn } from '@/app/guards'
import { safeRedirect } from '@/app/safe-redirect'
import { AuthPage } from '@/features/auth'
import { HealthStatus } from '@/features/system'

export const Route = createFileRoute('/login')({
  // Only same-app paths are kept (see safeRedirect).
  validateSearch: (search: Record<string, unknown>): { redirect?: string } => ({
    redirect: safeRedirect(search.redirect),
  }),
  beforeLoad: ({ context, search }) => redirectIfSignedIn(context.auth, search.redirect),
  component: LoginPage,
})

function LoginPage() {
  const { redirect } = Route.useSearch()
  const { t } = useTranslation()

  return (
    <CenteredLayout>
      <AuthPage redirectTo={redirect} />
      <p className="text-center text-sm text-slate-600">{t('app.tagline')}</p>
      <HealthStatus />
    </CenteredLayout>
  )
}
