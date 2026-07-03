// EL-087: loading indicators. The branded VU-meter (app-loading-meter) fills a
// data region while its read is in flight, then content replaces it. We delay
// the API response so the meter is reliably observable.
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE, SEED } from '../constants';

test.use({ storageState: ADMIN_STORAGE_STATE });

test.describe('loading indicators', () => {
  test('events list shows the VU-meter while loading, then the content', async ({ page }) => {
    // Hold the events read open so the meter is on screen long enough to assert.
    await page.route('**/api/events', async route => {
      if (route.request().method() === 'GET') {
        await new Promise(r => setTimeout(r, 800));
      }
      await route.continue();
    });

    await page.goto('/admin/events');

    // While loading: the branded meter is visible, the table is not.
    await expect(page.locator('app-loading-meter')).toBeVisible();

    // Once the read resolves: content replaces the meter.
    await expect(page.getByText(SEED.event.name)).toBeVisible();
    await expect(page.locator('app-loading-meter')).toHaveCount(0);
  });
});
