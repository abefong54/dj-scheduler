// Custom test fixture. The whole suite shares one seeded database (workers=1),
// so every test resets it to the deterministic baseline in
// backend/seed/seed_test.sql before running. That keeps mutating tests
// (create/delete event, confirm slot, …) from leaking into one another.
import { test as base, expect } from '@playwright/test';
import { execSync } from 'node:child_process';
import { resolve } from 'node:path';

const REPO_ROOT = resolve(__dirname, '../..');
const RESET_CMD =
  'docker compose -f docker-compose.test.yml exec -T db ' +
  'psql -q -U eventlineup -d eventlineup < backend/seed/seed_test.sql';

/** Reset the test database to the seeded baseline. */
export function resetDb(): void {
  execSync(RESET_CMD, { cwd: REPO_ROOT, stdio: 'pipe' });
}

export const test = base.extend<{ seededDb: void }>({
  seededDb: [
    async ({}, use) => {
      resetDb();
      await use();
    },
    { auto: true },
  ],
});

export { expect };
