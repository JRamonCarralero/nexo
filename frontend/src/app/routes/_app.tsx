import { createFileRoute } from '@tanstack/react-router'

import { AppShell } from '@/app/AppShell'
import { requireAuth } from '@/app/guards'

// Every page inside this layout needs a signed-in person.
export const Route = createFileRoute('/_app')({
  beforeLoad: ({ context, location }) => requireAuth(context.auth, location),
  component: AppShell,
})
