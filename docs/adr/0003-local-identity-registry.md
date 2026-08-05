# ADR 0003: Local Identity Registry Over Custom IdP Claims

## Status

Accepted (Partially superseded by [ADR-0007](./0007-api-key-m2m-authentication.md) for M2M identity resolution)

## Context

The system is multi-tenant. Every API request must be associated with an `application_id` to enforce data partitioning and S3 IAM role isolation. There are two actor types (see ADR-0002), each requiring a different resolution path:

- **M2M actors:** The external IdP issues standard OIDC JWTs containing a `client_id` claim. There is no mapping between `client_id` and `application_id` carried in the token itself.
- **Human actors:** The IdP issues JWTs containing a `sub` claim. There is no mapping between the user and any application embedded in the token.

Two approaches to solving this were considered:

**Option A — Custom IdP claims:** Enrich the JWT at the IdP level with a custom claim (e.g., `application_id` or `tenant_id`). The API service reads the claim directly from the token.

**Option B — Local identity registry:** Maintain mapping tables in the service's own database. The API service resolves `client_id` → `application_id` and `sub` → application access at request time using these tables.

Option A is explicitly ruled out by the project's hard constraints: *"DO NOT store or manage custom application/tenant claims in the IdP. The IdP should remain decoupled from service logic."* Additionally, encoding tenant identity into the IdP would couple every tenant lifecycle operation (onboarding, offboarding, role changes) to IdP administration.

## Decision

We will maintain two registry tables in the service database to resolve IdP identities to application context:

**`application_identities`** — Maps M2M OAuth clients to their application:

```sql
CREATE TABLE application_identities (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    idp_client_id  TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (application_id, idp_client_id)
);
```

**`user_application_access`** — Maps human users to the application(s) they may access, with a role:

```sql
CREATE TABLE user_application_access (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    email          TEXT NOT NULL,
    idp_user_sub   TEXT,
    role           TEXT NOT NULL DEFAULT 'viewer',
    status         TEXT NOT NULL DEFAULT 'pending',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    activated_at   TIMESTAMPTZ,
    last_login_at  TIMESTAMPTZ,
    UNIQUE (application_id, email),
    UNIQUE (application_id, idp_user_sub)
);
```

The `M2MAuth` middleware resolves `client_id` → `application_id` via `application_identities`. The `UserAuth` middleware resolves `sub` → access record via `user_application_access` (see ADR-0004 for the first-login activation flow).

All database access to these tables must go through `sqlc`-generated queries per project conventions.

## Consequences

### Positive

- **IdP decoupling:** The IdP has no knowledge of application structure, tenant lifecycle, or roles. These concerns remain entirely within the service boundary.
- **Operational independence:** Onboarding, offboarding, and role changes are database operations, not IdP configuration changes.
- **Auditability:** Registry tables can carry `created_at`, `activated_at`, and `last_login_at` metadata that would not be available if relying solely on token claims.
- **Extensibility:** `application_identities` supports multiple IdP clients per application (e.g., separate credentials per environment) without schema changes.

### Negative

- **Additional DB lookup per request:** Every authenticated request incurs one extra query to resolve the identity. This should be mitigated with connection pooling and short-lived in-process caching if it becomes a bottleneck.
- **Registry must be kept consistent:** If an application is deleted, its identity records must be cleaned up. `ON DELETE CASCADE` on the foreign keys handles this automatically.
