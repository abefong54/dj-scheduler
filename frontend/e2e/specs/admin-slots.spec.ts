// Event detail — the slot editor: add, edit (EL-035 PATCH), delete, conflict
// detection (DJ double-booking + stage overlap), and adding a stage.
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE, SEED } from '../constants';

test.use({ storageState: ADMIN_STORAGE_STATE });

const DETAIL = `/admin/events/${SEED.event.id}`;
const slotsApi = `/api/events/${SEED.event.id}/slots`;

test.describe('admin · slots', () => {
  test('adds a new slot', async ({ page }) => {
    await page.goto(DETAIL);
    await page.getByTestId('add-slot-trigger').click();
    await page.getByTestId('add-stage-select').selectOption(SEED.stages.side);
    await page.getByTestId('add-dj-select').selectOption(SEED.djBeta.id);
    await page.getByTestId('add-start').fill('22:30');

    const post = page.waitForResponse(
      (r) => r.url().includes(slotsApi) && r.request().method() === 'POST',
    );
    await page.getByTestId('add-slot-save').click();
    await post;

    await expect(page.getByText('22:30')).toBeVisible();
  });

  test('edits a slot via PATCH and reflects the new start time', async ({ page }) => {
    await page.goto(DETAIL);
    await page.getByTestId(`slot-edit-${SEED.slots.dj2}`).click();
    await page.getByTestId(`edit-start-${SEED.slots.dj2}`).fill('21:30');

    const patch = page.waitForResponse(
      (r) => r.url().includes(`${slotsApi}/${SEED.slots.dj2}`) && r.request().method() === 'PATCH',
    );
    await page.getByTestId(`edit-save-${SEED.slots.dj2}`).click();
    await patch;

    await expect(page.getByTestId(`slot-start-${SEED.slots.dj2}`)).toHaveText('21:30');
  });

  test('deletes a slot after confirming the dialog', async ({ page }) => {
    await page.goto(DETAIL);
    await expect(page.getByTestId(`slot-row-${SEED.slots.dj2}`)).toBeVisible();

    await page.getByTestId(`slot-delete-${SEED.slots.dj2}`).click();
    await page.getByTestId('dialog-confirm-btn').click();

    await expect(page.getByTestId(`slot-row-${SEED.slots.dj2}`)).toHaveCount(0);
  });

  test('blocks a double-booked DJ with an inline conflict error', async ({ page }) => {
    await page.goto(DETAIL);
    // DJ Testa is already booked 20:00–21:00 (Main). Re-book her at 20:00 on Side.
    await page.getByTestId('add-slot-trigger').click();
    await page.getByTestId('add-stage-select').selectOption(SEED.stages.side);
    await page.getByTestId('add-dj-select').selectOption(SEED.djTesta.id);
    await page.getByTestId('add-start').fill('20:00');
    await page.getByTestId('add-slot-save').click();

    await expect(page.getByTestId('add-conflict-error')).toContainText(SEED.djTesta.name);
  });

  test('blocks an overlapping stage booking with an inline conflict error', async ({ page }) => {
    await page.goto(DETAIL);
    // Main Stage already has a set 20:00–21:00. Book another set there at 20:00.
    await page.getByTestId('add-slot-trigger').click();
    await page.getByTestId('add-stage-select').selectOption(SEED.stages.main);
    await page.getByTestId('add-dj-select').selectOption(SEED.djBeta.id);
    await page.getByTestId('add-start').fill('20:00');
    await page.getByTestId('add-slot-save').click();

    await expect(page.getByTestId('add-conflict-error')).toContainText('Main Stage');
  });

  // EL-042: an uncertified student DJ is flagged for the target genre with an
  // inline warning (the teacher can still override by saving).
  test('warns when assigning an uncertified DJ to a genre', async ({ page }) => {
    await page.goto(DETAIL);
    await page.getByTestId('add-slot-trigger').click();
    // DJ Testa plays house but holds no certifications (seed default), so House
    // is a genre she isn't cleared for.
    await page.getByTestId('add-genre-select').selectOption('house');
    await page.getByTestId('add-dj-select').selectOption(SEED.djTesta.id);

    await expect(page.getByTestId('add-cert-warning')).toContainText(SEED.djTesta.name);
  });

  test('adds a stage', async ({ page }) => {
    await page.goto(DETAIL);
    await page.getByTestId('add-stage-btn').click();
    await expect(page.getByTestId('add-stage-modal')).toBeVisible();
    await page.getByTestId('stage-name-input').fill('Test Stage');
    await page.getByTestId('stage-create-btn').click();

    await expect(page.getByText('Test Stage')).toBeVisible();
  });
});
