# **Email Ingestion Gateway: Active Implementation Roadmap**

This document tracks active and upcoming development phases. Completed phases are archived in [task-list-archive.md](file:///C:/CodeWorkspace/projects/email-ingestion/docs/task-list-archive.md).

## **Project Status Overview**

**Current Phase:** Phase 6 (Application API & Brokered S3 Access Control)  
**Overall Progress:** 5 / 8 Phases Completed (62.5%)

| Phase | Status | Key Scope | Archive Details |
| :--- | :--- | :--- | :--- |
| **Phase 1: Local Foundations** | Completed | Bootstrap, Config & SQLC Generation | [View Details](file:///C:/CodeWorkspace/projects/email-ingestion/docs/task-list-archive.md#phase-1-local-development-foundations--data-engine) |
| **Phase 2: Perimeter SMTP & Cache** | Completed | SMTP Daemon & Address Validation | [View Details](file:///C:/CodeWorkspace/projects/email-ingestion/docs/task-list-archive.md#phase-2-in-memory-address-caching--perimeter-smtp-handshake) |
| **Phase 3: S3 Spooling & Edge Proxy** | Completed | S3 Direct Uploads, Redis Outbox & Echo Proxy | [View Details](file:///C:/CodeWorkspace/projects/email-ingestion/docs/task-list-archive.md#phase-3-stateless-s3-spooling-architecture) |
| **Phase 4: MIME Engine & Workers** | Completed | Redis Consumer Group Pool & MIME Parsing | [View Details](file:///C:/CodeWorkspace/projects/email-ingestion/docs/task-list-archive.md#phase-4-spool-queue-worker--mime-parsing-engine) |
| **Phase 5: Secure Webhook Dispatch** | Completed | SSRF Guard Handshake & Jitter Retries | [View Details](file:///C:/CodeWorkspace/projects/email-ingestion/docs/task-list-archive.md#phase-5-secure-webhook--callback-dispatch-engine) |
| **Phase 6: REST API & STS S3** | **IN PROGRESS** | OIDC JWT Auth & Brokered AWS STS Presigned URLs | *Active Below* |
| **Phase 7: Management Dashboard** | Pending | Vue.js SPA & Developer Management Console | *Active Below* |
| **Phase 8: Deployment & Infra** | Pending | Docker Multi-Stage, Traefik & MX DNS | *Active Below* |

---

## **Phase 6: Application API & Brokered S3 Access Control**

*Expose secure REST endpoints and authorize user file-access namespaces using brokered AWS STS IAM role assumption.*

### **6.1 Authentication & REST Endpoints**

* [ ] Implement JWT OIDC verification middleware utilizing a cached JWKS endpoint provider.  
* [ ] Create core routing pathways:  
  * `POST /api/v1/addresses` (Provision new assigned 10-char routing paths).  
  * `GET /api/v1/application` (Retrieve configurations and active scopes).  
  * `GET /api/v1/emails` (List history logs).  
* **Verification Checkpoint**: Query these endpoints with both valid and expired OIDC access tokens to verify signature enforcement.

### **6.2 Brokered IAM Role Assumption (S3 Downloader)**

* [ ] Implement the S3 download endpoint: `GET /api/v1/emails/{emailId}/attachments/{attachmentId}`.  
* [ ] Inside the Go handler:  
  1. Validate JWT. Resolve the active request client to their internal application identity.  
  2. Query Postgres to fetch the application's unique `aws_iam_role_arn`.  
  3. Call AWS STS `AssumeRole` using the Go AWS SDK.  
  4. Using the returned transient credentials, instantiate a scoped S3 client.  
  5. Generate a short-lived S3 Presigned URL.  
  6. Return the presigned URL to the client.  
* **Verification Checkpoint**: Authenticate as Tenant A and request a download link for an attachment. Verify the link works and matches S3 storage paths. Attempt to modify the URL path to Tenant B's folder and confirm AWS S3 rejects the request immediately.

---

## **Phase 7: Management Dashboard (Vue.js SPA)**

*Provide the developer portal interface in `frontend/apps/dashboard` to tie all system components together.*

> [!IMPORTANT]
> **Strict Requirement:** Vue 3 Router MUST use **file-based routing** via `src/pages/`. Do NOT define static route arrays manually in `src/router/index.js`.  
> **API Dual Principal Model (ADR-0002 & ADR-0004):** Dashboard SPA API requests MUST use the **`/app/v1/...`** prefix (with explicit `:app_id` scope in URL: `/app/v1/applications/:app_id/...`), authenticated via `UserAuth` OIDC Bearer tokens. Do NOT use M2M API `/api/v1/...` endpoints for dashboard operations.

### **7.1 OIDC Auth Provider & HTTP Client Subsystem**

* [ ] Implement OIDC PKCE authentication handler (`src/services/authService.js` / `src/services/oidcService.js`) utilizing `oidc-client-ts` (`UserManager`) interfacing with `Apogee-dev` IdP.  
* [ ] Configure HTTP client module (`src/services/apiService.js`) with base URL `/app/v1`, automatic `Authorization: Bearer <token>` header injection, 401 response handling, and automatic token refresh managed by `oidc-client-ts`.  
* [ ] Implement `useAuthStore` in `src/stores/auth.js` to manage user identity, OIDC tokens (via `oidc-client-ts`), active `app_id` selection, login, and logout state using `src/services/` modules.  
* **Verification Checkpoint**: Run dev server (`pnpm dev`), verify OIDC authentication handshake via `oidc-client-ts`, token storage, and Bearer token attachment to outgoing `/app/v1/...` API requests.

### **7.2 Core Pinia Stores & API Client Integration**

* [x] Implement `useAppStore` (`src/stores/application.js`) for tenant details, API keys, and global settings (`GET /app/v1/applications/:app_id`).  
* [x] Implement `useAddressStore` (`src/stores/addresses.js`) for provisioning (`POST /app/v1/applications/:app_id/addresses`) and toggling active 10-character email local-part routing addresses.  
* [x] Implement `useEmailStore` (`src/stores/emails.js`) for paginated email logs (`GET /app/v1/applications/:app_id/emails`) and fetching S3 presigned attachment URLs (`GET /app/v1/applications/:app_id/emails/{emailId}/attachments/{attachmentId}`).  
* [x] Implement `useWebhookStore` (`src/stores/webhooks.js`) for endpoint challenge verification (`PUT /app/v1/applications/:app_id/webhook`), job history logs, and manual outbox re-delivery requests (`POST /app/v1/applications/:app_id/webhook/jobs/{id}/redeliver`).  
* **Verification Checkpoint**: Test store actions against backend mock/live `/app/v1/...` API endpoints and confirm reactive state updates.

### **7.3 Dashboard Navigation Shell & Global UI Layout Specs**

* [x] **App Layout Frame (`src/Main.vue` & `src/layouts/DashboardLayout.vue`)**:
  * `v-layout`: Root container wrapper (`DashboardLayout.vue`).
  * `v-navigation-drawer`: Left-hand vertical navigation drawer (collapsible, fixed):
    * Top section: Gateway branding ("Email Ingestion Gateway" with `mdi-email-multiple` icon).
    * Navigation links (`NavigationLinks.vue`):
      * Dashboard Overview (`mdi-view-dashboard-outline` → `/`)
      * Routing Addresses (`mdi-email-multiple-outline` → `/addresses`)
      * Ingestion Logs (`mdi-table-clock` → `/emails`)
      * Webhook Console (`mdi-webhook` → `/webhooks`)
      * API Keys & Security (`mdi-key-chain` → `/settings`)
    * Bottom section (`NavigationAddOns.vue`): Dark/light theme switcher (`mdi-theme-light-dark`) and user avatar profile with logout menu (`v-menu`).
  * `v-main` + `v-container fluid`: Dynamic content workspace area:
    * Top tenant selector header (`AppSelector.vue`): `v-select` bound to `useAppStore` for switching active `app_id` with active tenant status badge (`v-chip`).
    * Router view slot (`<slot />` / `<router-view />`) wrapped inside content sheet.
  * `v-snackbar`: Global toast notification bar in layout frame for API response feedback (success/error) managed via central notification store.
* [x] **Shared Component Primitives (`src/components/`)**:
  * `StatusChip.vue`: Standardized status badge component using `v-chip` (Green for `ACTIVE`/`SUCCESS`, Red for `DEAD`/`FAILED`, Yellow for `PENDING`/`PROCESSING`, Grey for `INACTIVE`).  
  * `ConfirmDialog.vue`: Reusable `v-dialog` confirmation modal for key regeneration or deletion actions.  
  * `CodePreview.vue`: Syntax-highlighted code block component (`v-card` wrapper) for JSON bodies and headers.  
  * `StatsWidget.vue`: Metric summary card component with icon, value, label, and trend indicator.  
* **Verification Checkpoint**: Confirm sidebar navigation, dark/light theme toggle, and responsive layout across desktop and mobile viewports.

### **7.4 Vue 3 File-Based Page Routes & Detailed Screen Specifications (`src/pages/`)**

* [x] `src/pages/index.vue`: **Dashboard Overview Screen**
  * **Top Metrics Grid (`v-row`)**: 4 equal-width `v-col` cards (`StatsWidget.vue`):
    1. Total Ingested Emails (`mdi-email-outline`, stat number, +X% trend).  
    2. Active Routing Addresses (`mdi-at`, active/total count).  
    3. Webhook Delivery Success Rate (`mdi-webhook`, percentage, success badge).  
    4. Failed Outbox Jobs (`mdi-alert-circle-outline`, failure count, warning badge).  
  * **Main Content Layout (`v-row`)**:
    * Left Column (8 cols): **Recent Ingestion Stream (`v-card`)** — Compact `v-data-table` listing 5 most recent emails (`Received At`, `From`, `Subject`, `Local Part`), with a "View All Logs" link button.  
    * Right Column (4 cols): **Quick Actions & Gateway Status (`v-card`)** — Quick action buttons ("Provision Address", "Test Webhook"), active AWS IAM Role ARN badge, and system health status.  

* [x] `src/pages/addresses/index.vue`: **Address Management Panel**
  * **Header Toolbar (`v-row`)**: Screen title ("Assigned Email Addresses"), search text field (`v-text-field` `mdi-magnify`), status filter (`v-btn-toggle` All/Active/Inactive), and primary button ("Provision New Address" `v-btn` color="primary" `mdi-plus`).  
  * **Data Table (`v-card`)**: `v-data-table-server` displaying:
    * Columns: `Local Part` (font-monospace), `Email Address` (`[local_part]@ingest.domain.com` with copy button), `Description`, `Status` (`StatusChip.vue`), `Created At`, `Actions` (Active toggle `v-switch`).  
  * **Provision Modal (`v-dialog`)**: Form with Description field (`v-text-field`), info alert explaining 10-char system generation, and "Generate Address" submit button.  

* [x] `src/pages/emails/index.vue`: **Ingested Email Log Analyzer**
  * **Filter Toolbar (`v-card`)**: Search field (`From`/`Subject`), Local Part filter (`v-select`), Date picker, and Auto-refresh toggle switch (`v-switch`).  
  * **Data Table (`v-card`)**: `v-data-table-server` displaying:
    * Columns: `Received At` (formatted timestamp), `Local Part` (`StatusChip.vue`), `From`, `Subject` (truncated with tooltip), `Ref Token` (`v-chip`), `Attachments` (badge count), `Actions` (`v-btn` icon `mdi-eye` navigating to `/emails/[id]`).  

* [x] `src/pages/emails/[id].vue`: **Email Detail & Attachment Downloader**
  * **Header Bar**: Back button (`v-btn` `mdi-arrow-left`), Subject title, Message-ID chip (`v-chip` font-monospace), Received timestamp.  
  * **Split Layout (`v-row`)**:
    * Left Column (4 cols): **Metadata Card (`v-card`)** — `From` address, Envelope `To`, `Reference Token`, S3 Key Prefix path, Ingestion UUID.  
    * Right Column (8 cols): **Content & Attachments Card (`v-card`)**:
      * Tabs header (`v-tabs`): `HTML Body`, `Text Body`, `Raw JSON (contents.json)`.  
      * Tab items (`v-window-item`): Sanitized HTML body iframe, Text body block (`CodePreview.vue`), JSON contents block (`CodePreview.vue`).  
      * **Attachments Section (`v-divider` + list)**: Attachment cards displaying filename, mime type (`contentType`), byte size, and a primary button ("Download Attachment" `v-btn` `mdi-download`) that calls `GET /app/v1/applications/:app_id/emails/:id/attachments/:attachment_id` to acquire the STS presigned S3 link and open the secure download.  

* [x] `src/pages/webhooks/index.vue`: **Webhook Config & Delivery Sandbox Console**
  * **Top Card: Webhook Endpoint & Secret Settings (`v-card`)**:
    * Endpoint URL input (`v-text-field`), Max Retries slider (`v-slider` 1-10), Webhook Secret field (`whsec_...` masked with reveal/copy button).  
    * Actions: "Save Configuration" (`v-btn` color="primary"), "Test & Verify Endpoint" (`v-btn` color="secondary" `mdi-lightning-bolt` triggering `PUT /app/v1/applications/:app_id/webhook` challenge handshake).  
    * Handshake Alert (`v-alert`): Shows challenge verification result status.  
  * **Bottom Card: Webhook Delivery Outbox Sandbox (`v-card`)**:
    * Toolbar: Title ("Outbox Delivery Attempt History"), Auto-refresh toggle, Status filter chips (`All`, `PENDING`, `SUCCESS`, `FAILED`, `DEAD`).  
    * Data Table (`v-data-table-server`): Columns `Job ID`, `Attempt #`, `HTTP Status` (`StatusChip.vue`), `Duration` (`ms`), `Executed At`, `Is Retry` (boolean chip), `Actions`: "View Payload" modal button, and "Re-deliver Webhook" button (`v-btn` icon `mdi-refresh` calling `POST /app/v1/applications/:app_id/webhook/jobs/:id/redeliver`).  
    * **Payload & Response Modal (`v-dialog`)**: Side-by-side or tabbed request JSON payload and client HTTP response body inside `CodePreview.vue`.  

* [x] `src/pages/settings/index.vue`: **API Keys & Security Settings**
  * **API Key Card (`v-card`)**: Active API key (`eg_live_a1b2...` masked with reveal/copy toggle), "Regenerate API Key" button (`v-btn` color="error" opening `ConfirmDialog.vue` to call `POST /app/v1/applications/:app_id/api-keys`).  
  * **AWS S3 & IAM Security Card (`v-card`)**: Mapped AWS IAM Role ARN (`arn:aws:iam::...`), S3 Storage Bucket Prefix path (`s3://bucket/apps/{app_id}/`), STS presigned URL TTL duration (15 minutes).  

* [x] `src/pages/[...all].vue`: **404 Not Found View**
  * Centered card (`v-card`), 404 error icon (`mdi-alert-octagon-outline`), error message, and "Back to Dashboard" button (`v-btn` color="primary" `to="/"`).  

* **Verification Checkpoint**: Navigate between all generated routes, verify route parameters, dynamic paths (`/emails/:id`), and automatic typed route generation in `src/typed-router.d.ts`.

### **7.5 Operational Sandbox & Real-Time Log Inspector**

* [ ] Build interactive JSON payload inspector in `src/pages/webhooks/index.vue` and `src/pages/emails/[id].vue` to display formatted headers, body, and attempt logs.  
* [ ] Add auto-refresh / polling toggle for delivery sandbox log table to monitor live outbox worker job execution.  
* [ ] Add toast notification system (Vuetify `v-snackbar`) for success/error feedback across all user actions.  
* **Verification Checkpoint**: Perform manual webhook re-delivery from sandbox log table and verify UI updates immediately upon completion.

### **7.6 End-to-End System Integration & Verification**

* [ ] Run full system flow: Log in via OIDC PKCE -> Select/fetch app scope (`/app/v1/applications/:app_id`) -> Provision new 10-character address -> Send email via SMTP -> Verify ingestion in Email Logs -> Download attachment via STS presigned S3 link -> Verify webhook payload delivered -> Trigger manual re-delivery from Sandbox.  
* [ ] Run `pnpm lint` and `pnpm build` in `frontend/apps/dashboard` to verify zero build or linting errors.  
* **Verification Checkpoint**: Confirm clean `pnpm build` output and zero lint errors.

### **7.7 Webhook Outbox Delivery Jobs & Re-delivery API Engine**

*Implement backend persistence, service layer, Echo HTTP handlers, and routing pathways for `GET /app/v1/applications/:app_id/webhook/jobs` and `POST /app/v1/applications/:app_id/webhook/jobs/:job_id/redeliver`.*

* [x] **SQL Queries & Generation (`backend/pkg/database/public/query.public.sql`)**:
  * Implement `ListWebhookJobsByApplication`:
    ```sql
    -- name: ListWebhookJobsByApplication :many
    SELECT wj.id, wj.application_id, wj.ingested_email_id, wj.status, wj.retry_count, wj.next_delivery_at, wj.created_at,
           wl.http_status_code, wl.duration_ms, wl.attempt_number
    FROM webhook_delivery_jobs wj
    LEFT JOIN LATERAL (
      SELECT http_status_code, duration_ms, attempt_number
      FROM webhook_logs
      WHERE webhook_delivery_job_id = wj.id
      ORDER BY attempt_number DESC
      LIMIT 1
    ) wl ON TRUE
    WHERE wj.application_id = $1
      AND (sqlc.narg('status')::text IS NULL OR sqlc.narg('status')::text = '' OR wj.status = sqlc.narg('status')::text)
    ORDER BY wj.created_at DESC
    LIMIT $2 OFFSET $3;
    ```
  * Implement `GetWebhookJobByIDAndAppID`:
    ```sql
    -- name: GetWebhookJobByIDAndAppID :one
    SELECT * FROM webhook_delivery_jobs
    WHERE id = $1 AND application_id = $2
    LIMIT 1;
    ```
  * Implement `ResetWebhookJobForRedelivery`:
    ```sql
    -- name: ResetWebhookJobForRedelivery :one
    UPDATE webhook_delivery_jobs
    SET status = 'PENDING',
        retry_count = 0,
        next_delivery_at = CURRENT_TIMESTAMP
    WHERE id = $1 AND application_id = $2
    RETURNING *;
    ```
  * Implement `GetWebhookLogsByJobID`:
    ```sql
    -- name: GetWebhookLogsByJobID :many
    SELECT * FROM webhook_logs
    WHERE webhook_delivery_job_id = $1
    ORDER BY attempt_number DESC;
    ```
  * Run database code generation in `backend/`: `cd backend && sqlc generate`.

* [x] **Service Layer Enhancements (`backend/internal/service/webhook.go`)**:
  * Implement `ListJobs(ctx context.Context, appID uuid.UUID, limit, offset int32, status string)`:
    * Query `ListWebhookJobsByApplication` with parameters.
    * Return paginated job models with latest attempt telemetry.
  * Implement `RedeliverJob(ctx context.Context, appID, jobID uuid.UUID)`:
    * Validate job ownership for `appID` via `GetWebhookJobByIDAndAppID`.
    * Reset job status and delivery timestamp via `ResetWebhookJobForRedelivery`.
    * Notify active Redis outbox stream to trigger immediate background worker re-processing.

* [x] **Echo HTTP Handlers (`backend/internal/api/handler/webhook.go`)**:
  * Implement `HandleListWebhookJobs(svc *service.WebhookService) echo.HandlerFunc`:
    * Route: `GET /applications/:app_id/webhook/jobs`
    * Extract URL param `:app_id` and query string parameters `limit` (default: 50), `offset` (default: 0), `status` (optional filter).
    * Authorize tenant context via `CanAccessApplication(ctx, appID)`.
    * Return `http.StatusOK` with JSON array of webhook job records.
  * Implement `HandleRedeliverWebhookJob(svc *service.WebhookService) echo.HandlerFunc`:
    * Route: `POST /applications/:app_id/webhook/jobs/:job_id/redeliver`
    * Extract URL params `:app_id` and `:job_id`.
    * Authorize tenant context via `CanAccessApplication(ctx, appID)`.
    * Invoke `svc.RedeliverJob(ctx, appID, jobID)`.
    * Return `http.StatusOK` with response body `{"message": "Webhook delivery job re-queued successfully", "job_id": job_id, "status": "PENDING"}`.

* [x] **Router Registration (`backend/internal/api/router/router.go`)**:
  * Under `configureAppAPI` in the `/app/v1` group:
    ```go
    appGroup.GET("/applications/:app_id/webhook/jobs", handler.HandleListWebhookJobs(webhookService))
    appGroup.POST("/applications/:app_id/webhook/jobs/:job_id/redeliver", handler.HandleRedeliverWebhookJob(webhookService))
    ```

* **Verification Checkpoint**: Run `go test ./...` in `backend`, test `GET /app/v1/applications/:app_id/webhook/jobs` and `POST /app/v1/applications/:app_id/webhook/jobs/:job_id/redeliver` via authenticated API requests, and verify job status transitions to `PENDING` in PostgreSQL `webhook_delivery_jobs`.

---

## **Phase 8: Containerization & Infrastructure Deployment**

*Prepare the production-ready infrastructure stack, container routing profiles, and reverse proxy patterns.*

### **8.1 Docker Configurations & Build Stage**

* [ ] Write a multi-stage Dockerfile optimizing the Go application binary size and execution security (using a scratch or alpine base).  
* [ ] Implement the `docker-compose.yml` for production deployments, omitting Postgres container orchestration (since you utilize an external PostgreSQL instance), but including services for LocalStack/S3 and Traefik.

### **8.2 Traefik Routing & Production DNS**

* [ ] Create a Traefik routing configuration to handle automated Let's Encrypt SSL/TLS certificates and expose API ports securely.  
* [ ] Configure your system's public DNS MX records to point to the ingestion server's public IP address.  
* [ ] Implement clean TXT configurations (such as standard SPF strings, DKIM keys, and a basic `_dmarc` DMARC policy record) to prepare your domain for safe inbound validation.  
* **Verification Checkpoint**: Deploy the production stack via `docker compose up -d` and confirm that external mail clients can perform TLS handshakes and route traffic through Traefik to your Go SMTP Daemon.