import { Link, Outlet } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { EmailVerificationBanner, LanguageSwitcher, UserMenu } from '@/features/account'
import { WorkspaceNav } from '@/features/workspaces'

/** The frame of every page for a signed-in person: header, navigation and content. */
export function AppShell() {
  const { t } = useTranslation()

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      <a
        href="#main"
        className="sr-only rounded-md bg-white px-3 py-2 focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-10"
      >
        {t('common.skipToContent')}
      </a>
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex min-h-14 max-w-5xl flex-wrap items-center gap-x-4 gap-y-2 px-4 py-2">
          <Link to="/" className="text-lg font-bold tracking-tight">
            Nexo
          </Link>
          <WorkspaceNav />
          <div className="ml-auto flex items-center gap-3">
            <LanguageSwitcher />
            <UserMenu />
          </div>
        </div>
      </header>
      <EmailVerificationBanner />
      <main id="main" className="mx-auto max-w-5xl px-4 py-8">
        <Outlet />
      </main>
    </div>
  )
}
