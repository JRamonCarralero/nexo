import { useQueryClient } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { InvitationsPanel } from '@/features/invitations'
import type { Member, Role } from '@/shared/api/generated/model'
import { useGetMe } from '@/shared/api/generated/users/users'
import {
  getGetWorkspaceQueryKey,
  getListMembersQueryKey,
  getListWorkspacesQueryKey,
  useGetWorkspace,
  useListMembers,
  useRemoveMember,
  useUpdateMember,
} from '@/shared/api/generated/workspaces/workspaces'
import { useErrorMessage } from '@/shared/api/use-error-message'
import { useFormatDate } from '@/shared/format'
import { roles, useRoleLabel } from '@/shared/roles'
import { Alert, Button, Card, Select, Spinner } from '@/shared/ui'

/** The people of a workspace. Admins also manage roles and invitations here. */
export function MembersPage({ workspaceId }: { workspaceId: string }) {
  const { t } = useTranslation()
  const roleLabel = useRoleLabel()
  const formatDate = useFormatDate()
  const errorMessage = useErrorMessage()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const { data: workspace } = useGetWorkspace(workspaceId)
  const { data: me } = useGetMe()
  const members = useListMembers(workspaceId)
  const [confirming, setConfirming] = useState<string | null>(null)

  const refresh = () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: getListMembersQueryKey(workspaceId) }),
      // My own role may have changed: it decides what this page offers.
      queryClient.invalidateQueries({ queryKey: getGetWorkspaceQueryKey(workspaceId) }),
    ])

  const updateMember = useUpdateMember({ mutation: { onSuccess: refresh } })
  const removeMember = useRemoveMember({
    mutation: {
      onSuccess: async (_data, variables) => {
        setConfirming(null)
        if (variables.userId === me?.id) {
          // I left: this workspace is no longer mine.
          await queryClient.invalidateQueries({ queryKey: getListWorkspacesQueryKey() })
          await navigate({ to: '/' })
        } else {
          await refresh()
        }
      },
    },
  })

  if (members.isPending || !workspace) {
    return (
      <p role="status" className="flex items-center gap-2 text-slate-600">
        <Spinner className="size-4" /> {t('common.loading')}
      </p>
    )
  }
  if (members.error) return <Alert tone="error">{errorMessage(members.error)}</Alert>

  const isAdmin = workspace.myRole === 'admin'
  const actionError = updateMember.error ?? removeMember.error

  const changeRole = (member: Member, role: Role) => {
    if (role !== member.role)
      updateMember.mutate({ workspaceId, userId: member.userId, data: { role } })
  }

  return (
    <div className="flex flex-col gap-6">
      <header>
        <h1 className="text-2xl font-semibold tracking-tight">{t('workspaces.members.title')}</h1>
        <p className="mt-1 text-sm text-slate-600">
          {t('workspaces.members.subtitle', { name: workspace.name })}
        </p>
      </header>

      {actionError && <Alert tone="error">{errorMessage(actionError)}</Alert>}

      <Card className="overflow-x-auto p-0">
        <table className="w-full text-left text-sm">
          <caption className="sr-only">{t('workspaces.members.title')}</caption>
          <thead className="border-b border-slate-200 text-slate-600">
            <tr>
              <th scope="col" className="px-4 py-3 font-medium">
                {t('workspaces.members.person')}
              </th>
              <th scope="col" className="px-4 py-3 font-medium">
                {t('workspaces.members.role')}
              </th>
              <th scope="col" className="px-4 py-3 font-medium">
                {t('workspaces.members.joined')}
              </th>
              <th scope="col" className="px-4 py-3">
                <span className="sr-only">{t('workspaces.members.actions')}</span>
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-200">
            {members.data.items.map((member) => {
              const isSelf = member.userId === me?.id
              const name = member.name || member.email
              const busy =
                (updateMember.isPending && updateMember.variables?.userId === member.userId) ||
                (removeMember.isPending && removeMember.variables?.userId === member.userId)

              return (
                <tr key={member.userId}>
                  <td className="px-4 py-3">
                    <p className="font-medium">
                      {name}
                      {isSelf && (
                        <span className="ml-2 font-normal text-slate-500">
                          {t('workspaces.members.you')}
                        </span>
                      )}
                    </p>
                    {member.name && <p className="text-slate-600">{member.email}</p>}
                  </td>
                  <td className="px-4 py-3">
                    {isAdmin ? (
                      <Select
                        aria-label={t('workspaces.members.changeRoleOf', { name })}
                        value={member.role}
                        disabled={busy}
                        onChange={(e) => changeRole(member, e.target.value as Role)}
                        className="h-8 w-auto"
                      >
                        {roles.map((role) => (
                          <option key={role} value={role}>
                            {roleLabel(role)}
                          </option>
                        ))}
                      </Select>
                    ) : (
                      roleLabel(member.role)
                    )}
                  </td>
                  <td className="px-4 py-3 text-slate-600">{formatDate(member.joinedAt)}</td>
                  <td className="px-4 py-3 text-right">
                    {(isSelf || isAdmin) &&
                      (confirming === member.userId ? (
                        <span className="inline-flex items-center gap-2">
                          <span className="text-slate-700">{t('common.areYouSure')}</span>
                          <Button
                            size="sm"
                            variant="danger"
                            loading={busy}
                            onClick={() =>
                              removeMember.mutate({ workspaceId, userId: member.userId })
                            }
                          >
                            {t('common.confirm')}
                          </Button>
                          <Button size="sm" variant="ghost" onClick={() => setConfirming(null)}>
                            {t('common.cancel')}
                          </Button>
                        </span>
                      ) : (
                        <Button
                          size="sm"
                          variant="secondary"
                          onClick={() => setConfirming(member.userId)}
                          aria-label={
                            isSelf
                              ? t('workspaces.members.leaveLabel', { workspace: workspace.name })
                              : t('workspaces.members.removeLabel', { name })
                          }
                        >
                          {isSelf ? t('common.leave') : t('common.remove')}
                        </Button>
                      ))}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </Card>

      {isAdmin && <InvitationsPanel workspaceId={workspaceId} />}
    </div>
  )
}
