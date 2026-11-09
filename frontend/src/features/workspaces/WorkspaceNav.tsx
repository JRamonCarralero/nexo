import { Link, useNavigate, useParams } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { useListWorkspaces } from '@/shared/api/generated/workspaces/workspaces'
import { Select } from '@/shared/ui'

const NEW_WORKSPACE = '__new__'

const linkClass =
  'rounded-md px-3 py-1.5 text-sm text-slate-700 hover:bg-slate-100 ' +
  'focus-visible:outline-2 focus-visible:outline-slate-900'

/** Workspace picker and navigation of the current workspace, for the header. */
export function WorkspaceNav() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { workspaceId } = useParams({ strict: false })
  const { data } = useListWorkspaces()

  const workspaces = data?.items ?? []
  if (workspaces.length === 0) return null

  return (
    <>
      <Select
        aria-label={t('workspaces.switcher.label')}
        value={workspaceId ?? ''}
        onChange={(e) => {
          const id = e.target.value
          if (id === NEW_WORKSPACE) void navigate({ to: '/workspaces/new' })
          else void navigate({ to: '/workspaces/$workspaceId', params: { workspaceId: id } })
        }}
        className="h-8 w-auto max-w-48"
      >
        {!workspaceId && <option value="">{t('workspaces.switcher.choose')}</option>}
        {workspaces.map((w) => (
          <option key={w.id} value={w.id}>
            {w.name}
          </option>
        ))}
        <option value={NEW_WORKSPACE}>{t('workspaces.switcher.new')}</option>
      </Select>

      {workspaceId && (
        <nav aria-label={t('workspaces.nav.label')} className="flex items-center gap-1">
          <Link
            to="/workspaces/$workspaceId"
            params={{ workspaceId }}
            activeOptions={{ exact: true }}
            className={linkClass}
            activeProps={{ 'aria-current': 'page', className: 'font-medium text-slate-900' }}
          >
            {t('workspaces.nav.home')}
          </Link>
          <Link
            to="/workspaces/$workspaceId/members"
            params={{ workspaceId }}
            className={linkClass}
            activeProps={{ 'aria-current': 'page', className: 'font-medium text-slate-900' }}
          >
            {t('workspaces.nav.members')}
          </Link>
        </nav>
      )}
    </>
  )
}
