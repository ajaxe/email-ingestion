# ADR 0002: Dual Principal Model — Separate API Surfaces for M2M and Human Actors

## Status

Accepted

## Context

The Email Ingestion Gateway serves two fundamentally different caller types:

1. **Tenant SaaS Applications (M2M):** A registered SaaS app calling the API programmatically to consume ingested emails, download attachments, and manage webhooks. These callers authenticate using the OAuth2 `client_credentials` grant, obtaining a JWT where the `client_id` claim identifies the application. There is no human user involved.

2. **Dashboard Users (Human):** A person operating the Vue.js management dashboard to configure addresses, view statistics, and manage their application's webhook settings. These callers authenticate via `authorization_code` / PKCE grant, obtaining a JWT where the `sub` claim identifies the individual user.

These two actor types have different JWT shapes, different trust models, different authorization requirements, and different scoping semantics. Treating them through a single API surface with shared middleware would require complex conditional logic and introduce security ambiguity.

## Decision

We will maintain two distinct API route groups, each with its own authentication middleware:

- **`/api/v1/...`** — The M2M tenant API. Authenticated by `M2MAuth` middleware, which validates the bearer JWT, extracts the `client_id`, and resolves it to an `application_id` via the `application_identities` registry table (see ADR-0003). The application identity is injected into the Echo request context.

- **`/dashboard/v1/...`** — The human dashboard API. Authenticated by `UserAuth` middleware, which validates the bearer JWT, extracts the `sub` and `email` claims, and resolves the user's authorization for the requested `application_id` via the `user_application_access` registry table (see ADR-0003 and ADR-0004).

The two groups are registered independently in `router.go` and share no middleware chain.

## Consequences

### Positive

- **Clear separation of concerns:** Each API surface evolves independently. M2M API changes do not affect dashboard clients, and vice versa.
- **Security clarity:** Each middleware enforces the appropriate trust model for its actor type without conditional branching.
- **Independent versioning:** The M2M API and dashboard API can be versioned and deprecated separately.
- **Least privilege by surface:** M2M clients cannot access dashboard routes, and dashboard users cannot call M2M-only operations, enforced by route group boundaries.

### Negative

- **Duplicated handler logic:** Some operations (e.g., listing emails) may appear in both surfaces, requiring parallel handler implementations or a shared service layer that both call into.
- **Two middleware implementations to maintain:** Both `M2MAuth` and `UserAuth` must stay in sync with JWKS key rotation and any upstream IdP changes.
