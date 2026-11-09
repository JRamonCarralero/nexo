import { useTranslation } from 'react-i18next'

/** Formats an ISO date (e.g. from the API) as a short date in the current language. */
export function useFormatDate() {
  const { i18n } = useTranslation()
  return (iso: string) =>
    new Intl.DateTimeFormat(i18n.language, { dateStyle: 'medium' }).format(new Date(iso))
}
