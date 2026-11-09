import type { Invitation, Member, Role, Workspace } from '@/shared/api/generated/model'

export const WS = '0192f7a4-0000-7000-8000-0000000000aa'
export const ME = '0192f7a4-0000-7000-8000-000000000001'
export const LUIS = '0192f7a4-0000-7000-8000-000000000002'

export function workspace(overrides: Partial<Workspace> = {}): Workspace {
  return {
    id: WS,
    name: 'Producto',
    myRole: 'admin',
    createdAt: '2026-11-01T10:00:00Z',
    ...overrides,
  }
}

export function member(userId: string, name: string, role: Role = 'member'): Member {
  return {
    userId,
    email: `${name.toLowerCase()}@example.com`,
    name,
    avatarUrl: '',
    role,
    joinedAt: '2026-11-02T10:00:00Z',
  }
}

export function invitation(overrides: Partial<Invitation> = {}): Invitation {
  return {
    id: '0192f7a4-0000-7000-8000-0000000000bb',
    kind: 'email',
    email: 'bob@example.com',
    role: 'member',
    expiresAt: '2026-11-20T10:00:00Z',
    createdAt: '2026-11-13T10:00:00Z',
    ...overrides,
  }
}

export const me = {
  id: ME,
  email: 'ana@example.com',
  emailVerified: true,
  name: 'Ana',
  avatarUrl: '',
  isSuperadmin: false,
  createdAt: '2026-10-01T10:00:00Z',
}
