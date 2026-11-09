import { createFileRoute } from '@tanstack/react-router'

import { CenteredLayout } from '@/app/CenteredLayout'
import { requireAuth } from '@/app/guards'
import { AcceptInvitationPage } from '@/features/invitations'

export const Route = createFileRoute('/invitations/$token')({
  beforeLoad: ({ context, location }) => requireAuth(context.auth, location),
  component: InvitationRoute,
})

function InvitationRoute() {
  const { token } = Route.useParams()

  return (
    <CenteredLayout>
      <AcceptInvitationPage token={token} />
    </CenteredLayout>
  )
}
