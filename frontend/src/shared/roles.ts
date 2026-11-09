import { useTranslation } from 'react-i18next'

import type { Role } from '@/shared/api/generated/model'

/** Roles in the order they are offered. */
export const roles: Role[] = ['admin', 'member', 'guest']

export function useRoleLabel() {
  const { t } = useTranslation()
  return (role: Role) => t(`roles.${role}`)
}
