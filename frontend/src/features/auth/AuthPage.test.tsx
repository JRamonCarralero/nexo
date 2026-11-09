import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { AuthError } from '@/shared/platform/auth'
import { createFakeAuth } from '@/test/fake-auth'
import { me, workspace, WS } from '@/test/fixtures'
import { mockApi } from '@/test/mock-api'
import { renderApp } from '@/test/render-app'

const health = { 'GET /api/v1/health': { body: { status: 'ok', version: 'x', checks: {} } } }
const afterLogin = {
  'GET /api/v1/workspaces': { body: { items: [workspace()] } },
  [`GET /api/v1/workspaces/${WS}`]: { body: workspace() },
  'GET /api/v1/me': { body: me },
}

function signedOutApp(path = '/login') {
  const auth = createFakeAuth(null)
  return renderApp({ path, auth })
}

describe('AuthPage', () => {
  it('signs in and goes to the workspace', async () => {
    mockApi({ ...health, ...afterLogin })
    const { auth, history } = signedOutApp()
    const user = userEvent.setup()

    await user.type(await screen.findByLabelText('Email'), 'ana@example.com')
    await user.type(screen.getByLabelText('Contraseña'), 'secret123')
    await user.click(screen.getByRole('button', { name: 'Iniciar sesión' }))

    expect(auth.signInWithPassword).toHaveBeenCalledWith('ana@example.com', 'secret123')
    await waitFor(() => expect(history.location.pathname).toBe(`/workspaces/${WS}`))
  })

  it('goes back to where the person was heading', async () => {
    mockApi({ ...health, ...afterLogin })
    const { history } = signedOutApp('/login?redirect=%2Fworkspaces%2Fabc%2Fmembers')
    const user = userEvent.setup()

    await user.type(await screen.findByLabelText('Email'), 'ana@example.com')
    await user.type(screen.getByLabelText('Contraseña'), 'secret123')
    await user.click(screen.getByRole('button', { name: 'Iniciar sesión' }))

    await waitFor(() => expect(history.location.pathname).toBe('/workspaces/abc/members'))
  })

  it('validates the fields before calling the identity provider', async () => {
    mockApi(health)
    const { auth } = signedOutApp()
    const user = userEvent.setup()

    await user.type(await screen.findByLabelText('Email'), 'not-an-email')
    await user.click(screen.getByRole('button', { name: 'Iniciar sesión' }))

    expect(await screen.findByText('Escribe un email válido.')).toBeInTheDocument()
    expect(screen.getByText('Este campo es obligatorio.')).toBeInTheDocument()
    expect(screen.getByLabelText('Email')).toHaveAttribute('aria-invalid', 'true')
    expect(auth.signInWithPassword).not.toHaveBeenCalled()
  })

  it('shows a translated message when the credentials are wrong', async () => {
    mockApi(health)
    const { auth } = signedOutApp()
    auth.signInWithPassword.mockRejectedValueOnce(new AuthError('INVALID_CREDENTIALS'))
    const user = userEvent.setup()

    await user.type(await screen.findByLabelText('Email'), 'ana@example.com')
    await user.type(screen.getByLabelText('Contraseña'), 'wrong-password')
    await user.click(screen.getByRole('button', { name: 'Iniciar sesión' }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'El email o la contraseña no son correctos.',
    )
  })

  it('creates an account with name, email and password', async () => {
    mockApi({ ...health, ...afterLogin })
    const { auth, history } = signedOutApp()
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: 'Crear una cuenta' }))
    await user.type(screen.getByLabelText('Nombre'), 'Ana García')
    await user.type(screen.getByLabelText('Email'), 'ana@example.com')
    await user.type(screen.getByLabelText('Contraseña'), 'secret123')
    await user.click(screen.getByRole('button', { name: 'Crear cuenta' }))

    expect(auth.signUpWithPassword).toHaveBeenCalledWith(
      'ana@example.com',
      'secret123',
      'Ana García',
    )
    await waitFor(() => expect(history.location.pathname).toBe(`/workspaces/${WS}`))
  })

  it('asks for a password of at least 8 characters when creating an account', async () => {
    mockApi(health)
    const { auth } = signedOutApp()
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: 'Crear una cuenta' }))
    await user.type(screen.getByLabelText('Nombre'), 'Ana')
    await user.type(screen.getByLabelText('Email'), 'ana@example.com')
    await user.type(screen.getByLabelText('Contraseña'), 'short')
    await user.click(screen.getByRole('button', { name: 'Crear cuenta' }))

    expect(
      await screen.findByText('La contraseña debe tener al menos 8 caracteres.'),
    ).toBeInTheDocument()
    expect(auth.signUpWithPassword).not.toHaveBeenCalled()
  })

  it('continues with Google', async () => {
    mockApi({ ...health, ...afterLogin })
    const { auth, history } = signedOutApp()
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: 'Continuar con Google' }))

    expect(auth.signInWithGoogle).toHaveBeenCalled()
    await waitFor(() => expect(history.location.pathname).toBe(`/workspaces/${WS}`))
  })

  it('does not show an error when the Google window is closed', async () => {
    mockApi(health)
    const { auth } = signedOutApp()
    auth.signInWithGoogle.mockRejectedValueOnce(new AuthError('CANCELLED'))
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: 'Continuar con Google' }))

    await waitFor(() => expect(auth.signInWithGoogle).toHaveBeenCalled())
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('sends a password reset link', async () => {
    mockApi(health)
    const { auth } = signedOutApp()
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: '¿Has olvidado tu contraseña?' }))
    await user.type(screen.getByLabelText('Email'), 'ana@example.com')
    await user.click(screen.getByRole('button', { name: 'Enviar enlace' }))

    expect(auth.sendPasswordReset).toHaveBeenCalledWith('ana@example.com')
    expect(await screen.findByText(/Si existe una cuenta con ese email/)).toBeInTheDocument()
  })
})
