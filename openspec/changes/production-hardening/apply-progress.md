# Apply Progress: production-hardening — Slices 1-5 (data-integrity + operator-security + integration-confidence + operable-delivery + portfolio-documentation)

**Branches**: `slice/1-data-integrity` (baseline `4e8b406`) · `slice/2-operator-security` (from slice/1 head `8b2ebbe`) · `slice/3-integration-confidence` (from slice/2 head `05288ff`) · `slice/4-operable-delivery` (from slice/3 head `8c2347d`) · `slice/5-portfolio-documentation` (from slice/4 head `731db1d`)
**Mode**: Strict TDD (go test ./..., `strict_tdd: true` en openspec/config.yaml) — slice 5 is documentation-only: TDD RED/GREEN evidence is explicit `N/A` with structural/manual documentation verification instead of invented executable tests.
**Delivery**: chained slices — PR 1 (data-integrity), PR 2 (operator-security), PR 3 (integration-confidence), PR 4 (operable-delivery), PR 5 (portfolio-documentation); per `delivery_strategy=ask-on-risk` resolved by the orchestrator to slice execution. Chain strategy: `stacked-to-main` (slice/5 branch from slice/4 head; no push/PR — local reviewable boundaries only).
**Status**: Phase 1 (tasks 1.1–1.7), Phase 2 (tasks 2.1–2.4), Phase 3 (tasks 3.1–3.4), Phase 4 (tasks 4.1–4.3) and Phase 5 (tasks 5.1–5.2) COMPLETE — all 20 tasks done. Tree clean, all checks green.

## Slice 5 (portfolio-documentation) — TDD Cycle Evidence

Slice 5 changes only documentation (README + design docs). No executable behavior was added or modified; per the strict-tdd.md evidence contract, RED/GREEN/TRIANGULATE/REFACTOR rows are explicit `N/A` (no executable behavior exists to test) and the work unit is evidenced by structural/manual documentation verification instead.

| Task | Artifact | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|----------|-------|------------|-----|-------|-------------|----------|
| 5.1 | `README.md` (rewritten) | Documentation | ✅ full suite green (baseline: `go test ./...` 6/6 packages ok on `731db1d`) | N/A — docs only, no executable behavior | N/A — verified by structural review: every command/path/env var/route/claim grounded in repo + specs + apply-progress (see Work Unit Evidence) | N/A — single documentation deliverable; triangulation not applicable | ✅ improved over previous README: corrected stale API table, added auth/CSRF/health/container sections |
| 5.2 | `doc/design-production-hardening.md` (new) + `doc/2026-04-29-sistema-estimulos-incentivos-diseno.md` (corrected) | Documentation | ✅ full suite green (same baseline) | N/A — docs only, no executable behavior | N/A — verified by structural review (see Work Unit Evidence) | N/A — single design deliverable | ✅ preserved useful original design doc while correcting now-false claims (golang-migrate/validator/testify, PostgreSQL migration, stale project structure, multi-tenant open question) so the doc set has no contradictory claims |

### Slice 5 — Work Unit Evidence

