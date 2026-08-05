# Estimulos & Incentivos — Human Capital Analysis

A single-operator web application for human capital analysis and behavioral intervention
management, built on the Fogg behavior model (**B = M × A × P**). HR keeps a descriptive
analysis of the workforce and manages personalized **incentives** (motivation),
**nudges** (ability), and **stimuli** (prompts) with per-employee calibrated thresholds.

This is a portfolio-ready MVP: hardened for data integrity, operator security, reproducible
delivery, and safe local/container operation — deliberately without enterprise features
(see [Limitations & non-goals](#limitations--non-goals)).

> **What this is not:** this is not an enterprise HR platform. It has no SSO, no multi-tenant
> isolation, no multi-role RBAC, and no external integrations. It is designed for exactly one
> trusted operator running one instance.

## Quick path

1. `go run ./cmd/server` (or `docker compose up --build` for the container path).
2. Open `http://localhost:8080` — you land on the operator login.
3. Sign in with the documented development credentials (below) and explore the dashboard,
   employee list, incentives, and stimuli pages.
4. Confirm the server is healthy: `curl http://localhost:8080/healthz` → `ok`.

On first start with an empty database the app creates the schema and loads demo data
(6 employees, 8 incentives, 6 nudges). See [Demo data & seed](#demo-data--seed).

## Table of contents

- [Quick path](#quick-path)
- [Stack](#stack)
- [Architecture](#architecture)
- [Security boundary & threat model](#security-boundary--threat-model)
- [Persistence & backup](#persistence--backup)
- [Demo data & seed](#demo-data--seed)
- [Configuration](#configuration)
- [HTTP surface](#http-surface)
- [Deployment](#deployment)
- [Verification](#verification)
- [Limitations & non-goals](#limitations--non-goals)
- [License](#license)

## Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.26 (CGO-free, static binary) |
| HTTP | `net/http` with the Go 1.22+ enhanced mux |
| Database | SQLite via `modernc.org/sqlite` (pure Go, no CGO) |
| Templates | `html/template` (stdlib), parsed once at startup |
| Frontend | HTMX 2.0 + vanilla CSS |
| Logging | `log/slog` (stdlib), structured text events |
| Testing | `testing` (stdlib) |

## Architecture

The codebase follows a strict layered architecture, in dependency order:

```
cmd/server          → entry point, config, lifecycle, health, shutdown
internal/
  handler/          → HTTP layer: routes, HTML/JSON responses, auth + CSRF middleware
  service/          → application orchestration: validation, workflows, seeding
  domain/           → pure entities and invariants (Empleado, PerfilMAP, Umbral, Estimulo…)
  engine/           → business logic: threshold calibration, recommendation, risk, analysis
  store/sqlite/     → persistence: repositories, migrations, transactions
web/templates/      → Go html/template pages and HTMX partials
```

Request flow for any browser or API call:

```
HTTP → auth + CSRF middleware → handler DTO → service validation
    → Store.WithTx → tx-aware repositories → SQLite constraints
    → JSON / HTML response
```

- **domain** owns `Validate()` on every entity — the single source of invariants used by
  services, seed data, and tests (SQL constraints remain the final integrity boundary).
- **service** orchestrates domain + engine + store; all multi-row workflows run inside one
  SQLite transaction.
- **handler** maps routes to services, renders buffered HTML (generic 500 on render errors —
  no internal details leak), and applies the security middleware chain.
- **store/sqlite** owns connections, embedded migrations, and `Store.WithTx`.

See [`doc/design-production-hardening.md`](doc/design-production-hardening.md) for the full
architectural decisions, and the original concept document
[`doc/2026-04-29-sistema-estimulos-incentivos-diseno.md`](doc/2026-04-29-sistema-estimulos-incentivos-diseno.md)
for the behavioral model rationale.

## Security boundary & threat model

### Scope of trust

| Boundary | Assumption |
|----------|-----------|
| Operators | Exactly **one** trusted operator configured via environment variables |
| Network | The instance runs on a host/network the operator trusts (single-instance, single-operator scope) |
| Mutations | All state-changing requests require a valid session **and** a session-bound CSRF token |
| Persistence | The SQLite file is treated as private data (backup and access control are operator responsibilities) |

### Authentication

- Credentials and session secret come from `OPERATOR_USER`, `OPERATOR_PASSWORD`,
  `SESSION_SECRET` (deploy-time configuration — never hardcoded secrets in the image).
- Sessions are signed HMAC tokens in a cookie named `session` with
  **`Secure`**, **`HttpOnly`**, and **`SameSite=Lax`**; default TTL is 24 hours.
- Credential comparison and signature verification use constant-time operations
  (no user enumeration, no signature side channels).
- Missing sessions redirect browsers to `/login` and return a generic JSON `401` to APIs.
- `GET /login`, `POST /login`, and `POST /logout` are the only paths outside the
  auth/CSRF chain (`/logout` clears the cookie even for expired sessions).

### CSRF

- Every mutating method (anything except `GET`/`HEAD`/`OPTIONS`) requires a CSRF token
  derived from the current session: `X-CSRF-Token` header (injected by the base layout for
  HTMX requests) or a `_csrf` hidden form field.
- Rejected tokens return a generic JSON `403` with **no state change**.
- Chain order is fixed: `RequireAuth(CSRFProtect(...))` — CSRF validation needs the
  authenticated session to compare against.

### Threat model & accepted risks

| Risk | Status |
|------|--------|
| Stolen session cookie | **Accepted.** Sessions are stateless: logout clears the cookie client-side only, so a stolen token stays valid until its TTL (24 h). Fine for single-operator local scope; rotate `SESSION_SECRET` if compromised. |
| Plain-HTTP traffic | `COOKIE_SECURE` defaults to `true`, so the cookie is only sent over HTTPS. For plain-HTTP local demo, `compose.yaml` sets `COOKIE_SECURE=false` explicitly. Keep it `true` behind TLS. |
| Brute-force login | Out of scope: no rate limiting or account lockout. The operator boundary limits exposure. |
| Multi-user abuse | Non-goal: there is exactly one operator identity; there is no RBAC beyond that. |
| Nudge template defects | Known pre-existing rendering defects on `/nudges` pages (see [Known issues](#known-issues)). |
| SQLite concurrency | Single-process, single-instance design (see [Persistence](#persistence--backup)). |

## Persistence & backup

- **Engine:** SQLite via `modernc.org/sqlite` (pure Go). The database file lives at
  `DB_PATH` (default `estimulos.db`); the container path is `/data/estimulos.db` on a
  named volume.
- **Schema:** created by embedded, idempotent migrations (`000001_initial_schema`,
  `000002_hardening`) tracked in `schema_migrations`. Foreign keys are enabled per
  connection, with `CHECK`, `UNIQUE`, and `ON DELETE CASCADE` constraints plus access indexes.
- **Concurrency model:** the pool is intentionally **one connection** (`SetMaxOpenConns(1)`);
  this is a single-instance, single-operator app. Multi-process access to the same file is
  unsupported.
- **Backup expectations:**
  - The operator is responsible for backups. The simplest safe backup is a copy of the SQLite
    file/volume: stop the container (or `docker compose stop app`) and copy
    `/data/estimulos.db`, or use `sqlite3 estimulos.db ".backup backup.db"` while running.
  - Back up before any migration deployment (the hardening migration rebuilds tables).
  - Restore = stop the instance, replace the file, start the instance; migrations run only
    for unregistered versions.

## Demo data & seed

On startup, `Service.Seed` inserts demo data **only when the database is empty**
(`SELECT count(*) FROM empleados` = 0):

- **6 employees** with MAP profiles and initial thresholds: María García (Senior Developer),
  Juan Pérez (Junior Developer), Ana López (Tech Lead), Carlos Ruiz (Sales Manager),
  Laura Díaz (HR Specialist), Pedro Torres (UX Designer).
- **8 incentives** across identity, benefits, training, and corporate-project categories.
- **6 nudges** across defaults, social-proof, framing, and friction types.

The seed is deterministic, runs inside the startup sequence (30 s bound), and fails startup
on any error. To reset the demo: stop the server, delete the `estimulos.db` file (or the
Docker volume), and start again.

## Configuration

All configuration is via environment variables. Invalid values fail startup **before**
any traffic is served (fail-fast).

| Variable | Default | Purpose |
|----------|---------|---------|
| `PORT` | `8080` | HTTP port (validated 1–65535) |
| `DB_PATH` | `estimulos.db` | SQLite file path |
| `APP_ROOT` | (auto-detected) | Repo root containing `web/`; required in containers where the source path doesn't exist |
| `SHUTDOWN_TIMEOUT` | `15s` | Graceful-shutdown drain bound |
| `OPERATOR_USER` | `admin` *(dev)* | Operator username |
| `OPERATOR_PASSWORD` | `dev-admin-password-change-me` *(dev)* | Operator password |
| `SESSION_SECRET` | dev-only value *(dev)* | HMAC session signing secret (≥ 16 chars) |
| `COOKIE_SECURE` | `true` | Cookie `Secure` flag; set `false` only for plain-HTTP local demo |

> Development-only credentials are **not secrets**. The server logs a startup warning when
> they are in use; set all three (`OPERATOR_USER`, `OPERATOR_PASSWORD`, `SESSION_SECRET`)
> before any production-ish deployment.

## HTTP surface

### Health (public, outside the auth chain)

| Method | Path | Meaning |
|--------|------|---------|
| `GET` | `/healthz` | Liveness — process alive → `ok` |
| `GET` | `/readyz` | Readiness — templates loaded and DB ping succeeds → `ready`, otherwise `503 not ready` |

These power the Docker `HEALTHCHECK` and compose healthcheck; they bypass auth on purpose.

### HTML pages (authenticated)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/login` | Operator login form (public) |
| `POST` | `/login` | Authenticate and set session cookie (public) |
| `POST` | `/logout` | Clear session cookie (public) |
| `GET` | `/` | Dashboard with analysis and metrics |
| `GET` | `/empleados` | Employee list |
| `GET` | `/empleados/{id}` | Employee detail with MAP profile |
| `GET` | `/incentivos` | Incentive catalog |
| `GET` | `/incentivos/{id}` | Incentive detail |
| `GET` | `/nudges` | Nudge panel (⚠️ known render defect — see below) |
| `GET` | `/nudges/{id}` | Nudge detail (⚠️ known render defect) |
| `GET` | `/estimulos` | Stimuli (filter via `?estado=pendiente`) |

### API (authenticated; mutations additionally require CSRF)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/empleados` | List employees |
| `POST` | `/api/empleados` | Create employee |
| `GET` | `/api/empleados/{id}` | Get employee |
| `PUT` | `/api/empleados/{id}` | Update employee |
| `DELETE` | `/api/empleados/{id}` | Delete employee (cascades to profile, thresholds, stimuli, history) |
| `GET` | `/api/empleados/{id}/detail` | Composite employee detail (JSON) |
| `PUT` | `/api/empleados/{id}/perfil` | Update MAP profile |
| `POST` | `/api/empleados/import` | Import employees from CSV |
| `GET` | `/api/empleados/template` | Download CSV template |
| `GET` | `/api/analisis` | Descriptive analysis |
| `GET` | `/api/riesgos` | Risk zone |
| `POST` | `/api/recomendar/{id}` | Recommend intervention |
| `GET` | `/api/incentivos` | List incentives |
| `POST` | `/api/incentivos` | Create incentive |
| `GET` | `/api/incentivos/{id}` | Get incentive |
| `PUT` | `/api/incentivos/{id}` | Update incentive |
| `DELETE` | `/api/incentivos/{id}` | Delete incentive |
| `GET` | `/api/incentivos/elegibles/{empleadoId}` | Eligible incentives |
| `POST` | `/api/incentivos/{id}/elegibilidades` | Add eligibility rule |
| `DELETE` | `/api/elegibilidades/{id}` | Remove eligibility rule |
| `GET` | `/api/nudges` | List nudges |
| `POST` | `/api/nudges` | Create nudge |
| `GET` | `/api/nudges/{id}` | Get nudge |
| `PUT` | `/api/nudges/{id}` | Update nudge |
| `PUT` | `/api/nudges/{id}/toggle` | Toggle nudge active state |
| `GET` | `/api/estimulos` | List stimuli (`?estado=`) |
| `GET` | `/api/estimulos/{id}` | Get stimulus |
| `POST` | `/api/estimulos/{id}/aplicar` | Apply stimulus (idempotent, atomic) |

HTMX partial endpoints (`/api/empleados/new-form`, `/api/empleados/import-form`,
`/api/empleados/{id}/perfil-form`, `/api/estimulos/{id}/apply-form`, `/api/incentivos/form`,
`/api/nudges/form`, `/api/dashboard/stats|riesgos|distribucion|efectividad`) power the
in-page forms and dashboard panels.

## Deployment

### Local

```bash
# One-shot (creates estimulos.db next to the repo by default)
go run ./cmd/server

# Or build a binary and run it from any directory
go build -o server ./cmd/server
APP_ROOT=/path/to/repo ./server          # cwd-independent template resolution
```

`APP_ROOT` only needs to be set when the source path is not available (e.g. running the
binary from a foreign working directory or a copy without the repo layout).

### Container (Docker Compose — recommended for demo)

```bash
docker compose up --build
# → http://localhost:8080  (health: docker compose ps shows healthy)
```

`compose.yaml` provides a single `app` service (`estimulos-incentivos:local`) on port 8080,
a named volume `estimulos-data:/data` for the SQLite file, a `wget /readyz` healthcheck, and
`restart: unless-stopped`. Environment defaults mirror the dev credentials above; override
them via `OPERATOR_USER` / `OPERATOR_PASSWORD` / `SESSION_SECRET` in your shell or an `.env`
file. The image is multi-stage and CGO-free (`golang:1.26.2-alpine` build → `alpine:3.21`
runtime with `ca-certificates` and `tzdata`), with `web/` copied in and `APP_ROOT=/app`.

Stop and reset:

```bash
docker compose down
docker compose down -v          # only to also delete the SQLite volume (demo reset)
```

## Verification

All checks run from the repository root and are exactly what CI executes:

```bash
go test ./... -count=1 -timeout 120s   # all packages pass (6 packages, unit + SQLite + httptest + template suites)
go build ./...                          # compiles cleanly
go vet ./...                            # static analysis clean
gofmt -l .                              # must print NOTHING (format gate)
```

Runtime smoke checks:

```bash
curl -s http://localhost:8080/healthz   # → ok
curl -s http://localhost:8080/readyz    # → ready (503 "not ready" until templates/DB are up)
```

## Limitations & non-goals

### Explicit non-goals (out of scope by design)

- **No SSO/OIDC** — single operator credentials only.
- **No multi-tenancy** — one instance = one organization.
- **No multi-role RBAC** — exactly one operator identity.
- **No external integrations** — no ERP, payroll, email, or Slack.
- **No background jobs** — everything runs in the request path or at startup.
- **No PostgreSQL** — SQLite is the only database.
- **No Kubernetes** — single container with Compose.
- **No analytics/ML** — descriptive analysis only; the Fogg model needs no ML.
- **No browser E2E tests** — test coverage is unit/integration/httptest/template level.

### Known issues (honest inventory)

- **Nudge list/detail pages render a generic 500** (`error interno`): the pre-existing
  `nudges/_card.html` template has no `{{define "nudge-card"}}` block that
  `nudges/list.html` invokes, and `nudges/detail.html` executes against a map root.
  This defect is **not fixed by this change**; `PUT /api/nudges/{id}/toggle` falls back to a
  JSON response and still works. Fixing it is a dedicated follow-up.
- **Stateless sessions**: logout clears the cookie client-side; a captured token remains
  valid server-side until its 24 h TTL.
- **`COOKIE_SECURE`**: defaults to `true`; set `false` only for plain-HTTP local demos
  (as `compose.yaml` does) — keep `true` behind TLS.
- **Single-instance SQLite**: one process, one connection; concurrent writers on the same
  file are unsupported.
- **Employee deletion is permanent**: `DELETE /api/empleados/{id}` removes the employee and
  cascades to their MAP profile, thresholds, stimuli, and stimulus history via
  `ON DELETE CASCADE`; there is no soft-delete or undo.

### Tested behavior of the employee list

The employee list (`GET /empleados`) renders deterministically: each row shows name, role,
department, email, a "View profile" action, a "Recommend" action, and **exactly one** delete
action (a `🗑` button issuing `hx-delete="/api/empleados/{id}"` with an `hx-confirm` prompt
and `hx-swap="delete"`). This matches the tested render outcome — previously the template
emitted two delete actions per row and an orphaned table cell; that defect was fixed and is
covered by render tests. Deletion still requires a valid session and CSRF token.

## License

Private — portfolio MVP for internal demonstration.
