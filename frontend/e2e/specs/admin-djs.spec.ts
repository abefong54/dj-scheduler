// Admin DJ roster: add a DJ, delete a DJ (confirm dialog), and the shared
// data-table search filtering.
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE, SEED } from '../constants';

test.use({ storageState: ADMIN_STORAGE_STATE });

test.describe('admin · DJs', () => {
  test('lists the seeded DJs', async ({ page }) => {
    await page.goto('/admin/djs');
    await expect(page.getByText(SEED.djTesta.name)).toBeVisible();
    await expect(page.getByText(SEED.djBeta.name)).toBeVisible();
  });

  test('adds a DJ', async ({ page }) => {
    await page.goto('/admin/djs');
    await page.getByTestId('dj-name-input').fill('DJ Gamma');
    await page.getByTestId('dj-genres-input').fill('disco, funk');
    await page.getByTestId('dj-add-btn').click();

    await expect(page.getByText('DJ Gamma')).toBeVisible();
  });

  test('deletes a DJ after confirming the dialog', async ({ page }) => {
    await page.goto('/admin/djs');
    await page.getByTestId(`dj-delete-${SEED.djBeta.id}`).click();
    await page.getByTestId('dialog-confirm-btn').click();

    await expect(page.getByTestId(`dj-delete-${SEED.djBeta.id}`)).toHaveCount(0);
  });

  test('filters the DJ table via search', async ({ page }) => {
    await page.goto('/admin/djs');
    await page.locator('.dt-search').fill('Testa');

    await expect(page.getByTestId(`dj-delete-${SEED.djTesta.id}`)).toBeVisible();
    await expect(page.getByTestId(`dj-delete-${SEED.djBeta.id}`)).toHaveCount(0);
  });
});
