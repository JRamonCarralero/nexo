import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'

import en from './locales/en.json'
import es from './locales/es.json'

export const supportedLanguages = ['es', 'en'] as const
export type Language = (typeof supportedLanguages)[number]

function detectLanguage(): Language {
  const browser = typeof navigator === 'undefined' ? '' : navigator.language
  return browser.toLowerCase().startsWith('es') ? 'es' : 'en'
}

void i18n.use(initReactI18next).init({
  resources: { es: { translation: es }, en: { translation: en } },
  lng: detectLanguage(),
  fallbackLng: 'en',
  supportedLngs: supportedLanguages,
  interpolation: { escapeValue: false }, // React already escapes
})

export default i18n
