import { useQueryClient } from '@tanstack/react-query'
import { Link, useNavigate, useRouter } from '@tanstack/react-router'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'

import {
  useAcceptInvitation,
  useGetInvitation,
} from '@/shared/api/generated/invitations/invitations'
import { getListWorkspacesQueryKey } from '@/shared/api/generated/workspaces/workspaces'
import { useErrorMessage } from '@/shared/api/use-error-message'
import { useAuth, useAuthService } from '@/shared/platform/auth'
import { useRoleLabel } from '@/shared/roles'
import { Alert, Button, Card, Spinner } from '@/shared/ui'

/** What the person who opens an invitation link sees: who invites them, and the way to accept. */
export function AcceptInvitationPage({ token }: { token: string }) {
  const { t } = useTranslation()
  const roleLabel = useRoleLabel()
  const errorMessage = useErrorMessage()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const router = useRouter()
  const auth = useAuthService()
  const { user } = useAuth()
  const preview = useGetInvitation(token, { query: { retry: false } })
  const accept = useAcceptInvitation()
  const { reset: resetAccept } = accept

  // An "email not verified" error is stale once the account becomes verified.
  const emailVerified = user?.emailVerified
  useEffect(() => resetAccept(), [emailVerified, resetAccept])

  if (preview.isPending) {
    return (
      <p role="status" className="flex items-center gap-2 text-slate-600">
        <Spinner className="size-4" /> {t('common.loading')}
      </p>
    )
  }
  if (preview.error) {
    return (
      <Card className="w-full max-w-md">
        <Alert tone="error">{errorMessage(preview.error)}</Alert>
        <HomeLink />
      </Card>
    )
  }

  const invitation = preview.data
  if (invitation.status !== 'pending') {
    return (
      <Card className="w-full max-w-md">
        <h1 className="text-xl font-semibold">
          {t('invitations.accept.unavailableTitle', { workspace: invitation.workspaceName })}
        </h1>
        <Alert tone="warning" className="mt-4">
          {t(`invitations.accept.status.${invitation.status}`)}
        </Alert>
        <HomeLink />
      </Card>
    )
  }

  const wrongAccount =
    invitation.kind === 'email' &&
    user?.email != null &&
    invitation.email != null &&
    user.email.toLowerCase() !== invitation.email.toLowerCase()

  const switchAccount = async () => {
    await auth.signOut()
    router.history.push(`/login?redirect=${encodeURIComponent(`/invitations/${token}`)}`)
  }

  const onAccept = () =>
    accept.mutate(
      { token },
      {
        onSuccess: async (workspace) => {
          await queryClient.invalidateQueries({ queryKey: getListWorkspacesQueryKey() })
          await navigate({ to: '/workspaces/$workspaceId', params: { workspaceId: workspace.id } })
        },
      },
    )

  return (
    <Card className="w-full max-w-md">
      <h1 className="text-2xl font-semibold tracking-tight">
        {t('invitations.accept.title', { workspace: invitation.workspaceName })}
      </h1>
      <p className="mt-2 text-slate-700">
        {invitation.invitedBy
          ? t('invitations.accept.text', {
              inviter: invitation.invitedBy,
              role: roleLabel(invitation.role),
            })
          : t('invitations.accept.textAnonymous', { role: roleLabel(invitation.role) })}
      </p>

      <div className="mt-6 flex flex-col gap-4">
        {wrongAccount && (
          <Alert tone="warning">
            <p>
              {t('invitations.accept.wrongAccount', {
                invited: invitation.email,
                current: user?.email,
              })}
            </p>
            <Button size="sm" variant="secondary" className="mt-2" onClick={switchAccount}>
              {t('invitations.accept.switchAccount')}
            </Button>
          </Alert>
        )}
        {accept.error && <Alert tone="error">{errorMessage(accept.error)}</Alert>}
        <Button onClick={onAccept} loading={accept.isPending} disabled={wrongAccount}>
          {t('invitations.accept.button')}
        </Button>
      </div>
    </Card>
  )
}

function HomeLink() {
  const { t } = useTranslation()
  return (
    <Link to="/" className="mt-4 inline-block text-sm font-medium underline underline-offset-2">
      {t('common.goHome')}
    </Link>
  )
}
