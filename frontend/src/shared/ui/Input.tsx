import type { InputHTMLAttributes, SelectHTMLAttributes } from 'react'

import { cn } from './cn'

const control =
  'block w-full rounded-md border border-slate-300 bg-white px-3 text-sm text-slate-900 ' +
  'placeholder:text-slate-400 focus-visible:outline-2 focus-visible:outline-offset-0 ' +
  'focus-visible:outline-slate-900 disabled:cursor-not-allowed disabled:bg-slate-100 ' +
  'aria-[invalid=true]:border-red-600 aria-[invalid=true]:focus-visible:outline-red-600'

export function Input({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return <input className={cn(control, 'h-10', className)} {...props} />
}

/** A native select: accessible and keyboard friendly out of the box. */
export function Select({ className, ...props }: SelectHTMLAttributes<HTMLSelectElement>) {
  return <select className={cn(control, 'h-10 pr-8', className)} {...props} />
}
