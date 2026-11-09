import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { invitation, LUIS, me, member, ME, workspace, WS } from '@/test/fixtures'
import { mockApi, problem } from '@/test/mock-api'
import { renderApp } from '@/test/render-app'

describe('creating a workspace', () => {
  it('creates it and opens it', async () => {
    const created = workspace({ id: 'new-ws', name: 'Mi equipo' })
    const api = mockApi({
      'GET /api/v1/workspaces': { body: { items: [] } },
      'POST /api/v1/workspaces': { status: 201, body: created },
      'GET /api/v1/workspaces/new-ws': { body: created },
    })
    const { history } = renderApp({ path: '/workspaces/new' })
    const user = userEvent.setup()

    await user.type(await screen.findByLabelText('Nombre del espacio'), '  Mi equipo ')
    await user.click(screen.getByRole('button', { name: 'Crear espacio' }))

    await waitFor(() => expect(history.location.pathname).toBe('/workspaces/new-ws'))
    expect(api.called('POST /api/v1/workspaces')[0]?.body).toEqual({ name: 'Mi equipo' })
    expect(await screen.findByRole('heading', { name: 'Mi equipo' })).toBeInTheDocument()
  })

  it('requires a name', async () => {
    const api = mockApi({ 'GET /api/v1/workspaces': { body: { items: [] } } })
    renderApp({ path: '/workspaces/new' })
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: 'Crear espacio' }))

    expect(await screen.findByText('Este campo es obligatorio.')).toBeInTheDocument()
    expect(api.called('POST /api/v1/workspaces')).toHaveLength(0)
  })

  it('explains the error when the API rejects the name', async () => {
    mockApi({
      'GET /api/v1/workspaces': { body: { items: [workspace()] } },
      'POST /api/v1/workspaces': problem(400, 'INVALID_WORKSPACE_NAME'),
    })
    renderApp({ path: '/workspaces/new' })
    const user = userEvent.setup()

    await user.type(await screen.findByLabelText('Nombre del espacio'), 'x')
    await user.click(screen.getByRole('button', { name: 'Crear espacio' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('entre 1 y 100 caracteres')
  })
})

describe('workspace pages', () => {
  it('lets people switch workspace and create a new one from the header', async () => {
    const other = workspace({ id: 'other', name: 'Otro' })
    mockApi({
      'GET /api/v1/workspaces': { body: { items: [workspace(), other] } },
      [`GET /api/v1/workspaces/${WS}`]: { body: workspace() },
      'GET /api/v1/workspaces/other': { body: other },
    })
    const { history } = renderApp({ path: `/workspaces/${WS}` })
    const user = userEvent.setup()

    const switcher = await screen.findByRole('combobox', { name: 'Espacio de trabajo' })
    await user.selectOptions(switcher, 'Otro')
    await waitFor(() => expect(history.location.pathname).toBe('/workspaces/other'))

    await user.selectOptions(
      screen.getByRole('combobox', { name: 'Espacio de trabajo' }),
      '+ Nuevo espacio…',
    )
    await waitFor(() => expect(history.location.pathname).toBe('/workspaces/new'))
  })

  it('says so when the workspace does not exist or the person is not a member', async () => {
    mockApi({
      'GET /api/v1/workspaces': { body: { items: [workspace()] } },
      'GET /api/v1/workspaces/unknown': problem(404, 'WORKSPACE_NOT_FOUND'),
    })
    renderApp({ path: '/workspaces/unknown' })

    expect(await screen.findByRole('alert')).toHaveTextContent('no existe o no tienes acceso')
    expect(screen.getByRole('link', { name: 'Volver a mis espacios' })).toBeInTheDocument()
  })
})

describe('members page', () => {
  const baseRoutes = (role: 'admin' | 'member' | 'guest') => ({
    'GET /api/v1/workspaces': { body: { items: [workspace({ myRole: role })] } },
    [`GET /api/v1/workspaces/${WS}`]: { body: workspace({ myRole: role }) },
    'GET /api/v1/me': { body: me },
    [`GET /api/v1/workspaces/${WS}/members`]: {
      body: { items: [member(ME, 'Ana', 'admin'), member(LUIS, 'Luis', 'member')] },
    },
    [`GET /api/v1/workspaces/${WS}/invitations`]: { body: { items: [invitation()] } },
  })

  it('lets admins change roles, remove people and invite', async () => {
    const api = mockApi({
      ...baseRoutes('admin'),
      [`PATCH /api/v1/workspaces/${WS}/members/${LUIS}`]: { body: member(LUIS, 'Luis', 'guest') },
    })
    renderApp({ path: `/workspaces/${WS}/members` })
    const user = userEvent.setup()

    expect(await screen.findByRole('heading', { name: 'Miembros' })).toBeInTheDocument()
    expect(await screen.findByText('Invitar personas')).toBeInTheDocument()
    const table = screen.getByRole('table')
    expect(within(table).getByText('Ana')).toBeInTheDocument()
    expect(within(table).getByText('(tú)')).toBeInTheDocument()

    await user.selectOptions(
      screen.getByRole('combobox', { name: 'Cambiar el rol de Luis' }),
      'Invitado/a',
    )

    await waitFor(() =>
      expect(api.called(`PATCH /api/v1/workspaces/${WS}/members/${LUIS}`)[0]?.body).toEqual({
        role: 'guest',
      }),
    )
    expect(screen.getByRole('button', { name: 'Quitar a Luis' })).toBeInTheDocument()
  })

  it('asks for confirmation before removing someone', async () => {
    const api = mockApi({
      ...baseRoutes('admin'),
      [`DELETE /api/v1/workspaces/${WS}/members/${LUIS}`]: { status: 204 },
    })
    renderApp({ path: `/workspaces/${WS}/members` })
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: 'Quitar a Luis' }))
    expect(api.called(`DELETE /api/v1/workspaces/${WS}/members/${LUIS}`)).toHaveLength(0)

    await user.click(screen.getByRole('button', { name: 'Confirmar' }))
    await waitFor(() =>
      expect(api.called(`DELETE /api/v1/workspaces/${WS}/members/${LUIS}`)).toHaveLength(1),
    )
  })

  it('can be cancelled', async () => {
    const api = mockApi(baseRoutes('admin'))
    renderApp({ path: `/workspaces/${WS}/members` })
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: 'Quitar a Luis' }))
    await user.click(screen.getByRole('button', { name: 'Cancelar' }))

    expect(screen.getByRole('button', { name: 'Quitar a Luis' })).toBeInTheDocument()
    expect(api.calls.some((c) => c.method === 'DELETE')).toBe(false)
  })

  it('explains that a workspace needs an admin', async () => {
    mockApi({
      ...baseRoutes('admin'),
      [`PATCH /api/v1/workspaces/${WS}/members/${ME}`]: problem(409, 'LAST_ADMIN'),
    })
    renderApp({ path: `/workspaces/${WS}/members` })
    const user = userEvent.setup()

    await user.selectOptions(
      await screen.findByRole('combobox', { name: 'Cambiar el rol de Ana' }),
      'Miembro',
    )

    expect(await screen.findByRole('alert')).toHaveTextContent('necesita al menos un administrador')
  })

  it('shows members a read-only list, without management or invitations', async () => {
    mockApi(baseRoutes('member'))
    renderApp({ path: `/workspaces/${WS}/members` })

    expect(await screen.findByRole('heading', { name: 'Miembros' })).toBeInTheDocument()
    await screen.findByText('Luis')
    expect(screen.queryByRole('combobox', { name: /Cambiar el rol/ })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /Quitar a/ })).not.toBeInTheDocument()
    expect(screen.queryByText('Invitar personas')).not.toBeInTheDocument()
    expect(screen.getAllByRole('combobox').length).toBe(2) // header: workspace and language only
  })

  it('lets anyone leave and then returns to their workspaces', async () => {
    const api = mockApi({
      ...baseRoutes('member'),
      [`DELETE /api/v1/workspaces/${WS}/members/${ME}`]: { status: 204 },
    })
    const { history } = renderApp({ path: `/workspaces/${WS}/members` })
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: 'Salir de Producto' }))
    await user.click(screen.getByRole('button', { name: 'Confirmar' }))

    await waitFor(() =>
      expect(api.called(`DELETE /api/v1/workspaces/${WS}/members/${ME}`)).toHaveLength(1),
    )
    await waitFor(() => expect(history.location.pathname).not.toContain('/members'))
  })

  it('shows the role of each person', async () => {
    mockApi(baseRoutes('guest'))
    renderApp({ path: `/workspaces/${WS}/members` })

    const row = (await screen.findByText('Luis')).closest('tr')!
    expect(within(row).getByText('Miembro')).toBeInTheDocument()
  })
})
