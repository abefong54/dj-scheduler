# E2E tests (Playwright, full-stack)

These specs run against a **real** backend + Postgres (not mocks) plus the Angular
dev server. They cover the public schedule, the token-gated DJ portal, and the
authenticated admin (events, slots, DJs), plus i18n.

## Run locally

The test backend uses port **8080**, so stop any dev backend first (the dev
`docker-compose.yml` stack also uses 8080).

```bash
# from frontend/
npm run e2e:up      # build + start the throwaway backend/Postgres, then seed
npm run e2e         # run the Playwright suite (starts `ng serve` itself)
npm run e2e:down    # stop the stack and delete its volume
```

Other scripts: `e2e:headed` (watch it run), `e2e:report` (open the HTML report),
`e2e:reset` (re-seed to baseline), `e2e:token` (print an organizer JWT).

## How it works

- **docker-compose.test.yml** (repo root) — throwaway Postgres (host port 5433) +
  Go API (8080) with a fixed test `JWT_SECRET` and dummy Google creds.
- **global-setup.ts** — waits for the API, mints an organizer JWT (HS256, same
  secret the API verifies with), and writes it as storage state so admin specs
  start logged in. No real Google OAuth.
- **fixtures.ts** — resets the DB to `backend/seed/seed_test.sql` before every
  test, so the shared database stays deterministic (the suite runs with 1 worker).
- **constants.ts** — the seeded fixture IDs/token the specs assert against; keep
  in sync with `seed_test.sql`.

## Auth contexts

| Context | How the test authenticates |
|---|---|
| Admin (`/admin/**`) | minted JWT injected via storage state (`test.use({ storageState })`) |
| DJ portal (`/dj/portal`) | the seeded raw token in the URL query |
| Public (`/events/:id`) | none |

## Adding tests

Prefer `getByRole` / `getByText`; use `getByTestId(...)` for the interactive
elements that already carry `data-testid` (slot rows, dialog buttons, form
inputs, DJ-portal confirm/flag). Assert on seeded **data** (names) rather than
translated UI strings where possible, so specs stay i18n-agnostic.

## CI

`.github/workflows/e2e.yml` runs this suite on PRs to `main`. Make it a required
status check so deploys (Netlify on push to `main`) are gated on it passing.
