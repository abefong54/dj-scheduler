// A logged-in visitor who lands on the public marketing page needs a way into
// their console. On the landing route the top-bar primary CTA becomes "Go to
// console" when — and only when — the visitor is authenticated (otherwise it is
// the prospect-facing "Book a demo").
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE } from '../constants';

test.describe('landing · console entry (signed in)', () => {
  test.use({ storageState: ADMIN_STORAGE_STATE });

  test('the top bar links into the console and Book-a-demo is gone', async ({ page }) => {
    await page.goto('/');

    await expect(page.getByTestId('nav-book-demo')).toHaveCount(0);

    const cta = page.getByTestId('nav-console');
    await expect(cta).toBeVisible();

    await cta.click();
    await expect(page).toHaveURL(/\/admin\/events/);
  });
});

test.describe('landing · console entry (signed out)', () => {
  // Force an empty session — no minted JWT — regardless of any default state.
  test.use({ storageState: { cookies: [], origins: [] } });

  test('an anonymous visitor sees Book-a-demo and no console link', async ({ page }) => {
    await page.goto('/');

    await expect(page.getByTestId('nav-console')).toHaveCount(0);
    await expect(page.getByTestId('nav-book-demo')).toBeVisible();
  });
});
