import { useTranslation } from 'react-i18next'

import { supportedLanguages, type Language } from '@/shared/platform/i18n'
import { Select } from '@/shared/ui'

const names: Record<Language, string> = { es: 'Español', en: 'English' }

export function LanguageSwitcher() {
  const { t, i18n } = useTranslation()
  const current = supportedLanguages.find((l) => i18n.language.startsWith(l)) ?? 'en'

  return (
    <Select
      aria-label={t('common.language')}
      value={current}
      onChange={(e) => void i18n.changeLanguage(e.target.value)}
      className="h-8 w-auto"
    >
      {supportedLanguages.map((lang) => (
        <option key={lang} value={lang}>
          {names[lang]}
        </option>
      ))}
    </Select>
  )
}
