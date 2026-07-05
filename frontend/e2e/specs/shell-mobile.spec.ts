// EL-093: the shared admin shell must be usable on a phone — the fixed 240px
// sidebar previously never collapsed and the topbar actions overflowed
// off-screen. At a 390px viewport the sidebar is an off-canvas drawer behind a
// ☰ button, the projected actions fold into a ⋯ menu, and nothing forces a
// horizontal page scroll.
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE, SEED } from '../constants';

test.use({ storageState: ADMIN_STORAGE_STATE, viewport: { width: 390, height: 844 } });

const noHorizontalOverflow = (page: import('@playwright/test').Page) =>
  page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1);

test.describe('responsive admin shell — mobile (EL-093)', () => {
  test('no horizontal page scroll on the events page', async ({ page }) => {
    await page.goto('/admin/events');
    await expect(page.locator('.shell-hamburger')).toBeVisible();
    expect(await noHorizontalOverflow(page)).toBe(true);
  });

  test('the sidebar is off-canvas until the hamburger opens it', async ({ page }) => {
    await page.goto('/admin/events');
    const sidebar = page.locator('.shell-sidebar');

    // Off-canvas: translated fully out of the viewport. toBeInViewport auto-
    // retries, so it settles past the slide animation rather than sampling a
    // mid-transition frame.
    await expect(sidebar).not.toBeInViewport();

    await page.locator('.shell-hamburger').click();
    await expect(sidebar).toHaveClass(/shell-sidebar-open/);
    await expect(sidebar).toBeInViewport();

    // Backdrop dismisses it.
    await page.locator('.shell-backdrop').click();
    await expect(sidebar).not.toHaveClass(/shell-sidebar-open/);
    await expect(sidebar).not.toBeInViewport();
  });

  test('projected topbar actions are reachable via the ⋯ menu', async ({ page }) => {
    await page.goto(`/admin/events/${SEED.event.id}`);
    await expect(page.locator('.shell-hero, .detail-hero').first()).toBeVisible();

    // The export action is off in the overflow menu, hidden until opened.
    const exportBtn = page.getByTestId('event-export');
    await expect(exportBtn).toBeHidden();

    await page.locator('.shell-actions-toggle').click();
    await expect(page.locator('.shell-actions-group')).toHaveClass(/open/);
    await expect(exportBtn).toBeVisible();

    expect(await noHorizontalOverflow(page)).toBe(true);
  });
});
