import { safeRedirect } from './safe-redirect'

describe('safeRedirect', () => {
  it.each([
    ['/', '/'],
    ['/workspaces/abc/members', '/workspaces/abc/members'],
    ['/invitations/abc?x=1#top', '/invitations/abc?x=1#top'],
  ])('accepts the in-app path %s', (input, expected) => {
    expect(safeRedirect(input)).toBe(expected)
  })

  it.each([
    undefined,
    null,
    42,
    '',
    'workspaces',
    'https://evil.example',
    '//evil.example',
    '/\\evil.example',
    'javascript:alert(1)',
    '/ok\nhttps://evil.example',
    '/a\\b',
  ])('ignores %s', (input) => {
    expect(safeRedirect(input)).toBeUndefined()
  })
})
