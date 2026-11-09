/**
 * Validates the `redirect` search param used after signing in. Only paths of
 * this same app are accepted ("/invitations/abc"): anything else (another
 * site, "//evil.com", "/\evil.com", "javascript:…") is ignored. Without this,
 * a crafted login link could send people to a malicious site right after they
 * authenticate (an open redirect).
 */
export function safeRedirect(value: unknown): string | undefined {
  if (typeof value !== 'string') return undefined
  if (!value.startsWith('/') || value.startsWith('//') || value.startsWith('/\\')) return undefined
  // Control characters and backslashes can be normalized by browsers into a host.
  // eslint-disable-next-line no-control-regex
  if (/[\u0000-\u001f\\]/.test(value)) return undefined
  return value
}
