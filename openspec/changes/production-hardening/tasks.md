# Tasks: Production Hardening

## Review Workload Forecast

Estimated changed lines: ~2,200–2,600 (code + tests + docs + CI); suggested split: 5 PRs, one per slice.

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Focused test command | Runtime harness | Rollback boundary |
|------|------|-----------|----------------------|-----------------|-------------------|
| 1 | data-integrity (validators, tx, idempotency, migrations) | PR 1 | `go test ./internal/...` | `go test ./...` — real SQLite temp DBs | Revert domain/service/store + migration pair |
| 2 | operator-security (auth, CSRF, middleware) | PR 2 | `go test ./internal/handler/` | `go test ./internal/handler/` (httptest) | Revert handler/auth/middleware + main.go wiring |
| 3 | integration-confidence (concurrency, httptest, template tests, list.html) | PR 3 | `go test ./...` | `go test ./...` | Revert test files + list.html |
| 4 | operable-delivery (lifecycle, health, Docker, CI) | PR 4 | `go vet ./... && gofmt -l .` | `docker compose up` + `curl /healthz` | Revert main.go/Dockerfile/compose/CI |
| 5 | portfolio-documentation (README + design docs) | PR 5 | N/A — docs, no executable behavior | N/A — docs only; manual walkthrough | Revert README/doc files |

## Phase 1: Data Integrity

- [x] 1.1 RED: `internal/domain/*_test.go` — Validate() rejects invalid ranges/enums/cross-field
- [x] 1.2 GREEN: add Validate() to Empleado, PerfilMAP, Umbral, Estimulo (`empleado.go`, `perfil_map.go`, `umbral.go`, `estimulo.go`)
- [x] 1.3 RED: sqlite tests — WithTx rollback, tx-aware CRUD on real DB
- [x] 1.4 GREEN: `Store.WithTx` + `CreateEmpleadoTx`/`CreatePerfilMAPTx`/`CreateUmbralTx`/`TransitionEstimuloTx`
- [x] 1.5 RED: service tests — `CreateEmpleadoConPerfil` rollback; ApplyEstimulo repeat/concurrent (one winner)
- [x] 1.6 GREEN: fix `CreateEmpleadoConPerfil` (use WithTx); `ApplyEstimulo` conditional UPDATE + RowsAffected + `ApplyResult`; Validate() first
- [x] 1.7 Migrations: `000002_hardening.up.sql` (FK/CHECK/UNIQUE/index after backfill) + `.down.sql` (data-preserving)

## Phase 2: Operator Security

- [ ] 2.1 RED: handler tests — 401 unauthenticated, cookie flags, bad CSRF rejected, no state change
- [ ] 2.2 GREEN: `internal/handler/auth.go` — env operator creds + session secret, login/logout, Secure/HttpOnly/SameSite=Lax cookie
- [ ] 2.3 GREEN: `internal/handler/middleware.go` — auth middleware, mutation authorization, session-bound CSRF, safe errors
- [ ] 2.4 GREEN: wire middleware + CSRF into `handler.go`, `cmd/server/main.go`, template forms

## Phase 3: Integration Confidence

- [ ] 3.1 RED: SQLite tests — concurrent apply one-winner; mid-tx failure no partial state
- [ ] 3.2 RED: httptest suite — auth/CSRF/success/error routes; persistence unchanged on rejected requests
- [ ] 3.3 RED: template test — `list.html` renders deterministically, parseable, exactly one delete action per row
- [ ] 3.4 GREEN: remove duplicate delete markup in `web/templates/empleados/list.html`

## Phase 4: Operable Delivery

- [ ] 4.1 `cmd/server/main.go` — startup validation (fail before traffic), slog, /healthz + /readyz, bounded shutdown, cwd-independent templates
- [ ] 4.2 `Dockerfile` + `compose.yaml` — single-instance demo, SQLite volume, healthcheck
- [ ] 4.3 `.github/workflows/ci.yml` — go test/build/vet + `gofmt -l .`, pinned to repo root, fail on non-zero (threat matrix: PR commands)

## Phase 5: Portfolio Documentation

- [ ] 5.1 `README.md` — architecture layers, threat model + single-operator scope, persistence/backup, demo/seed, deployment, verification, non-goals
- [ ] 5.2 `doc/*design*.md` — decisions, migration/rollback, limitations; walkthrough matches tested behavior
