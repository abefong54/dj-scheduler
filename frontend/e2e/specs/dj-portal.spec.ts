// DJ portal (US-010 / US-011): a DJ opens their token link, sees their slots,
// and confirms or flags each one. Token-gated, no login.
import { test, expect } from '../fixtures';
import { SEED } from '../constants';

const PORTAL_URL = `/dj/portal?token=${SEED.djTesta.portalToken}`;
const slotId = SEED.slots.dj1; // DJ Testa's seeded slot

test.describe('DJ portal', () => {
  test('loads the DJ and their seeded slot', async ({ page }) => {
    await page.goto(PORTAL_URL);
    await expect(page.getByRole('heading', { name: SEED.djTesta.name })).toBeVisible();
    await expect(page.getByTestId(`confirm-slot-${slotId}`)).toBeVisible();
  });

  test('confirming a slot shows the confirmed status and disables Confirm', async ({ page }) => {
    await page.goto(PORTAL_URL);
    // Wait for the PATCH so we assert on the server-acknowledged state.
    const patch = page.waitForResponse(
      (r) => r.url().includes(`/api/dj/portal/slots/${slotId}`) && r.request().method() === 'PATCH',
    );
    await page.getByTestId(`confirm-slot-${slotId}`).click();
    await patch;

    await expect(page.getByTestId(`slot-status-${slotId}`)).toBeVisible();
    await expect(page.getByTestId(`confirm-slot-${slotId}`)).toBeDisabled();
  });

  test('flagging a slot shows the flagged status and disables Flag', async ({ page }) => {
    await page.goto(PORTAL_URL);
    const patch = page.waitForResponse(
      (r) => r.url().includes(`/api/dj/portal/slots/${slotId}`) && r.request().method() === 'PATCH',
    );
    await page.getByTestId(`flag-slot-${slotId}`).click();
    await patch;

    await expect(page.getByTestId(`slot-status-${slotId}`)).toBeVisible();
    await expect(page.getByTestId(`flag-slot-${slotId}`)).toBeDisabled();
  });

  test('an invalid token shows the error state, not the schedule', async ({ page }) => {
    await page.goto('/dj/portal?token=not-a-real-token');
    await expect(page.getByRole('heading', { name: SEED.djTesta.name })).toBeHidden();
    await expect(page.getByTestId(`confirm-slot-${slotId}`)).toHaveCount(0);
  });
});
