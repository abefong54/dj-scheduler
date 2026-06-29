// Admin event management: list, create, delete (with confirm dialog), clone.
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE, SEED } from '../constants';

test.use({ storageState: ADMIN_STORAGE_STATE });

test.describe('admin · events', () => {
  test('lists the seeded event', async ({ page }) => {
    await page.goto('/admin/events');
    await expect(page.getByText(SEED.event.name)).toBeVisible();
  });

  test('creates a new event and shows it in the list', async ({ page }) => {
    await page.goto('/admin/events/new');
    await page.getByTestId('event-name-input').fill('Created By Test');
    await page.getByTestId('event-venue-input').fill('QA Venue');
    await page.getByTestId('event-start-date').fill('2026-09-10');
    await page.getByTestId('event-end-date').fill('2026-09-10');
    await page.getByTestId('event-genres-input').fill('techno, house');
    await page.getByTestId('event-create-btn').click();

    await expect(page).toHaveURL(/\/admin\/events$/);
    await expect(page.getByText('Created By Test')).toBeVisible();
  });

  test('deletes an event after confirming the dialog', async ({ page }) => {
    await page.goto('/admin/events');
    await expect(page.getByTestId(`event-card-${SEED.event.id}`)).toBeVisible();

    await page.getByTestId(`event-delete-${SEED.event.id}`).click();
    await page.getByTestId('dialog-confirm-btn').click();

    await expect(page.getByTestId(`event-card-${SEED.event.id}`)).toHaveCount(0);
  });

  test('clones an event into a new one', async ({ page }) => {
    await page.goto('/admin/events');
    await page.getByTestId(`event-clone-${SEED.event.id}`).click();
    await page.getByTestId('dialog-confirm-btn').click();

    // Clone navigates to the new event's detail page — a different id than the source.
    await expect(page).toHaveURL(/\/admin\/events\/[0-9a-f-]{36}$/);
    expect(page.url()).not.toContain(SEED.event.id);
  });
});
