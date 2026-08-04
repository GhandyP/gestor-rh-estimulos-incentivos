# Design: Production Hardening

## Technical Approach

Harden the existing CGO-free layered application in dependency order: **data-integrity → operator-security → integration-confidence → operable-delivery → portfolio-documentation**. Preserve `domain → engine → service → handler → store/sqlite`; use strict TDD and real `modernc.org/sqlite`. No framework, identity platform, or product capability is added.

## Architecture Decisions

| Decision | Choice | Rejected | Rationale |
|---|---|---|---|
| Domain validation | Add `Validate() error` to domain aggregates/value types and call it in service commands before store access. | Handler-only or SQL-only validation | One invariant source serves API, HTML, seed, and tests while SQL remains the final integrity boundary. |
| Transactions | `Store.WithTx` owns begin/commit/rollback; every repository operation used inside workflows accepts `*sql.Tx`. | Nested DB calls or callbacks using `s.db` | The current `CreateEmpleadoConPerfil` is falsely atomic because it ignores `tx`; one connection and one transaction are required. |
| Stimulus idempotency | `ApplyEstimulo` performs `UPDATE ... WHERE estado='pendiente'`, checks `RowsAffected`, then writes history and recalibrates in the same transaction; return `ApplyResult{Applied bool, Conflict bool}`. | Pre-read then update | Conditional transition is race-safe; a loser gets a deterministic conflict and cannot create side effects. |
| Operator security | Environment-configured single operator credentials and session secret; `Secure`, `HttpOnly`, `SameSite=Lax` cookie, auth middleware, mutating-route authorization, session-bound CSRF token. | Hardcoded secret, SSO/OIDC, multi-role RBAC | Fits the stated single-instance scope and keeps secrets deploy-time configurable. |
| Templates/lifecycle | Parse templates once from an explicit root or `embed.FS`; `cmd/server` validates config/assets, uses `slog`, health/readiness, signal-driven bounded shutdown. | Per-request relative parsing | Startup fails before traffic and behavior is independent of cwd. |

## Data Flow

```text
HTTP → auth/CSRF middleware → handler DTO → service validation
    → Store.WithTx → tx-aware repositories → SQLite constraints
    → JSON/HTML response
```

Employee creation inserts employee, `PerfilMAP`, and `Umbral` atomically. Stimulus application conditionally transitions state, inserts one history point, recalculates and updates the threshold, then commits; any error rolls back all writes.

## File Changes

| Slice | Files | Change |
|---|---|---|
| 1 data-integrity | `internal/domain/*.go`, `internal/service/service.go`, `internal/store/sqlite/{store,empleados,estimulos}.go`, `internal/store/sqlite/migrations/*.sql` | Central validators; tx-aware CRUD/workflows; affected-row contract; FK/CHECK/UNIQUE/index migration with forward/backfill and documented rollback preserving rows. |
| 2 security | `internal/handler/{handler,middleware,auth}.go`, `cmd/server/main.go`, templates/forms | Config/session/auth/CSRF middleware, safe error mapping, protected mutations and tokens. |
| 3 confidence | `internal/store/sqlite/*_test.go`, `internal/service/*_test.go`, `internal/handler/*_test.go`, `web/templates/empleados/list.html` | Real SQLite rollback/concurrency/double-apply tests; `httptest` auth/CSRF; deterministic template tests; remove duplicate delete markup. |
| 4 delivery | `cmd/server/main.go`, `Dockerfile`, `compose.yaml`, `.github/workflows/ci.yml` | Startup/lifecycle/logging/health and reproducible container/CI checks. |
| 5 documentation | `README.md`, `doc/*design*.md` | Architecture, threat model, persistence/backup, demo, deployment, verification, limitations. |

## Interfaces / Contracts

```go
func (s *Store) WithTx(ctx context.Context, fn func(*sql.Tx) error) error
func (s *Store) CreateEmpleadoTx(ctx context.Context, tx *sql.Tx, e *domain.Empleado) error
func (s *Store) CreatePerfilMAPTx(ctx context.Context, tx *sql.Tx, p *domain.PerfilMAP) error
func (s *Store) CreateUmbralTx(ctx context.Context, tx *sql.Tx, u *domain.Umbral) error
func (s *Store) TransitionEstimuloTx(ctx context.Context, tx *sql.Tx, id int64, at time.Time) (bool, error)
type ApplyResult struct { Applied, Conflict bool }
func (s *Service) ApplyEstimulo(ctx context.Context, id int64, respuesta float64) (ApplyResult, error)
```

## Testing Strategy

Unit tests cover every validator and result contract. SQLite integration tests use `t.TempDir()` or isolated in-memory databases, migrations/FKs/constraints, injected intermediate failures, rollback, repeated apply, and concurrent apply (one commit, one conflict). `httptest` covers unauthenticated/authenticated routes, cookie flags, missing/malformed/mismatched CSRF, unchanged persistence, and safe 4xx/5xx errors. Template tests parse/render repeatedly and assert one valid delete action per employee. E2E browser tests are explicitly out of scope.

## Threat Matrix

| Boundary | Applicability | Safe/failure behavior and RED test |
|---|---|---|
| Documentation-like paths | N/A: no executable-file classification | No execution boundary; no test. |
| Git repository selection | N/A: no VCS automation | No repository selection; no test. |
| Commit state | N/A: Git is not initialized | CI does not commit; no test. |
| Push state | N/A: no push automation | No destination/ref resolution; no test. |
| PR commands | Applicable: CI invokes shell checks | Pin commands to repository root; fail clearly on non-zero `go test/build/vet/gofmt`; RED test checks CI contains all four commands and no composed user input. |

## Migration / Rollout

Use numbered forward migrations: enable foreign keys per connection, add indexes and constraints after validating/backfilling existing rows, and preserve legacy values through explicit normalization. Rollback is a reverse, data-preserving migration (drop only newly added indexes/constraints after dependent data is validated), tested on a copy; deploy one instance and back up the SQLite file before migration.

## Open Questions

None blocking. The apply phase must preserve `delivery_strategy=ask-on-risk` and remain split into dependency-aligned slices because the change exceeds the 400-line review budget.
