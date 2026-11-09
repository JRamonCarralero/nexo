import { zodResolver } from '@hookform/resolvers/zod'
import { useQueryClient } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { z } from 'zod'

import {
  getListWorkspacesQueryKey,
  useCreateWorkspace,
  useListWorkspaces,
} from '@/shared/api/generated/workspaces/workspaces'
import { useErrorMessage } from '@/shared/api/use-error-message'
import { useFieldError } from '@/shared/forms'
import { Alert, Button, Card, Field, Input } from '@/shared/ui'

const schema = z.object({
  name: z.string().trim().min(1, 'validation.required').max(100, 'validation.nameTooLong'),
})

/** Creates a workspace; the creator becomes its first admin. */
export function CreateWorkspaceForm() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const errorMessage = useErrorMessage()
  const fieldError = useFieldError()
  const { data: existing } = useListWorkspaces()
  const create = useCreateWorkspace()
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm({ resolver: zodResolver(schema) })

  const isFirst = existing?.items.length === 0

  const onSubmit = handleSubmit((values) =>
    create.mutate(
      { data: { name: values.name } },
      {
        onSuccess: async (workspace) => {
          await queryClient.invalidateQueries({ queryKey: getListWorkspacesQueryKey() })
          await navigate({ to: '/workspaces/$workspaceId', params: { workspaceId: workspace.id } })
        },
      },
    ),
  )

  return (
    <Card className="mx-auto w-full max-w-lg">
      <h1 className="text-2xl font-semibold tracking-tight">
        {isFirst ? t('workspaces.create.firstTitle') : t('workspaces.create.title')}
      </h1>
      <p className="mt-1 text-sm text-slate-600">
        {isFirst ? t('workspaces.create.firstSubtitle') : t('workspaces.create.subtitle')}
      </p>

      <form onSubmit={onSubmit} noValidate className="mt-6 flex flex-col gap-4">
        {create.error && <Alert tone="error">{errorMessage(create.error)}</Alert>}
        <Field label={t('workspaces.create.name')} error={fieldError(errors.name)}>
          {(control) => (
            <Input
              placeholder={t('workspaces.create.namePlaceholder')}
              autoFocus
              {...control}
              {...register('name')}
            />
          )}
        </Field>
        <Button type="submit" loading={create.isPending}>
          {t('workspaces.create.submit')}
        </Button>
      </form>
    </Card>
  )
}
