import type { Problem } from './generated/model'

/**
 * Error thrown for any non-2xx response. `code` is the stable code from the
 * backend's AppError (e.g. TASK_NOT_FOUND) and is what the UI translates.
 */
export class ApiError<TBody = Problem> extends Error {
  readonly status: number
  readonly code: string
  readonly traceId?: string
  readonly problem?: TBody

  constructor(status: number, code: string, problem?: TBody, traceId?: string) {
    super(code)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.problem = problem
    this.traceId = traceId
  }
}

/** Codes used when the failure happens before getting a Problem from the API. */
export const NETWORK_ERROR = 'NETWORK_ERROR'
export const UNEXPECTED_RESPONSE = 'UNEXPECTED_RESPONSE'

/** Used by orval as the error type of every generated hook. */
export type ErrorType<TBody> = ApiError<TBody>

/**
 * Returns the current ID token (or null when signed out). With forceRefresh it
 * asks the identity provider for a new one. Registered at startup by the auth
 * adapter, so this module does not depend on any identity SDK.
 */
export type TokenProvider = (forceRefresh?: boolean) => Promise<string | null>

let tokenProvider: TokenProvider | null = null

export function setTokenProvider(provider: TokenProvider | null) {
  tokenProvider = provider
}

/** Code of the 401 the API answers for an expired or invalid token. */
const INVALID_TOKEN = 'INVALID_TOKEN'

/** Mutator used by every generated request (see orval.config.ts). */
export const httpClient = async <T>(url: string, init?: RequestInit): Promise<T> => {
  const res = await send(url, init, false)
  const body: unknown = res.status === 204 ? undefined : await res.json().catch(() => undefined)

  if (!res.ok) {
    if (isProblem(body)) {
      throw new ApiError(res.status, body.code, body, body.traceId)
    }
    throw new ApiError(
      res.status,
      UNEXPECTED_RESPONSE,
      undefined,
      res.headers.get('X-Trace-Id') ?? undefined,
    )
  }
  return body as T
}

/**
 * Sends the request with the bearer token. If the API says the token is
 * invalid (it may have expired), it asks for a fresh one and tries once more.
 */
async function send(
  url: string,
  init: RequestInit | undefined,
  refreshed: boolean,
): Promise<Response> {
  const headers = new Headers({ Accept: 'application/json, application/problem+json' })
  new Headers(init?.headers).forEach((value, key) => headers.set(key, value))

  const token = tokenProvider ? await tokenProvider(refreshed) : null
  if (token) headers.set('Authorization', `Bearer ${token}`)

  let res: Response
  try {
    res = await fetch(url, { ...init, headers })
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') throw err
    throw new ApiError(0, NETWORK_ERROR)
  }

  if (res.status === 401 && token && !refreshed) {
    const problem: unknown = await res
      .clone()
      .json()
      .catch(() => undefined)
    if (isProblem(problem) && problem.code === INVALID_TOKEN) return send(url, init, true)
  }
  return res
}

function isProblem(body: unknown): body is Problem {
  return typeof body === 'object' && body !== null && typeof (body as Problem).code === 'string'
}
