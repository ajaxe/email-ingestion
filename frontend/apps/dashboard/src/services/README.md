# Services Directory (`src/services`)

This directory contains service-level modules and infrastructure handlers for the frontend SPA:

- `apiService.js` / `api.js`: HTTP client module, base URL configuration (`/app/v1`), header interceptors, and error handling.
- `authService.js`: Authentication service handling token persistence, credentials verification, and auth state helpers.
- `oidcService.js`: `oidc-client-ts` integration wrapper managing PKCE authorization flows, silent refresh, and user sessions.
