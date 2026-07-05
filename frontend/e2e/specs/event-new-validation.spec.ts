// Regression: the "Create" button on the new-event form used to silently do
// nothing when a required field was empty (no navigation, no message). On a
// phone the fiddly native date picker often leaves Start empty, so the button
// looked dead — "works in the webview but not on my iPhone". Pressing Create now
// surfaces per-field validation instead of no-oping, and still creates once the
// required fields are filled.
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE } from '../constants';

test.use({ storageState: ADMIN_STORAGE_STATE, viewport: { width: 390, height: 844 } });

test.describe('new event — required-field validation (mobile)', () => {
  test('pressing Create with the start date empty shows an error and does not navigate', async ({ page }) => {
    await page.goto('/admin/events/new');

    await page.getByTestId('event-name-input').fill('No Date Fest');
    await page.getByTestId('event-venue-input').fill('QA Arena');
    // Deliberately leave the required start date empty (the iOS date-picker case).

    await page.getByTestId('event-create-btn').click();

    // The button must not silently no-op: feedback appears and we stay on the form.
    await expect(page.getByTestId('event-start-error')).toBeVisible();
    await expect(page).toHaveURL(/\/admin\/events\/new$/);
  });

  test('filling the start date clears the error and Create then works', async ({ page }) => {
    await page.goto('/admin/events/new');

    await page.getByTestId('event-name-input').fill('Now Complete Fest');
    await page.getByTestId('event-venue-input').fill('QA Arena');
    await page.getByTestId('event-create-btn').click();
    await expect(page.getByTestId('event-start-error')).toBeVisible();

    // Supplying the missing field clears its error…
    await page.getByTestId('event-start-date').fill('2026-09-10');
    await expect(page.getByTestId('event-start-error')).toHaveCount(0);

    // …and now Create actually creates the event and returns to the list.
    await page.getByTestId('event-create-btn').click();
    await expect(page).toHaveURL(/\/admin\/events$/);
    await expect(page.getByText('Now Complete Fest')).toBeVisible();
  });
});
