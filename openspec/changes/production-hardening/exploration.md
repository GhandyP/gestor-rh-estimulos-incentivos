## Exploration: production-hardening

### Current State

The MVP is a layered Go application using `net/http`, `database/sql`, pure-Go `modernc.org/sqlite`, `html/template`, and HTMX. The domain and engine are separated from the service, handler, and SQLite store, but the application is not yet production-ready for a portfolio deployment.

Evidence from the repository:

- `cmd/server/main.go` opens SQLite, seeds demo data, starts `ListenAndServe`, and defers database close. It has server timeouts but no signal handling, graceful shutdown, health endpoint, structured logging, or configurable template root.
- `internal/service/service.go:27-58` creates an employee, MAP profile, and threshold through three independent writes. A partial failure leaves inconsistent data. `ApplyEstimulo` similarly performs status update, history insert, and threshold recalibration as independent writes.
- `internal/store/sqlite/estimulos.go:113-119` guards the update with `estado='pendiente'` but does not inspect affected rows; a repeated application can continue into history and recalibration even when the stimulus was already applied.
- `internal/store/sqlite/empleados.go:29-40` exposes `CreateEmpleadoConPerfil`, but its callback calls store methods that use `s.db`, not the supplied transaction, so it does not provide the intended atomicity. Foreign keys/cascades are also not explicitly enabled in `store.go` or the migration.
- Validation is distributed or absent: handlers decode JSON and call services directly; `UpdatePerfilMAP`, employee creation, incentive creation, and nudge creation do not centrally enforce ranges, required fields, enum values, or cross-field rules.
- `internal/handler/handler.go` has no authentication, authorization, CSRF protection, middleware, request IDs, or consistent error envelope. HTML templates are parsed from relative `web/templates/...` paths on every request, coupling runtime behavior to the working directory.
- `web/templates/empleados/list.html:34-41` contains duplicated/malformed delete-button markup and an extra closing `</td>`, confirming the known HTML gap despite the gap reports declaring completion.
- Tests currently consist of `internal/engine/engine_test.go` with 9 unit tests. No SQLite store integration tests, handler/HTTP tests, template rendering tests, or end-to-end tests are present. The cached project context reports 46% engine coverage, `go vet ./...` and `go build ./...` as available, and no E2E layer.
- The reports are inconsistent about build status: `backend-report.md` says build/tests passed, while `backend-gaps-report.md` reports a toolchain timeout. This must be reverified during implementation/verification rather than treated as a stable baseline.
- The repository contains no `Dockerfile`, Compose file, CI workflow, or deployment manifest. `README.md` documents local `go run` only and currently describes the MVP as private/internal.
- The project is not a Git repository. OpenSpec files can be created and reviewed on the filesystem, but commits, branches, PR slices, review history, and reproducible delivery receipts cannot exist until the user explicitly consents to `git init`.

For portfolio-ready scope, this change should harden the existing single-instance MVP without turning it into a multi-tenant HR platform.

### Affected Areas

- `cmd/server/main.go` — lifecycle, configuration, logging, health/readiness, and shutdown.
- `internal/domain/*.go` — centralized validation rules and domain invariants for employee, MAP profile, threshold, incentive, nudge, and stimulus values.
- `internal/service/service.go` — transactional application use cases, duplicate-application behavior, validation orchestration, and error semantics.
- `internal/store/sqlite/*.go` and `internal/store/sqlite/migrations/` — transaction-aware repository methods, affected-row checks, constraints/indexes, foreign-key enforcement, and migration coverage.
- `internal/handler/handler.go` — authentication/authorization boundary, CSRF protection for state-changing browser requests, request/error handling, and health route.
- `web/templates/empleados/list.html` — remove duplicate delete markup and verify rendered HTML.
- `web/templates/*` and template loading code — embedded templates or an explicit configurable template root with deterministic startup failure.
- New tests under `internal/store/sqlite`, `internal/service`, and `internal/handler` — SQLite integration, transaction rollback, concurrent/repeated apply, HTTP routes, CSRF/auth, and template rendering.
- New deployment/documentation files such as `Dockerfile`, `compose.yaml`, CI workflow, and updated `README.md` — reproducible local/container delivery and portfolio presentation.

### Approaches

1. **Incremental hardening in layered slices** — preserve the current architecture, add domain validation and transaction-aware store APIs, then secure, test, operate, deploy, and document it.
   - Pros: respects existing boundaries; each slice has a clear verification target and rollback boundary; minimizes framework/dependency growth; supports SQLite now and later PostgreSQL migration.
   - Cons: requires temporary compatibility work while service and store APIs evolve; the lack of Git prevents normal slice branches/PRs until consent is given.
   - Effort: High

2. **Rewrite around a new web framework or ORM** — replace handler/store plumbing while adding hardening features.
   - Pros: could provide middleware, validation, and repository abstractions faster in the short term.
   - Cons: expands MVP risk, abandons the existing stdlib/CGO-free constraint, makes the portfolio change harder to review, and does not automatically solve business transaction semantics or tests.
   - Effort: High