| Evidence | Value |
|---|---|
| Focused test command / result | N/A — no executable behavior in this slice (docs only). Baseline suites remain green: `go test ./... -count=1 -timeout 120s` → 6/6 packages `ok` (cmd/server 1.154s, domain 0.014s, engine 0.018s, handler 0.950s, service 0.451s, store/sqlite 0.774s). `go build ./...` OK · `go vet ./...` OK · `gofmt -l .` empty |
| Runtime harness command / scenario and exact result | N/A — no runtime boundary introduced. Referenced runtime behaviors (health endpoints, login, employee delete, apply-stimulus idempotency, nudge toggle JSON fallback, shutdown drain) are already covered by the slice 1–4 suites and the slice-4 binary/Docker smoke evidence; the docs describe exactly those tested outcomes |
| Manual / structural documentation verification | 1) Every env var table entry matches `cmd/server/config.go` + `internal/handler/auth.go` (PORT/DB_PATH/APP_ROOT/SHUTDOWN_TIMEOUT/OPERATOR_*/SESSION_SECRET/COOKIE_SECURE, defaults and validation). 2) Every route in the README tables cross-checked against `handler.RegisterRoutes` (all 49 `mux.HandleFunc` registrations incl. HTMX partials and `DELETE /api/empleados/{id}`) and `server.go` (`/healthz`, `/readyz` top-level mux). 3) Demo-data counts (6 empleados, 8 incentivos, 6 nudges) match `Service.Seed`; seed-on-empty semantics verified in `Seed()`. 4) Delete-cascade claim verified in `000002_hardening.up.sql` (`ON DELETE CASCADE`) + `Store.DeleteEmpleado`. 5) Nudge defect claims verified in `web/templates/nudges/` (`list.html` invokes `{{template "nudge-card" .}}`, `_card.html` has no such define; `detail.html` map root) — documented as NOT fixed. 6) Compose volume reset command corrected to `docker compose down -v` after `docker compose config` resolved sanitized project name `estimuloseincentivos`. 7) Internal doc links resolve: `doc/design-production-hardening.md` and `doc/2026-04-29-sistema-estimulos-incentivos-diseno.md` exist and are linked from the README |
| Rollback boundary | Revert `a130c0d` (or `git revert a130c0d`): removes README rewrite + new design doc + design-doc corrections without touching application source, tests, Docker, CI, or migrations. The follow-up SDD-marking commit (`docs(sdd): mark Phase 5 ...`) reverts to the pre-slice-5 tasks.md/apply-progress state |

### Slice 5 — Deviations & Notes

1. **Docs language**: README and the new design doc were written in English (technical documentation default per the language domain contract). The original Spanish design doc was preserved and corrected in place (Spanish) to avoid a rewrite destroying its conceptual value.
2. **`doc/*design*.md` target**: no existing file matched the `*design*` glob (the original is `...-diseno.md`, Spanish for "design"), so `doc/design-production-hardening.md` was created as the as-built design doc; the original was corrected rather than duplicated.
3. **No unsupported claims**: the docs state the employee delete behavior and tested render outcomes as tested, and explicitly document that the nudge list/detail render defect is NOT fixed (generic 500) while `PUT /api/nudges/{id}/toggle` works via JSON fallback.
4. **Compose volume name**: README initially referenced `estimulos-incentivos_estimulos-data`; `docker compose config` resolved the sanitized project name to `estimuloseincentivos`, so the reset command uses `docker compose down -v` (robust to project-name sanitization).
5. **Changed-line count**: this slice is 705 changed lines (572 insertions / 133 deletions). It stays inside the assigned slice budget (max 2500) and matches the planned PR 5 of the tasks.md forecast (~2,200–2,600 total across 5 PRs); it is reported here for reviewer awareness.

