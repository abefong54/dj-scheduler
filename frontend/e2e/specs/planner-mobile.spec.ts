// EL-094: at phone width the event planner renders the lineup as stacked cards
// and add/edit happens in a dialog (not the 7-col inline table). Verifies the
// cards render, the dialogs drive the same create/update paths, and the page
// never scrolls sideways.
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE, SEED } from '../constants';

test.use({ storageState: ADMIN_STORAGE_STATE, viewport: { width: 390, height: 844 } });

const detail = `/admin/events/${SEED.event.id}`;

test.describe('planner mobile — cards + dialog (EL-094)', () => {
  test('renders stacked cards instead of the table, with no horizontal scroll', async ({ page }) => {
    await page.goto(detail);
    await expect(page.locator('.slot-card').first()).toBeVisible();
    await expect(page.locator('.slot-table')).toHaveCount(0);
    const overflow = await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth + 1,
    );
    expect(overflow).toBe(true);
  });

  test('adds a slot through the add dialog', async ({ page }) => {
    await page.goto(detail);
    const cards = page.locator('.slot-card');
    const before = await cards.count();

    await page.getByTestId('add-slot-trigger').click();
    await expect(page.getByTestId('add-slot-dialog')).toBeVisible();
    await page.getByTestId('dialog-add-stage').selectOption({ label: 'Main Stage' });
    await page.getByTestId('dialog-add-dj').selectOption({ label: SEED.djTesta.name });
    await page.getByTestId('dialog-add-start').fill('22:30');
    await page.getByTestId('dialog-add-save').click();

    await expect(cards).toHaveCount(before + 1);
    await expect(page.locator('.slot-card', { hasText: '22:30' })).toBeVisible();
  });

  test('edits a slot through the edit dialog', async ({ page }) => {
    await page.goto(detail);
    await page.locator('[data-testid^="slot-edit-"]').first().click();
    await expect(page.getByTestId('edit-slot-dialog')).toBeVisible();

    await page.getByTestId('dialog-edit-start').fill('23:15');
    await page.getByTestId('dialog-edit-save').click();

    await expect(page.getByTestId('edit-slot-dialog')).toBeHidden();
    await expect(page.locator('.slot-card', { hasText: '23:15' })).toBeVisible();
  });
});
