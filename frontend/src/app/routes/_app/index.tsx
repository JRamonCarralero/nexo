import { createFileRoute } from '@tanstack/react-router'

import { WorkspaceRedirect } from '@/features/workspaces'

export const Route = createFileRoute('/_app/')({
  component: WorkspaceRedirect,
})
