// EL-085: the Soundcheck landing lead form. A CTA reveals the form; a valid
// submit posts to POST /api/leads and shows the success state.
import { test, expect } from '../fixtures';

test.describe('landing · lead form', () => {
  test('a CTA opens the lead form and a valid submit shows success', async ({ page }) => {
    await page.goto('/');

    // A CTA button (openLeadForm) reveals the form.
    await page.locator('.lp-btn-amber').first().click();

    const form = page.locator('app-lead-form');
    await expect(form).toBeVisible();

    await form.locator('input[name="name"]').fill('E2E Prospect');
    await form.locator('input[name="organization"]').fill('E2E DJ School');
    await form.locator('input[name="email"]').fill('prospect@e2e.test');
    await form.locator('textarea[name="message"]').fill('Interested in a demo.');

    await form.locator('button[type="submit"]').click();

    // POST /api/leads succeeds → the form swaps to the success message.
    await expect(page.locator('.lead-form-success')).toBeVisible();
  });

  test('submit stays disabled until name + valid email are present', async ({ page }) => {
    await page.goto('/');
    await page.locator('.lp-btn-amber').first().click();

    const form = page.locator('app-lead-form');
    const submit = form.locator('button[type="submit"]');
    await expect(submit).toBeDisabled();

    await form.locator('input[name="name"]').fill('Someone');
    await form.locator('input[name="email"]').fill('not-an-email');
    await expect(submit).toBeDisabled();

    await form.locator('input[name="email"]').fill('someone@e2e.test');
    await expect(submit).toBeEnabled();
  });
});
