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

  // EL-099: the add form is no longer an always-visible inline card — it opens in
  // a modal from the topbar "+ Add" button (uniform with the Events "+ New").
  test('adds a DJ via the topbar add modal', async ({ page }) => {
    await page.goto('/admin/djs');
    // No inline form on load; the fields live inside the modal.
    await expect(page.getByTestId('dj-name-input')).toHaveCount(0);

    await page.getByTestId('shell-new-btn').click();
    await expect(page.getByTestId('add-dj-modal')).toBeVisible();

    await page.getByTestId('dj-name-input').fill('DJ Gamma');
    await page.getByTestId('dj-genres-input').fill('disco, funk');
    await page.getByTestId('dj-add-btn').click();

    await expect(page.getByText('DJ Gamma')).toBeVisible();
    // Modal closes on a successful add.
    await expect(page.getByTestId('add-dj-modal')).toHaveCount(0);
  });

  test('the add modal opens from the topbar and Cancel closes it without adding', async ({ page }) => {
    await page.goto('/admin/djs');
    await page.getByTestId('shell-new-btn').click();
    await expect(page.getByTestId('add-dj-modal')).toBeVisible();

    await page.getByTestId('dj-add-cancel').click();
    await expect(page.getByTestId('add-dj-modal')).toHaveCount(0);
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

  // EL-020: edit a DJ's certifications + student status from the roster.
  test('edits a DJ — sets a certification and graduate status', async ({ page }) => {
    await page.goto('/admin/djs');
    await page.getByTestId(`dj-edit-${SEED.djTesta.id}`).click();
    await expect(page.getByTestId('dj-edit-panel')).toBeVisible();

    await page.getByTestId('dj-edit-cert-House').click();
    await page.getByTestId('dj-edit-graduate').click();
    await page.getByTestId('dj-edit-save').click();

    // Panel closes and the new certification chip appears on the roster row.
    await expect(page.getByTestId('dj-edit-panel')).toHaveCount(0);
    await expect(
      page.getByTestId(`dj-cert-${SEED.djTesta.id}`).filter({ hasText: 'House' }),
    ).toBeVisible();
  });
});
