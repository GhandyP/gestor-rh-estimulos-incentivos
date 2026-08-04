# Delta for portfolio-documentation

## ADDED Requirements

### Requirement: Complete portfolio contract

README and design documentation MUST describe the layered architecture (domain → engine → service → handler → store), security boundary and threat model, SQLite persistence and backup expectations, demo/seed behavior, deployment steps, verification commands, and explicit limitations/non-goals.

#### Scenario: New operator can deploy safely
- GIVEN a reader follows the documented local or container path
- WHEN they configure and start the application
- THEN they can identify persistence, health, security assumptions, and verification steps

#### Scenario: Threat model states boundaries
- GIVEN a reader reviews the security documentation
- WHEN they assess deployment suitability
- THEN single-operator scope, trusted-host assumptions, protected mutations, and excluded enterprise features are explicit

### Requirement: Correct portfolio UX

Documentation and employee-list presentation MUST describe and demonstrate the tested behavior without claiming unsupported functionality, and MUST preserve the stated non-goals: no SSO/OIDC, multi-tenancy, multi-role RBAC, external integrations, jobs, PostgreSQL, Kubernetes, ML, or browser E2E.

#### Scenario: Demo matches implementation
- GIVEN the documented demo data and employee list are used
- WHEN the portfolio walkthrough is followed
- THEN displayed actions and outcomes match the tested application behavior

#### Scenario: Non-goals remain explicit
- GIVEN a reader searches the feature scope
- WHEN they review the README or design docs
- THEN all excluded capabilities are clearly identified as separate future work
