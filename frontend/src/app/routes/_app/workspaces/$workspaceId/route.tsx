import { createFileRoute } from '@tanstack/react-router'

import { WorkspaceLayout } from '@/features/workspaces'

export const Route = createFileRoute('/_app/workspaces/$workspaceId')({
  component: Layout,
})

function Layout() {
  const { workspaceId } = Route.useParams()
  return <WorkspaceLayout workspaceId={workspaceId} />
}
