// EL-096: motion & view-transition foundation. Verifies, in a real browser, that
// (1) the calm motion tokens resolve on :root, and (2) client-side route changes
// go through the View Transitions API instead of hard-cutting — the "smooth
// between views" feel. Asserts final/observable state, never animation frames.
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE } from '../constants';

test.use({ storageState: ADMIN_STORAGE_STATE });

test.describe('view transitions + motion tokens (EL-096)', () => {
  test('motion tokens are defined on :root', async ({ page }) => {
    await page.goto('/admin/events');
    const tokens = await page.evaluate(() => {
      const s = getComputedStyle(document.documentElement);
      return {
        dur2: s.getPropertyValue('--dur-2').trim(),
        easeOut: s.getPropertyValue('--ease-out-soft').trim(),
      };
    });
    expect(tokens.dur2).toBe('240ms');
    expect(tokens.easeOut).not.toBe('');
  });

  test('client-side navigation uses the View Transitions API', async ({ page }) => {
    // Count startViewTransition calls across SPA navigations. addInitScript runs
    // once at the initial document load; the counter then persists through
    // client-side route changes (no full reload).
    await page.addInitScript(() => {
      const w = window as unknown as { __vtCalls: number };
      w.__vtCalls = 0;
      const orig = document.startViewTransition?.bind(document);
      if (orig) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (document as any).startViewTransition = (cb: any) => {
          w.__vtCalls++;
          return orig(cb);
        };
      }
    });

    await page.goto('/admin/events');
    await expect(page.locator('.shell-sidebar')).toBeVisible();

    await page.locator('[data-nav="djs"]').click();
    await expect(page).toHaveURL(/\/admin\/djs$/);

    const calls = await page.evaluate(() => (window as unknown as { __vtCalls: number }).__vtCalls);
    expect(calls).toBeGreaterThan(0);
  });

  test('navigation still works under reduced motion', async ({ page }) => {
    await page.emulateMedia({ reducedMotion: 'reduce' });
    await page.goto('/admin/events');
    await page.locator('[data-nav="djs"]').click();
    await expect(page).toHaveURL(/\/admin\/djs$/);
    await expect(page.locator('.shell-sidebar')).toBeVisible();
  });
});
