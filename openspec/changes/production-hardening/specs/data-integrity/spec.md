# Delta for data-integrity

## ADDED Requirements

### Requirement: Validated and atomic domain workflows

The system MUST reject invalid required fields, ranges, enum values, and cross-field combinations before persistence. Employee creation MUST atomically persist employee, PerfilMAP, and Umbral; stimulus application MUST atomically persist application, history, and recalibration.

#### Scenario: Invalid input is rejected
- GIVEN a request contains an invalid domain value
- WHEN the service validates it
- THEN it returns a validation error and persists no partial record

#### Scenario: Intermediate failure rolls back
- GIVEN one operation in either workflow fails
- WHEN the workflow completes
- THEN all writes in that workflow are rolled back

### Requirement: Constraint-backed idempotency

The system MUST enforce foreign keys, applicable uniqueness/check constraints, and indexes through forward migrations. Applying one stimulus more than once, including concurrently, MUST produce one successful state transition, one history entry, and one recalibration.

#### Scenario: Repeat application is harmless
- GIVEN a stimulus is already applied
- WHEN it is applied again
- THEN no duplicate history or recalibration is created

#### Scenario: Concurrent application has one winner
- GIVEN concurrent requests target the same pending stimulus
- WHEN both attempt application
- THEN exactly one commits and the other reports the defined conflict/idempotent result
