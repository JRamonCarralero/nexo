import { screen } from '@testing-library/react'

import { jsonResponse, renderWithProviders } from '@/test/render'

import { HealthStatus } from './HealthStatus'

describe('HealthStatus', () => {
  afterEach(() => vi.unstubAllGlobals())

  it('shows the API version when the service is healthy', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(
          jsonResponse({ status: 'ok', version: '1.2.3', checks: { database: 'ok' } }),
        ),
    )

    renderWithProviders(<HealthStatus />)

    expect(await screen.findByText('API conectada · v1.2.3')).toBeInTheDocument()
  })

  it('shows a translated message and the trace ID when the service is down', async () => {
    const problem = {
      title: 'Service Unavailable',
      status: 503,
      code: 'SERVICE_UNAVAILABLE',
      traceId: 'trace-1',
    }
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(jsonResponse(problem, 503, 'application/problem+json')),
    )

    renderWithProviders(<HealthStatus />)

    expect(await screen.findByText('API no disponible')).toBeInTheDocument()
    expect(screen.getByText(/El servicio no está disponible/)).toBeInTheDocument()
    expect(screen.getByText('Código de seguimiento: trace-1')).toBeInTheDocument()
  })
})