### Recommendation

Use incremental hardening and include all listed production-hardening recommendations, but explicitly limit the product to one trusted HR operator per deployed instance; authentication is a minimal application boundary, not a full identity platform.

Recommended slices and dependency order:

1. **Data integrity and domain invariants** — central validation; atomic employee + profile + threshold creation; atomic stimulus application + history + recalibration; affected-row/idempotency protection; foreign keys, constraints, indexes, and migration tests. This is the foundation for every later slice.
2. **Security boundary** — minimal operator authentication, authorization for mutating routes, CSRF tokens for browser state changes, secure cookie/session configuration, and safe request/error handling. Depends on stable mutation boundaries from slice 1.
3. **Integration confidence** — real SQLite tests with `t.TempDir()`/in-memory databases, rollback and repeat-apply cases, HTTP `httptest` route tests, auth/CSRF cases, template parsing/rendering, and regression coverage for the employee list HTML. Depends on the APIs established by slices 1–2; should be developed alongside them under strict TDD.
4. **Operability and reproducible delivery** — `log/slog`, startup validation, `/healthz` (and optionally `/readyz`), graceful signal shutdown, embedded or explicitly configured templates, Dockerfile, Compose for local demonstration, CI commands for test/build/vet/gofmt, and deployment configuration. Depends on stable startup and test behavior.
5. **Documentation and portfolio UX** — consolidate README and design documentation, document threat model and limitations, provide demo/deployment instructions, explain seed data and database persistence, and polish the corrected employee list behavior. Depends on the final operational contract and tested routes.

Scope included: authentication/CSRF, transactional integrity, duplicate-application prevention, centralized validation, HTML correction, SQLite + HTTP integration tests, graceful shutdown, deterministic template loading, structured logging, health endpoint, Docker/Compose/CI, and README/documentation consolidation.

Non-goals: SSO/OIDC, multi-tenancy, RBAC beyond the single operator role, external ERP/payroll integrations, email/Slack delivery, background job scheduling, PostgreSQL migration itself, Kubernetes/managed cloud infrastructure, advanced analytics/ML, and a full browser E2E suite. Those would require separate changes and product decisions.

### Risks

- Authentication and CSRF can introduce session/cookie behavior that must be tested without making the single-instance MVP appear to support enterprise identity or multi-tenancy.
- Transaction correctness is easy to fake: repository methods must execute through the same `*sql.Tx`, and tests must prove rollback after each intermediate failure.
- Repeated or concurrent stimulus application requires an atomic state transition and an explicit API contract (idempotent success versus conflict); merely adding a pre-read is race-prone.
- Existing migration code is effectively one-time/manual (`000001` plus an ad hoc column check); adding constraints or foreign-key behavior needs a forward migration and a rollback/data-safety plan.
- Relative template paths, automatic demo seeding, and ignored handler errors can make container/readiness behavior differ from local development.
- Reports disagree about Go build stability; implementation must capture fresh commands and distinguish environment/toolchain failures from application failures.
- The expected change exceeds the 400-line review budget. The dependency slices above should become chained/stacked delivery units after Git is initialized; until then, the filesystem can preserve boundaries but cannot provide branch or PR isolation.
- No Git repository means no commit-based audit trail, branch protection, or CI trigger. The change can be implemented and verified locally, but delivery as a portfolio PR is blocked until explicit Git initialization and remote/review decisions.

### Acceptance Definition

The MVP is ready for portfolio presentation and deployment when:

- Invalid domain data is rejected consistently, and all employee/stimulus workflows are atomic with rollback tests.
- A stimulus cannot be applied twice, including under concurrent requests; history and threshold recalibration occur exactly once on success.
- Mutating browser requests require authenticated operator access and valid CSRF protection; unauthorized and invalid-CSRF requests are rejected.
- SQLite integration and HTTP tests pass with deterministic template rendering, including the corrected employee list HTML and key happy/error paths.
- The server exposes a meaningful health endpoint, logs structured startup/request/error events, shuts down gracefully, and loads templates independently of the caller's working directory.
- A clean checkout can run the documented test/build/lint commands and start through the documented container/local path; CI reproduces those checks.
- README and design docs state architecture, threat/security limits, persistence/backup expectations, demo data behavior, deployment steps, and explicit non-goals.
- Fresh `go test ./...`, `go build ./...`, `go vet ./...`, and `gofmt -l .` results are recorded. Git/branch/PR delivery is marked pending until user consent to initialize Git.

### Ready for Proposal

Yes. The orchestrator should create a proposal that preserves the five-slice dependency order, records the single-operator portfolio scope and non-goals, forecasts a high 400-line budget risk, and asks for Git initialization consent before branch/PR delivery. The next phase is `sdd-propose`; after proposal, run `sdd-spec` and `sdd-design` before task planning.
