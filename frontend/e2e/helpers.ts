import AxeBuilder from '@axe-core/playwright'
import { expect, type Page } from '@playwright/test'

/** The accounts and password created by `make db-seed` (backend/internal/devseed). */
export const seed = {
  password: 'nexo1234',
  ana: 'ana@nexo.local', // admin of "Lanzamiento" and "Marketing"
  marta: 'marta@nexo.local', // guest in "Lanzamiento"
  workspace: 'Lanzamiento',
}

/** An email that has never been used, so tests can be run again on the same database. */
export function uniqueEmail(name: string): string {
  return `${name}.${Date.now().toString(36)}${Math.random().toString(36).slice(2, 6)}@e2e.nexo.local`
}

/** Creates an account from the sign-in page and waits until it leaves it. */
export async function signUp(page: Page, name: string, email: string, password = 'e2e-password-1') {
  await page.getByRole('button', { name: 'Crear una cuenta' }).click()
  await page.getByLabel('Nombre').fill(name)
  await page.getByLabel('Email').fill(email)
  await page.getByLabel('Contraseña').fill(password)
  await page.getByRole('button', { name: 'Crear cuenta', exact: true }).click()
  await expect(page).not.toHaveURL(/\/login/)
}

/** Signs in from the sign-in page and waits until it leaves it. */
export async function signIn(page: Page, email: string, password: string) {
  await page.getByLabel('Email').fill(email)
  await page.getByLabel('Contraseña').fill(password)
  await page.getByRole('button', { name: 'Iniciar sesión', exact: true }).click()
  await expect(page).not.toHaveURL(/\/login/)
}

export async function signOut(page: Page) {
  await page.getByRole('button', { name: 'Cerrar sesión' }).click()
  await expect(page).toHaveURL(/\/login/)
}

/** Fails with the list of WCAG 2.1 A/AA violations axe finds on the current page. */
export async function expectAccessible(page: Page) {
  const { violations } = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
    .analyze()
  expect(
    violations.map(
      (v) => `${v.id}: ${v.help} (${v.nodes.map((n) => n.target.join(' ')).join(' | ')})`,
    ),
  ).toEqual([])
}
