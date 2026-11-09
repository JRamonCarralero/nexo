import { safeStorage } from '@/shared/storage'

const KEY = 'nexo.lastWorkspace'

/** The workspace the person was in last time, to open it straight away. */
export const lastWorkspace = {
  get: () => safeStorage.get(KEY),
  set: (id: string) => safeStorage.set(KEY, id),
}
