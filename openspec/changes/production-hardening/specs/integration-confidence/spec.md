# Delta for integration-confidence

## ADDED Requirements

### Requirement: Reproducible persistence and HTTP confidence

The test suite MUST use real SQLite databases to verify migrations, foreign keys, rollback, constraints, concurrent/repeated application, and atomic workflows. HTTP tests MUST use `httptest` to cover authentication, authorization, CSRF, success, and error routes without browser E2E.

#### Scenario: SQLite failures prove rollback
- GIVEN a transaction fails after an earlier write
- WHEN the integration test reopens or queries the database
- THEN no partial state remains

#### Scenario: HTTP boundary rejects unsafe requests
- GIVEN an unauthenticated or invalid-CSRF `httptest` request
- WHEN it invokes a mutation route
- THEN the response is rejected and persistence is unchanged

### Requirement: Deterministic rendering regression coverage

Tests MUST parse and render templates deterministically and MUST verify the employee-list output contains valid, non-duplicated mutation markup and preserves key happy and error paths.

#### Scenario: Template rendering is stable
- GIVEN the same fixture data and template input
- WHEN the template is rendered repeatedly
- THEN the output is parseable and equivalent

#### Scenario: Employee list has one action
- GIVEN an employee list with deletable records
- WHEN it is rendered
- THEN each record has exactly one valid delete action and no malformed cell structure
