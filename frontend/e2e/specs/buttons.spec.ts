// Button standard (app-button): the admin action buttons became icon-only, so
// these specs guard that the icon conversion kept each action operable and
// accessibly labelled, and exercise the icon-driven flows the refactor added
// (copy-portal-link toggle, view-public, export). The add/edit/delete flows
// themselves stay covered by admin-djs.spec.ts / admin-slots.spec.ts via the
// preserved data-testids.
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE, SEED } from '../constants';

test.use({ storageState: ADMIN_STORAGE_STATE, permissions: ['clipboard-write'] });

test.describe('button standard · DJ roster', () => {
  test('row actions are icon buttons with accessible names', async ({ page }) => {
    await page.goto('/admin/djs');

    // Icon-only buttons carry no visible text, so each must expose an aria-label
    // (i18n-agnostic check: just that it is present and non-empty).
    for (const action of ['dj-edit', 'dj-copy', 'dj-delete']) {
      const btn = page.getByTestId(`${action}-${SEED.djTesta.id}`);
      await expect(btn).toBeVisible();
      await expect(btn).toHaveAttribute('aria-label', /\S/);
      // The glyph is a real inline icon, not text.
      await expect(btn.locator('svg')).toBeVisible();
    }
  });

  test('edit icon button opens the edit panel', async ({ page }) => {
    await page.goto('/admin/djs');
    await page.getByTestId(`dj-edit-${SEED.djTesta.id}`).click();
    await expect(page.getByTestId('dj-edit-panel')).toBeVisible();
  });

  test('delete icon button opens the confirm dialog and removes the DJ', async ({ page }) => {
    await page.goto('/admin/djs');
    await page.getByTestId(`dj-delete-${SEED.djBeta.id}`).click();
    await expect(page.getByTestId('dialog-confirm-btn')).toBeVisible();
    await page.getByTestId('dialog-confirm-btn').click();
    await expect(page.getByTestId(`dj-delete-${SEED.djBeta.id}`)).toHaveCount(0);
  });

  test('copy-portal-link button swaps to the "copied" check icon', async ({ page }) => {
    await page.goto('/admin/djs');
    const copyBtn = page.getByTestId(`dj-copy-${SEED.djTesta.id}`);

    // Before: the link glyph.
    await expect(copyBtn.locator('svg')).toHaveAttribute('data-icon', 'link');

    await copyBtn.click();

    // After copying, the icon toggles to the check glyph (and reverts later).
    await expect(copyBtn.locator('svg')).toHaveAttribute('data-icon', 'check');
  });
});

test.describe('button standard · event detail', () => {
  const DETAIL = `/admin/events/${SEED.event.id}`;

  test('slot row actions are icon buttons with accessible names', async ({ page }) => {
    await page.goto(DETAIL);
    for (const action of ['slot-edit', 'slot-delete']) {
      const btn = page.getByTestId(`${action}-${SEED.slots.dj2}`);
      await expect(btn).toBeVisible();
      await expect(btn).toHaveAttribute('aria-label', /\S/);
      await expect(btn.locator('svg')).toBeVisible();
    }
  });

  test('view-public icon button opens the public schedule in a new tab', async ({ page }) => {
    await page.goto(DETAIL);
    const popupPromise = page.waitForEvent('popup');
    await page.getByTestId('event-view-public').click();
    const popup = await popupPromise;
    await expect(popup).toHaveURL(new RegExp(`/events/${SEED.event.id}`));
  });

  test('export icon button downloads the schedule as .xlsx', async ({ page }) => {
    await page.goto(DETAIL);
    const downloadPromise = page.waitForEvent('download');
    await page.getByTestId('event-export').click();
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toMatch(/\.xlsx$/);
  });
});
