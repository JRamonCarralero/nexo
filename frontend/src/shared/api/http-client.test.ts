import { jsonResponse } from '@/test/render'

import {
  ApiError,
  httpClient,
  NETWORK_ERROR,
  setTokenProvider,
  UNEXPECTED_RESPONSE,
} from './http-client'

describe('httpClient', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    setTokenProvider(null)
  })

  it('returns the body of a successful response', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse({ status: 'ok' })))

    await expect(httpClient('/api/v1/health')).resolves.toEqual({ status: 'ok' })
  })

  it('throws an ApiError with the code and trace ID of a Problem response', async () => {
    const problem = { title: 'Not Found', status: 404, code: 'TASK_NOT_FOUND', traceId: 'abc' }
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(jsonResponse(problem, 404, 'application/problem+json')),
    )

    const err = await httpClient('/x').catch((e: unknown) => e)

    expect(err).toBeInstanceOf(ApiError)
    expect(err).toMatchObject({ status: 404, code: 'TASK_NOT_FOUND', traceId: 'abc' })
  })

  it('uses UNEXPECTED_RESPONSE when the error body is not a Problem', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('<html>', { status: 502 })))

    await expect(httpClient('/x')).rejects.toMatchObject({ status: 502, code: UNEXPECTED_RESPONSE })
  })

  it('uses NETWORK_ERROR when the request cannot be sent', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))

    await expect(httpClient('/x')).rejects.toMatchObject({ status: 0, code: NETWORK_ERROR })
  })

  describe('authentication', () => {
    const problem = (code: string) =>
      jsonResponse(
        { title: 'Unauthorized', status: 401, code, traceId: 't' },
        401,
        'application/problem+json',
      )
    const bearerOf = (fetchMock: ReturnType<typeof vi.fn>, call: number) =>
      new Headers((fetchMock.mock.calls[call]?.[1] as RequestInit).headers).get('Authorization')

    it('sends the bearer token of the registered provider', async () => {
      const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ ok: true }))
      vi.stubGlobal('fetch', fetchMock)
      setTokenProvider(async () => 'token-1')

      await httpClient('/x')

      expect(bearerOf(fetchMock, 0)).toBe('Bearer token-1')
    })

    it('sends no Authorization header when signed out', async () => {
      const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ ok: true }))
      vi.stubGlobal('fetch', fetchMock)
      setTokenProvider(async () => null)

      await httpClient('/x')

      expect(bearerOf(fetchMock, 0)).toBeNull()
    })

    it('keeps the headers the caller sets', async () => {
      const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ ok: true }))
      vi.stubGlobal('fetch', fetchMock)

      await httpClient('/x', { method: 'POST', headers: { 'Content-Type': 'application/json' } })

      const headers = new Headers((fetchMock.mock.calls[0]?.[1] as RequestInit).headers)
      expect(headers.get('Content-Type')).toBe('application/json')
      expect(headers.get('Accept')).toContain('application/json')
    })

    it('retries once with a fresh token when the API says the token is invalid', async () => {
      const fetchMock = vi
        .fn()
        .mockResolvedValueOnce(problem('INVALID_TOKEN'))
        .mockResolvedValueOnce(jsonResponse({ ok: true }))
      vi.stubGlobal('fetch', fetchMock)
      const provider = vi.fn(async (forceRefresh?: boolean) => (forceRefresh ? 'fresh' : 'stale'))
      setTokenProvider(provider)

      await expect(httpClient('/x')).resolves.toEqual({ ok: true })

      expect(provider).toHaveBeenNthCalledWith(1, false)
      expect(provider).toHaveBeenNthCalledWith(2, true)
      expect(bearerOf(fetchMock, 1)).toBe('Bearer fresh')
    })

    it('does not retry forever: a second INVALID_TOKEN is reported', async () => {
      const fetchMock = vi.fn().mockImplementation(async () => problem('INVALID_TOKEN'))
      vi.stubGlobal('fetch', fetchMock)
      setTokenProvider(async () => 'token')

      await expect(httpClient('/x')).rejects.toMatchObject({ status: 401, code: 'INVALID_TOKEN' })
      expect(fetchMock).toHaveBeenCalledTimes(2)
    })

    it('does not retry other 401 errors', async () => {
      const fetchMock = vi.fn().mockResolvedValue(problem('EMAIL_REQUIRED'))
      vi.stubGlobal('fetch', fetchMock)
      setTokenProvider(async () => 'token')

      await expect(httpClient('/x')).rejects.toMatchObject({ code: 'EMAIL_REQUIRED' })
      expect(fetchMock).toHaveBeenCalledTimes(1)
    })
  })
})
