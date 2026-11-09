import { createFileRoute } from '@tanstack/react-router'

import { CreateWorkspaceForm } from '@/features/workspaces'

export const Route = createFileRoute('/_app/workspaces/new')({
  component: CreateWorkspaceForm,
})
