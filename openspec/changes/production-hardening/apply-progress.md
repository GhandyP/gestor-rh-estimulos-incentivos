# Apply Progress: production-hardening — Slice 1 (data-integrity)

**Branch**: `slice/1-data-integrity` (baseline `4e8b406`)
**Mode**: Strict TDD (go test ./..., `strict_tdd: true` en openspec/config.yaml)
**Delivery**: chained slice — PR 1 of 5 (data-integrity), per `delivery_strategy=ask-on-risk` resolved by the orchestrator to slice execution. No remote: local commits only.
**Status**: All Phase 1 tasks (1.1–1.7) complete. Tree clean, all checks green.

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 1.1+1.2 | `internal/domain/validate_test.go` | Unit | N/A (new package tests) | ✅ Written first — compile fails: `Validate undefined` | ✅ 32/32 cases pass | ✅ 7+ cases per entity (valid + each rejection + cross-field) | ✅ gofmt, shared table harness |
| 1.3+1.4 | `internal/store/sqlite/store_test.go` | Integration (real SQLite, t.TempDir) | ✅ engine+domain baseline pass | ✅ Written first — compile fails: WithTx signature + missing Tx methods | ✅ 7/7 tests pass | ✅ commit, rollback, conditional transition, missing stimulus | ✅ insert logic extracted behind `execer`/`queryer` |
| 1.5+1.6 | `internal/service/service_test.go` | Integration (real store + SQLite trigger failure injection) | ✅ prior suite green | ✅ Written first — compile fails: `ApplyEstimulo returns 1 value` | ✅ 7/7 tests pass | ✅ 8-way concurrent one-winner; double apply; invalid respuesta | ✅ service.go gofmt'ed (pre-existing drift) |
| 1.7 | `internal/store/sqlite/migrations_test.go` | Integration (legacy + fresh real DBs) | ✅ prior suite green | ✅ Embedded RED: legacy insert with missing parent SUCCEEDS pre-migration, rejected post | ✅ 5/5 tests pass | ✅ idempotency, FK, CHECK, orphan backfill, down | ✅ runner extracted: `recordMigration`/`applyPending` |

## Work Unit Evidence

| Evidence | Value |
|---|---|
| Focused test command / result | `go test ./internal/domain/` ok · `go test ./internal/store/sqlite/` 12/12 ok · `go test ./internal/service/` 7/7 ok |
| Runtime harness | `go test ./...` (28 tests) ok, `go build ./...` ok, `go vet ./...` ok; migration replayed on a copy of the real demo DB `estimulos.db` via sqlite3: rows preserved 6/6/6/2/1, `PRAGMA foreign_key_check` empty, FK insert rejected (error 19) |
| Rollback boundary | Revert commits 43c7814→1c71f82 or `000002_hardening.down.sql` for schema; domain/service/store changes revert without touching slices 2–5 |

## Completed Tasks

- [x] 1.1 RED domain validator tests
- [x] 1.2 GREEN `Validate()` on Empleado, PerfilMAP, Umbral, Estimulo
- [x] 1.3 RED sqlite tx tests (real DB)
- [x] 1.4 GREEN `Store.WithTx` + tx-aware CRUD + `TransitionEstimuloTx`
- [x] 1.5 RED service tests (rollback, repeat, concurrent one-winner)
- [x] 1.6 GREEN atomic `CreateEmpleado`, idempotent `ApplyEstimulo` (`ApplyResult{Applied,Conflict}`, Validate-first)
- [x] 1.7 Migrations `000002_hardening` up/down + runner + FK per connection

## Files Changed (slice 1)

| File | Action |
|------|--------|
| `internal/domain/{empleado,estimulo,perfil_map,umbral}.go` | Modified — added `Validate() error` |
| `internal/domain/validate_test.go` | Created — table-driven validator tests |
| `internal/store/sqlite/store.go` | Modified — `WithTx` per design contract, migration runner (`schema_migrations`), FK pragma per connection |
| `internal/store/sqlite/empleados.go` | Modified — `CreateEmpleadoTx`, `CreatePerfilMAPTx`, fixed `CreateEmpleadoConPerfil` to write through tx |
| `internal/store/sqlite/estimulos.go` | Modified — `CreateUmbralTx`, `TransitionEstimuloTx`, tx variants for umbral/historial; removed old `ApplyEstimulo` |
| `internal/store/sqlite/store_test.go` | Created — 7 tx integration tests |
| `internal/store/sqlite/migrations/000002_hardening.up.sql` | Created — backfill + rebuilds with FK/CHECK/UNIQUE + 3 indexes |
| `internal/store/sqlite/migrations/000002_hardening.down.sql` | Created — data-preserving rollback |
| `internal/store/sqlite/migrations_test.go` | Created — 5 migration tests |
| `internal/service/service.go` | Modified — atomic `CreateEmpleado`, `ApplyResult`, idempotent `ApplyEstimulo`, Validate-first in commands |
| `internal/service/service_test.go` | Created — 7 service tests |
| `internal/handler/handler.go` | Modified — adapted `ApplyEstimuloAPI` to `ApplyResult` (conflict → `status: conflict`) |

## Deviations from Design

1. **`CreateEmpleadoConPerfil` fixed at store level** and service `CreateEmpleado` made atomic (employee+PerfilMAP+Umbral in one `WithTx`) — the service previously had no `CreateEmpleadoConPerfil`; the store method is now exercised by tests.
2. **Additional tx variants** (`GetUmbralTx`, `AddHistorialTx`, `GetHistorialTx`, `UpdateUmbralTx`) beyond the four listed contracts — required so `ApplyEstimulo` history+recalibration run inside the SAME transaction per spec.
3. **`ON DELETE CASCADE`** on child FKs (perfiles_map, umbrales, estimulos, historial_estimulos) — preserves the existing UI delete-employee flow; with RESTRICT the delete endpoint would break once FKs are enforced.
4. `WithTx` signature changed to the design contract `func(tx *sql.Tx) error` (was ctx-carrying).

## Remaining Tasks (other slices, untouched)

- [ ] Phase 2 (operator-security), Phase 3 (integration-confidence), Phase 4 (operable-delivery), Phase 5 (portfolio-documentation)

## Risks

- **Delete semantics**: with FKs enforced, `DeleteEmpleado` cascades (deliberate). Deleting an incentivo with elegibilidades now fails with a constraint error until a policy is chosen in a later phase.
- **Migration failure on legacy data**: CHECK constraints (`email`/`nombre` non-empty) make the migration fail loudly if a legacy DB has empty values instead of silently normalizing — fail-safe by design.
- **`Estimulo.Validate` requires non-empty `contenido`** — manual estimulo creation without content is now rejected; only the engine's `Recomendar` path creates estimulos today.
- **gofmt drift pre-existing** in `internal/engine/*`, `internal/domain/incentivo.go`, `internal/domain/nudge.go` — untouched (out of scope per slice instructions).
