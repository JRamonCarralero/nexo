import '@testing-library/jest-dom/vitest'

import { cleanup, configure } from '@testing-library/react'
import { afterEach, beforeEach, vi } from 'vitest'

import i18n from '@/shared/platform/i18n'

// Routes are code-split: the first render of each loads a chunk, which can take
// over a second on a busy machine (a shared CI runner). Waiting longer costs
// nothing when things are fast and removes false failures when they are not.
configure({ asyncUtilTimeout: 5000 })

beforeEach(async () => {
  await i18n.changeLanguage('es')
})

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
  window.localStorage.clear()
})
