// Token helpers for the student-features E2E specs. Mirrors global-setup.ts: HS256
// JWTs signed in JS over the same JWT_SECRET the test API verifies with, so no Go
// dependency. Organizer tokens match token.Sign; student tokens match
// token.SignStudent (role=student, sub=dj id, school_id).
import { createHmac } from 'node:crypto';
import { BASE_URL, JWT_SECRET, SEED } from './constants';

function b64url(input: string): string {
  return Buffer.from(input).toString('base64url');
}

function signJwt(payload: Record<string, unknown>): string {
  const header = b64url(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
  const body = b64url(JSON.stringify(payload));
  const sig = createHmac('sha256', JWT_SECRET).update(`${header}.${body}`).digest('base64url');
  return `${header}.${body}.${sig}`;
}

const nowSec = (): number => Math.floor(Date.now() / 1000);

/** Organizer JWT for authenticated API setup calls (same claims as global-setup). */
export function adminToken(): string {
  const t = nowSec();
  return signJwt({
    organizer_id: SEED.organizerId,
    email: 'e2e-organizer@eventlineup.local',
    iat: t,
    exp: t + 3600,
  });
}

/** role=student JWT for the seeded student DJ (sub = dj id, school_id scoped). */
export function studentToken(
  djId: string = SEED.djTesta.id,
  schoolId: string = SEED.schoolId,
): string {
  const t = nowSec();
  return signJwt({ sub: djId, role: 'student', school_id: schoolId, iat: t, exp: t + 3600 });
}

/** Playwright storage state that starts a page authenticated as the seeded student. */
export function studentStorageState() {
  return {
    cookies: [],
    origins: [{ origin: BASE_URL, localStorage: [{ name: 'student_jwt', value: studentToken() }] }],
  };
}
