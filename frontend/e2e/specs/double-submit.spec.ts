// BUG-008: a double-clicked Create button must not create two records. The
// SubmitGuard disables the button (and ignores re-entry) while the request is in
// flight. We hold the POST open so the second click lands during that window.
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE } from '../constants';

test.use({ storageState: ADMIN_STORAGE_STATE });

test.describe('double-submit guard', () => {
  test('double-clicking "Add DJ" creates exactly one DJ', async ({ page }) => {
    await page.route('**/api/djs', async route => {
      if (route.request().method() === 'POST') {
        await new Promise(r => setTimeout(r, 600));
      }
      await route.continue();
    });

    await page.goto('/admin/djs');
    await page.getByTestId('dj-name-input').fill('Dup Guard DJ');
    await page.getByTestId('dj-genres-input').fill('house');

    const add = page.getByTestId('dj-add-btn');
    await add.click(); // fires the (held) POST and disables the button
    await add.click({ force: true }); // second physical click during the in-flight window → ignored

    // Roster reloads after the single create — the DJ appears exactly once.
    await expect(page.getByText('Dup Guard DJ')).toHaveCount(1);
  });

  test('double-clicking "Create Event" creates exactly one event', async ({ page }) => {
    await page.route('**/api/events', async route => {
      if (route.request().method() === 'POST') {
        await new Promise(r => setTimeout(r, 600));
      }
      await route.continue();
    });

    await page.goto('/admin/events/new');
    await page.getByTestId('event-name-input').fill('Dup Guard Event');
    await page.getByTestId('event-venue-input').fill('Test Venue');
    await page.getByTestId('event-start-date').fill('2026-12-01');
    await page.getByTestId('event-end-date').fill('2026-12-01');

    const create = page.getByTestId('event-create-btn');
    await create.click();
    await create.click({ force: true });

    // Success navigates to the list; the new event appears exactly once.
    await expect(page).toHaveURL(/\/admin\/events$/);
    await expect(page.getByText('Dup Guard Event')).toHaveCount(1);
  });
});