## Slice 4 (operable-delivery) — TDD Cycle Evidence (preserved)

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 4.1 (templates) | `internal/handler/templates_set_test.go` | Template set (parse+execute, real files) | ✅ full suite green (43+ handler tests) | ✅ Written first — compile fails: `TemplateSet`, `ParseTemplates`, `SetTemplates` undefined | ✅ 8/8 tests pass | ✅ every registered page loads; missing asset (`estimulos/_table.html`) fails load; unknown page errors; partial root render (`stats` >42<); cwd-independent via `t.Chdir`; deterministic double-render | ✅ `templateFuncs` moved to templates.go; pageDefs sorted for deterministic iteration; gofmt clean |
| 4.1 (config) | `cmd/server/config_test.go` | Unit | ✅ prior suite green | ✅ Written first — compile fails: `Config`/`ConfigFromEnv` undefined | ✅ 4/4 tests pass | ✅ defaults (DB_PATH/8080/15s), custom values, 6 invalid PORTs, 4 invalid SHUTDOWN_TIMEOUTs | ✅ envString helper extracted |
| 4.1 (root) | `cmd/server/root_test.go` | Unit (cwd-independence) | ✅ prior suite green | ✅ Written first — compile fails: `resolveAppRoot` undefined | ✅ 4/4 tests pass | ✅ repo-root found + go.mod/web/templates present; `t.Chdir` arbitrary dir → still found; APP_ROOT override; relative APP_ROOT normalized to absolute | ✅ — |
| 4.1 (health) | `cmd/server/health_test.go` | Integration (real SQLite + real templates, httptest) | ✅ prior suite green | ✅ Written first — compile fails: `newHealthMux` undefined | ✅ 4/4 tests pass | ✅ healthz 200 "ok"; readyz 200 "ready"; readyz 503 "not ready" with DB closed; readyz 503 with templates nil | ✅ — |
| 4.1 (lifecycle) | `cmd/server/lifecycle_test.go` | Integration (real server: real SQLite, real templates, real HTTP) | ✅ prior suite green | ✅ Written first — compile fails: `run`, `shutdownServer` undefined | ✅ 6/6 tests pass | ✅ full run: readyz poll 200 → cancel → returns nil ≤ bound → listener closed (dial refused) → log has starting/server listening/request/shutdown complete; startup fails on missing templates (ERROR event, never listens); startup fails on unopenable DB; in-flight request drained within bound (150ms handler, 3s bound, elapsed 100ms–3s, request 200); bound exceeded → forced close + error ≤ 2s | ✅ shutdownServer extracted (Background-based timeout, not the canceled signal ctx — avoids instant-expired deadline); statusRecorder/requestLogger separated |
| 4.2 | `cmd/server/delivery_test.go` | Structural (delivery artifacts) | ✅ prior suite green | ✅ Written first — files don't exist (`leer Dockerfile: no such file`) | ✅ 3/3 tests pass | ✅ Dockerfile pinned golang:1.26.2-alpine + alpine:3.21, CGO_ENABLED=0, COPY web/, APP_ROOT=/app, HEALTHCHECK, EXPOSE, DB_PATH; compose single instance (1 container_name, 1 image), 8080:8080, estimulos-data:/data volume, healthcheck, no deploy/replicas; .dockerignore excludes .git/openspec/*.db//server | ✅ — |
| 4.3 | `cmd/server/ci_test.go` | Structural (threat matrix: PR commands) | ✅ prior suite green | ✅ Written first — `.github/workflows/ci.yml` missing (RED: file read fails) | ✅ 3/3 tests pass | ✅ all four commands present (`go test ./...`, `go build ./...`, `go vet ./...`, `gofmt -l .`); pinned to repo root (`working-directory: ${{ github.workspace }}`); triggers push+pull_request, go-version-file; no `git commit`/`git push`/`git config`; no `github.event` composed input | ✅ — |

### Slice 4 — Test-Count Reconciliation (preserved)

Stable count: `31 new top-level test functions (23 cmd/server + 8 template-set) plus table-driven subtests`. The TDD table above retains the 32 Phase 4 acceptance cases as scenario evidence. The prior executor summary's `49 new tests` figure is not used as evidence because it is unsupported by the durable acceptance table and repository top-level function count.

### Slice 4 — Work Unit Evidence (preserved)

| Evidence | Value |
|---|---|
| Focused test command / result | `go test ./internal/handler/ -run 'TestParseTemplates\|TestTemplateSet\|TestHandlerSetTemplates'` — 8/8 ok · `go test ./cmd/server/` — 27 tests ok (config 4, root 4, health 4, lifecycle 6, delivery 3, ci 3, helpers) |
| Runtime harness command / result | Binary smoke: `go build -o /tmp/opencode/smoke-server ./cmd/server` + run from foreign cwd (`/tmp/opencode/smoke`) with `APP_ROOT=<repo>` → `/healthz` 200, `/readyz` `ready`, `/login` 200 (template set render), SIGTERM → `shutting down, draining active requests` → `shutdown complete` → process exited, port freed. Docker: `docker compose config -q` OK → `docker compose up -d --build` → health `healthy` → `/healthz` 200, `/readyz` `ready` → volume `estimulos-data` contains seeded `estimulos.db` (81920 bytes) → `docker compose down` + `rm -f` clean |
| Rollback boundary | Revert `fffc0f8`→`1828e9e` (5 commits) or `git revert` per commit: `feat(handler)` (templates set + render wiring), `feat(server)` (lifecycle), `feat(ops)` (Docker/compose), `ci:` (workflow), `style:` (gofmt). TemplateSet refactor reverts without touching store/service/domain logic; Docker/CI files are additive |

