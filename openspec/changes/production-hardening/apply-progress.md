# Apply Progress: production-hardening — Slices 1-3 (data-integrity + operator-security + integration-confidence)

**Branches**: `slice/1-data-integrity` (baseline `4e8b406`) · `slice/2-operator-security` (from slice/1 head `8b2ebbe`) · `slice/3-integration-confidence` (from slice/2 head `05288ff`)
**Mode**: Strict TDD (go test ./..., `strict_tdd: true` en openspec/config.yaml)
**Delivery**: chained slices — PR 1 (data-integrity), PR 2 (operator-security), PR 3 (integration-confidence); per `delivery_strategy=ask-on-risk` resolved by the orchestrator to slice execution. No remote: local commits only.
**Status**: Phase 1 (tasks 1.1–1.7), Phase 2 (tasks 2.1–2.4) and Phase 3 (tasks 3.1–3.4) COMPLETE. Tree clean, all checks green.

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 1.1+1.2 | `internal/domain/validate_test.go` | Unit | N/A (new package tests) | ✅ Written first — compile fails: `Validate undefined` | ✅ 32/32 cases pass | ✅ 7+ cases per entity (valid + each rejection + cross-field) | ✅ gofmt, shared table harness |
| 1.3+1.4 | `internal/store/sqlite/store_test.go` | Integration (real SQLite, t.TempDir) | ✅ engine+domain baseline pass | ✅ Written first — compile fails: WithTx signature + missing Tx methods | ✅ 7/7 tests pass | ✅ commit, rollback, conditional transition, missing stimulus | ✅ insert logic extracted behind `execer`/`queryer` |
| 1.5+1.6 | `internal/service/service_test.go` | Integration (real store + SQLite trigger failure injection) | ✅ prior suite green | ✅ Written first — compile fails: `ApplyEstimulo returns 1 value` | ✅ 7/7 tests pass | ✅ 8-way concurrent one-winner; double apply; invalid respuesta | ✅ service.go gofmt'ed (pre-existing drift) |
| 1.7 | `internal/store/sqlite/migrations_test.go` | Integration (legacy + fresh real DBs) | ✅ prior suite green | ✅ Embedded RED: legacy insert with missing parent SUCCEEDS pre-migration, rejected post | ✅ 5/5 tests pass | ✅ idempotency, FK, CHECK, orphan backfill, down | ✅ runner extracted: `recordMigration`/`applyPending` |
| 2.1+2.2 (auth) | `internal/handler/auth_test.go` | Integration (httptest, real SQLite) | ✅ prior suite green (28 tests) | ✅ Written first — compile fails: `Authenticator undefined`, `New(svc, auth)` arity | ✅ 7/7 tests pass | ✅ 2 login-failure cases (user vs password), tampered token, logout lifecycle | ✅ session token nonce extracted (fixes same-second collision) |
| 2.1+2.3 (CSRF) | `internal/handler/csrf_test.go` | Integration (httptest, real SQLite) | ✅ auth tests green | ✅ Written first — compile fails: `Middleware undefined`; then runtime RED: wrong middleware order returned 403 always | ✅ 5/5 tests pass | ✅ absent/malformed/mismatched/valid-header/form-field; no-state-change via count | ✅ chain order fixed to `RequireAuth(CSRFProtect(next))` |
| 2.4 | `internal/handler/templates_security_test.go` | Structural (cwd-independent reads) | ✅ full suite green | ✅ Written first — assertions on not-yet-wired templates fail | ✅ 3/3 tests pass | ✅ meta tag + header injection + login + 6 mutation forms | ✅ page data maps share `csrfFor(r)` helper |
| 3.1 | `internal/store/sqlite/store_test.go` | Integration (real SQLite, t.TempDir) | ✅ full suite green (43 tests) | ✅ Written first — scenario not covered at store level (concurrency only at service layer) | ✅ 2/2 tests pass | ✅ 8-way concurrent one-winner + multi-entity rollback across 3 tables | ✅ gofmt |
| 3.2 | `internal/handler/routes_test.go` | Integration (httptest, real SQLite) | ✅ full suite green | ✅ Written first — success/error routes + PUT-rejection persistence not covered after slice 2 | ✅ 4/4 tests pass (6 success cases + 4 error cases + malformed body + PUT rejected) | ✅ success list/detail, 400/404/405 safe statuses, malformed body, second mutation verb | ✅ gofmt |
| 3.3+3.4 | `internal/handler/templates_render_test.go` + `web/templates/empleados/list.html` | Template render (parse+execute, cwd-independent) | ✅ full suite green | ✅ TRUE RED: rendered list had 2 delete actions per row and 6 open/7 close cells | ✅ RED→GREEN: tests fail pre-fix, pass post-fix (byte-identical re-render, 1 delete/row, 6/6 cells, key pages reproducible) | ✅ 3 employees fixture + 4 key pages + login | ✅ gofmt |

## Work Unit Evidence

