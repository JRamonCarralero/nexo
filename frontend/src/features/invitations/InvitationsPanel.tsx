import { useTranslation } from 'react-i18next'

import { Card } from '@/shared/ui'

import { InviteForm } from './InviteForm'
import { PendingInvitations } from './PendingInvitations'

/** Invite people and manage pending invitations. Only admins get this. */
export function InvitationsPanel({ workspaceId }: { workspaceId: string }) {
  const { t } = useTranslation()

  return (
    <Card aria-labelledby="invitations-title" className="flex flex-col gap-6">
      <div>
        <h2 id="invitations-title" className="text-lg font-medium">
          {t('invitations.title')}
        </h2>
        <p className="mt-1 text-sm text-slate-600">{t('invitations.subtitle')}</p>
      </div>
      <InviteForm workspaceId={workspaceId} />
      <PendingInvitations workspaceId={workspaceId} />
    </Card>
  )
}