## TDD Cycle Evidence (cumulative, slices 1–3 preserved)

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 1.1+1.2 | `internal/domain/validate_test.go` | Unit | N/A (new package tests) | ✅ Written first — compile fails: `Validate undefined` | ✅ 32/32 cases pass | ✅ 7+ cases per entity (valid + each rejection + cross-field) | ✅ gofmt, shared table harness |
| 1.3+1.4 | `internal/store/sqlite/store_test.go` | Integration (real SQLite, t.TempDir) | ✅ engine+domain baseline pass | ✅ Written first — compile fails: WithTx signature + missing Tx methods | ✅ 7/7 tests pass | ✅ commit, rollback, conditional transition, missing stimulus | ✅ insert logic extracted behind `execer`/`queryer` |
| 1.5+1.6 | `internal/service/service_test.go` | Integration (real store + SQLite trigger failure injection) | ✅ prior suite green | ✅ Written first — compile fails: `ApplyEstimulo returns 1 value` | ✅ 7/7 tests pass | ✅ 8-way concurrent one-winner; double apply; invalid respuesta | ✅ service.go gofmt'ed |
| 1.7 | `internal/store/sqlite/migrations_test.go` | Integration (legacy + fresh real DBs) | ✅ prior suite green | ✅ Embedded RED: legacy insert with missing parent SUCCEEDS pre-migration, rejected post | ✅ 5/5 tests pass | ✅ idempotency, FK, CHECK, orphan backfill, down | ✅ runner extracted: `recordMigration`/`applyPending` |
| 2.1+2.2 (auth) | `internal/handler/auth_test.go` | Integration (httptest, real SQLite) | ✅ prior suite green (28 tests) | ✅ Written first — compile fails: `Authenticator undefined`, `New(svc, auth)` arity | ✅ 7/7 tests pass | ✅ 2 login-failure cases (user vs password), tampered token, logout lifecycle | ✅ session token nonce extracted |
| 2.1+2.3 (CSRF) | `internal/handler/csrf_test.go` | Integration (httptest, real SQLite) | ✅ auth tests green | ✅ Written first — compile fails: `Middleware undefined`; then runtime RED: wrong middleware order returned 403 always | ✅ 5/5 tests pass | ✅ absent/malformed/mismatched/valid-header/form-field; no-state-change via count | ✅ chain order fixed to `RequireAuth(CSRFProtect(next))` |
| 2.4 | `internal/handler/templates_security_test.go` | Structural (cwd-independent reads) | ✅ full suite green | ✅ Written first — assertions on not-yet-wired templates fail | ✅ 3/3 tests pass | ✅ meta tag + header injection + login + 6 mutation forms | ✅ page data maps share `csrfFor(r)` helper |
| 3.1 | `internal/store/sqlite/store_test.go` | Integration (real SQLite, t.TempDir) | ✅ full suite green (43 tests) | ✅ Written first — scenario not covered at store level | ✅ 2/2 tests pass | ✅ 8-way concurrent one-winner + multi-entity rollback across 3 tables | ✅ gofmt |
| 3.2 | `internal/handler/routes_test.go` | Integration (httptest, real SQLite) | ✅ full suite green | ✅ Written first — success/error routes + PUT-rejection persistence not covered after slice 2 | ✅ 4/4 tests pass | ✅ success list/detail, 400/404/405 safe statuses, malformed body, second mutation verb | ✅ gofmt |
| 3.3+3.4 | `internal/handler/templates_render_test.go` + `web/templates/empleados/list.html` | Template render (parse+execute, cwd-independent) | ✅ full suite green | ✅ TRUE RED: rendered list had 2 delete actions per row and 6 open/7 close cells | ✅ RED→GREEN: tests fail pre-fix, pass post-fix | ✅ 3 employees fixture + 4 key pages + login | ✅ gofmt |

## Completed Tasks (cumulative — all 20)

