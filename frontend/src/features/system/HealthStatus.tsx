import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import { useGetHealth } from '@/shared/api/generated/system/system'
import { useErrorMessage } from '@/shared/api/use-error-message'

export function HealthStatus() {
  const { t } = useTranslation()
  const errorMessage = useErrorMessage()
  const { data, error, isPending } = useGetHealth({ query: { retry: false } })

  if (isPending) {
    return <StatusPill tone="neutral">{t('system.checking')}</StatusPill>
  }
  if (error) {
    return (
      <div className="flex flex-col items-center gap-1" role="alert">
        <StatusPill tone="error">{t('system.down')}</StatusPill>
        <p className="text-sm text-slate-600">{errorMessage(error)}</p>
        {error.traceId && (
          <p className="font-mono text-xs text-slate-400">
            {t('traceId', { traceId: error.traceId })}
          </p>
        )}
      </div>
    )
  }
  return <StatusPill tone="ok">{t('system.ok', { version: data.version })}</StatusPill>
}

const tones = {
  neutral: 'bg-slate-100 text-slate-700',
  ok: 'bg-emerald-100 text-emerald-800',
  error: 'bg-red-100 text-red-800',
} as const

function StatusPill({ tone, children }: { tone: keyof typeof tones; children: ReactNode }) {
  return (
    <span className={`rounded-full px-3 py-1 text-sm font-medium ${tones[tone]}`} role="status">
      {children}
    </span>
  )
}
