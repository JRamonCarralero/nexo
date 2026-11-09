import { createFileRoute } from '@tanstack/react-router'

import { MembersPage } from '@/features/workspaces'

export const Route = createFileRoute('/_app/workspaces/$workspaceId/members')({
  component: Members,
})

function Members() {
  const { workspaceId } = Route.useParams()
  return <MembersPage workspaceId={workspaceId} />
}
