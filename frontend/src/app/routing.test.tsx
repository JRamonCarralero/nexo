import { screen, waitFor } from '@testing-library/react'

import { ana, createFakeAuth } from '@/test/fake-auth'
import { me, workspace, WS } from '@/test/fixtures'
import { mockApi } from '@/test/mock-api'
import { renderApp } from '@/test/render-app'

describe('route guards', () => {
  it('sends signed-out visitors to the login page, remembering where they were going', async () => {
    mockApi({ 'GET /api/v1/health': { body: { status: 'ok', version: 'x', checks: {} } } })

    const { history } = renderApp({ path: '/workspaces/abc/members', auth: createFakeAuth(null) })

    expect(await screen.findByRole('heading', { name: 'Inicia sesión' })).toBeInTheDocument()
    expect(history.location.pathname).toBe('/login')
    expect(history.location.search).toContain('redirect=%2Fworkspaces%2Fabc%2Fmembers')
  })

  it('protects invitation links too', async () => {
    mockApi({ 'GET /api/v1/health': { body: { status: 'ok', version: 'x', checks: {} } } })

    const { history } = renderApp({ path: '/invitations/some-token', auth: createFakeAuth(null) })

    expect(await screen.findByRole('heading', { name: 'Inicia sesión' })).toBeInTheDocument()
    expect(history.location.search).toContain('redirect=%2Finvitations%2Fsome-token')
  })

  it('ignores a redirect to another site', async () => {
    mockApi({ 'GET /api/v1/health': { body: { status: 'ok', version: 'x', checks: {} } } })

    const { history } = renderApp({
      path: '/login?redirect=https://evil.example',
      auth: createFakeAuth(null),
    })

    expect(await screen.findByRole('heading', { name: 'Inicia sesión' })).toBeInTheDocument()
    expect(history.location.search).not.toContain('evil')
  })

  it('takes people who are already signed in away from the login page', async () => {
    mockApi({
      'GET /api/v1/workspaces': { body: { items: [workspace()] } },
      [`GET /api/v1/workspaces/${WS}`]: { body: workspace() },
      'GET /api/v1/me': { body: me },
    })

    const { history } = renderApp({ path: '/login', auth: createFakeAuth(ana) })

    await waitFor(() => expect(history.location.pathname).toBe(`/workspaces/${WS}`))
  })

  it('lands on the last workspace used, or the first one', async () => {
    const other = workspace({ id: 'other', name: 'Otro' })
    mockApi({
      'GET /api/v1/workspaces': { body: { items: [workspace(), other] } },
      [`GET /api/v1/workspaces/${WS}`]: { body: workspace() },
      'GET /api/v1/workspaces/other': { body: other },
      'GET /api/v1/me': { body: me },
    })
    window.localStorage.setItem('nexo.lastWorkspace', 'other')

    const { history } = renderApp({ path: '/' })

    await waitFor(() => expect(history.location.pathname).toBe('/workspaces/other'))
    expect(await screen.findByRole('heading', { name: 'Otro' })).toBeInTheDocument()
  })

  it('asks people with no workspace to create their first one', async () => {
    mockApi({ 'GET /api/v1/workspaces': { body: { items: [] } } })

    const { history } = renderApp({ path: '/' })

    expect(
      await screen.findByRole('heading', { name: 'Crea tu primer espacio de trabajo' }),
    ).toBeInTheDocument()
    expect(history.location.pathname).toBe('/workspaces/new')
  })

  it('shows a not-found page for unknown addresses', async () => {
    mockApi({})

    renderApp({ path: '/nothing/here' })

    expect(await screen.findByRole('heading', { name: 'Página no encontrada' })).toBeInTheDocument()
  })
})
