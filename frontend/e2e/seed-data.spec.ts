import { expect, test } from '@playwright/test'

import { expectAccessible, seed, signIn } from './helpers'

// These tests use the example data of `make db-seed`.

test('an admin sees the members and the pending invitation', async ({ page }) => {
  await page.goto('/login')
  await signIn(page, seed.ana, seed.password)

  await page
    .getByRole('combobox', { name: 'Espacio de trabajo' })
    .selectOption({ label: seed.workspace })
  await expect(page.getByRole('heading', { name: seed.workspace })).toBeVisible()
  await expectAccessible(page)

  await page.getByRole('link', { name: 'Miembros', exact: true }).click()
  await expect(page.getByRole('row', { name: /Luis Pérez/ })).toBeVisible()
  await expect(page.getByRole('row', { name: /Marta Soler/ })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Invitar personas' })).toBeVisible()
  await expect(page.getByText('nuevo@nexo.local')).toBeVisible()
  await expectAccessible(page)
})

test('a guest sees the members but cannot invite', async ({ page }) => {
  await page.goto('/login')
  await signIn(page, seed.marta, seed.password)

  await page.getByRole('link', { name: 'Miembros', exact: true }).click()
  await expect(page.getByRole('row', { name: /Ana García/ })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Invitar personas' })).toHaveCount(0)
  await expectAccessible(page)
})

test('wrong credentials show an error and stay on the sign-in page', async ({ page }) => {
  await page.goto('/login')
  await page.getByLabel('Email').fill(seed.ana)
  await page.getByLabel('Contraseña').fill('not-the-password')
  await page.getByRole('button', { name: 'Iniciar sesión', exact: true }).click()

  await expect(page.getByRole('alert')).toContainText('El email o la contraseña no son correctos.')
  await expect(page).toHaveURL(/\/login/)
  await expectAccessible(page)
})
