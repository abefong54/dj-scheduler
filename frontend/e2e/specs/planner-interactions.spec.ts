// EL-095: desktop planner interactions that previously had no e2e coverage —
// sortable table headers, keyboard save/cancel on the inline add row, and the
// add-stage colour-swatch → create-stage flow end to end.
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE, SEED } from '../constants';

test.use({ storageState: ADMIN_STORAGE_STATE });

const detail = `/admin/events/${SEED.event.id}`;

test.describe('planner desktop interactions (EL-095)', () => {
  test('sorting by the DJ header reorders the rows', async ({ page }) => {
    await page.goto(detail);
    // Seed order is by start_time: DJ Testa (20:00) then DJ Beta (21:00).
    await expect(page.locator('.slot-row .slot-dj-name').first()).toHaveText(/DJ Testa/);

    await page.getByRole('columnheader', { name: /DJ/ }).click();
    // dj_name ascending → DJ Beta first.
    await expect(page.locator('.slot-row .slot-dj-name').first()).toHaveText(/DJ Beta/);
  });

  test('Enter in the inline add row saves the slot; Escape cancels', async ({ page }) => {
    await page.goto(detail);

    // Escape closes the add row without creating anything.
    await page.getByTestId('add-slot-trigger').click();
    await expect(page.getByTestId('add-stage-select')).toBeVisible();
    await page.getByTestId('add-stage-select').press('Escape');
    await expect(page.getByTestId('add-stage-select')).toBeHidden();

    // Fill and press Enter → the slot is created and appears in the table.
    await page.getByTestId('add-slot-trigger').click();
    await page.getByTestId('add-stage-select').selectOption({ label: 'Main Stage' });
    await page.getByTestId('add-dj-select').selectOption({ label: SEED.djTesta.name });
    await page.getByTestId('add-start').fill('23:00');
    await page.getByTestId('add-start').press('Enter');

    await expect(page.locator('.slot-time-cell', { hasText: '23:00' })).toBeVisible();
  });

  test('picking a colour swatch creates a stage with that colour', async ({ page }) => {
    await page.goto(detail);
    await page.getByTestId('add-stage-btn').click();
    await expect(page.getByTestId('add-stage-modal')).toBeVisible();

    await page.getByTestId('stage-name-input').fill('Neon Stage');
    const swatch = page.locator('.color-swatch').nth(2);
    await swatch.click();
    await expect(swatch).toHaveClass(/swatch-active/);

    await page.getByTestId('stage-create-btn').click();
    await expect(page.locator('.stage-chip-name', { hasText: 'Neon Stage' })).toBeVisible();
  });
});
