# Phase 7.1 Architecture & Implementation Analysis: Admin Password Auth with Future OIDC Migration

## 1. Executive Summary

This document provides a detailed architectural analysis and implementation strategy for **Phase 7.1 (Management Dashboard Authentication)**. 

To enable rapid local development and immediate usability without requiring an active external Identity Provider (IdP), the system will initially support a **Static Admin Password Authentication** mechanism configured in `backend/config.yaml`. 

The design is explicitly structured around an **Abstracted Identity & Authentication Subsystem** so that migrating to OIDC (e.g., `Apogee-dev` IdP) in a later phase requires **zero breaking changes** to the frontend route guards, HTTP client interceptors, or backend domain logic.

---

## 2. Alignment with Architecture Decision Records (ADRs)

| ADR | Principle | Alignment Strategy |
| :--- | :--- | :--- |
| **ADR-0002** | Dual Principal Model | Dashboard authentication operates strictly under the `/app/v1/...` API group via `UserAuth` middleware. M2M API (`/api/v1/...`) remains entirely isolated and uses `M2MAuth`. |
| **ADR-0003** | Local Identity Registry | Admin identity maps to a system user record in `users` (`idp_user_sub = 'static:admin'`). Permissions and application access are resolved via `user_application_access`. |
| **ADR-0004** | `application_id` Placement | Dashboard routes enforce explicit `:app_id` scoping in URLs (`/app/v1/applications/:app_id/...`), with `UserAuth` verifying access rights. |
| **ADR-0005** | JIT User Provisioning | On password login (or startup), the system ensures the admin identity is linked to the primary tenant application, mirroring the JIT activation flow used for OIDC logins. |

---

## 3. Architecture & Technical Design

```
+-----------------------------------------------------------------------------------+
|                                 FRONTEND (Vue.js SPA)                             |
|                                                                                   |
|  [ Login Screen ] --(1) POST /app/v1/auth/login--> [ Auth Store / Axios Module ]  |
|                                                            |                      |
|                                                            v                      |
|                                                 Authorization: Bearer <JWT>       |
+------------------------------------------------------------|----------------------+
                                                             |
                                                             v
+-----------------------------------------------------------------------------------+
|                                 BACKEND (Go / Echo)                               |
|                                                                                   |
|  /app/v1/auth/login                /app/v1/applications/:app_id/...              |
|          |                                         |                              |
|          v                                         v                              |
|  [ Login Handler ]                         [ UserAuth Middleware ]                |
|  Verify Password against Config                     |                             |
|  Generate Signed Dashboard JWT                      v                             |
|                                            [ Token Verifier Interface ]           |
|                                            /                          \           |
|                     (provider: "password") /                            \ (provider: "oidc")
|                                           v                              v        |
|                              PasswordTokenVerifier              OIDCTokenVerifier |
|                              (HMAC-SHA256 Local Key)            (JWKS Public Key) |
+-----------------------------------------------------------------------------------+
```

---

## 4. Backend Implementation Plan

### 4.1 Configuration Schema Extension

Add an `auth` section to `backend/pkg/config/config.go` and `backend/config.example.yaml`:

```yaml
auth:
  provider: "password" # Options: "password" | "oidc"
  admin:
    email: "admin@example.com"
    password: "admin-password-change-me" # Or env EM_AUTH_ADMIN_PASSWORD
  jwt_secret: "dev-dashboard-jwt-secret-key-change-in-prod" # Or env EM_AUTH_JWT_SECRET
  token_ttl_hours: 24
  oidc:
    issuer: "https://auth.apogee.dev"
    client_id: "email-ingestion-dashboard"
    jwks_uri: "https://auth.apogee.dev/.well-known/jwks.json"
```

### 4.2 Standardized Dashboard JWT Claims

Regardless of authentication provider (`password` or `oidc`), tokens passed to `/app/v1/...` will share a unified claims layout:

```go
type UserClaims struct {
    jwt.RegisteredClaims
    UserID   string   `json:"user_id"`
    Email    string   `json:"email"`
    Sub      string   `json:"sub"`      // 'static:admin' for password auth, IdP sub for OIDC
    Provider string   `json:"provider"` // 'password' or 'oidc'
    Roles    []string `json:"roles"`
}
```

### 4.3 Pluggable `TokenVerifier` Interface

Define a clean interface for token validation in `backend/internal/service/auth.go`:

```go
type TokenVerifier interface {
    VerifyToken(ctx context.Context, tokenString string) (*UserClaims, error)
}
```

1. **`PasswordTokenVerifier`**: Validates tokens signed via HMAC-SHA256 (`jwt_secret`).
2. **`OIDCTokenVerifier`**: Validates RSA/ECDSA tokens using cached JWKS key sets fetched from `oidc.jwks_uri`.

