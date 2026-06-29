// Global setup: ensure the backend is up, then mint an organizer JWT and write
// it as Playwright storage state so admin specs start authenticated.
//
// The token is signed here in JS (node:crypto) rather than shelling out to the
// Go mint CLI, so the test harness has no Go dependency. It is an HS256 JWT over
// the same JWT_SECRET the test API verifies with — identical wire format to what
// backend/internal/token.Sign produces.
import { createHmac } from 'node:crypto';
import { mkdirSync, writeFileSync } from 'node:fs';
import { dirname } from 'node:path';
import { ADMIN_STORAGE_STATE, BACKEND_URL, BASE_URL, JWT_SECRET, SEED } from './constants';

function b64url(input: string): string {
  return Buffer.from(input).toString('base64url');
}

function signJwt(payload: Record<string, unknown>, secret: string): string {
  const header = b64url(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
  const body = b64url(JSON.stringify(payload));
  const sig = createHmac('sha256', secret).update(`${header}.${body}`).digest('base64url');
  return `${header}.${body}.${sig}`;
}

async function waitForBackend(timeoutMs = 30_000): Promise<void> {
  const deadline = Date.now() + timeoutMs;
  let lastErr = '';
  while (Date.now() < deadline) {
    try {
      const res = await fetch(`${BACKEND_URL}/healthz`);
      if (res.ok) return;
      lastErr = `HTTP ${res.status}`;
    } catch (e) {
      lastErr = (e as Error).message;
    }
    await new Promise((r) => setTimeout(r, 1000));
  }
  throw new Error(
    `Backend not reachable at ${BACKEND_URL}/healthz (${lastErr}).\n` +
      'Start the E2E stack first:  npm run e2e:up   (stop any dev backend on :8080).',
  );
}

export default async function globalSetup(): Promise<void> {
  await waitForBackend();

  const now = Math.floor(Date.now() / 1000);
  const token = signJwt(
    {
      organizer_id: SEED.organizerId,
      email: 'e2e-organizer@eventlineup.local',
      iat: now,
      exp: now + 24 * 60 * 60,
    },
    JWT_SECRET,
  );

  const state = {
    cookies: [],
    origins: [{ origin: BASE_URL, localStorage: [{ name: 'jwt', value: token }] }],
  };
  mkdirSync(dirname(ADMIN_STORAGE_STATE), { recursive: true });
  writeFileSync(ADMIN_STORAGE_STATE, JSON.stringify(state, null, 2));
}
