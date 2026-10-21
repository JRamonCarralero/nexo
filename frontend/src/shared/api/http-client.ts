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

/** Mutator used by every generated request (see orval.config.ts). */
export const httpClient = async <T>(url: string, init?: RequestInit): Promise<T> => {
  let res: Response
  try {
    res = await fetch(url, {
      ...init,
      headers: { Accept: 'application/json, application/problem+json', ...init?.headers },
    })
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') throw err
    throw new ApiError(0, NETWORK_ERROR)
  }

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

function isProblem(body: unknown): body is Problem {
  return typeof body === 'object' && body !== null && typeof (body as Problem).code === 'string'
}
