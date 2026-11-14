import { expect, test } from '@playwright/test'

import { expectAccessible } from './helpers'

// Guards the guard: an accessibility check that never fails would give false
// confidence to every other test in this folder.
test('the accessibility check catches a violation', async ({ page }) => {
  await page.setContent('<main><img src="x.png"><button></button></main>')

  await expect(expectAccessible(page)).rejects.toThrow(/image-alt|button-name/)
})
