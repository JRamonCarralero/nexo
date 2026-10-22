import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { HealthStatus } from '@/features/system'

export const Route = createFileRoute('/')({
  component: HomePage,
})

function HomePage() {
  const { t } = useTranslation()

  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-6 bg-slate-50 p-4 text-slate-900">
      <h1 className="text-5xl font-bold tracking-tight">Nexo</h1>
      <p className="text-lg text-slate-600">{t('app.tagline')}</p>
      <HealthStatus />
    </main>
  )
}
