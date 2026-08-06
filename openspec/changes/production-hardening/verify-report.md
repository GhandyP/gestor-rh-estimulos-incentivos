```yaml
schema: gentle-ai.verify-result/v1
evidence_revision: sha256:020273f20bca5638cc4d0e2a8341b0316f8283716469e881989ae3093698a876
verdict: fail
blockers: 4
critical_findings: 4
requirements: 8/10
scenarios: 16/20
test_command: go test ./... -count=1 -timeout 120s
test_exit_code: 0
test_output_hash: sha256:4b080a6a5b1a3a0ee7997f5ab3a02869b9b755eb6eff53b02032a4b34cbcf35b
build_command: go build ./...
build_exit_code: 0
build_output_hash: sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
```

## Verification Report

**Status**: failed
**Change**: `production-hardening`
**Version**: N/A — five OpenSpec delta specifications
**Mode**: Strict TDD
**Project root**: `/home/ghandy/Documents/Estimulos-e-incentivos`
**Artifact store**: OpenSpec only

### Executive Summary

All 20 implementation tasks are checked, the current runtime proof passed for all six Go packages, and build, vet, and formatting checks passed. Strict scenario verification fails because the two documentation requirements contain four scenarios with no covering runtime test; structural/manual documentation evidence exists but cannot satisfy the strict runtime-evidence rule.

### Artifact Completeness

| Artifact | Status | Path |
|----------|--------|------|
| Proposal | ✅ Read | `openspec/changes/production-hardening/proposal.md` |
| Specifications | ✅ Read — 5 files | `openspec/changes/production-hardening/specs/` |
| Design | ✅ Read | `openspec/changes/production-hardening/design.md` |
| Tasks | ✅ Read — 20/20 checked | `openspec/changes/production-hardening/tasks.md` |
| Apply progress | ✅ Read — cumulative slices 1–5 preserved | `openspec/changes/production-hardening/apply-progress.md` |
| Verify report | ✅ Candidate admitted before persistence | `openspec/changes/production-hardening/verify-report.md` |

### Requirement and Scenario Counts

Counts are taken from the five retrieved specs: 2 requirements and 4 scenarios per spec, for 10 requirements and 20 scenarios total.

| Spec | Requirements | Scenarios | Requirements fully runtime-verified | Scenarios runtime-verified |
|------|--------------|-----------|-------------------------------------|----------------------------|
| `data-integrity` | 2 | 4 | 2 | 4 |
| `operator-security` | 2 | 4 | 2 | 4 |
| `integration-confidence` | 2 | 4 | 2 | 4 |
| `operable-delivery` | 2 | 4 | 2 | 4 |
| `portfolio-documentation` | 2 | 4 | 0 | 0 |
| **Total** | **10** | **20** | **8** | **16** |

### Completeness

| Metric | Value |
|--------|-------|
| Tasks total | 20 |
| Tasks complete | 20 |
| Tasks incomplete | 0 |
| Requirements fully verified | 8/10 |
| Scenarios with passing runtime coverage | 16/20 |

### Task Coverage

| Task | Result | Evidence |
|------|--------|----------|
| 1.1 | ✅ | `internal/domain/validate_test.go` validator cases |
| 1.2 | ✅ | `Validate()` on `Empleado`, `PerfilMAP`, `Umbral`, and `Estimulo` |
| 1.3 | ✅ | `internal/store/sqlite/store_test.go` transaction rollback/CRUD tests |
| 1.4 | ✅ | `Store.WithTx` and transaction-aware repository methods |
| 1.5 | ✅ | `internal/service/service_test.go` rollback, repeated apply, concurrent apply |
| 1.6 | ✅ | Service validation-first workflow and conditional `ApplyResult` transition |
| 1.7 | ✅ | `000002_hardening` migration and `migrations_test.go` |
| 2.1 | ✅ | `auth_test.go`, `csrf_test.go`: unauthenticated and invalid-CSRF rejection |
| 2.2 | ✅ | `internal/handler/auth.go` environment credentials and protected cookie |
| 2.3 | ✅ | `internal/handler/middleware.go` auth, mutation authorization, and CSRF |
| 2.4 | ✅ | Handler/server wiring and CSRF-bearing templates |
| 3.1 | ✅ | Real SQLite concurrent winner and multi-entity rollback tests |
| 3.2 | ✅ | `routes_test.go` and CSRF `httptest` success/error/persistence tests |
| 3.3 | ✅ | Deterministic template render and employee-list regression tests |
| 3.4 | ✅ | `web/templates/empleados/list.html` has one delete action per row |
| 4.1 | ✅ | Config, root, health, lifecycle, and template-set tests; current suite passed |
| 4.2 | ✅ | Dockerfile/Compose structural tests; historical slice-4 live smoke evidence |
| 4.3 | ✅ | CI structural tests plus current test/build/vet/gofmt proof |
| 5.1 | ✅* | README content structurally checked in apply progress; no runtime behavior |
| 5.2 | ✅* | Design documentation structurally checked in apply progress; no runtime behavior |
| **Total** | **20/20 checked** | `*` documentation tasks are complete but their spec scenarios lack runtime covering tests |

