// EL-112 support: a public Privacy Policy reachable from the landing footer, and
// an email-consent notice on the login screen — both back the LINE Login
// email-permission review.
import { test, expect } from '../fixtures';

test.describe('privacy policy (EL-112)', () => {
  test('is reachable from the landing footer', async ({ page }) => {
    await page.goto('/');
    const link = page.getByTestId('footer-privacy-link');
    await expect(link).toBeVisible();
    await link.click();
    await expect(page).toHaveURL(/\/privacy$/);
    await expect(page.getByTestId('privacy-title')).toBeVisible();
    await expect(page.getByTestId('privacy-section-email')).toBeVisible();
  });

  test('login shows the email-consent notice linking the policy', async ({ page }) => {
    await page.goto('/login');
    const consent = page.getByTestId('login-consent');
    await expect(consent).toBeVisible();
    await expect(consent.getByRole('link')).toHaveAttribute('href', '/privacy');
  });
});
