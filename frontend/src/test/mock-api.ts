import { vi } from 'vitest'

export interface MockRequest {
  method: string
  path: string
  body: unknown
}

export interface MockResponse {
  status?: number
  body?: unknown
}

type Route = MockResponse | ((request: MockRequest) => MockResponse)

/** A Problem Details response, like the API sends for errors. */
export function problem(status: number, code: string): MockResponse {
  return { status, body: { title: 'Error', status, code, traceId: 'trace-test' } }
}

/**
 * Replaces fetch with a fake API. Routes are keyed "METHOD /path" (with the
 * /api/v1 prefix); the value is the response or a function building it.
 * Requests nobody expected are recorded in `unhandled` and answered with 500,
 * so a test can assert that the app only called what it should.
 */
export function mockApi(routes: Record<string, Route>) {
  const calls: MockRequest[] = []
  const unhandled: string[] = []

  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: string | URL | Request, init?: RequestInit) => {
      const method = (init?.method ?? 'GET').toUpperCase()
      const path = new URL(String(input), 'http://localhost').pathname
      const body = typeof init?.body === 'string' ? (JSON.parse(init.body) as unknown) : undefined
      const request = { method, path, body }
      calls.push(request)

      const route = routes[`${method} ${path}`]
      if (!route) {
        unhandled.push(`${method} ${path}`)
        return json(
          { title: 'Unexpected', status: 500, code: 'UNEXPECTED_IN_TEST', traceId: 't' },
          500,
        )
      }
      const { status = 200, body: responseBody } =
        typeof route === 'function' ? route(request) : route
      return status === 204 ? new Response(null, { status }) : json(responseBody, status)
    }),
  )

  return {
    calls,
    unhandled,
    called: (key: string) => calls.filter((c) => `${c.method} ${c.path}` === key),
  }
}

function json(body: unknown, status: number) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': status >= 400 ? 'application/problem+json' : 'application/json' },
  })
}
