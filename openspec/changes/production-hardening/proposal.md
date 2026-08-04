# Proposal: Production Hardening

## Intent

Make the existing MVP ready for portfolio presentation and deployment without adding product functionality. Harden it while preserving the layered CGO-free Go architecture.

## Scope

### In Scope
1. **Data integrity:** validation; atomic employee/ProfileMAP/Umbral and Estimulo/history/recalibration workflows; duplicate prevention; SQLite FKs, constraints, indexes, and migrations.
2. **Security boundary:** single trusted operator authentication/authorization, CSRF, secure cookies, and safe request/error handling.
3. **Integration confidence:** SQLite rollback/concurrency/repeat-apply tests, `httptest` auth/CSRF routes, template rendering, and employee-list regression coverage.
4. **Operability and delivery:** `slog`, startup validation, health, graceful shutdown, deterministic templates, Dockerfile, Compose, CI, and reproducible checks.
5. **Documentation and portfolio UX:** README/design consolidation, threat model, limitations, persistence, demo, deployment, and corrected employee-list presentation.

### Out of Scope

SSO/OIDC, multi-tenancy, RBAC beyond one operator, ERP/payroll/email/Slack integrations, background jobs, PostgreSQL migration, Kubernetes, analytics/ML, and browser E2E.

## Capabilities

### New Capabilities
- `data-integrity`: Invariants, transactions, and SQLite constraints.
- `operator-security`: Single-operator auth, authorization, and CSRF.
- `integration-confidence`: SQLite, HTTP, template, and regression tests.
- `operable-delivery`: Lifecycle, health, logging, templates, containers, and CI.
- `portfolio-documentation`: Documentation and portfolio UX contract.

### Modified Capabilities
- None; no existing source specs are present.

## Approach

Harden current layers in dependency order with strict TDD (`go test ./...`), preserving domain → engine → service → handler → store and `modernc.org/sqlite`. Scope exceeds 400 lines: `delivery_strategy = ask-on-risk`; ask about chained delivery before apply. `git init` is pending consent.

## Affected Areas

| Area | Impact |
|------|--------|
| `internal/domain`, `service`, `store/sqlite` | Validation, transactions, migrations |
| `handler`, `cmd/server`, `web/templates` | Security, lifecycle, health, UX |
| Tests, Docker/Compose/CI, docs | Confidence and reproducible delivery |

## Risks

- **High:** Transaction/concurrent-apply errors; use same-`sql.Tx` APIs and rollback tests.
- **High:** Review workload exceeds budget; ask before apply and use dependency-aligned slices.
- **Medium:** Security implies enterprise identity; document single-operator limits.

## Rollback Plan

Revert each slice independently with data-preserving migration rollback. Keep delivery/docs separable. Do not initialize Git without consent.

## Dependencies

- User consent for `git init`; fresh migration/toolchain results.

## Success Criteria

- [ ] Invalid data is rejected; employee/stimulus workflows are atomic.
- [ ] Stimuli cannot apply twice, including concurrently; history/recalibration occur once.
- [ ] Authenticated operator and valid CSRF protect browser mutations.
- [ ] SQLite/HTTP/template tests pass; health, structured logs, graceful shutdown, and deterministic templates work.
- [ ] Local/container delivery and CI reproduce `go test ./...`, `go build ./...`, `go vet ./...`, and `gofmt -l .`.
- [ ] README/design docs cover architecture, security, persistence, demo, deployment, and non-goals; Git delivery remains pending consent.
