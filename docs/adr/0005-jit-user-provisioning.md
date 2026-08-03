# ADR 0005: Just-in-Time User Provisioning via Email Correlation

## Status

Accepted

## Context

Dashboard users must be authorized to access one or more applications before they can use the dashboard API. A formal invite/accept flow (sending invite emails, tracking invite tokens, expiry logic) adds significant surface area for a feature that is not yet a priority.

The `user_application_access` table (see ADR-0003) holds the user-to-application mapping. Before a user's first login, their IdP `sub` identifier is unknown — the IdP does not embed application context in the JWT, and there is no pre-login handshake between the service and the IdP.

The IdP JWT does carry standard OIDC claims including `email` (with an `email_verified` flag), which is known at the time of manual pre-provisioning.

## Decision

We will use a **just-in-time (JIT) provisioning** pattern where a user record is pre-provisioned with only their email address, and the IdP `sub` is populated on first successful login.

**Pre-provisioning (admin action):**
An operator manually inserts a partial record into `user_application_access` specifying the target `application_id`, the user's `email`, and their `role`. The `idp_user_sub` field is left `NULL` and the `status` is set to `'pending'`.

**First login flow (two-phase lookup in `UserAuth` middleware):**

1. Validate the bearer JWT via JWKS. Reject if invalid.
2. Reject if `email_verified` claim is `false`. Unverified emails must never be used as a correlation key.
3. Attempt lookup by `(application_id, idp_user_sub)` → **fast path**. If found and `status = 'active'`, authorize and proceed.
4. On miss, attempt lookup by `(application_id, email)` where `status = 'pending'` → **activation path**.
5. If found: atomically `UPDATE` the record setting `idp_user_sub`, `status = 'active'`, and `activated_at`. Use `WHERE idp_user_sub IS NULL` as a guard against concurrent first-login races.
6. If not found at either step: return `403 Forbidden`.

**All subsequent logins:** Only the fast path (step 3) is exercised. The slow activation path is a one-time operation per user.

**The required sqlc queries are:**

```sql
-- name: GetUserAccessBySub :one
SELECT * FROM user_application_access
WHERE application_id = $1 AND idp_user_sub = $2 AND status = 'active';

-- name: GetPendingUserAccessByEmail :one
SELECT * FROM user_application_access
WHERE application_id = $1 AND email = $2 AND status = 'pending';

-- name: ActivateUserAccess :one
UPDATE user_application_access
SET idp_user_sub = $2, status = 'active', activated_at = $3, last_login_at = $3
WHERE id = $1 AND idp_user_sub IS NULL
RETURNING *;

-- name: UpdateLastLogin :exec
UPDATE user_application_access SET last_login_at = now() WHERE id = $1;
```

## Consequences

### Positive

- **No invite flow required:** Onboarding a user requires a single database insert. No API endpoint, no email sending infrastructure, no token expiry management.
- **Self-healing on email change (post-activation):** Once `idp_user_sub` is populated, all subsequent lookups use `sub`. A user's email changing in the IdP does not break their access.
- **Race-safe activation:** The `WHERE idp_user_sub IS NULL` guard on the `UPDATE` ensures that concurrent first-login requests do not corrupt the record. The second concurrent request will fall through to the now-active fast path.
- **Auditable lifecycle:** `created_at`, `activated_at`, and `last_login_at` timestamps provide a full provisioning history without additional infrastructure.

### Negative

- **Email as a trusted correlation key:** The security of first-login activation depends entirely on the IdP correctly setting `email_verified = true` only for verified addresses. This must be validated strictly; accepting unverified emails would allow an attacker to claim any pre-provisioned identity.
- **No self-service onboarding:** Users cannot register themselves. An operator must manually insert the pre-provisioning record. This is acceptable for the current operational model but will require a proper admin API if self-service is introduced later.
- **Stale email in DB post-activation:** If a user's email changes in the IdP after activation, the `email` column in `user_application_access` will be stale. This does not affect authorization (which uses `sub`), but may cause confusion in admin tooling. The `last_login_at` update can optionally refresh the stored email claim to mitigate this.
