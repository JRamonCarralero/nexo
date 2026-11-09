import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'

import { App } from '@/app/App'
import { createFirebaseAuth, readFirebaseConfig } from '@/shared/platform/auth'
import '@/shared/platform/i18n'
import '@/styles.css'

// The one place where the identity provider is chosen and created.
const auth = createFirebaseAuth(readFirebaseConfig())

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App auth={auth} />
  </StrictMode>,
)
