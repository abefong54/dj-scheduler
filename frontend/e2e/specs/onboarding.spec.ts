// EL-106/107: the onboarding guard runs on every admin route and calls /api/me.
// A user who already has a school (the seeded organizer owns "Test Academy")
// must NOT be bounced to /onboarding, and the sidebar must show their school
// name where "THE BOOTH" used to be. Requires the EL-105 /api/me backend.
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE } from '../constants';

test.use({ storageState: ADMIN_STORAGE_STATE });

test.describe('onboarding guard + school name (EL-106/107)', () => {
  test('an onboarded user reaches the console (not redirected to onboarding)', async ({ page }) => {
    await page.goto('/admin/events');
    await expect(page).toHaveURL(/\/admin\/events$/);
    await expect(page.locator('.shell-sidebar')).toBeVisible();
  });

  test('the sidebar shows the school name instead of the workspace label', async ({ page }) => {
    await page.goto('/admin/events');
    // The seeded organizer owns "Test Academy" (backend/seed/seed_test.sql).
    await expect(page.locator('.shell-sidebar')).toContainText('Test Academy');
  });
});
