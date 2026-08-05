# ADR 0007: Gateway-Managed API Keys for M2M Authentication

## Status

Accepted (Supersedes ADR-0003 for M2M Identity Resolution)

## Context

ADR-0002 established a Dual Principal Model separating tenant SaaS application callers (M2M) from human operators (Dashboard Users). Under ADR-0003, M2M authentication relied on external IdP-issued OAuth2 JWTs, using a local `application_identities` table to map the token's `client_id` claim to an `application_id`.

While ADR-0003 avoided custom IdP claims, M2M authentication remained coupled to the central IdP in several ways:
1. **Administrative Coupling:** Onboarding a new tenant application required registering client credentials in the external IdP before mapping the `client_id` in the Gateway database.
2. **Runtime Coupling:** M2M callers had to perform an OAuth2 `client_credentials` grant handshake with the external IdP to obtain a JWT before calling the Gateway.
3. **Integration Overhead:** Third-party SaaS callers had to implement token caching and refresh logic.

Replacing `client_id` mapping with Gateway-managed API keys moves M2M authentication entirely inside the service boundary, completing the decoupling from the external IdP for M2M traffic.

## Decision

We will replace the `application_identities` table and IdP-based JWT authentication for M2M callers with a Gateway-managed **API Key system**.

- **M2M Authentication:** Authenticated via `Authorization: Bearer <api_key>` or `X-API-Key: <api_key>` headers. The `M2MAuth` middleware resolves the key to an `application_id` via a new `api_keys` registry table.
- **Human Authentication:** Remains unchanged. Dashboard users continue to authenticate via external IdP JWTs resolved through the `user_application_access` table (defined in ADR-0003 and ADR-0004).

### Recommended Implementation Standard

#### 1. Database Schema (`api_keys`)

```sql
CREATE TABLE api_keys (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    key_prefix     TEXT NOT NULL,
    key_hash       TEXT NOT NULL UNIQUE,
    scopes         TEXT[] NOT NULL DEFAULT '{}',
    expires_at     TIMESTAMPTZ,
    last_used_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_app_id ON api_keys(application_id);
```

#### 2. Key Format & Secret Scanning Security
- API keys will follow a structured format with an environment-aware prefix: `eig_live_<random_bytes>` or `eig_test_<random_bytes>`.
- The `key_prefix` (first 12 characters, e.g., `eig_live_a1b2`) is stored unhashed for display/identification in the management UI.
- The full key is hashed using **SHA-256** (`key_hash`) before persistence. Raw keys are returned to the user **once upon creation** and never stored.

#### 3. Middleware Resolution Flow (`M2MAuth`)
1. Extract the raw key from `Authorization: Bearer <key>` or `X-API-Key: <key>`.
2. Compute the SHA-256 hash of the incoming key.
3. Query `api_keys` via `sqlc` for a matching `key_hash`.
4. Validate that the key is active and not expired (`expires_at IS NULL OR expires_at > now()`).
5. Asynchronously update `last_used_at` timestamp to avoid blocking request latency.
6. Inject the resolved `application_id` into the request context.

#### 4. Dashboard Management Surface
Human operators manage API keys for their application via the `/dashboard/v1/...` API:
- `POST /dashboard/v1/applications/:app_id/api-keys` (Generate new key)
- `GET /dashboard/v1/applications/:app_id/api-keys` (List active key metadata and `key_prefix`)
- `DELETE /dashboard/v1/applications/:app_id/api-keys/:key_id` (Instantly revoke key)

All queries must be generated via `sqlc` per project conventions.

## Consequences

### Positive

- **Total IdP Decoupling for M2M:** Zero administrative or runtime dependency on the external IdP for tenant M2M API traffic.
- **Simplified Client Integration:** SaaS callers do not need OAuth2 token acquisition, caching, or token refresh logic.
- **Self-Service Management:** Application owners can create, label, rotate, and revoke API keys self-service via the management dashboard.
- **Leak Detection:** Structured key prefixes enable automated detection by repository secret scanners (e.g., GitHub Secret Scanning).

### Negative

- **Gateway Key Management Responsibility:** The Gateway backend assumes responsibility for secure key generation, SHA-256 hashing, and revocation tracking.
- **Static Credential Exposure:** Unlike short-lived 1-hour OAuth JWTs, API keys are persistent unless explicitly created with expiration dates or rotated by users.
