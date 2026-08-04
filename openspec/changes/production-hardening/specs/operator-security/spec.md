# Delta for operator-security

## ADDED Requirements

### Requirement: Single-operator security boundary

The system MUST authenticate exactly one configured trusted operator, authorize protected routes only for that operator, use secure session cookies, and avoid exposing sensitive request or error data. It MUST NOT imply SSO/OIDC, multi-tenancy, or multi-role RBAC.

#### Scenario: Unauthenticated mutation is rejected
- GIVEN a browser request lacks valid operator authentication
- WHEN it targets a protected mutation
- THEN the system rejects it without changing state

#### Scenario: Authenticated operator is allowed
- GIVEN the configured operator has a valid session
- WHEN it requests an authorized route
- THEN the request proceeds and the session cookie uses the configured secure protections

### Requirement: CSRF-protected browser mutations

The system MUST require a valid CSRF token for every state-changing browser request and MUST reject missing, malformed, or mismatched tokens without persistence.

#### Scenario: Valid token permits mutation
- GIVEN an authenticated operator has a matching CSRF token
- WHEN the operator submits a browser mutation
- THEN the mutation is processed

#### Scenario: Invalid token is rejected
- GIVEN an authenticated operator submits a missing or invalid token
- WHEN the browser mutation is received
- THEN the system returns a client error and changes no data
