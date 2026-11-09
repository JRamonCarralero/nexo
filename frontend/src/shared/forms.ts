import { useTranslation } from 'react-i18next'

/**
 * Turns a form field error into a translated message. Validation schemas use
 * translation keys as messages (e.g. 'validation.required'), so errors follow
 * the language chosen at the moment they are shown.
 */
export function useFieldError() {
  const { t } = useTranslation()
  return (error?: { message?: string }) => (error?.message ? t(error.message) : undefined)
}