- [x] 1.1 RED domain validator tests
- [x] 1.2 GREEN `Validate()` on Empleado, PerfilMAP, Umbral, Estimulo
- [x] 1.3 RED sqlite tx tests (real DB)
- [x] 1.4 GREEN `Store.WithTx` + tx-aware CRUD + `TransitionEstimuloTx`
- [x] 1.5 RED service tests (rollback, repeat, concurrent one-winner)
- [x] 1.6 GREEN atomic `CreateEmpleado`, idempotent `ApplyEstimulo` (`ApplyResult{Applied,Conflict}`, Validate-first)
- [x] 1.7 Migrations `000002_hardening` up/down + runner + FK per connection
- [x] 2.1 RED handler tests — 401 unauthenticated, cookie flags, bad CSRF rejected without state change
- [x] 2.2 GREEN `internal/handler/auth.go` — env operator creds + session secret, login/logout, Secure/HttpOnly/SameSite=Lax cookie
- [x] 2.3 GREEN `internal/handler/middleware.go` — auth middleware, mutation authorization, session-bound CSRF, safe errors
- [x] 2.4 GREEN wire middleware + CSRF into `handler.go`, `cmd/server/main.go`, template forms
- [x] 3.1 RED store-level concurrent one-winner + multi-entity rollback (no partial state in 3 tables)
- [x] 3.2 RED httptest success/error routes + persistence unchanged on rejected PUT/malformed body
- [x] 3.3 RED deterministic template render tests (list.html + key pages; RED caught duplicate delete + orphan `</td>`)
- [x] 3.4 GREEN `list.html` fixed — exactly one delete action per row, 6/6 balanced cells
- [x] 4.1 GREEN `cmd/server/{main,config,root,health,server}.go` + `internal/handler/templates.go` — startup validation (fail before traffic), slog startup/request/error events, /healthz + /readyz, bounded graceful shutdown (drain + force-close), templates parsed once from explicit root (cwd-independent)
- [x] 4.2 GREEN `Dockerfile` + `compose.yaml` + `.dockerignore` — reproducible single-instance demo, SQLite volume, healthcheck (verified live via docker compose)
- [x] 4.3 GREEN `.github/workflows/ci.yml` — repo-root-pinned workflow running go test/build/vet + gofmt gate, fail on non-zero, no commit/push, no composed user input
- [x] 5.1 `README.md` — architecture layers, threat model + single-operator scope, persistence/backup, demo/seed, deployment, verification, non-goals
- [x] 5.2 `doc/*design*.md` — decisions, migration/rollback, limitations; walkthrough matches tested behavior

## Files Changed

### Slice 5 (portfolio-documentation)

| File | Action |
|------|--------|
| `README.md` | Rewritten (English) — quick path, stack, layered architecture + request flow, security boundary + threat model + accepted risks, persistence/backup expectations, demo/seed, env-var reference, full HTTP surface (health + HTML + API incl. HTMX partials), local + container deployment, exact verification commands, explicit non-goals, honest known issues, tested employee-list delete behavior |
| `doc/design-production-hardening.md` | Created (English) — ADR table (8 decisions), validation behavior per entity, transaction/rollback behavior, migration/rollback behavior, auth/CSRF boundaries, lifecycle/health/shutdown, cwd-independent templates, Docker/CI delivery, limitations, tested-behavior walkthrough, references |
| `doc/2026-04-29-sistema-estimulos-incentivos-diseno.md` | Modified (Spanish, preserved) — status note linking to the new design doc; corrected constraints (PostgreSQL portability → single-instance non-goals), persistence diagram, stack table (removed golang-migrate/go-playground/validator/testify), project structure to the as-built layout, multi-tenant open question → resolved non-goal |

### Slice 4 (operable-delivery) — preserved

