import type { HTMLAttributes } from 'react'

import { cn } from './cn'

export function Card({ className, ...props }: HTMLAttributes<HTMLElement>) {
  return (
    <section
      className={cn('rounded-lg border border-slate-200 bg-white p-6 shadow-sm', className)}
      {...props}
    />
  )
}
