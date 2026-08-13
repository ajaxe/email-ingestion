# Email Ingestion Gateway: Completed Tasks Archive

This document archives completed implementation phases and verification checkpoints for the Email Ingestion Gateway project. For active and pending roadmap tasks, refer to [task-list.md](file:///C:/CodeWorkspace/projects/email-ingestion/docs/task-list.md).

---

## Phase 1: Local Development Foundations & Data Engine

*Before building active server daemons, establish the local environment boundaries, schemas, and typed data wrappers, leveraging your existing running PostgreSQL instance.*

### 1.1 Local Project Bootstrap

* [x] Initialize the project repository with a clean Go module structure (`go mod init`).  
* [x] ~~Create a local environment configuration file (.env or .env.local) to define development parameters~~:  
  * DB_DSN (pointing to your existing running PostgreSQL instance).  
  * S3_BUCKET (configured for local folder-spool testing or a dedicated development AWS bucket).  
* [x] ~~Implement a system configuration parser in Go (e.g., using cleanenv or standard library os.LookupEnv) to validate these database connections on application startup~~.  
* [x] Use viper lib to take path to _config.yaml_ file with defaults while accepting environment overrides with prefix `EM_`.

* **Verification Checkpoint**: Run a simple main.go ping script that parses your .env file and successfully establishes a database connection pool (sql.DB) to your existing PostgreSQL instance.

### 1.2 DB Engine & SQLC Generation

* [x] Write the PostgreSQL schema in schema.sql outlining all relational tables, indexes, and custom enum types (spool_status, webhook_status).  
* [x] Draft SQL query patterns in query.sql for all state modifications (such as transactional outbox locks, polling queries, address allocations).  
* [x] Setup and run sqlc generate to generate safe, strongly-typed Go query files.  
* **Verification Checkpoint**: Apply the schemas to your running PostgreSQL instance and verify that the SQLC compiled Go files match database columns and types flawlessly.

## Phase 2: In-Memory Address Caching & Perimeter SMTP Handshake

*Accept inbound TCP packets and block unauthorized messages at the perimeter before storing any payload files on disk.*

### 2.1 SMTP Daemon Integration

* [x] Import github.com/emersion/go-smtp and instantiate a basic SMTP server listening on local port 2525.  
* [x] Set up the TLS handshake configurations and map standard debug logging formats.  
* [x] Implement Go SMTP Backend and Session interfaces to intercept standard hooks (Mail, Rcpt, Data).  
* **Verification Checkpoint**: Use telnet localhost 2525 or netcat to verify the server negotiates a connection and responds to standard EHLO greetings.

### 2.2 Inbound Address Validation Layer (RAM Cache + DB Index)

* [x] Implement a lightweight in-memory cache layer (using go-cache, ristretto, or simple thread-safe Redis clients).  
* [x] Inside the Rcpt() connection hook:  
  1. Strip sub-address parameters (e.g., extracting a8f3g9j2k1 from a8f3g9j2k1+token123@domain.com).  
  2. Query the fast in-memory cache first.  
  3. Fall back to a targeted PostgreSQL query using the SQLC generated index search.  
  4. Cache the result aside and return 550 User Unknown to the client if the email local-part is unassigned.  
* [x] Implement envelope-level SPF check inside the Mail() connection hook utilizing ~~github.com/emersion/go-msgauth/spf~~ github.com/mileusna/spf based on the sender's connecting TCP IP.  
* **Verification Checkpoint**: Send test handshakes with both registered and fake email addresses. Validate that fake addresses are immediately dropped with a 550 SMTP error code during the socket connection.

## Phase 3: Stateless S3 Spooling Architecture

*Once an email is accepted at the socket level, stream it directly to S3 via multipart uploads to guarantee data durability and preserve stateless Gateway nodes.*

### 3.1 Spool Queue Database Schema

* [x] Define the `spool_status` ENUM (e.g., `'PENDING'`, `'PROCESSING'`, `'FAILED'`) in your PostgreSQL schema (`public.sql`).
* [x] Create the `inbound_spool_queue` table with the following columns:
  * `id` (UUID, Primary Key)
  * `s3_object_key` (VARCHAR, the path to the raw `.eml` in S3)
  * `status` (spool_status ENUM, default `'PENDING'`)
  * `attempt_count` (INTEGER, default 0, to track parsing retry attempts)
  * `last_error_message` (TEXT, to store worker MIME parsing errors)
  * `created_at` (TIMESTAMPTZ, default NOW())
  * `updated_at` (TIMESTAMPTZ, default NOW())
* [x] Generate the corresponding Go models using `sqlc generate`.

### 3.2 Direct S3 Stream Archiving

* [x] Implement the Data() SMTP hook. Inside, generate a secure UUID for the transaction.  
* [x] Setup an `s3manager.Uploader` connected to the SMTP `io.Reader` stream.  
* [x] Stream the raw MIME payload directly to an S3 spool object key (e.g., `s3://bucket/spool/{uuid}.eml`) using concurrent chunked uploads (e.g., 5MB parts) ~~over a VPC Gateway Endpoint~~.  
* [x] Get library github.com/emersion/go-msgauth to implement DKIM
* [x] Simultaneously pipe the stream into a single-pass DKIM signature checker using a Go io.TeeReader or MultiWriter wrapper.  
* [x] ~~Configure an S3 Bucket Lifecycle Rule to abort incomplete multipart uploads after 1 day.~~ infrastructure related, to be done later.
* **Verification Checkpoint**: Send a large mock email containing attachments. Verify the Go process memory (RAM) consumption remains flat while the `.eml` file is successfully assembled in the S3 bucket.

### 3.3 Atomic Outbox Enqueueing

* [x] Publish a JSON payload containing the S3 object key (`s3://bucket/spool/{uuid}.eml`) to a Redis stream.  
* [x] Only after a successful S3 upload and Redis stream publish, return 250 OK to the sender.  
* **Verification Checkpoint**: Send an email. Verify that a message containing the S3 object key is published to the Redis stream and the connection gracefully terminates.

### 3.4 SMTP Edge Proxy & Ingestion API Refactor

*Isolate the SMTP daemon into a lightweight stateless proxy and migrate core ingestion logic to an Echo-based HTTP API.*

* [x] **Add Framework:** Run `go get github.com/labstack/echo/v4` and initialize the Echo router in `cmd/api.go`.
* [x] **`internal/ingest` (Shared Contracts):** Create this base package to hold shared constants (e.g., `HeaderEdgeToken = "X-Edge-Auth-Token"`, standard routes, and JSON error structures) to prevent magic strings across client/server boundaries.
* [x] **`internal/ingest/client` (Lightweight Edge Client):** 
  * [x] Implement an `IngestClient` struct wrapping the standard `net/http` library.
  * [x] Add a `StreamPayload(ctx context.Context, reader io.Reader) error` method that performs an HTTP POST, piping the reader directly to the request body, and applies the edge authentication header.
* [x] **`internal/api/` (Unified HTTP Layer):**
  * [x] **`router/`**: Create router setup to mount middleware and API routes (e.g., `POST /internal/api/v1/ingest`).
  * [x] **`middleware/`**: Implement Edge Authentication middleware to validate the shared secret/token for ingest routes.
  * [x] **`handler/ingest.go`**: Create the Ingestion controller. Move the S3 `s3manager.Uploader` logic, DKIM validation, and Redis outbox enqueueing from the SMTP daemon into this handler. Connect `c.Request().Body` directly to the S3 uploader.
* [x] **Refactor SMTP Daemon (`internal/smtp/session.go`):** 
  * [x] Initialize `client.NewIngestClient()` during SMTP server startup.
  * [x] Update the `Data()` hook to pass its `io.Reader` directly to `client.StreamPayload()`.
  * [x] Handle HTTP status codes: return `250 OK` to the MTA on a successful API upload, and `4xx/5xx` SMTP errors on failure.
  * [x] Remove all S3, Redis, and direct database dependencies from the SMTP daemon to finalize its stateless proxy architecture.
* **Verification Checkpoint**: Send a mock email containing attachments via the local SMTP server. Verify the payload is proxied over HTTP to the Echo server and successfully spooled to S3 without the SMTP process importing the AWS SDK.

## Phase 4: Spool Queue Worker & MIME Parsing Engine

*Process spooled email files concurrently, parse nested attachments, and upload results securely.*

### 4.1 Redis Consumer Group Worker Pool

* [x] Implement a concurrent, multi-threaded worker pool utilizing Go channels and Goroutines.  
* [x] Initialize a Redis Consumer Group (`XGROUP CREATE`) to track message consumption across multiple worker nodes.  
* [x] Write a worker loop that blocks on `XREADGROUP` to consume new spool jobs from the Redis stream, ensuring each message is routed to exactly one thread.  
* [x] Implement a recovery loop utilizing `XPENDING` and `XCLAIM` to detect and retry messages that have stalled or failed to process.  
* **Verification Checkpoint**: Publish mock job JSON payloads directly to the Redis stream and verify that multiple running workers process them concurrently with zero collisions. Kill a worker during processing and ensure another worker re-claims the pending job.

### 4.2 MIME Engine & S3 Storage Ingestion

* [x] Inside the worker thread:  
  1. Download the raw `.eml` spool file from S3 into memory or a streaming reader.  
  2. Pass the data stream to the enmime parser (`enmime.ReadEnvelope`).  
  3. Extract standard text bodies, metadata headers, and attachments.  
  4. Store the parsed email body structure as a `contents.json` file in S3 inside the application folder path (`apps/{application_id}/...`).  
  5. Upload raw attachment binaries as separate files (`attachments/{id}.bin`).  
  6. Insert meta rows into the `ingested_emails` database table.  
  7. Handle processing results and errors: update `attempt_count` in the `inbound_spool_queue` database table. Use the processing error to decide whether to acknowledge (`XACK`) the message (e.g. fatal errors or max attempts reached) or leave it for retry.  
  8. On success, delete the raw spool `.eml` object from S3, update the `inbound_spool_queue` status to `COMPLETED`, and acknowledge the Redis stream message (`XACK`).  
* **Verification Checkpoint**: Send a test email containing multiple files (e.g., a PDF and an image). Verify the spool database row status is updated to `COMPLETED`, the Redis message is acknowledged, the raw `.eml` is deleted from S3, and S3 contains the mapped directory structure.

## Phase 5: Secure Webhook & Callback Dispatch Engine

*Notify client applications using a secure transaction outbox backed by strict SSRF defenses and replay protection.*

### 5.1 Callback Setup & SSRF Defense Handshake

* [x] Implement the webhook configuration and subscription logic inside the Application API.  
* [x] Create an outbound DNS-resolving network dialer that overrides standard lookups and drops connection targets resolving to private, loopback, or link-local address blocks (RFC 1918 limits).  
* [x] Implement the Challenge Handshake:  
  1. Generate a cryptographic hex challenge token.  
  2. Send an outbound POST containing the challenge to the candidate's webhook endpoint.  
  3. Verify the endpoint returns an HTTP 200 echoing the exact challenge.  
* **Verification Checkpoint**: Try registering http://127.0.0.1:5432/callback as a webhook destination. Verify the DNS hook blocks the setup. Register a public mock endpoint and verify the challenge handshakes cleanly.

### 5.2 Outbox Runner & Jitter Retries

* [x] Write the background loop to poll webhook_delivery_jobs.  
* [x] Formulate the payload containing parsed JSON content metadata.  
* [x] Implement HMAC-SHA256 signature generator. Append the signature in the custom header X-Gateway-Signature alongside the transmission timestamp.  
* [x] Implement Exponential Backoff with Full Jitter retry calculations to handle failed callback targets.  
* [x] Write audit attempts to the webhook_logs table.  
* **Verification Checkpoint**: Point webhooks to a test endpoint. Shut down the test endpoint to trigger failure retries. Verify that backoff wait times increase exponentially with randomized jitter intervals.

---

## Phase 6: Application API & Direct S3 Presigned Downloads

*Expose secure REST endpoints for M2M and Dashboard callers, authorizing attachment access via S3 presigned URLs.*

### 6.1 Authentication & REST Endpoints

* [x] Implement JWT OIDC verification middleware utilizing a cached JWKS endpoint provider.  
* [x] Create core routing pathways:  
  * `POST /api/v1/addresses` (Provision new assigned 10-char routing paths / `/app/v1/applications/:app_id/addresses`).  
  * `GET /api/v1/application` (Retrieve configurations and active scopes / `/app/v1/applications/:app_id`).  
  * `GET /api/v1/emails` (List history logs / `/app/v1/applications/:app_id/emails`).  
* **Verification Checkpoint**: Query these endpoints with both valid and expired OIDC access tokens to verify signature enforcement.

### 6.2 Direct S3 Presigned Attachment Downloader (Dual-Principal API)

* [x] **S3 Presigned Storage Service (`backend/internal/storage/`)**:
  * Extend `S3StorageService` with direct `GetObject` presigned URL generation using AWS SDK v2 (`s3.NewPresignClient`), bypassing AWS STS `AssumeRole`.
  * Generate short-lived presigned download URLs with configurable expiration (15-minute TTL).
  * Maintain LocalStack and custom endpoint compatibility (`S3BaseEndpoint` and `UsePathStyle`) for local dev environment.

* [x] **Dual-Principal API Endpoints & Routing (`backend/internal/api/` per ADR-0002 & ADR-0004)**:
  * **Dashboard SPA Surface (`/app/v1/...`)**:
    * Implement `GET /app/v1/applications/:app_id/emails/:email_id/attachments/:attachment_id` in `backend/internal/api/handler/application.go`.
    * Authenticated via `AppAuth` / `UserAuth` middleware with explicit `:app_id` URL scope validated by `CanAccessApplication(ctx, appID)`.
    * Verify email and attachment tenant ownership for `app_id`.
    * Generate presigned URL and return `dto.AttachmentURLResponse` (`attachment_id`, `download_url`, `expires_at`).
  * **M2M Tenant API Surface (`/api/v1/...`)**:
    * Implement `GET /api/v1/emails/:email_id/attachments/:attachment_id` in `backend/internal/api/handler/api_email.go`.
    * Authenticated via `M2MAuth` middleware with context-derived `app_id` (`c.Get("app_id")`), flat route without `:app_id` in URL.
    * Verify email and attachment tenant ownership for caller's `app_id`, generate presigned URL, and return `dto.AttachmentURLResponse`.

* [x] **Frontend Dashboard SPA Integration (`frontend/apps/dashboard/`)**:
  * Wire `useEmailStore.fetchAttachmentURL(emailId, attachmentId)` action calling `GET /app/v1/applications/:app_id/emails/:email_id/attachments/:attachment_id`.
  * Update attachment download buttons in `src/pages/emails/[id].vue` to request presigned download links and launch direct downloads in browser.
  * Add error handling and notification toasts (`v-snackbar`) for download link generation failures or expired links.

* **Verification Checkpoint**:
  * **Backend Tests**: Run `go test ./...` in `backend/` to verify presigned URL generation and endpoint handler logic.
  * **Tenant Isolation Checkpoint**: Authenticate as Tenant A (`app_id_a`) and request attachment URL; verify key path matches `apps/<app_id_a>/...`. Verify Tenant A cannot fetch attachment URLs for Tenant B's emails (returns 404 Not Found / 403 Forbidden).
  * **Dual-Principal Surface Verification**: Verify presigned URL retrieval across both `/app/v1/applications/:app_id/emails/:email_id/attachments/:attachment_id` (Dashboard session) and `/api/v1/emails/:email_id/attachments/:attachment_id` (M2M bearer token).

---

## Phase 7: Management Dashboard (Vue.js SPA)

*Provide the developer portal interface in `frontend/apps/dashboard` to tie all system components together.*

> [!IMPORTANT]
> **Strict Requirement:** Vue 3 Router MUST use **file-based routing** via `src/pages/`. Do NOT define static route arrays manually in `src/router/index.js`.  
> **API Dual Principal Model (ADR-0002 & ADR-0004):** Dashboard SPA API requests MUST use the **`/app/v1/...`** prefix (with explicit `:app_id` scope in URL: `/app/v1/applications/:app_id/...`), authenticated via `UserAuth` OIDC Bearer tokens. Do NOT use M2M API `/api/v1/...` endpoints for dashboard operations.

### 7.1 OIDC Auth Provider & HTTP Client Subsystem

* [x] Implement OIDC PKCE authentication handler (`src/services/authService.js` / `src/services/oidcService.js`) utilizing `oidc-client-ts` (`UserManager`) interfacing with `Apogee-dev` IdP.  
* [x] Configure HTTP client module (`src/services/apiService.js`) with base URL `/app/v1`, automatic `Authorization: Bearer <token>` header injection, 401 response handling, and automatic token refresh managed by `oidc-client-ts`.  
* [x] Implement `useAuthStore` in `src/stores/auth.js` to manage user identity, OIDC tokens (via `oidc-client-ts`), active `app_id` selection, login, and logout state using `src/services/` modules.  
* **Verification Checkpoint**: Run dev server (`pnpm dev`), verify OIDC authentication handshake via `oidc-client-ts`, token storage, and Bearer token attachment to outgoing `/app/v1/...` API requests.

### 7.2 Core Pinia Stores & API Client Integration

* [x] Implement `useAppStore` (`src/stores/application.js`) for tenant details, API keys, and global settings (`GET /app/v1/applications/:app_id`).  
* [x] Implement `useAddressStore` (`src/stores/addresses.js`) for provisioning (`POST /app/v1/applications/:app_id/addresses`) and toggling active 10-character email local-part routing addresses.  
* [x] Implement `useEmailStore` (`src/stores/emails.js`) for paginated email logs (`GET /app/v1/applications/:app_id/emails`) and fetching S3 presigned attachment URLs (`GET /app/v1/applications/:app_id/emails/{emailId}/attachments/{attachmentId}`).  
* [x] Implement `useWebhookStore` (`src/stores/webhooks.js`) for endpoint challenge verification (`PUT /app/v1/applications/:app_id/webhook`), job history logs, and manual outbox re-delivery requests (`POST /app/v1/applications/:app_id/webhook/jobs/{id}/redeliver`).  
* **Verification Checkpoint**: Test store actions against backend mock/live `/app/v1/...` API endpoints and confirm reactive state updates.

### 7.3 Dashboard Navigation Shell & Global UI Layout Specs

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

### 7.4 Vue 3 File-Based Page Routes & Detailed Screen Specifications (`src/pages/`)

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

### 7.5 Operational Sandbox & Real-Time Log Inspector

* [x] Build interactive JSON payload inspector in `src/pages/webhooks/index.vue` and `src/pages/emails/[id].vue` to display formatted headers, body, and attempt logs.  
* [x] Add auto-refresh / polling toggle for delivery sandbox log table to monitor live outbox worker job execution.  
* [x] Add toast notification system (Vuetify `v-snackbar`) for success/error feedback across all user actions.  
* **Verification Checkpoint**: Perform manual webhook re-delivery from sandbox log table and verify UI updates immediately upon completion.

### 7.6 End-to-End System Integration & Verification

* [x] Run full system flow: Log in via OIDC PKCE -> Select/fetch app scope (`/app/v1/applications/:app_id`) -> Provision new 10-character address -> Send email via SMTP -> Verify ingestion in Email Logs -> Download attachment via STS presigned S3 link -> Verify webhook payload delivered -> Trigger manual re-delivery from Sandbox.  
* [x] Run `pnpm lint` and `pnpm build` in `frontend/apps/dashboard` to verify zero build or linting errors.  
* **Verification Checkpoint**: Confirm clean `pnpm build` output and zero lint errors.

### 7.7 Webhook Outbox Delivery Jobs & Re-delivery API Engine

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

