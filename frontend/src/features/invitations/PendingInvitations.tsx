import { useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'

import {
  getListInvitationsQueryKey,
  useListInvitations,
  useRevokeInvitation,
} from '@/shared/api/generated/invitations/invitations'
import { useErrorMessage } from '@/shared/api/use-error-message'
import { useFormatDate } from '@/shared/format'
import { useRoleLabel } from '@/shared/roles'
import { Alert, Button, Spinner } from '@/shared/ui'

/** Invitations that can still be used, with a way to revoke each one. */
export function PendingInvitations({ workspaceId }: { workspaceId: string }) {
  const { t } = useTranslation()
  const roleLabel = useRoleLabel()
  const formatDate = useFormatDate()
  const errorMessage = useErrorMessage()
  const queryClient = useQueryClient()
  const { data, error, isPending } = useListInvitations(workspaceId)
  const revoke = useRevokeInvitation({
    mutation: {
      onSuccess: () =>
        queryClient.invalidateQueries({ queryKey: getListInvitationsQueryKey(workspaceId) }),
    },
  })

  if (isPending) {
    return (
      <p role="status" className="flex items-center gap-2 text-sm text-slate-600">
        <Spinner className="size-4" /> {t('common.loading')}
      </p>
    )
  }
  if (error) return <Alert tone="error">{errorMessage(error)}</Alert>

  return (
    <div>
      <h3 className="text-base font-medium">{t('invitations.pending.title')}</h3>
      {revoke.error && (
        <Alert tone="error" className="mt-2">
          {errorMessage(revoke.error)}
        </Alert>
      )}
      {data.items.length === 0 ? (
        <p className="mt-2 text-sm text-slate-600">{t('invitations.pending.empty')}</p>
      ) : (
        <ul className="mt-2 divide-y divide-slate-200 rounded-md border border-slate-200">
          {data.items.map((invitation) => (
            <li key={invitation.id} className="flex flex-wrap items-center gap-3 px-4 py-3">
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium">
                  {invitation.email ?? t('invitations.pending.linkItem')}
                </p>
                <p className="text-sm text-slate-600">
                  {roleLabel(invitation.role)} ·{' '}
                  {t('invitations.pending.expires', { date: formatDate(invitation.expiresAt) })}
                </p>
              </div>
              <Button
                size="sm"
                variant="secondary"
                loading={revoke.isPending && revoke.variables?.invitationId === invitation.id}
                onClick={() => revoke.mutate({ workspaceId, invitationId: invitation.id })}
                aria-label={t('invitations.pending.revokeLabel', {
                  name: invitation.email ?? t('invitations.pending.linkItem'),
                })}
              >
                {t('common.revoke')}
              </Button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
