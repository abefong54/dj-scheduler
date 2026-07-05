// On a phone the "add a new set" dialog must not have a dead Save button. The
// seeded event has multiple stages, so there is no default stage to pre-fill —
// pressing Save without picking one has to surface a stage-required error, not
// silently do nothing. The dialog's action buttons also use the shared
// <app-button> so they render uniformly (base .btn + variant).
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE, SEED } from '../constants';

test.use({ storageState: ADMIN_STORAGE_STATE, viewport: { width: 390, height: 844 } });

const DETAIL = `/admin/events/${SEED.event.id}`;
const slotsApi = `/api/events/${SEED.event.id}/slots`;

test.describe('admin · add a set on mobile', () => {
  test('Save with no stage surfaces an error instead of doing nothing', async ({ page }) => {
    await page.goto(DETAIL);

    // Below md the card list opens the add dialog (not the inline table row).
    await page.getByTestId('add-slot-trigger').click();
    await expect(page.getByTestId('add-slot-dialog')).toBeVisible();

    // Pressing Save with the stage unpicked must not POST, and must show why.
    let posted = false;
    page.on('request', (r) => {
      if (r.url().includes(slotsApi) && r.method() === 'POST') posted = true;
    });
    await page.getByTestId('dialog-add-save').click();

    await expect(page.getByTestId('dialog-stage-error')).toBeVisible();
    await expect(page.getByTestId('add-slot-dialog')).toBeVisible(); // stays open to fix
    expect(posted).toBe(false);
  });

  test('picking a stage clears the error and creates the set', async ({ page }) => {
    await page.goto(DETAIL);
    await page.getByTestId('add-slot-trigger').click();

    // Trigger the error first, then resolve it.
    await page.getByTestId('dialog-add-save').click();
    await expect(page.getByTestId('dialog-stage-error')).toBeVisible();

    await page.getByTestId('dialog-add-stage').selectOption(SEED.stages.side);
    await page.getByTestId('dialog-add-start').fill('23:15');

    const post = page.waitForResponse(
      (r) => r.url().includes(slotsApi) && r.request().method() === 'POST',
    );
    await page.getByTestId('dialog-add-save').click();
    await post; // the set was created — Save works

    // The add dialog stays open and resets for rapid entry (parity with the
    // desktop inline row): the stage-required error clears and the new set shows
    // in the card list behind it.
    await expect(page.getByTestId('dialog-stage-error')).toHaveCount(0);
    await expect(page.getByText('23:15')).toBeVisible();
  });

  test('the dialog action buttons render as uniform shared .btn buttons', async ({ page }) => {
    await page.goto(DETAIL);
    await page.getByTestId('add-slot-trigger').click();

    const save = page.getByTestId('dialog-add-save');
    await expect(save).toHaveClass(/\bbtn\b/);
    await expect(save).toHaveClass(/btn-primary/);
  });
});
