import { zodResolver } from '@hookform/resolvers/zod'
import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useForm, useWatch } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { z } from 'zod'

import type { InvitationCreated, InvitationKind, Role } from '@/shared/api/generated/model'
import {
  getListInvitationsQueryKey,
  useCreateInvitation,
} from '@/shared/api/generated/invitations/invitations'
import { useErrorMessage } from '@/shared/api/use-error-message'
import { useFieldError } from '@/shared/forms'
import { roles, useRoleLabel } from '@/shared/roles'
import { Alert, Button, Field, Input, Select } from '@/shared/ui'

const schema = z
  .object({
    kind: z.enum(['email', 'link']),
    email: z.string().trim(),
    role: z.enum(['admin', 'member', 'guest']),
  })
  .superRefine((value, ctx) => {
    if (value.kind === 'email') {
      if (value.email === '') {
        ctx.addIssue({ code: 'custom', path: ['email'], message: 'validation.required' })
      } else if (!z.email().safeParse(value.email).success) {
        ctx.addIssue({ code: 'custom', path: ['email'], message: 'validation.email' })
      }
    } else if (value.role === 'admin') {
      ctx.addIssue({ code: 'custom', path: ['role'], message: 'invitations.form.linkNoAdmin' })
    }
  })

type FormValues = z.input<typeof schema>

/** Invite by email (the link goes in an email) or create a shareable link. */
export function InviteForm({ workspaceId }: { workspaceId: string }) {
  const { t } = useTranslation()
  const roleLabel = useRoleLabel()
  const errorMessage = useErrorMessage()
  const fieldError = useFieldError()
  const queryClient = useQueryClient()
  const create = useCreateInvitation()
  const [created, setCreated] = useState<InvitationCreated>()
  const {
    register,
    handleSubmit,
    control,
    getValues,
    setValue,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { kind: 'email', email: '', role: 'member' },
  })
  const kind = useWatch({ control, name: 'kind' })

  const onKindChange = (next: InvitationKind) => {
    // A link can never grant admin: fall back to a role it can give.
    if (next === 'link' && getValues('role') === 'admin') setValue('role', 'member')
  }

  const onSubmit = handleSubmit((values) => {
    setCreated(undefined)
    create.mutate(
      {
        workspaceId,
        data: {
          kind: values.kind,
          role: values.role,
          ...(values.kind === 'email' ? { email: values.email } : {}),
        },
      },
      {
        onSuccess: async (result) => {
          setCreated(result)
          reset({ kind: values.kind, email: '', role: values.role })
          await queryClient.invalidateQueries({ queryKey: getListInvitationsQueryKey(workspaceId) })
        },
      },
    )
  })

  const rolesForKind: Role[] = kind === 'link' ? roles.filter((r) => r !== 'admin') : roles

  return (
    <form onSubmit={onSubmit} noValidate className="flex flex-col gap-4">
      {create.error && <Alert tone="error">{errorMessage(create.error)}</Alert>}

      <fieldset className="flex flex-col gap-2">
        <legend className="text-sm font-medium">{t('invitations.form.kind')}</legend>
        <div className="flex flex-wrap gap-4">
          {(['email', 'link'] as const).map((value) => (
            <label key={value} className="flex items-center gap-2 text-sm">
              <input
                type="radio"
                value={value}
                {...register('kind', { onChange: () => onKindChange(value) })}
              />
              {t(`invitations.form.kinds.${value}`)}
            </label>
          ))}
        </div>
      </fieldset>

      {kind === 'email' && (
        <Field label={t('invitations.form.email')} error={fieldError(errors.email)}>
          {(control) => (
            <Input type="email" autoComplete="off" {...control} {...register('email')} />
          )}
        </Field>
      )}

      <Field
        label={t('invitations.form.role')}
        hint={kind === 'link' ? t('invitations.form.linkRoleHint') : undefined}
        error={fieldError(errors.role)}
      >
        {(control) => (
          <Select {...control} {...register('role')}>
            {rolesForKind.map((role) => (
              <option key={role} value={role}>
                {roleLabel(role)}
              </option>
            ))}
          </Select>
        )}
      </Field>

      <Button type="submit" loading={create.isPending} className="self-start">
        {kind === 'email' ? t('invitations.form.sendEmail') : t('invitations.form.createLink')}
      </Button>

      {created && <CreatedNotice created={created} />}
    </form>
  )
}

function CreatedNotice({ created }: { created: InvitationCreated }) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)

  if (!created.url) {
    return (
      <Alert tone="success">
        {t('invitations.created.email', { email: created.invitation.email })}
      </Alert>
    )
  }

  const url = created.url
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(url)
      setCopied(true)
    } catch {
      /* the field stays selectable so it can be copied by hand */
    }
  }

  return (
    <Alert tone="success" title={t('invitations.created.linkTitle')}>
      <p>{t('invitations.created.linkWarning')}</p>
      <div className="mt-2 flex gap-2">
        <Input
          readOnly
          value={url}
          aria-label={t('invitations.created.linkLabel')}
          onFocus={(e) => e.currentTarget.select()}
        />
        <Button variant="secondary" onClick={copy}>
          {copied ? t('common.copied') : t('common.copy')}
        </Button>
      </div>
    </Alert>
  )
}
