import { Link, Outlet } from '@tanstack/react-router'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'

import { useGetWorkspace } from '@/shared/api/generated/workspaces/workspaces'
import { useErrorMessage } from '@/shared/api/use-error-message'
import { Alert, Spinner } from '@/shared/ui'

import { lastWorkspace } from './last-workspace'

/**
 * Wraps the pages of one workspace: waits for it to load, explains the error
 * if the person cannot open it (it does not exist or they are not a member)
 * and remembers it as the last one used.
 */
export function WorkspaceLayout({ workspaceId }: { workspaceId: string }) {
  const { t } = useTranslation()
  const errorMessage = useErrorMessage()
  const { data, error, isPending } = useGetWorkspace(workspaceId)

  useEffect(() => {
    if (data) lastWorkspace.set(data.id)
  }, [data])

  if (isPending) {
    return (
      <p role="status" className="flex items-center gap-2 text-slate-600">
        <Spinner className="size-4" /> {t('common.loading')}
      </p>
    )
  }
  if (error) {
    return (
      <Alert tone="error">
        <p>{errorMessage(error)}</p>
        <Link to="/" className="mt-2 inline-block font-medium underline underline-offset-2">
          {t('workspaces.backToMine')}
        </Link>
      </Alert>
    )
  }
  return <Outlet />
}
