import { useRouter } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { useAuth, useAuthService } from '@/shared/platform/auth'
import { Button } from '@/shared/ui'

/** Who is signed in, and the way out. */
export function UserMenu() {
  const { t } = useTranslation()
  const { user } = useAuth()
  const auth = useAuthService()
  const router = useRouter()

  if (!user) return null

  const signOut = async () => {
    await auth.signOut()
    router.history.push('/login')
  }

  return (
    <div className="flex items-center gap-2">
      <span
        className="hidden max-w-48 truncate text-sm text-slate-700 sm:inline"
        title={user.email ?? undefined}
      >
        {user.displayName || user.email}
      </span>
      <Button size="sm" variant="ghost" onClick={signOut}>
        {t('auth.actions.signOut')}
      </Button>
    </div>
  )
}
