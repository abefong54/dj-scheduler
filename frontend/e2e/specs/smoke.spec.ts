// Smoke tests — prove the full harness loop works end to end: seeded DB,
// minted admin storage state, dev-server proxy to the real backend, and all
// three auth contexts (public, admin, DJ-portal). These assert on seeded DATA
// (names), which is not translated, so they are i18n-agnostic.
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE, SEED } from '../constants';

test.describe('public + unauthenticated', () => {
  test('public schedule renders the seeded event', async ({ page }) => {
    await page.goto(`/events/${SEED.event.id}`);
    await expect(page.getByRole('heading', { name: SEED.event.name })).toBeVisible();
  });

  test('admin route redirects to /login when unauthenticated', async ({ page }) => {
    await page.goto('/admin/events');
    await expect(page).toHaveURL(/\/login$/);
  });

  test('DJ portal loads the DJ for a valid token', async ({ page }) => {
    await page.goto(`/dj/portal?token=${SEED.djTesta.portalToken}`);
    await expect(page.getByRole('heading', { name: SEED.djTesta.name })).toBeVisible();
  });
});

test.describe('authenticated admin', () => {
  test.use({ storageState: ADMIN_STORAGE_STATE });

  test('events list shows the seeded event', async ({ page }) => {
    await page.goto('/admin/events');
    await expect(page.getByText(SEED.event.name)).toBeVisible();
  });
});