| Evidence | Value |
|---|---|
| Focused test command / result | Slice 3: `go test ./internal/store/sqlite/ -run 'TransitionEstimuloTxConcurrentOneWinner\|MultiEntityWorkflowRollbackLeavesNoPartialState'` — 2/2 ok · `go test ./internal/handler/` — all ok (auth 7, CSRF 5, templates-security 3, routes 4, render 5) · slice 1: domain 32, store 12, service 7, migrations 5; slice 2: handler 15 |
| Runtime harness | `go test ./... -count=1 -timeout 120s` ok · `go build ./...` ok · `go vet ./...` ok · live server smoke test (slice 2) — see slice-2 evidence below |
| Rollback boundary | Slice 3: revert `6980051`→`8afb535` (3 commits) or `git revert 8afb535` for template fix; store/handler tests revert without touching slices 4–5; slice 1-2 rollback as previously recorded |

## Completed Tasks (cumulative)

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

## Files Changed

### Slice 1 (data-integrity)

| File | Action |
|------|--------|
| `internal/domain/{empleado,estimulo,perfil_map,umbral}.go` | Modified — added `Validate() error` |
| `internal/domain/validate_test.go` | Created — table-driven validator tests |
| `internal/store/sqlite/store.go` | Modified — `WithTx` per design contract, migration runner (`schema_migrations`), FK pragma per connection |
| `internal/store/sqlite/empleados.go` | Modified — `CreateEmpleadoTx`, `CreatePerfilMAPTx`, fixed `CreateEmpleadoConPerfil` to write through tx |
| `internal/store/sqlite/estimulos.go` | Modified — `CreateUmbralTx`, `TransitionEstimuloTx`, tx variants; removed old `ApplyEstimulo` |
| `internal/store/sqlite/store_test.go` | Created — 7 tx integration tests |
| `internal/store/sqlite/migrations/000002_hardening.{up,down}.sql` | Created — backfill + FK/CHECK/UNIQUE/index migration pair |
| `internal/store/sqlite/migrations_test.go` | Created — 5 migration tests |
| `internal/service/service.go` | Modified — atomic `CreateEmpleado`, `ApplyResult`, idempotent `ApplyEstimulo`, Validate-first |
| `internal/service/service_test.go` | Created — 7 service tests |
| `internal/handler/handler.go` | Modified — adapted `ApplyEstimuloAPI` to `ApplyResult` (conflict → `status: conflict`) |

### Slice 2 (operator-security)

| File | Action |
|------|--------|
| `internal/handler/auth.go` | Created — `AuthConfig`/`AuthConfigFromEnv` (OPERATOR_USER/PASSWORD, SESSION_SECRET, COOKIE_SECURE), `Authenticator` (constant-time creds, signed stateless session tokens with random nonce, session-bound CSRF derivation, Secure/HttpOnly/SameSite=Lax cookie) |
| `internal/handler/middleware.go` | Created — `RequireAuth` (401 JSON / 303 HTML), `CSRFProtect` (mutations only, X-CSRF-Token header or `_csrf` field), `Handler.Middleware` chain, safe-error JSON helpers |
| `internal/handler/handler.go` | Modified — `New(svc, auth)` signature, login/logout/login-page handlers, `csrfFor(r)` helper, `CSRFToken` in all page/form data (detail pages wrapped under `Detail`) |
| `internal/handler/auth_test.go` | Created — 7 httptest auth tests |
| `internal/handler/csrf_test.go` | Created — 5 httptest CSRF tests |
| `internal/handler/templates_security_test.go` | Created — 3 structural template-wiring tests |
| `cmd/server/main.go` | Modified — auth from env, startup warning for dev defaults (no values logged), `h.Middleware(mux)` as handler |
| `web/templates/login.html` | Created — standalone operator login page |
| `web/templates/base.html` | Modified — csrf meta tag, htmx `configRequest` X-CSRF-Token injection, logout form |
| `web/templates/empleados/{_form,_import_form,_perfil_form,detail}.html` | Modified — hidden `_csrf` inputs; detail wrapped under `Detail` |
| `web/templates/{incentivos/_form,incentivos/detail,nudges/_form,estimulos/_apply_form}.html` | Modified — hidden `_csrf` inputs; apply form under `Estimulo` |
| `.gitignore` | Fixed — `server` → `/server` (was ignoring all of `cmd/server/`) |

### Slice 3 (integration-confidence)

