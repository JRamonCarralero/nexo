import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { useGetWorkspace } from '@/shared/api/generated/workspaces/workspaces'
import { Card } from '@/shared/ui'

import { useRoleLabel } from '@/shared/roles'

/** The home of a workspace. Projects and tasks will live here (milestone M1). */
export function WorkspaceHome({ workspaceId }: { workspaceId: string }) {
  const { t } = useTranslation()
  const roleLabel = useRoleLabel()
  const { data: workspace } = useGetWorkspace(workspaceId)

  if (!workspace) return null

  return (
    <div className="flex flex-col gap-6">
      <header>
        <h1 className="text-2xl font-semibold tracking-tight">{workspace.name}</h1>
        <p className="mt-1 text-sm text-slate-600">
          {t('workspaces.home.yourRole', { role: roleLabel(workspace.myRole) })}
        </p>
      </header>

      <Card>
        <h2 className="text-lg font-medium">{t('workspaces.home.soonTitle')}</h2>
        <p className="mt-1 text-sm text-slate-600">{t('workspaces.home.soonText')}</p>
      </Card>

      <Card>
        <h2 className="text-lg font-medium">{t('workspaces.home.teamTitle')}</h2>
        <p className="mt-1 text-sm text-slate-600">{t('workspaces.home.teamText')}</p>
        <Link
          to="/workspaces/$workspaceId/members"
          params={{ workspaceId }}
          className="mt-3 inline-block text-sm font-medium underline underline-offset-2"
        >
          {t('workspaces.home.teamLink')}
        </Link>
      </Card>
    </div>
  )
}
