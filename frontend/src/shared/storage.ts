/**
 * localStorage that never throws: it can be unavailable or full (private
 * windows, blocked site data). Only for per-person conveniences such as the
 * last workspace or the language, never for anything that must persist.
 */
export const safeStorage = {
  get(key: string): string | null {
    try {
      return window.localStorage.getItem(key)
    } catch {
      return null
    }
  },
  set(key: string, value: string) {
    try {
      window.localStorage.setItem(key, value)
    } catch {
      /* ignore */
    }
  },
}
