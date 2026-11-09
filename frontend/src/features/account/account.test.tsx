import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import i18n from '@/shared/platform/i18n'
import { ana, createFakeAuth } from '@/test/fake-auth'
import { me, workspace, WS } from '@/test/fixtures'
import { mockApi } from '@/test/mock-api'
import { renderApp } from '@/test/render-app'

const routes = {
  'GET /api/v1/workspaces': { body: { items: [workspace()] } },
  [`GET /api/v1/workspaces/${WS}`]: { body: workspace() },
  'GET /api/v1/me': { body: me },
}

describe('email verification banner', () => {
  const unverified = { ...ana, emailVerified: false }

  it('asks people who signed up with a password to verify their email', async () => {
    mockApi(routes)
    renderApp({ path: `/workspaces/${WS}`, auth: createFakeAuth(unverified) })

    expect(await screen.findByText(/Verifica tu email \(ana@example.com\)/)).toBeInTheDocument()
  })

  it('stays hidden for verified accounts', async () => {
    mockApi(routes)
    renderApp({ path: `/workspaces/${WS}` })

    await screen.findByRole('heading', { name: 'Producto' })
    expect(screen.queryByText(/Verifica tu email/)).not.toBeInTheDocument()
  })

  it('resends the verification email', async () => {
    mockApi(routes)
    const auth = createFakeAuth(unverified)
    renderApp({ path: `/workspaces/${WS}`, auth })
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: 'Reenviar el email' }))

    expect(auth.sendVerificationEmail).toHaveBeenCalled()
    expect(
      await screen.findByText('Email enviado. Revisa tu bandeja de entrada.'),
    ).toBeInTheDocument()
  })

  it('disappears once the account is verified', async () => {
    mockApi(routes)
    const auth = createFakeAuth(unverified)
    auth.refreshUser.mockImplementationOnce(async () =>
      auth.setUser({ ...unverified, emailVerified: true }),
    )
    renderApp({ path: `/workspaces/${WS}`, auth })
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: 'Ya lo he verificado' }))

    await waitFor(() => expect(screen.queryByText(/Verifica tu email/)).not.toBeInTheDocument())
  })

  it('tells people when it is still not verified', async () => {
    mockApi(routes)
    renderApp({ path: `/workspaces/${WS}`, auth: createFakeAuth(unverified) })
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: 'Ya lo he verificado' }))

    expect(await screen.findByText(/Todavía no consta como verificado/)).toBeInTheDocument()
  })
})

describe('header', () => {
  it('shows who is signed in and signs out', async () => {
    mockApi({
      ...routes,
      'GET /api/v1/health': { body: { status: 'ok', version: 'x', checks: {} } },
    })
    const auth = createFakeAuth()
    const { history } = renderApp({ path: `/workspaces/${WS}`, auth })
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: 'Cerrar sesión' }))

    expect(auth.signOut).toHaveBeenCalled()
    await waitFor(() => expect(history.location.pathname).toBe('/login'))
  })

  it('changes the language and remembers it', async () => {
    mockApi(routes)
    renderApp({ path: `/workspaces/${WS}` })
    const user = userEvent.setup()

    await user.selectOptions(await screen.findByRole('combobox', { name: 'Idioma' }), 'English')

    expect(await screen.findByRole('link', { name: 'Members' })).toBeInTheDocument()
    expect(window.localStorage.getItem('nexo.language')).toBe('en')
    expect(document.documentElement.lang).toBe('en')
    await i18n.changeLanguage('es')
  })
})
