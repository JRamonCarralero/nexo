import { Navigate } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { useListWorkspaces } from '@/shared/api/generated/workspaces/workspaces'
import { useErrorMessage } from '@/shared/api/use-error-message'
import { Alert, Button, Spinner } from '@/shared/ui'

import { lastWorkspace } from './last-workspace'

/**
 * The landing page of a signed-in person: straight to their last workspace
 * (or the first one), or to creating their first one.
 */
export function WorkspaceRedirect() {
  const { t } = useTranslation()
  const errorMessage = useErrorMessage()
  const { data, error, isPending, refetch } = useListWorkspaces()

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
        <Button size="sm" variant="secondary" className="mt-2" onClick={() => void refetch()}>
          {t('common.retry')}
        </Button>
      </Alert>
    )
  }

  const workspaces = data.items
  if (workspaces.length === 0) return <Navigate to="/workspaces/new" replace />

  const target = workspaces.find((w) => w.id === lastWorkspace.get()) ?? workspaces[0]!
  return <Navigate to="/workspaces/$workspaceId" params={{ workspaceId: target.id }} replace />
}
