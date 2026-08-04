# Delta for operable-delivery

## ADDED Requirements

### Requirement: Observable and safe lifecycle

The server MUST validate startup configuration and template availability, emit structured `slog` startup/request/error events, expose meaningful health and readiness behavior, and gracefully shut down on termination signals within a bounded interval.

#### Scenario: Health reflects readiness
- GIVEN the server has started with a usable database and templates
- WHEN health/readiness endpoints are requested
- THEN they report healthy/readiness status

#### Scenario: Shutdown drains safely
- GIVEN the server receives a termination signal while serving requests
- WHEN shutdown runs
- THEN it stops accepting work, closes resources, and completes within the configured bound

### Requirement: Reproducible delivery artifacts

Templates MUST load deterministically independent of the caller working directory. Local/container delivery and CI MUST reproduce `go test ./...`, `go build ./...`, `go vet ./...`, and `gofmt -l .`; Docker and Compose MAY provide the single-instance demonstration path.

#### Scenario: Clean checkout passes checks
- GIVEN a clean checkout with the documented toolchain
- WHEN the four verification commands run locally or in CI
- THEN their results are reproducible and failures are actionable

#### Scenario: Missing startup asset fails safely
- GIVEN a required configuration or template is unavailable
- WHEN startup validates dependencies
- THEN startup fails clearly before serving traffic
