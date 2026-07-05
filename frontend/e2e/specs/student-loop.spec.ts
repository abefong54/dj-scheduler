// The full student payoff loop, end-to-end across the student + teacher contexts:
// teacher core challenge → student posts a mix → teacher passes → the mix appears
// in the student feed with its snapshotted badge. Exercises the real API.
import { test, expect } from '../fixtures';
import { request as pwRequest } from '@playwright/test';
import { ADMIN_STORAGE_STATE, BACKEND_URL } from '../constants';
import { adminToken, studentStorageState } from '../student-auth';

test('post → review → pass → feed shows the badge', async ({ browser }) => {
  // 1) Teacher creates a core (badge) challenge via the API (deterministic setup).
  const api = await pwRequest.newContext({
    extraHTTPHeaders: { Authorization: `Bearer ${adminToken()}` },
  });
  const created = await api.post(`${BACKEND_URL}/api/challenges`, {
    data: {
      type: 'core',
      title: 'E2E Blend',
      description: 'blend two tracks',
      badge_label: 'E2E Blend',
      badge_icon: '🎧',
      sort_order: 10,
    },
  });
  expect(created.ok()).toBeTruthy();

  // 2) The student sees it and posts a mix link.
  const studentCtx = await browser.newContext({ storageState: studentStorageState() });
  const sp = await studentCtx.newPage();
  await sp.goto('/student/challenges');
  // The seeded student has no assigned custom work, so the school's core set is
  // their challenge list; each is postable inline.
  await expect(sp.getByTestId('my-challenges').getByText('E2E Blend')).toBeVisible();
  await sp.getByTestId('open-post').first().click();
  await sp.getByTestId('post-url').fill('https://youtu.be/dQw4w9WgXcQ');
  await sp.getByTestId('post-submit').click();
  await expect(sp.getByTestId('status').first()).toContainText('posted');

  // 3) The teacher finds it in the review queue and passes it.
  const adminCtx = await browser.newContext({ storageState: ADMIN_STORAGE_STATE });
  const ap = await adminCtx.newPage();
  await ap.goto('/admin/review');
  await expect(ap.getByTestId('review-item')).toHaveCount(1);
  await ap.getByTestId('pass').click();
  await expect(ap.getByTestId('review-empty')).toBeVisible();

  // 4) The passed mix shows in the student's feed with the embed + earned badge.
  await sp.goto('/student/feed');
  await expect(sp.getByTestId('feed-item')).toHaveCount(1);
  await expect(sp.getByTestId('feed-item').getByTestId('media-embed-frame')).toBeVisible();
  await expect(sp.getByTestId('feed-badge')).toContainText('E2E Blend');

  await studentCtx.close();
  await adminCtx.close();
  await api.dispose();
});
