// Teacher challenge authoring against the real API (Track H). Admin storage state.
import { test, expect } from '../fixtures';
import { ADMIN_STORAGE_STATE } from '../constants';

test.use({ storageState: ADMIN_STORAGE_STATE });

test.describe('teacher · challenge authoring', () => {
  test('authors a core badge challenge that then appears in the list', async ({ page }) => {
    await page.goto('/admin/challenges');

    await page.getByTestId('type-core').click();
    await page.getByTestId('challenge-title').fill('First Blend');
    await page.getByTestId('badge-label').fill('First Blend');
    await page.getByTestId('badge-icon').fill('🎧');
    await page.getByTestId('challenge-create').click();

    await expect(page.getByTestId('challenge-list').getByText('First Blend')).toBeVisible();
  });

  test('authors a custom challenge assigned to a seeded student', async ({ page }) => {
    await page.goto('/admin/challenges');

    await page.getByTestId('type-custom').click();
    await page.getByTestId('challenge-title').fill('EQ transitions');
    // The assign dropdown is populated from the roster; pick the first student.
    await page.getByTestId('assign-student').selectOption({ index: 1 });
    await page.getByTestId('challenge-create').click();

    await expect(page.getByTestId('challenge-list').getByText('EQ transitions')).toBeVisible();
  });
});