### Build & Tests Execution

The following final-proof commands were executed once to completion in the repository root. Output hashes are SHA-256 over the exact combined stdout/stderr bytes.

| Check | Command | Exit code | Output SHA-256 | Result |
|-------|---------|-----------|----------------|--------|
| Tests | `go test ./... -count=1 -timeout 120s` | 0 | `sha256:4b080a6a5b1a3a0ee7997f5ab3a02869b9b755eb6eff53b02032a4b34cbcf35b` | ✅ Passed |
| Build | `go build ./...` | 0 | `sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` | ✅ Passed |
| Vet | `go vet ./...` | 0 | `sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` | ✅ Passed |
| Format | `gofmt -l .` | 0 | `sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` | ✅ Empty output |

**Exact test output**:

```text
ok  	estimulos-incentivos/cmd/server	1.677s
ok  	estimulos-incentivos/internal/domain	0.019s
ok  	estimulos-incentivos/internal/engine	0.030s
ok  	estimulos-incentivos/internal/handler	1.432s
ok  	estimulos-incentivos/internal/service	0.577s
ok  	estimulos-incentivos/internal/store/sqlite	0.808s
```

Build, vet, and gofmt produced no output. The empty-output digest is `sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.

**Coverage**: changed-file coverage was not rerun because the explicit final-proof command contract required exactly the four commands above. The cached package percentages in `openspec/config.yaml` are historical and are not used as current verification evidence.

### Spec Compliance Matrix

`COMPLIANT` means a covering test passed in the current runtime proof. The four documentation scenarios are `UNTESTED` under the strict rule even though their documents were manually and structurally reviewed.

| Spec requirement | Scenario | Passing runtime evidence | Result |
|------------------|----------|--------------------------|--------|
| Data integrity — validated and atomic domain workflows | Invalid input is rejected | `internal/service/service_test.go > TestCreateEmpleadoRejectsInvalidInputBeforePersist`; `TestApplyEstimuloRejectsInvalidRespuesta`; domain validator table tests | ✅ COMPLIANT |
| Data integrity — validated and atomic domain workflows | Intermediate failure rolls back | `internal/service/service_test.go > TestCreateEmpleadoRollsBackFullyOnIntermediateFailure`; `internal/store/sqlite/store_test.go > TestMultiEntityWorkflowRollbackLeavesNoPartialState` | ✅ COMPLIANT |
| Data integrity — constraint-backed idempotency | Repeat application is harmless | `internal/service/service_test.go > TestApplyEstimuloRepeatedHasOneWinner` | ✅ COMPLIANT |
| Data integrity — constraint-backed idempotency | Concurrent application has one winner | `internal/service/service_test.go > TestApplyEstimuloConcurrentHasExactlyOneWinner`; store-level concurrent transition test | ✅ COMPLIANT |
| Operator security — single-operator security boundary | Unauthenticated mutation is rejected | `internal/handler/auth_test.go > TestUnauthenticatedMutationRejectedWithoutStateChange` | ✅ COMPLIANT |
| Operator security — single-operator security boundary | Authenticated operator is allowed | `internal/handler/auth_test.go > TestLoginSuccessSetsSecureSessionCookie`; `TestAuthenticatedOperatorAllowedOnAuthorizedRoute` | ✅ COMPLIANT |
| Operator security — CSRF-protected browser mutations | Valid token permits mutation | `internal/handler/csrf_test.go > TestValidCSRFTokenPermitsMutation`; `TestCSRFViaFormFieldAccepted` | ✅ COMPLIANT |
| Operator security — CSRF-protected browser mutations | Invalid token is rejected | `internal/handler/csrf_test.go > TestCSRFAbsentRejectedWithoutStateChange`; `TestCSRFMalformedRejectedWithoutStateChange`; `TestCSRFMismatchedRejectedWithoutStateChange`; `routes_test.go > TestRejectedUpdateMutationDoesNotChangeState` | ✅ COMPLIANT |
| Integration confidence — reproducible persistence and HTTP confidence | SQLite failures prove rollback | `internal/store/sqlite/store_test.go > TestWithTxRollbackOnError`; `TestMultiEntityWorkflowRollbackLeavesNoPartialState`; migration tests | ✅ COMPLIANT |
| Integration confidence — reproducible persistence and HTTP confidence | HTTP boundary rejects unsafe requests | `internal/handler/auth_test.go`, `csrf_test.go`, and `routes_test.go` `httptest` cases | ✅ COMPLIANT |
| Integration confidence — deterministic rendering regression coverage | Template rendering is stable | `internal/handler/templates_set_test.go > TestTemplateSetExecutePageIsDeterministic`; `templates_render_test.go > TestEmpleadosListParseAndDeterministicRender`; `TestKeyPagesParseAndDeterministicRender` | ✅ COMPLIANT |
| Integration confidence — deterministic rendering regression coverage | Employee list has one action | `internal/handler/templates_render_test.go > TestEmpleadosListExactlyOneDeleteActionPerRow`; `TestEmpleadosListBalancedCellsPerRow` | ✅ COMPLIANT |
| Operable delivery — observable and safe lifecycle | Health reflects readiness | `cmd/server/health_test.go > TestHealthzReportsOK`; `TestReadyzReportsReady`; not-ready cases | ✅ COMPLIANT |
| Operable delivery — observable and safe lifecycle | Shutdown drains safely | `cmd/server/lifecycle_test.go > TestRunServesHealthAndShutsDownWithinBound`; `TestShutdownDrainsInFlightRequestWithinBound`; `TestShutdownForceClosesWhenBoundExceeded` | ✅ COMPLIANT |
| Operable delivery — reproducible delivery artifacts | Clean checkout passes checks | `cmd/server/ci_test.go > TestCIWorkflowRunsAllFourChecksAtRepoRoot` plus current test/build/vet/gofmt commands | ✅ COMPLIANT |
| Operable delivery — reproducible delivery artifacts | Missing startup asset fails safely | `cmd/server/lifecycle_test.go > TestRunStartupFailsOnMissingTemplates`; `TestRunStartupFailsOnUnopenableDatabase` | ✅ COMPLIANT |
| Portfolio documentation — complete portfolio contract | New operator can deploy safely | No runtime covering test; structural/manual review in `apply-progress.md` only | ❌ UNTESTED |
| Portfolio documentation — complete portfolio contract | Threat model states boundaries | No runtime covering test; README/design content reviewed statically | ❌ UNTESTED |
| Portfolio documentation — correct portfolio UX | Demo matches implementation | No runtime covering test; README/design and source claims cross-checked statically | ❌ UNTESTED |
| Portfolio documentation — correct portfolio UX | Non-goals remain explicit | No runtime covering test; README/design content reviewed statically | ❌ UNTESTED |

**Compliance summary**: 16/20 scenarios have passing runtime coverage; 4/20 documentation scenarios remain `UNTESTED` under Strict TDD verification.

### Correctness (Static Evidence)

| Requirement | Status | Notes |
|-------------|--------|-------|
| Validated and atomic domain workflows | ✅ Implemented | Domain validators, service validation-first paths, `WithTx`, tx-aware repositories, and rollback tests are present. |
| Constraint-backed idempotency | ✅ Implemented | Migration constraints/indexes and conditional `RowsAffected` transition are implemented and tested, including repeated/concurrent application. |
| Single-operator security boundary | ✅ Implemented | Environment-configured operator, signed sessions, secure cookie protections, authorization middleware, and generic error mapping are implemented and tested. |
| CSRF-protected browser mutations | ✅ Implemented | Session-bound header/form token validation protects all non-safe methods and rejected requests preserve state. |
| Reproducible persistence and HTTP confidence | ✅ Implemented | Real SQLite and `httptest` coverage passes across migrations, rollback, concurrency, auth, CSRF, and safe routes. |
| Deterministic rendering regression coverage | ✅ Implemented | Parse-once template sets and employee-list rendering tests pass. |
| Observable and safe lifecycle | ✅ Implemented | Startup validation, structured `slog`, health/readiness, bounded drain, and forced close behavior are tested. |
| Reproducible delivery artifacts | ✅ Implemented | Docker/Compose/CI structural checks pass and all four current verification commands pass. |
| Complete portfolio contract | ⚠️ Implemented statically; runtime verification incomplete | README and design documents contain the requested architecture, security, persistence, demo, deployment, verification, and limitation material, but no runtime test covers the documentation scenarios. |
| Correct portfolio UX | ⚠️ Implemented statically; runtime verification incomplete | Employee-list behavior is runtime-tested and documented accurately, but the documentation/non-goal scenarios have no runtime covering tests. |

### Design Coherence

| Design decision | Followed? | Evidence |
|----------------|-----------|----------|
| Domain validation is centralized in aggregate/value-type `Validate()` methods | ✅ Yes | `internal/domain/*.go` validators are called at service boundaries before writes. |
| Transactions own all multi-row workflows and repository calls use `*sql.Tx` | ✅ Yes | `Store.WithTx`, tx-aware employee/profile/threshold/history methods, and `ApplyEstimulo` use one transaction. |
| Stimulus application uses conditional state transition and deterministic conflict result | ✅ Yes | `UPDATE ... WHERE estado='pendiente'`, `RowsAffected`, `ApplyResult`, history, and recalibration are in one transaction. |
| Security is one environment-configured operator with signed secure sessions and session-bound CSRF | ✅ Yes | `auth.go`, `middleware.go`, route wiring, cookie tests, CSRF tests, and safe error tests. |
| Templates/lifecycle are startup-validated, parse-once, cwd-independent, observable, and bounded on shutdown | ✅ Yes | `TemplateSet`, `resolveAppRoot`, `run`, health mux, `slog`, and lifecycle tests. |
| Delivery remains CGO-free and single-instance | ✅ Yes | `modernc.org/sqlite`, pinned Docker stages, one Compose service/volume, and documented non-goals. |

### Strict TDD Compliance

| Check | Result | Details |
|-------|--------|---------|
| TDD evidence reported | ✅ | Cumulative `TDD Cycle Evidence` tables exist in `apply-progress.md` for slices 1–5. |
| All executable tasks have tests | ✅ | 18/18 executable tasks have test evidence; documentation tasks 5.1–5.2 are explicitly N/A by the apply contract. |
| RED confirmed (test files exist) | ✅ | Test files referenced by all executable work units exist in the repository; documentation rows correctly use N/A. |
| GREEN confirmed (tests pass) | ✅ | Current `go test ./... -count=1 -timeout 120s` passed all six packages. |
| Triangulation adequate | ✅ | Executable task evidence records multiple cases or explicit acceptance sets; documentation rows are N/A. |
| Safety net for modified files | ✅ | Apply progress records baseline/full-suite safety nets for executable and documentation slices. |

**TDD Compliance**: 6/6 applicable checks passed; the two documentation-only tasks correctly remain N/A for executable RED/GREEN evidence.

### Test Layer Distribution

The source inventory contains 80 top-level test functions across 16 test-bearing files, plus one helper-only test file. Counts below are top-level functions; table-driven subtests are additional cases.

| Layer | Tests | Files | Tools |
|-------|-------|-------|-------|
| Unit | 12 | 3 | Go stdlib `testing` |
| Integration | 62 | 11 | Real SQLite, `httptest`, `html/template`, Go stdlib `testing` |
| E2E | 0 | 0 | Not available and explicitly out of scope |
| Unknown/structural | 6 | 2 | Go stdlib file/content checks for Docker, Compose, and CI |
| **Total** | **80** | **16** | |

No test uses an unavailable integration or E2E tool. The structural and template tests are intentionally not presented as browser E2E coverage.

### Changed-File Coverage

Coverage analysis was not run because the user-authorized final proof was restricted to the exact four commands recorded above. No current changed-file line/branch percentages or uncovered-line ranges are claimed. The cached package coverage in `openspec/config.yaml` is historical and informational only.

### Assertion Quality

**Assertion quality**: ✅ No tautologies, orphan empty-only assertions, ghost loops, assertions without production/static behavior calls, smoke-only render assertions, or mock-heavy tests were found. Fixed non-empty fixtures and explicit row/count/status/content checks exercise real behavior. The parse-only nudge test is intentional documentation of a known pre-existing template defect, not a behavioral success claim.

### Quality Metrics

| Metric | Result |
|--------|--------|
| Linter / static analysis | ✅ `go vet ./...` exit 0; no output |
| Type checker / compiler | ✅ `go build ./...` exit 0; no output |
| Formatter | ✅ `gofmt -l .` exit 0; empty output |

### Documentation and Path Checks

| Check | Result | Evidence |
|-------|--------|----------|
| Repository root | ✅ | `git rev-parse --show-toplevel` → `/home/ghandy/Documents/Estimulos-e-incentivos` |
| Tracked paths containing whitespace | ✅ | `git ls-files` scan → 0 paths with whitespace |
| Active proposal/spec/design/tasks/apply paths | ✅ | All required OpenSpec files exist and were read |
| README design links | ✅ | Both linked design files exist under `doc/` |
| README verification commands | ✅ | Test, build, vet, and gofmt commands match the current proof/CI |
| README non-goals | ✅ | SSO/OIDC, multi-tenancy, multi-role RBAC, external integrations, jobs, PostgreSQL, Kubernetes, analytics/ML, and browser E2E are explicit |
| CI route and command claims | ✅ | 49 handler registrations and all four CI commands were cross-checked; CI is root-pinned and has no write automation |
| Demo seed claims | ✅ | Apply-progress structural review records 6 employees, 8 incentives, and 6 nudges matching `Service.Seed` |
| Historical path note | ⚠️ | `apply-progress.md` retains one historical reference to the pre-rename directory with spaces; it is not a tracked path or an active root claim and was not modified. |

### Runtime Evidence Status

| Evidence | Status | Treatment |
|----------|--------|-----------|
| Current Go final proof | ✅ Current | Four required commands executed once to completion; all exit 0. |
| Slice-4 native binary smoke | ✅ Previously recorded | Foreign-CWD binary, health/readiness, login, SIGTERM drain, log events, and port release are recorded in `apply-progress.md`; not rerun in this verification. |
| Slice-4 Docker Compose smoke | ✅ Previously recorded | `docker compose config`, build/up, healthy readiness, seeded volume, and cleanup are recorded in `apply-progress.md`; not rerun in this verification. |
| Browser E2E | ➖ Out of scope | No browser E2E tool or scenario is required by the proposal/specs. |

### Authority Evidence

Review authority was confirmed from the authoritative preflight and the read-only native status projection:

- Review lineage: `review-d768c42f28b1ae46`.
- Review gate: `allow`.
- Binding revision: `sha256:e3baea1380e20f0ff85b3350ce763bc879ac3ce00352495086f48631f7d8e98d`.
- Native status reason: `explicit bound compact authority exactly matches the current repository`.
- The inherited runtime attempt token and work unit were used as supplied; this executor did not call `sdd-attempt acquire`, `begin`, `reset`, or `settle`, and did not start or mutate review lifecycle state.

### Deviations and Accepted Risks

| Item | Classification | Verification treatment |
|------|----------------|------------------------|
| Stateless session logout leaves a stolen signed cookie valid until TTL | Accepted risk / WARNING | Documented in README/design; outside the single-operator MVP scope. |
| `COOKIE_SECURE=false` in Compose | Accepted risk / WARNING | Explicitly limited to plain-HTTP demo; production guidance keeps `true` behind TLS. |
| Login/logout are CSRF/auth allowlisted | Accepted risk / WARNING | Documented as low-risk because logout has no state-changing persistence and login establishes the session. |
| Nudge list/detail template defects remain | Accepted risk / WARNING | Documented as pre-existing and out of scope; toggle endpoint has JSON fallback. |
| Slice-4 gofmt normalization touched pre-existing drift | Design deviation / WARNING | Behavior-neutral and required for the new CI formatting gate; recorded in apply progress. |
| Documentation-only scenarios lack runtime tests | Verification blocker / CRITICAL | Four required scenarios are `UNTESTED`; structural/manual review cannot satisfy the strict runtime rule. |

### Issues Found

**CRITICAL**:

1. `portfolio-documentation / Complete portfolio contract / New operator can deploy safely` has no covering test that passes at runtime.
2. `portfolio-documentation / Complete portfolio contract / Threat model states boundaries` has no covering test that passes at runtime.
3. `portfolio-documentation / Correct portfolio UX / Demo matches implementation` has no covering test that passes at runtime.
4. `portfolio-documentation / Correct portfolio UX / Non-goals remain explicit` has no covering test that passes at runtime.

**WARNING**:

- The four critical documentation gaps are inherent to the docs-only slice and were explicitly recorded as structural/manual N/A evidence; they still block strict full-scenario verification.
- Stateless sessions, the plain-HTTP Compose cookie setting, login/logout CSRF allowlisting, and known nudge template defects remain accepted and documented risks.
- Historical slice-4 binary/Docker evidence was consumed from `apply-progress.md` and was not independently rerun here.

**SUGGESTION**:

- Add repository-level documentation contract tests or an explicitly configured manual-verification policy if these four documentation scenarios must become archive-ready under Strict TDD.
- Run changed-file coverage in a separately authorized verification pass if coverage percentages are required for release reporting.

### Verdict

**FAIL**

The implementation and all 20 tasks are complete, and every executable scenario has passing current runtime evidence. Verification is not archive-ready because four required documentation scenarios have no runtime covering tests under the active Strict TDD contract.