| File | Action |
|------|--------|
| `internal/handler/templates.go` | Created — `TemplateSet` + `ParseTemplates(root)` parse-once (all page/partial sets, deterministic sorted iteration, any missing file = startup error); `pageDefs` single source of truth; partials named by basename so root `Execute` works (fixes pre-existing `incomplete or empty template` 500s) |
| `internal/handler/handler.go` | Modified — `templates *TemplateSet` field + `SetTemplates` + `execute`/`renderHTML` helpers; all 24 render sites switched from per-request `ParseFiles` to the pre-parsed set (buffer-then-write, generic 500 on error) |
| `internal/handler/templates_set_test.go` | Created — 8 tests (load all pages, missing asset fails, no templates dir fails, deterministic render, partial root render, unknown page errors, cwd-independent, SetTemplates wiring) |
| `cmd/server/config.go` | Created — `Config` + `ConfigFromEnv` (PORT 1–65535, SHUTDOWN_TIMEOUT positive duration, DB_PATH/APP_ROOT defaults) |
| `cmd/server/root.go` | Created — `resolveAppRoot`: APP_ROOT override (containers) else walk-up from source path to go.mod (cwd-independent) |
| `cmd/server/health.go` | Created — `newHealthMux`: public `/healthz` (liveness) + `/readyz` (readiness: templates loaded + DB ping, 503 otherwise) |
| `cmd/server/server.go` | Created — `run(ctx, cfg, logger)`: full wiring (store→migrate→ping→seed→auth→templates→mux→health→slog request logging); `shutdownServer` bounded graceful shutdown (drain, force-close on bound exceeded); `statusRecorder`/`requestLogger` structured slog events |
| `cmd/server/main.go` | Rewritten — slog TextHandler, ConfigFromEnv validation, `signal.NotifyContext` (SIGINT/SIGTERM), `run()` delegate, structured error exit |
| `cmd/server/{config,root,health,lifecycle,testutil}_test.go` | Created — 27 tests (see evidence) |
| `Dockerfile` | Created — multi-stage, pinned golang:1.26.2-alpine build (CGO_ENABLED=0, static binary) + alpine:3.21 runtime with ca-certificates/tzdata, `COPY web/`, APP_ROOT=/app, DB_PATH=/data, VOLUME /data, HEALTHCHECK /readyz |
| `compose.yaml` | Created — single `app` service, 8080:8080, SQLite volume `estimulos-data:/data`, env overrides with dev defaults, healthcheck, restart unless-stopped, no deploy/replicas |
| `.dockerignore` | Created — excludes .git/openspec/*.db//server from build context |
| `.github/workflows/ci.yml` | Created — `ci` on push+pull_request, checkout+setup-go (go-version-file), steps test/build/vet + gofmt gate (fail on non-empty `gofmt -l .`), working-directory pinned to `${{ github.workspace }}`, contents:read, no commit/push |
| `cmd/server/{delivery,ci}_test.go` | Created — 6 structural artifact tests |
| `internal/domain/{incentivo,nudge}.go`, `internal/engine/{analisis,calibrador}.go` | Modified — gofmt-only normalization of PRE-EXISTING drift (32/32 identical lines realigned; required so the new CI gofmt gate passes; see Deviations) |

### Slices 1–3 (preserved — see previous revisions)

Slice 1: domain validators (`internal/domain/*.go` + `validate_test.go`), `store.go`/`empleados.go`/`estimulos.go` (WithTx + tx CRUD), migrations `000002_hardening.{up,down}.sql`, `service.go` (atomic/ApplyResult), handler apply adaptation. Slice 2: `auth.go`, `middleware.go`, `handler.go` wiring, `login.html`, base/mutation template CSRF, `.gitignore` fix. Slice 3: store concurrent/rollback tests, `routes_test.go`, `templates_render_test.go`, `list.html` fix.

## Deviations from Design

1. **Slice 5 — docs language**: README + new design doc in English (language domain contract); original Spanish design doc preserved and corrected in place rather than duplicated.
2. **Slice 5 — design-doc target**: created `doc/design-production-hardening.md` because no existing file matched `doc/*design*.md`; corrected the original `...-diseno.md` to remove contradictory claims (golang-migrate/validator/testify, PostgreSQL migration intent, stale structure, multi-tenant open question).
3. **Slice 5 — compose reset command**: uses `docker compose down -v` because the sanitized compose project name is `estimuloseincentivos` (directory `Estimulos e incentivos`), not `estimulos-incentivos`.
4. **Slice 4 — gofmt normalization of pre-existing drift (4 files)**: `internal/engine/{analisis,calibrador}.go` and `internal/domain/{incentivo,nudge}.go` had PRE-EXISTING formatting drift (documented in slices 1–3 as out of scope). The new CI deliverable (task 4.3) runs a gofmt gate that fails on non-empty `gofmt -l .` output, and spec scenario "Clean checkout passes checks" requires the four commands to reproduce cleanly. gofmt -w is behavior-neutral (verified: 32 insertions / 32 deletions of identical lines, only alignment changed; full suite green before and after). Normalized as a REQUIRED enabler of this slice's own CI gate; recorded here so reviewers can distinguish it from new drift.
5. **Slice 4 — all HTMX partial endpoints fixed from 500 to working render**: the previous per-request pattern `template.New("").ParseFiles(file)` + `Execute` on the empty root produced the runtime error `template: "" is an incomplete or empty template` (verified empirically) — every partial endpoint (stats/riesgos/distribucion/efectividad/recomendacion/forms/rows/table) returned 500. The TemplateSet names partial roots by basename (top-level `ParseFiles` semantics), making root `Execute` valid. Behavioral fix, not a regression; covered by `TestTemplateSetExecutePartialRendersRoot`.
6. **Slice 4 — page render errors now generic 500 (safe error mapping)**: previously parse errors leaked template paths (`Error al cargar templates: <path>`); `renderHTML` buffers then writes a generic `error interno` 500. Parse errors now happen at startup (fail-fast), not per-request.
7. **Slice 4 — `ToggleNudgeAPI` now falls back to JSON on execute error** instead of an empty 200, because the pre-existing defect (`nudges/_card.html` lacks `{{define "nudge-card"}}`) now surfaces as an execute error from the pre-parsed set. The underlying defect remains out of scope (documented since slice 3; fix belongs to a dedicated correction).
8. **Slice 4 — shutdown timeout uses context.Background() as parent**: the signal context is already canceled when shutdown starts; using it as parent of `context.WithTimeout` would produce an instantly-expired deadline and skip draining. `shutdownServer` builds its own bounded context.
9. **Slice 4 — `cmd/server` health endpoints bypass the auth/CSRF middleware chain** via a top-level mux (`/healthz`, `/readyz` public; everything else behind `RequireAuth(CSRFProtect(...))` + request logging). This is required for container healthchecks and orchestrator probes.
10. Deviations 1–8 from slices 1–3 remain as previously recorded (.gitignore defect, stateless sessions, middleware order, CSRF header-primary, Detail wrapping, slice-1 items, unknown-route GET, nudges template defect).

## Remaining Tasks

None — all 20 tasks (Phases 1–5) complete. Next SDD phase: `sdd-verify`.

## Risks

- **Stateless logout** (slice 2): stolen cookie valid until TTL (24h). Accepted for single-operator local scope; documented in README + design doc.
- **`COOKIE_SECURE=true` default**: compose sets `COOKIE_SECURE=false` for plain-HTTP demo; production over TLS should keep true. Documented.
- **Login/logout exempt from CSRF+auth (allowlisted)** — accepted low-risk; documented.
- **Nudges template defects** (`nudges/_card.html` missing `nudge-card` define; `nudges/detail.html` map root) — pre-existing, documented in README/design as NOT fixed; `ToggleNudgeAPI` returns JSON fallback; `/nudges` pages render generic 500 via `renderHTML`. Fix belongs to a dedicated correction.
- **`gofmt -l .` CI gate**: now clean (drift normalized); any future unformatted commit fails CI by design.
- **Docker image tags pinned to golang:1.26.2-alpine / alpine:3.21**: verified reachable and built successfully; tag updates are intentional, reviewable changes.
- **Delete semantics / migration data checks** from slice 1 as previously recorded; README documents deletion as permanent (CASCADE, no soft-delete).
- **Slice 5 doc size**: 705 changed lines (572+/133−) — within the assigned slice budget (max 2500) and the planned PR 5 of the forecast; reported for reviewer awareness.
