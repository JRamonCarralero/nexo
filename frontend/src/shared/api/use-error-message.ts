import { useCallback } from 'react'
import { useTranslation } from 'react-i18next'

import { ApiError } from './http-client'

/**
 * Turns any error into a message in the user's language, using the stable
 * `code` from the backend (errors.<CODE> in the locale files).
 */
export function useErrorMessage() {
  const { t } = useTranslation()

  return useCallback(
    (error: unknown): string => {
      const fallback = t('errors.default')
      if (error instanceof ApiError) {
        return t(`errors.${error.code}`, { defaultValue: fallback })
      }
      return fallback
    },
    [t],
  )
}
