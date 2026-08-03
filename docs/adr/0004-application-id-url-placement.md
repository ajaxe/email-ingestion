# ADR 0004: application_id Placement in API URLs

## Status

Accepted

## Context

All tenant data is partitioned by `application_id`. Every API handler requires this value to scope its database queries and S3 operations correctly. The question is where this value should come from in each request.

The current implementation (Phase 5.1, pre-auth) places `:app_id` as a URL path parameter on all tenant routes:

```
GET  /api/v1/applications/:app_id/emails
POST /api/v1/applications/:app_id/addresses
PUT  /api/v1/applications/:app_id/webhook
```

This was appropriate during early development when no authentication middleware existed. However, as authentication is introduced (Phase 6.1), the correct source of `application_id` differs between the two principal types established in ADR-0002.

Three options were considered:

**Option A — URL param for all routes:** Keep `:app_id` everywhere. Simple, RESTfully hierarchical. But for M2M callers, it is redundant with the token identity and creates BOLA/IDOR risk (OWASP API Top 10 #1) if not reconciled with the JWT.

**Option B — Context-derived for all routes:** Remove `:app_id` entirely, always derive from JWT via the local registry. Clean for M2M, but breaks for dashboard users who may need to select between multiple applications in a single session.

**Option C — Asymmetric by surface:** Remove `:app_id` from M2M routes (derive from token context); retain `:app_id` on dashboard routes (explicit user scope selection). Each surface gets the semantics appropriate to its actor type.

## Decision

We will apply **Option C**: asymmetric `application_id` sourcing by API surface.

**M2M tenant API (`/api/v1/...`):**
- No `:app_id` in the URL.
- `M2MAuth` middleware resolves `client_id` → `application_id` via `application_identities` and injects it into the Echo context via `c.Set("app_id", appID)`.
- Handlers retrieve it with `c.Get("app_id")`.
- A M2M client credential maps to exactly one application; there is no ambiguity.
- Routes become flat and token-scoped, analogous to Stripe's `/v1/charges` rather than `/v1/accounts/:id/charges`.

```
GET  /api/v1/emails
POST /api/v1/addresses
PUT  /api/v1/webhook
GET  /api/v1/application
```

**Dashboard API (`/dashboard/v1/...`):**
- `:app_id` is retained in the URL as an explicit scope selector.
- `UserAuth` middleware validates that the authenticated user (`sub`) has an entry in `user_application_access` for the requested `:app_id`. Requests for unauthorized applications are rejected with `403`.
- The `:app_id` is not an IDOR risk in this context because it is validated against the user's access record on every request.

```
GET  /dashboard/v1/applications
GET  /dashboard/v1/applications/:app_id/emails
PUT  /dashboard/v1/applications/:app_id/webhook
POST /dashboard/v1/applications/:app_id/addresses
```

**Migration path from Phase 5.1:**
The existing `/api/v1/applications/:app_id/...` routes are temporary, unauthenticated, and used for internal testing only. They will be replaced when Phase 6.1 auth middleware is implemented. These routes must not be exposed in any non-development environment.

## Consequences

### Positive

- **Structural IDOR elimination for M2M:** It is architecturally impossible for an M2M client to access another tenant's data by constructing a URL, because `application_id` never comes from the request.
- **Correct semantics per actor:** M2M tokens are scoped to one application implicitly; dashboard users select scope explicitly. The URL design reflects this accurately.
- **Non-breaking extensibility:** If multi-application M2M clients are ever required in the future, `/api/v1/applications/:app_id/...` routes can be added additively without changing existing clients.

### Negative

- **Asymmetry requires documentation:** Developers unfamiliar with the design may find it surprising that the two API surfaces handle `application_id` differently. This ADR serves as the canonical explanation.
- **Handler duplication risk:** Operations available on both surfaces (e.g., listing emails) must ensure both handler paths correctly source `application_id` from context vs. URL param respectively.