| File | Action |
|------|--------|
| `internal/store/sqlite/store_test.go` | Modified — `TestTransitionEstimuloTxConcurrentOneWinner` (8 goroutines, exactly 1 winner, 1 historial, estado aplicado) + `TestMultiEntityWorkflowRollbackLeavesNoPartialState` (failure after 3 writes → 0 rows in empleados/perfiles_map/umbrales) |
| `internal/handler/routes_test.go` | Created — `TestAuthenticatedSuccessRoutes` (6 JSON read endpoints), `TestErrorRoutesReturnSafeStatuses` (400/404/405, no internal details), `TestMalformedJSONBodyRejectedWithoutStateChange`, `TestRejectedUpdateMutationDoesNotChangeState` (PUT sin CSRF → 403, empleado intacto) |
| `internal/handler/templates_render_test.go` | Created — `TestEmpleadosListParseAndDeterministicRender`, `TestEmpleadosListExactlyOneDeleteActionPerRow`, `TestEmpleadosListBalancedCellsPerRow`, `TestKeyPagesParseAndDeterministicRender` (dashboard/incentivos/estimulos/login), `TestNudgesListParses` (parse-only, defect documentado) |
| `web/templates/empleados/list.html` | Fixed — removed duplicated delete `<button>` y `</td>` huérfano; una sola acción de borrado por fila, celdas 6/6 balanceadas |

## Deviations from Design

1. **`.gitignore` defect fixed (slice 2)**: pattern `server` ignored the whole `cmd/server/` directory — `cmd/server/main.go` had never been tracked. Anchored to `/server` so the built binary stays ignored but `cmd/server` is committed. Pre-existing repo defect, discovered because task 2.4 mandates main.go wiring.
2. **Stateless sessions**: session tokens are signed, stateless cookies (expiry + nonce + HMAC). Logout clears the client cookie but does NOT revoke already-issued tokens (revocation would need server-side state); documented as accepted limitation for a single-operator local tool. The nonce was added after the CSRF mismatch test exposed same-second token collisions.
3. **Middleware order**: chain is `RequireAuth(CSRFProtect(next))` — CSRFProtect needs the authenticated session in context; the reverse order (initially implemented) made every mutation return 403. Caught by the httptest suite.
4. **CSRF via header is the primary mechanism**: HTMX forms use `json-enc`, which `r.FormValue("_csrf")` cannot parse; the global `htmx:configRequest` header injection (meta tag) covers every HTMX mutation including `hx-delete`/`hx-put` buttons; hidden `_csrf` inputs are added to dedicated forms as the non-JS fallback.
5. **Detail pages wrapped under `Detail`** (and apply form under `Estimulo`, perfil form under `Perfil`) so `base.html` can render `{{.CSRFToken}}` from map data.
6. Slice 1 deviations (CreateEmpleado atomicity, extra tx variants, ON DELETE CASCADE, WithTx signature) remain as previously recorded.
7. **Slice 3 — unknown-route GET**: the mux registers `GET /` (dashboard), so an unknown GET path falls through to the dashboard and returns 500 when templates are cwd-relative (slice-4 concern). The httptest error-route table covers the real boundary instead: 400 invalid ID, 404 missing resource, 405 method-not-allowed on unmapped paths (POST + valid CSRF reaches the mux). No handler code was changed in slice 3.
8. **Slice 3 — pre-existing nudges template defect (NOT fixed, out of scope)**: `web/templates/nudges/_card.html` lacks `{{define "nudge-card"}}`, but `nudges/list.html` invokes `{{template "nudge-card" .}}`. html/template resolves `{{template}}` references at execution even in un-taken branches, so `/nudges` renders an empty 200 (handler ignores the ExecuteTemplate error) and `ToggleNudgeAPI`'s `ExecuteTemplate(w, "nudge-card", ...)` errors. Covered as parse-only (`TestNudgesListParses`); the fix (adding the define) belongs to a later slice or a dedicated correction.

## Remaining Tasks (other slices, untouched)

- [ ] Phase 4 (operable-delivery), Phase 5 (portfolio-documentation)

## Risks

- **Stateless logout** (slice 2): a stolen session cookie remains valid until TTL (24h default) even after logout. Accepted for single-operator local scope; documented in code.
- **`COOKIE_SECURE=true` default**: over plain-HTTP deployments (e.g., non-localhost) browsers will not store the session cookie; must set `COOKIE_SECURE=false` for such demos or serve TLS (health/readiness and Docker are slice 4).
- **Login is exempt from CSRF** and **logout is exempt from auth+CSRF** (allowlisted) — login CSRF is a recognized low-risk pattern; logout CSRF impact is nil.
- **Pre-existing gofmt drift** in `internal/engine/*`, `internal/domain/incentivo.go`, `internal/domain/nudge.go` — untouched (out of scope per slice instructions).
- **Pre-existing**: `nudges/detail.html` renders `{{with .}}` against a map root (`Nudge` key), so `.Nombre`/`.Descripcion` fields render empty — pre-existing bug observed while wrapping data; out of slice-2 scope (noted for a later phase). **Slice 3 adds**: `nudges/_card.html` missing `{{define "nudge-card"}}` breaks `/nudges` render and `ToggleNudgeAPI` — needs a later correction.
- **Templates still cwd-relative** (`web/templates/...`): slice 4 makes startup cwd-independent (embed/root). Structural and render tests resolve the repo root via `runtime.Caller` so they pass regardless of cwd.
- **Delete semantics / migration data checks** from slice 1 remain as previously recorded.
