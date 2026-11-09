import { createFileRoute } from '@tanstack/react-router'

import { WorkspaceHome } from '@/features/workspaces'

export const Route = createFileRoute('/_app/workspaces/$workspaceId/')({
  component: Home,
})

function Home() {
  const { workspaceId } = Route.useParams()
  return <WorkspaceHome workspaceId={workspaceId} />
}
