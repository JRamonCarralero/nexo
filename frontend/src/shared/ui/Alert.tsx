import type { ReactNode } from 'react'

import { cn } from './cn'

const tones = {
  error: 'border-red-200 bg-red-50 text-red-900',
  warning: 'border-amber-200 bg-amber-50 text-amber-900',
  success: 'border-emerald-200 bg-emerald-50 text-emerald-900',
  info: 'border-slate-200 bg-slate-50 text-slate-900',
} as const

interface AlertProps {
  tone?: keyof typeof tones
  title?: string
  className?: string
  children?: ReactNode
}

/** A message box. Errors are announced immediately (role="alert"), the rest politely. */
export function Alert({ tone = 'info', title, className, children }: AlertProps) {
  return (
    <div
      role={tone === 'error' ? 'alert' : 'status'}
      className={cn('rounded-md border px-4 py-3 text-sm', tones[tone], className)}
    >
      {title && <p className="font-medium">{title}</p>}
      {children && <div className={cn(title && 'mt-1')}>{children}</div>}
    </div>
  )
}
