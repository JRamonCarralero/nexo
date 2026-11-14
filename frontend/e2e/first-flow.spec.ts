import { expect, test } from '@playwright/test'

import { expectAccessible, signOut, signUp, uniqueEmail } from './helpers'

// The first milestone, end to end: register, create a workspace, invite
// someone, and that person joins.
test('registers, creates a workspace and invites someone who joins', async ({ page }) => {
  const workspace = `Equipo e2e ${Date.now().toString(36)}`

  // 1. Someone arrives without an account and is sent to sign in.
  await page.goto('/')
  await expect(page).toHaveURL(/\/login/)
  await expect(page.getByRole('heading', { name: 'Inicia sesión' })).toBeVisible()
  await expectAccessible(page)

  // 2. They create an account and land on the creation of their first workspace.
  await signUp(page, 'Olga Prueba', uniqueEmail('olga'))
  await expect(
    page.getByRole('heading', { name: 'Crea tu primer espacio de trabajo' }),
  ).toBeVisible()
  await expectAccessible(page)

  // 3. They create it and see its home.
  await page.getByLabel('Nombre del espacio').fill(workspace)
  await page.getByRole('button', { name: 'Crear espacio' }).click()
  await expect(page.getByRole('heading', { name: workspace })).toBeVisible()
  await expectAccessible(page)

  // 4. They open the members and create an invitation link.
  await page.getByRole('link', { name: 'Miembros', exact: true }).click()
  await expect(page.getByRole('heading', { name: 'Miembros' })).toBeVisible()
  await page.getByLabel('Con un enlace').check()
  await page.getByRole('button', { name: 'Crear enlace' }).click()
  const link = await page.getByLabel('Enlace de invitación', { exact: true }).inputValue()
  expect(link).toMatch(/\/invitations\/[\w-]{43}$/)
  await expectAccessible(page)

  // 5. Someone else opens the link, creates an account and joins.
  await signOut(page)
  await page.goto(link)
  await expect(page).toHaveURL(/\/login/) // the link is kept to come back to it
  await signUp(page, 'Pablo Invitado', uniqueEmail('pablo'))
  await expect(page.getByRole('heading', { name: `Te han invitado a ${workspace}` })).toBeVisible()
  await expectAccessible(page)
  await page.getByRole('button', { name: 'Unirme al espacio' }).click()
  await expect(page.getByRole('heading', { name: workspace })).toBeVisible()

  // 6. Both are now members.
  await page.getByRole('link', { name: 'Miembros', exact: true }).click()
  await expect(page.getByRole('row', { name: /Olga Prueba/ })).toBeVisible()
  await expect(page.getByRole('row', { name: /Pablo Invitado/ })).toBeVisible()
})

test('shows the sign-up and password recovery forms accessibly', async ({ page }) => {
  await page.goto('/login')

  await page.getByRole('button', { name: 'Crear una cuenta' }).click()
  await expect(page.getByRole('heading', { name: 'Crea tu cuenta' })).toBeVisible()
  await expectAccessible(page)

  await page.getByRole('button', { name: 'Iniciar sesión', exact: true }).click()
  await page.getByRole('button', { name: '¿Has olvidado tu contraseña?' }).click()
  await expect(page.getByRole('heading', { name: 'Recupera tu contraseña' })).toBeVisible()
  await expectAccessible(page)
})