A factory function initializes the active verifier based on `cfg.Auth.Provider`.

### 4.4 Endpoints & Handlers

Register authentication endpoints under `/app/v1/auth`:

* **`POST /app/v1/auth/login`**:
  * Request Body: `{ "email": "admin@example.com", "password": "..." }`
  * Action: Compare credentials against `cfg.Auth.Admin`. Ensure admin record exists in DB `users`. Generate JWT.
  * Response: `{ "token": "<JWT>", "user": { "id": "...", "email": "...", "sub": "static:admin" } }`
* **`GET /app/v1/auth/me`**:
  * Returns authenticated user details and list of accessible `application_id` records.
* **`POST /app/v1/auth/logout`**:
  * Returns `200 OK` (client clears token from local storage).

---

## 5. Frontend Implementation Plan (`frontend/apps/dashboard`)

### 5.1 Auth Pinia Store (`src/stores/auth.js`)

Manages login state, token persistence, and active application context:

* State: `token`, `user`, `authProvider` ("password" | "oidc"), `isAuthenticated`.
* Actions:
  * `login(email, password)`: Calls `POST /app/v1/auth/login`, stores token in `localStorage`, updates state.
  * `fetchUser()`: Calls `GET /app/v1/auth/me` on app boot to restore session.
  * `logout()`: Clears storage and resets store state.

### 5.2 Axios Interceptor Module (`src/plugins/axios.js`)

* **Request Interceptor**: Reads `authStore.token` and attaches header:  
  `Authorization: Bearer <token>`
* **Response Interceptor**: Catches `401 Unauthorized`. Clears token and redirects user to `/login`.

### 5.3 Login Page (`src/pages/login.vue`)

* Clean, responsive Vuetify 3 form with Email & Password input fields.
* Form validation and feedback alerts for invalid credentials.
* Automatic redirect to `/` (Overview Dashboard) upon successful authentication.

### 5.4 Vue Router Navigation Guard

In `src/router/guard.js` (or integrated in router setup):

```javascript
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  if (to.meta.requiresAuth !== false && !authStore.isAuthenticated) {
    return next({ name: '/login' })
  }
  next()
})
```

---

## 6. Migration Strategy: Upgrading to OIDC

When transitioning from Password Auth to OIDC in the future, the migration involves:

1. **Backend Config Change**:
   Set `auth.provider: "oidc"` in `config.yaml`.
2. **Backend Verifier Switch**:
   The factory instantiates `OIDCTokenVerifier` instead of `PasswordTokenVerifier`.
3. **Frontend Mode Switch**:
   * If `authProvider == "oidc"`, the frontend login action initiates an OIDC PKCE redirect flow to `Apogee-dev` IdP instead of showing the static password form.
   * On callback at `/auth/callback`, the PKCE authorization code is exchanged for an OIDC access token.
   * `/app/v1/...` API endpoints continue receiving `Authorization: Bearer <token>` headers seamlessly.

---

## 7. Pros, Cons & Security Considerations

### Benefits
* **Zero External Dependencies for Local Dev**: Developers can run and test the complete SPA and API locally without spinning up or mocking an OIDC IdP server.
* **Seamless Migration**: Upgrading to OIDC requires zero refactoring of business handlers or Vue components.
* **Strict ADR Compliance**: Keeps M2M API (`/api/v1`) and Dashboard API (`/app/v1`) completely decoupled.

### Security Best Practices
* **Config Security**: Password in `config.yaml` can be overridden via `EM_AUTH_ADMIN_PASSWORD` environment variable to prevent committing secrets to source control.
* **Token TTL**: Dashboard JWTs enforce short lifetimes (e.g. 24 hours max).
* **Environment Scoping**: In `production` environment mode, backend should log a strong warning if static password auth is enabled instead of OIDC.

---

## 8. Summary of Action Items for Phase 7.1 Implementation

1. **Backend**:
   * [ ] Update `backend/pkg/config/config.go` & `backend/config.example.yaml` with `auth` config block.
   * [ ] Implement `TokenVerifier` interface (`PasswordTokenVerifier` & `OIDCTokenVerifier`).
   * [ ] Implement `/app/v1/auth/login` and `/app/v1/auth/me` handlers.
   * [ ] Attach `UserAuth` middleware using `TokenVerifier` to `/app/v1/...` routes in `router.go`.
2. **Frontend**:
   * [ ] Implement `src/stores/auth.js` Pinia store.
   * [ ] Configure Axios HTTP client with Bearer token injection and 401 response interceptor.
   * [ ] Build `src/pages/login.vue` with email & password authentication UI.
   * [ ] Setup router navigation guards for protected routes.
