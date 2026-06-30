// Console side nav (AdminShell sidebar) — EL-067. The sidebar must be present
// and consistent on every admin page (the DJs page previously rendered none),
// and the brand must appear exactly once: in the top bar, not the sidebar.
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE } from '../constants';

test.use({ storageState: ADMIN_STORAGE_STATE });

test.describe('console side nav (EL-067)', () => {
  test('events page renders the sidebar with Events active', async ({ page }) => {
    await page.goto('/admin/events');
    await expect(page.locator('.shell-sidebar')).toBeVisible();
    await expect(page.locator('[data-nav="events"]')).toHaveClass(/shell-nav-active/);
  });

  test('djs page renders the sidebar with DJs active', async ({ page }) => {
    await page.goto('/admin/djs');
    await expect(page.locator('.shell-sidebar')).toBeVisible();
    await expect(page.locator('[data-nav="djs"]')).toHaveClass(/shell-nav-active/);
  });

  test('sidebar persists when navigating Events → DJs via the side nav', async ({ page }) => {
    await page.goto('/admin/events');
    await page.locator('[data-nav="djs"]').click();
    await expect(page).toHaveURL(/\/admin\/djs$/);
    await expect(page.locator('.shell-sidebar')).toBeVisible();
    await expect(page.locator('[data-nav="djs"]')).toHaveClass(/shell-nav-active/);
  });

  test('brand appears once — in the top bar, not the sidebar', async ({ page }) => {
    await page.goto('/admin/events');
    await expect(page.locator('.navbar-brand')).toHaveText('EventLineup');
    await expect(page.locator('.shell-sidebar .shell-wordmark')).toHaveCount(0);
  });
});
