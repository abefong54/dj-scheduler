// Shared constants for the E2E suite. The IDs/token below must stay in sync with
// backend/seed/seed_test.sql — they are the deterministic fixtures the tests
// assert against.
import { resolve } from 'node:path';

export const BASE_URL = 'http://localhost:4200';
export const BACKEND_URL = 'http://localhost:8080';

// Must match JWT_SECRET in docker-compose.test.yml so a token minted here is
// accepted by the test API.
export const JWT_SECRET = 'e2e-test-jwt-secret-do-not-use-in-prod-0123456789';

// Where global-setup writes the authenticated admin storage state.
export const ADMIN_STORAGE_STATE = resolve(__dirname, '.auth/admin.json');

// Seeded fixtures (see backend/seed/seed_test.sql).
export const SEED = {
  organizerId: '00000000-0000-0000-0000-000000000001',
  event: {
    id: '00000000-0000-0000-0000-0000000000e1',
    name: 'E2E Fest',
    venue: 'Test Arena',
  },
  djTesta: {
    id: '00000000-0000-0000-0000-0000000000d1',
    name: 'DJ Testa',
    // Raw portal token whose SHA-256 hash is seeded for this DJ.
    portalToken: 'e2e-portal-token-dj1',
  },
  djBeta: {
    id: '00000000-0000-0000-0000-0000000000d2',
    name: 'DJ Beta',
  },
  stages: {
    main: '00000000-0000-0000-0000-0000000005a1',
    side: '00000000-0000-0000-0000-0000000005a2',
  },
  slots: {
    dj1: '00000000-0000-0000-0000-000000005101',
    dj2: '00000000-0000-0000-0000-000000005102',
  },
} as const;
