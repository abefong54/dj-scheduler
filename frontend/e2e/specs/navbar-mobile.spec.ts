// The global top bar: the Soundcheck logo always links home, and below the lg
// breakpoint the language toggles + sign-in collapse behind a ☰ dropdown while
// the primary CTA stays in the bar.
import { test, expect } from '../fixtures';

test.describe('navbar · mobile (signed out)', () => {
  // Anonymous session and a phone-sized viewport.
  test.use({ storageState: { cookies: [], origins: [] }, viewport: { width: 390, height: 844 } });

  test('the logo links back to the landing page', async ({ page }) => {
    await page.goto('/login');
    await page.getByTestId('brand-wordmark').click();
    await expect(page).toHaveURL(/\/$/);
  });

  test('secondary items are hidden until the hamburger opens the dropdown', async ({ page }) => {
    await page.goto('/');
    const toggle = page.getByTestId('nav-menu-toggle');
    await expect(toggle).toBeVisible();

    // Collapsed: sign-in lives in the closed dropdown, so it is not visible.
    await expect(page.getByTestId('signin')).toBeHidden();

    await toggle.click();
    await expect(toggle).toHaveAttribute('aria-expanded', 'true');
    await expect(page.getByTestId('nav-menu')).toHaveClass(/navbar-menu-open/);
    await expect(page.getByTestId('signin')).toBeVisible();
  });

  test('the language toggle works from inside the dropdown, and picking a language closes it', async ({ page }) => {
    await page.goto('/');
    const toggle = page.getByTestId('nav-menu-toggle');
    await toggle.click();

    const zh = page.getByTestId('nav-menu').getByRole('button', { name: '中文' });
    await zh.click();

    // Picking a language also closes the dropdown (it's the same as tapping
    // any other item that leaves the menu, e.g. Sign in). The button is now
    // hidden (display: none), so re-open the menu to check the selection stuck.
    await expect(page.getByTestId('nav-menu')).not.toHaveClass(/navbar-menu-open/);
    await toggle.click();
    await expect(page.getByTestId('nav-menu').getByRole('button', { name: '中文' })).toHaveClass(/lang-btn-active/);
  });

  test('the primary CTA stays in the bar, not the dropdown', async ({ page }) => {
    await page.goto('/');
    // Book-a-demo is visible without opening the menu (it is not inside it).
    await expect(page.getByTestId('nav-book-demo')).toBeVisible();
  });
});

test.describe('navbar · desktop (signed out)', () => {
  test.use({ storageState: { cookies: [], origins: [] }, viewport: { width: 1280, height: 800 } });

  test('items are inline and the hamburger is hidden', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByTestId('nav-menu-toggle')).toBeHidden();
    // Sign-in is visible inline without any toggle.
    await expect(page.getByTestId('signin')).toBeVisible();
    await expect(page.getByTestId('nav-book-demo')).toBeVisible();
  });

  test('the logo still links home', async ({ page }) => {
    await page.goto('/login');
    await page.getByTestId('brand-wordmark').click();
    await expect(page).toHaveURL(/\/$/);
  });
});
