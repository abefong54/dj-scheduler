// EL-111: the login page offers "Log in with LINE" (a second working SSO next to
// Google) and shows a friendly message when a LINE sign-in returns no email.
// The button is gated on environment.lineLoginEnabled (on in dev/e2e).
import { test, expect } from '../fixtures';

test.describe('login — LINE sign-in (EL-111)', () => {
  test('offers a LINE login that heads to the backend /auth/line route', async ({ page }) => {
    await page.goto('/login');

    const lineBtn = page.getByTestId('line-signin');
    await expect(lineBtn).toBeVisible();

    // Clicking navigates to the backend LINE redirect route — assert the request
    // is made (we don't drive the real external LINE consent, same as Google).
    const [req] = await Promise.all([
      page.waitForRequest('**/auth/line'),
      lineBtn.click(),
    ]);
    expect(req.url()).toContain('/auth/line');
  });

  test('shows the no-email message when redirected back with ?error=no_email', async ({ page }) => {
    await page.goto('/login?error=no_email');
    await expect(page.getByTestId('login-error')).toBeVisible();
  });
});
