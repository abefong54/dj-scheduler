// i18n: the EN / 中文 language switch in the top bar re-translates the UI. The
// primary nav now lives in the Console sidebar (EL-067), which only renders on
// authenticated admin pages — so this runs against /admin/events, not /login.
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE } from '../constants';

test.use({ storageState: ADMIN_STORAGE_STATE });

test.describe('i18n', () => {
  test('toggling to 中文 and back re-translates the side nav', async ({ page }) => {
    await page.goto('/admin/events');

    // Default is English — the sidebar "Events" nav link.
    await expect(page.getByRole('link', { name: 'Events' })).toBeVisible();

    // Switch to Traditional Chinese.
    await page.getByRole('button', { name: '中文' }).click();
    await expect(page.getByRole('link', { name: '活動' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Events' })).toHaveCount(0);

    // Switch back to English.
    await page.getByRole('button', { name: 'EN' }).click();
    await expect(page.getByRole('link', { name: 'Events' })).toBeVisible();
  });
});
