// i18n: the EN / 中文 language switch in the navbar re-translates the UI.
import { test, expect } from '../fixtures';

test.describe('i18n', () => {
  test('toggling to 中文 and back re-translates the nav', async ({ page }) => {
    await page.goto('/login');

    // Default is English.
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
