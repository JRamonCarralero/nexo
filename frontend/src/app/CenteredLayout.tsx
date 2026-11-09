import { Link } from '@tanstack/react-router'
import type { ReactNode } from 'react'

import { EmailVerificationBanner, LanguageSwitcher } from '@/features/account'

/** A simple centered page, for the screens outside the app frame (sign in, invitations). */
export function CenteredLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col bg-slate-50 text-slate-900">
      <header className="flex items-center justify-between px-4 py-3">
        <Link to="/" className="text-lg font-bold tracking-tight">
          Nexo
        </Link>
        <LanguageSwitcher />
      </header>
      <EmailVerificationBanner />
      <main className="flex flex-1 flex-col items-center justify-center gap-6 p-4">{children}</main>
    </div>
  )
}
