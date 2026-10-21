import { jsonResponse } from '@/test/render'

import { ApiError, httpClient, NETWORK_ERROR, UNEXPECTED_RESPONSE } from './http-client'

describe('httpClient', () => {
  afterEach(() => vi.unstubAllGlobals())

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
})
