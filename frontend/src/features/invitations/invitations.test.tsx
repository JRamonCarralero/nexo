import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { createFakeAuth } from '@/test/fake-auth'
import { invitation, LUIS, me, member, ME, workspace, WS } from '@/test/fixtures'
import { mockApi, problem } from '@/test/mock-api'
import { renderApp } from '@/test/render-app'

const TOKEN = 'A'.repeat(43)

const adminRoutes = {
  'GET /api/v1/workspaces': { body: { items: [workspace()] } },
  [`GET /api/v1/workspaces/${WS}`]: { body: workspace() },
  'GET /api/v1/me': { body: me },
  [`GET /api/v1/workspaces/${WS}/members`]: {
    body: { items: [member(ME, 'Ana', 'admin'), member(LUIS, 'Luis')] },
  },
}

describe('inviting people', () => {
  it('sends an invitation by email', async () => {
    const api = mockApi({
      ...adminRoutes,
      [`GET /api/v1/workspaces/${WS}/invitations`]: { body: { items: [] } },
      [`POST /api/v1/workspaces/${WS}/invitations`]: {
        status: 201,
        body: { invitation: invitation({ email: 'bob@example.com' }) },
      },
    })
    renderApp({ path: `/workspaces/${WS}/members` })
    const user = userEvent.setup()

    await user.type(await screen.findByLabelText('Email de la persona'), 'bob@example.com')
    await user.click(screen.getByRole('button', { name: 'Enviar invitación' }))

    expect(await screen.findByText('Invitación enviada a bob@example.com.')).toBeInTheDocument()
    expect(api.called(`POST /api/v1/workspaces/${WS}/invitations`)[0]?.body).toEqual({
      kind: 'email',
      email: 'bob@example.com',
      role: 'member',
    })
  })

  it('validates the email before sending', async () => {
    const api = mockApi({
      ...adminRoutes,
      [`GET /api/v1/workspaces/${WS}/invitations`]: { body: { items: [] } },
    })
    renderApp({ path: `/workspaces/${WS}/members` })
    const user = userEvent.setup()

    await user.type(await screen.findByLabelText('Email de la persona'), 'nope')
    await user.click(screen.getByRole('button', { name: 'Enviar invitación' }))

    expect(await screen.findByText('Escribe un email válido.')).toBeInTheDocument()
    expect(api.called(`POST /api/v1/workspaces/${WS}/invitations`)).toHaveLength(0)
  })

  it('shows the link once and offers to copy it', async () => {
    const url = `http://localhost:5173/invitations/${TOKEN}`
    const api = mockApi({
      ...adminRoutes,
      [`GET /api/v1/workspaces/${WS}/invitations`]: { body: { items: [] } },
      [`POST /api/v1/workspaces/${WS}/invitations`]: {
        status: 201,
        body: { invitation: invitation({ kind: 'link', email: undefined, role: 'guest' }), url },
      },
    })
    renderApp({ path: `/workspaces/${WS}/members` })
    const user = userEvent.setup()

    await user.click(await screen.findByLabelText('Con un enlace'))
    await user.selectOptions(screen.getByLabelText('Rol'), 'Invitado/a')
    await user.click(screen.getByRole('button', { name: 'Crear enlace' }))

    expect(await screen.findByLabelText('Enlace de invitación')).toHaveValue(url)
    expect(screen.getByText(/no se volverá a mostrar/)).toBeInTheDocument()
    expect(api.called(`POST /api/v1/workspaces/${WS}/invitations`)[0]?.body).toEqual({
      kind: 'link',
      role: 'guest',
    })

    await user.click(screen.getByRole('button', { name: 'Copiar' }))
    expect(await navigator.clipboard.readText()).toBe(url) // user-event provides the clipboard
    expect(await screen.findByRole('button', { name: '¡Copiado!' })).toBeInTheDocument()
  })

  it('does not offer the admin role for links, and says why', async () => {
    mockApi({
      ...adminRoutes,
      [`GET /api/v1/workspaces/${WS}/invitations`]: { body: { items: [] } },
    })
    renderApp({ path: `/workspaces/${WS}/members` })
    const user = userEvent.setup()

    // By email, the admin role is available.
    const roleSelect = await screen.findByLabelText('Rol')
    expect(within(roleSelect).getByRole('option', { name: 'Administrador/a' })).toBeInTheDocument()

    await user.selectOptions(roleSelect, 'Administrador/a')
    await user.click(screen.getByLabelText('Con un enlace'))

    expect(
      within(screen.getByLabelText('Rol')).queryByRole('option', { name: 'Administrador/a' }),
    ).toBeNull()
    expect(screen.getByLabelText('Rol')).toHaveValue('member')
    expect(screen.getByText(/no puede dar el rol de administrador/)).toBeInTheDocument()
  })

  it('shows the API error when the email cannot be sent', async () => {
    mockApi({
      ...adminRoutes,
      [`GET /api/v1/workspaces/${WS}/invitations`]: { body: { items: [] } },
      [`POST /api/v1/workspaces/${WS}/invitations`]: problem(503, 'EMAIL_DELIVERY_FAILED'),
    })
    renderApp({ path: `/workspaces/${WS}/members` })
    const user = userEvent.setup()

    await user.type(await screen.findByLabelText('Email de la persona'), 'bob@example.com')
    await user.click(screen.getByRole('button', { name: 'Enviar invitación' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('No se ha podido enviar el email')
  })

  it('lists pending invitations and revokes them', async () => {
    const pending = invitation({ email: 'bob@example.com' })
    const api = mockApi({
      ...adminRoutes,
      [`GET /api/v1/workspaces/${WS}/invitations`]: { body: { items: [pending] } },
      [`DELETE /api/v1/workspaces/${WS}/invitations/${pending.id}`]: { status: 204 },
    })
    renderApp({ path: `/workspaces/${WS}/members` })
    const user = userEvent.setup()

    expect(await screen.findByText('bob@example.com')).toBeInTheDocument()
    await user.click(
      screen.getByRole('button', { name: 'Revocar la invitación de bob@example.com' }),
    )

    await waitFor(() =>
      expect(api.called(`DELETE /api/v1/workspaces/${WS}/invitations/${pending.id}`)).toHaveLength(
        1,
      ),
    )
  })

  it('says when there is nothing pending', async () => {
    mockApi({
      ...adminRoutes,
      [`GET /api/v1/workspaces/${WS}/invitations`]: { body: { items: [] } },
    })
    renderApp({ path: `/workspaces/${WS}/members` })

    expect(await screen.findByText('No hay invitaciones pendientes.')).toBeInTheDocument()
  })
})

describe('accepting an invitation', () => {
  const preview = (overrides = {}) => ({
    workspaceName: 'Producto',
    kind: 'email',
    email: 'ana@example.com',
    role: 'member',
    invitedBy: 'Luis',
    expiresAt: '2026-11-20T10:00:00Z',
    status: 'pending',
    ...overrides,
  })

  it('shows who invites and joins the workspace', async () => {
    const api = mockApi({
      [`GET /api/v1/invitations/${TOKEN}`]: { body: preview() },
      [`POST /api/v1/invitations/${TOKEN}/accept`]: { body: workspace({ myRole: 'member' }) },
      'GET /api/v1/workspaces': { body: { items: [workspace({ myRole: 'member' })] } },
      [`GET /api/v1/workspaces/${WS}`]: { body: workspace({ myRole: 'member' }) },
    })
    const { history } = renderApp({ path: `/invitations/${TOKEN}` })
    const user = userEvent.setup()

    expect(
      await screen.findByRole('heading', { name: 'Te han invitado a Producto' }),
    ).toBeInTheDocument()
    expect(screen.getByText('Luis te invita a unirte como Miembro.')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Unirme al espacio' }))

    await waitFor(() => expect(history.location.pathname).toBe(`/workspaces/${WS}`))
    expect(api.called(`POST /api/v1/invitations/${TOKEN}/accept`)).toHaveLength(1)
  })

  it('warns when the person is signed in with another email', async () => {
    mockApi({
      [`GET /api/v1/invitations/${TOKEN}`]: { body: preview({ email: 'bob@example.com' }) },
    })
    renderApp({ path: `/invitations/${TOKEN}` })

    expect(
      await screen.findByText(
        /es para bob@example.com, pero has iniciado sesión como ana@example.com/,
      ),
    ).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Unirme al espacio' })).toBeDisabled()
  })

  it('lets the person switch account and come back to the invitation', async () => {
    mockApi({
      [`GET /api/v1/invitations/${TOKEN}`]: { body: preview({ email: 'bob@example.com' }) },
      'GET /api/v1/health': { body: { status: 'ok', version: 'x', checks: {} } },
    })
    const { auth, history } = renderApp({ path: `/invitations/${TOKEN}` })
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: 'Usar otra cuenta' }))

    expect(auth.signOut).toHaveBeenCalled()
    await waitFor(() => expect(history.location.pathname).toBe('/login'))
    expect(decodeURIComponent(history.location.search)).toContain(`/invitations/${TOKEN}`)
  })

  it.each([
    ['expired', 'ha caducado'],
    ['revoked', 'ha sido revocada'],
    ['accepted', 'ya se ha utilizado'],
  ])('explains that a %s invitation cannot be used', async (status, text) => {
    mockApi({ [`GET /api/v1/invitations/${TOKEN}`]: { body: preview({ status }) } })
    renderApp({ path: `/invitations/${TOKEN}` })

    expect(await screen.findByText(new RegExp(text))).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Unirme al espacio' })).not.toBeInTheDocument()
  })

  it('explains an unknown invitation', async () => {
    mockApi({ [`GET /api/v1/invitations/${TOKEN}`]: problem(404, 'INVITATION_NOT_FOUND') })
    renderApp({ path: `/invitations/${TOKEN}` })

    expect(await screen.findByRole('alert')).toHaveTextContent('no existe o ya no es válida')
  })

  it('asks to verify the email when the API requires it', async () => {
    const auth = createFakeAuth({
      uid: 'u1',
      email: 'ana@example.com',
      emailVerified: false,
      displayName: 'Ana',
    })
    mockApi({
      [`GET /api/v1/invitations/${TOKEN}`]: { body: preview() },
      [`POST /api/v1/invitations/${TOKEN}/accept`]: problem(403, 'EMAIL_NOT_VERIFIED'),
    })
    renderApp({ path: `/invitations/${TOKEN}`, auth })
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: 'Unirme al espacio' }))

    const alerts = await screen.findAllByRole('alert')
    expect(alerts.some((a) => /Verifica tu dirección de email/.test(a.textContent ?? ''))).toBe(
      true,
    )
    // The banner offers the way to fix it, right on this page.
    expect(screen.getByRole('button', { name: 'Ya lo he verificado' })).toBeInTheDocument()
  })

  it('drops the "verify your email" error once the account is verified', async () => {
    const unverified = {
      uid: 'u1',
      email: 'ana@example.com',
      emailVerified: false,
      displayName: 'Ana',
    }
    const auth = createFakeAuth(unverified)
    mockApi({
      [`GET /api/v1/invitations/${TOKEN}`]: { body: preview() },
      [`POST /api/v1/invitations/${TOKEN}/accept`]: problem(403, 'EMAIL_NOT_VERIFIED'),
    })
    renderApp({ path: `/invitations/${TOKEN}`, auth })
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: 'Unirme al espacio' }))
    expect(
      await screen.findByText('Verifica tu dirección de email para aceptar la invitación.'),
    ).toBeInTheDocument()

    auth.setUser({ ...unverified, emailVerified: true })

    await waitFor(() =>
      expect(
        screen.queryByText('Verifica tu dirección de email para aceptar la invitación.'),
      ).not.toBeInTheDocument(),
    )
  })
})
