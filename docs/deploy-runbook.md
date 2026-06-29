# EventLineup — Production Deploy Runbook

Sprint P (EL-018 Neon · EL-017 Railway · EL-040 host headers). This is the
**human-gated** half of the sprint: account creation and secret entry are done by
you in the Neon / Railway / Netlify dashboards. The code, config, and this
runbook are deploy-ready in the repo. **Never commit real secrets** — every value
below is entered in a dashboard, not a file.

## Architecture

```
Browser ──HTTPS──▶ Netlify (Angular SPA, static)
                      │  XHR (apiUrl)
                      ▼
                   Railway (Go API, Docker)
                      │  DATABASE_URL (pooled)
                      ▼
                   Neon (Postgres)
```

- Frontend and API are **cross-origin** (different domains). CORS on the API is
  pinned to `FRONTEND_URL`; the SPA's CSP `connect-src` is pinned to the API URL.

## Coordination / ordering

Run migrations and deploy **after** the feature agents' migrations (cards EL-049,
cert-scheduling Sprint S) are merged to `main`, so Neon gets the complete schema.
This sprint does **not** edit migration files — it only runs them.

Do the steps in this order: **EL-018 (Neon) → EL-017 (Railway) → EL-040 / Netlify env**.

---

## EL-018 — Neon Postgres (production)

1. Create a project at <https://neon.tech> (region close to your Railway region).
2. From the dashboard, copy **two** connection strings:
   - **Pooled** (host contains `-pooler`) — for the running app (`DATABASE_URL`).
   - **Direct** (no `-pooler`) — for running migrations (DDL is happier on a
     direct connection than through the pooler).
   Both must include `?sslmode=require`.
3. Run all migrations against the **direct** connection. The `cmd/migrate` tool is
   idempotent (tracks applied files in a `schema_migrations` table) and applies
   every `backend/migrations/*.sql` in order:
   ```bash
   cd backend
   DATABASE_URL='postgresql://USER:PASS@HOST/DB?sslmode=require' go run ./cmd/migrate
   ```
   Expect `apply 001_init.sql … apply 008_line_notify.sql` then
   `N migration(s) applied`. Re-running prints `skip …` for each and
   `nothing to migrate`.
4. Verify:
   ```bash
   psql 'postgresql://USER:PASS@HOST/DB?sslmode=require' -c 'SELECT 1;'
   psql 'postgresql://USER:PASS@HOST/DB?sslmode=require' -c 'SELECT name FROM schema_migrations ORDER BY name;'
   ```
   The second should list all 8 migration files.

> ⚠️ The EL-018 ticket text predates later migrations and mentions only
> `001`/`002`. The real set is `001`–`008`; `cmd/migrate` runs whatever is in the
> folder, so just run it — don't hand-pick files.

**EL-018 done when:** Neon project exists, all 8 migrations applied, `SELECT 1`
succeeds, and the pooled `DATABASE_URL` is ready to paste into Railway (next step).

---

## EL-017 — Railway (Go API, production)

1. Create a Railway project → **Deploy from GitHub repo** → this repo.
2. **Root directory:** set the service root to `backend/` (the `Dockerfile` and
   `railway.toml` live there). Railway builds with the Dockerfile (`railway.toml`
   → `builder = "DOCKERFILE"`).
3. Set environment variables (Service → Variables). **Do not** set `PORT` —
   Railway injects it and the app reads it.

   | Variable | Value | Notes |
   |---|---|---|
   | `DATABASE_URL` | Neon **pooled** string | from EL-018, `?sslmode=require` |
   | `FRONTEND_URL` | `https://<your-site>.netlify.app` | exact origin, no trailing slash — CORS pins to it |
   | `JWT_SECRET` | `openssl rand -base64 48` | ≥32 chars or the app refuses to start |
   | `LINE_NOTIFY_ENCRYPTION_KEY` | `openssl rand -hex 32` | 64 hex chars; **required** at startup |
   | `GOOGLE_CLIENT_ID` | from Google Cloud Console | |
   | `GOOGLE_CLIENT_SECRET` | from Google Cloud Console | |
   | `GOOGLE_REDIRECT_URL` | `https://<railway-url>/auth/google/callback` | must match Google console exactly |
   | `COOKIE_SECURE` | `true` | marks the OAuth state cookie Secure (HTTPS) |

   See `backend/template.env` for the annotated list. All eight non-`PORT` vars
   are required — the app fails fast on startup if any are missing/malformed.
4. The healthcheck is `GET /healthz` (`railway.toml`) — unauthenticated liveness.
   (It was previously `/api/events`, which always 401s; fixed in EL-017.)
5. Update **Google Cloud Console → Credentials → OAuth client**:
   - Authorized redirect URI: `https://<railway-url>/auth/google/callback`
   - Authorized JS origin: your Netlify URL.
6. Verify after deploy:
   ```bash
   curl -i https://<railway-url>/healthz          # 200 {"status":"ok"}
   curl -i https://<railway-url>/api/events        # 401 (auth required) — EXPECTED, proves the route + DB are up
   ```
   To confirm a DB-backed authed route (AC: `/api/djs` 200), sign in through the
   app and load the roster, or mint a token locally against the same JWT_SECRET
   (`backend/cmd/mintdevtoken`) and `curl -H "Authorization: Bearer <jwt>"`.

**EL-017 done when:** deploy succeeds from `main`, `/healthz` is 200, an authed
`/api/djs` returns 200 against Neon, CORS allows the Netlify origin, and the
Railway URL is ready for Netlify (next step).

---

## EL-040 — Netlify env + host-layer security headers

Code for EL-040 (strict CSP, security headers, the keep-localStorage decision) is
already in the repo: `frontend/public/_headers` (authoritative CSP),
`frontend/netlify.toml` (placeholder substitution), and a new Go
security-headers middleware on the API. The only dashboard step is the env var.

1. Netlify → Site settings → Environment variables → add:
   - `RAILWAY_API_URL` = `https://<railway-url>` (no trailing slash).
   The build's `sed` step rewrites the `YOUR_RAILWAY_URL` placeholder into both
   `environment.prod.ts` (`apiUrl`) and `public/_headers` (CSP `connect-src`).
2. Trigger a deploy (clear cache & deploy).
3. Verify:
   ```bash
   curl -I https://<your-site>.netlify.app/        # check headers below
   ```
   Confirm the response carries:
   - `Content-Security-Policy:` with `connect-src 'self' https://<railway-url>`
     (the placeholder is gone) and `script-src 'self'`.
   - `Referrer-Policy: no-referrer`, `X-Content-Type-Options: nosniff`,
     `X-Frame-Options: DENY`.
   Then in the app: log in via Google, load a page (fonts render), and run an
   Excel export — all must work with no CSP violations in the browser console.

> If the page is blank with a thrown error about `YOUR_RAILWAY_URL`, the
> `RAILWAY_API_URL` env var wasn't set at build time (EL-017 guard in `main.ts`).

**EL-040 done when:** the CSP and security headers are served on the live site,
the placeholder is substituted, and login / API / fonts / xlsx export all work
under the policy.

---

## Secrets checklist

- [ ] No secret value committed to the repo (all entered in dashboards).
- [ ] `JWT_SECRET`, `LINE_NOTIFY_ENCRYPTION_KEY`, Google client secret, and the
      Neon password exist **only** in Railway/Neon dashboards.
- [ ] `COOKIE_SECURE=true` in Railway prod.
- [ ] Google OAuth redirect URI updated to the Railway callback URL.
