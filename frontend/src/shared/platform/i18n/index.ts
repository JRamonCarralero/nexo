import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'

import { safeStorage } from '@/shared/storage'

import en from './locales/en.json'
import es from './locales/es.json'

export const supportedLanguages = ['es', 'en'] as const
export type Language = (typeof supportedLanguages)[number]

const STORAGE_KEY = 'nexo.language'

function isSupported(value: string | null): value is Language {
  return supportedLanguages.some((l) => l === value)
}

/** The language the person chose before, or else the browser's (Spanish if it starts with "es"). */
function initialLanguage(): Language {
  const stored = safeStorage.get(STORAGE_KEY)
  if (isSupported(stored)) return stored
  const browser = typeof navigator === 'undefined' ? '' : navigator.language
  return browser.toLowerCase().startsWith('es') ? 'es' : 'en'
}

void i18n.use(initReactI18next).init({
  resources: { es: { translation: es }, en: { translation: en } },
  lng: initialLanguage(),
  fallbackLng: 'en',
  supportedLngs: supportedLanguages,
  interpolation: { escapeValue: false }, // React already escapes
})

// Remember the choice and keep the page's language attribute in step (screen readers use it).
i18n.on('languageChanged', (language) => {
  if (isSupported(language)) safeStorage.set(STORAGE_KEY, language)
  document.documentElement.lang = language
})
document.documentElement.lang = i18n.language

export default i18n
