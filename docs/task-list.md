# **Email Ingestion Gateway: Progressive Implementation Roadmap**

This document outlines a high-level, production-grade implementation roadmap. The tasks are strictly ordered so that each completed item provides the baseline infrastructure, code libraries, or data layers required for the subsequent step.

## **Phase 1: Local Development Foundations & Data Engine**

*Before building active server daemons, establish the local environment boundaries, schemas, and typed data wrappers, leveraging your existing running PostgreSQL instance.*

### **1.1 Local Project Bootstrap**

* [x] Initialize the project repository with a clean Go module structure (go mod init).  
* [x] ~~Create a local environment configuration file (.env or .env.local) to define development parameters~~:  
  * DB_DSN (pointing to your existing running PostgreSQL instance).  
  * S3_BUCKET (configured for local folder-spool testing or a dedicated development AWS bucket).  
* [x] ~~Implement a system configuration parser in Go (e.g., using cleanenv or standard library os.LookupEnv) to validate these database connections on application startup~~.  
* [x] Use viper lib to take path to _config.yaml_ file with defaults while accepting environment overrides with prefix `EM_`.

* **Verification Checkpoint**: Run a simple main.go ping script that parses your .env file and successfully establishes a database connection pool (sql.DB) to your existing PostgreSQL instance.

### **1.2 DB Engine & SQLC Generation**

* [x] Write the PostgreSQL schema in schema.sql outlining all relational tables, indexes, and custom enum types (spool_status, webhook_status).  
* [x] Draft SQL query patterns in query.sql for all state modifications (such as transactional outbox locks, polling queries, address allocations).  
* [x] Setup and run sqlc generate to generate safe, strongly-typed Go query files.  
* **Verification Checkpoint**: Apply the schemas to your running PostgreSQL instance and verify that the SQLC compiled Go files match database columns and types flawlessly.

## **Phase 2: In-Memory Address Caching & Perimeter SMTP Handshake**

*Accept inbound TCP packets and block unauthorized messages at the perimeter before storing any payload files on disk.*

### **2.1 SMTP Daemon Integration**

* [x] Import github.com/emersion/go-smtp and instantiate a basic SMTP server listening on local port 2525.  
* [x] Set up the TLS handshake configurations and map standard debug logging formats.  
* [x] Implement Go SMTP Backend and Session interfaces to intercept standard hooks (Mail, Rcpt, Data).  
* **Verification Checkpoint**: Use telnet localhost 2525 or netcat to verify the server negotiates a connection and responds to standard EHLO greetings.

### **2.2 Inbound Address Validation Layer (RAM Cache + DB Index)**

* [x] Implement a lightweight in-memory cache layer (using go-cache, ristretto, or simple thread-safe Redis clients).  
* [x] Inside the Rcpt() connection hook:  
  1. Strip sub-address parameters (e.g., extracting a8f3g9j2k1 from a8f3g9j2k1+token123@domain.com).  
  2. Query the fast in-memory cache first.  
  3. Fall back to a targeted PostgreSQL query using the SQLC generated index search.  
  4. Cache the result aside and return 550 User Unknown to the client if the email local-part is unassigned.  
* [x] Implement envelope-level SPF check inside the Mail() connection hook utilizing ~~github.com/emersion/go-msgauth/spf~~ github.com/mileusna/spf based on the sender's connecting TCP IP.  
* **Verification Checkpoint**: Send test handshakes with both registered and fake email addresses. Validate that fake addresses are immediately dropped with a 550 SMTP error code during the socket connection.

## **Phase 3: Stateless S3 Spooling Architecture**

*Once an email is accepted at the socket level, stream it directly to S3 via multipart uploads to guarantee data durability and preserve stateless Gateway nodes.*

### **3.1 Spool Queue Database Schema**

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

### **3.2 Direct S3 Stream Archiving**

* [x] Implement the Data() SMTP hook. Inside, generate a secure UUID for the transaction.  
* [x] Setup an `s3manager.Uploader` connected to the SMTP `io.Reader` stream.  
* [x] Stream the raw MIME payload directly to an S3 spool object key (e.g., `s3://bucket/spool/{uuid}.eml`) using concurrent chunked uploads (e.g., 5MB parts) ~~over a VPC Gateway Endpoint~~.  
* [x] Get library github.com/emersion/go-msgauth to implement DKIM
* [x] Simultaneously pipe the stream into a single-pass DKIM signature checker using a Go io.TeeReader or MultiWriter wrapper.  
* [x] ~~Configure an S3 Bucket Lifecycle Rule to abort incomplete multipart uploads after 1 day.~~ infrastructure related, to be done later.
* **Verification Checkpoint**: Send a large mock email containing attachments. Verify the Go process memory (RAM) consumption remains flat while the `.eml` file is successfully assembled in the S3 bucket.

### **3.3 Atomic Outbox Enqueueing**

* [x] Publish a JSON payload containing the S3 object key (`s3://bucket/spool/{uuid}.eml`) to a Redis stream.  
* [x] Only after a successful S3 upload and Redis stream publish, return 250 OK to the sender.  
* **Verification Checkpoint**: Send an email. Verify that a message containing the S3 object key is published to the Redis stream and the connection gracefully terminates.

### **3.4 SMTP Edge Proxy & Ingestion API Refactor**

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

## **Phase 4: Spool Queue Worker & MIME Parsing Engine**

*Process spooled email files concurrently, parse nested attachments, and upload results securely.*

### **4.1 Redis Consumer Group Worker Pool**

* [x] Implement a concurrent, multi-threaded worker pool utilizing Go channels and Goroutines.  
* [x] Initialize a Redis Consumer Group (`XGROUP CREATE`) to track message consumption across multiple worker nodes.  
* [x] Write a worker loop that blocks on `XREADGROUP` to consume new spool jobs from the Redis stream, ensuring each message is routed to exactly one thread.  
* [x] Implement a recovery loop utilizing `XPENDING` and `XCLAIM` to detect and retry messages that have stalled or failed to process.  
* **Verification Checkpoint**: Publish mock job JSON payloads directly to the Redis stream and verify that multiple running workers process them concurrently with zero collisions. Kill a worker during processing and ensure another worker re-claims the pending job.

### **4.2 MIME Engine & S3 Storage Ingestion**

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

## **Phase 5: Secure Webhook & Callback Dispatch Engine**

*Notify client applications using a secure transaction outbox backed by strict SSRF defenses and replay protection.*

### **5.1 Callback Setup & SSRF Defense Handshake**

* [x] Implement the webhook configuration and subscription logic inside the Application API.  
* [x] Create an outbound DNS-resolving network dialer that overrides standard lookups and drops connection targets resolving to private, loopback, or link-local address blocks (RFC 1918 limits).  
* [x] Implement the Challenge Handshake:  
  1. Generate a cryptographic hex challenge token.  
  2. Send an outbound POST containing the challenge to the candidate's webhook endpoint.  
  3. Verify the endpoint returns an HTTP 200 echoing the exact challenge.  
* **Verification Checkpoint**: Try registering http://127.0.0.1:5432/callback as a webhook destination. Verify the DNS hook blocks the setup. Register a public mock endpoint and verify the challenge handshakes cleanly.

### **5.2 Outbox Runner & Jitter Retries**

* [x] Write the background loop to poll webhook_delivery_jobs.  
* [x] Formulate the payload containing parsed JSON content metadata.  
* [x] Implement HMAC-SHA256 signature generator. Append the signature in the custom header X-Gateway-Signature alongside the transmission timestamp.  
* [x] Implement Exponential Backoff with Full Jitter retry calculations to handle failed callback targets.  
* [x] Write audit attempts to the webhook_logs table.  
* **Verification Checkpoint**: Point webhooks to a test endpoint. Shut down the test endpoint to trigger failure retries. Verify that backoff wait times increase exponentially with randomized jitter intervals.

## **Phase 6: Application API & Brokered S3 Access Control**

*Expose secure REST endpoints and authorize user file-access namespaces using brokered AWS STS IAM role assumption.*

### **6.1 Authentication & REST Endpoints**

* [ ] Implement JWT OIDC verification middleware utilizing a cached JWKS endpoint provider.  
* [ ] Create core routing pathways:  
  * POST /api/v1/addresses (Provision new assigned 10-char routing paths).  
  * GET /api/v1/application (Retrieve configurations and active scopes).  
  * GET /api/v1/emails (List history logs).  
* **Verification Checkpoint**: Query these endpoints with both valid and expired OIDC access tokens to verify signature enforcement.

### **6.2 Brokered IAM Role Assumption (S3 Downloader)**

* [ ] Implement the S3 download endpoint: GET /api/v1/emails/{emailId}/attachments/{attachmentId}.  
* [ ] Inside the Go handler:  
  1. Validate JWT. Resolve the active request client to their internal application identity.  
  2. Query Postgres to fetch the application's unique aws_iam_role_arn.  
  3. Call AWS STS AssumeRole using the Go AWS SDK.  
  4. Using the returned transient credentials, instantiate a scoped S3 client.  
  5. Generate a short-lived S3 Presigned URL.  
  6. Return the presigned URL to the client.  
* **Verification Checkpoint**: Authenticate as Tenant A and request a download link for an attachment. Verify the link works and matches S3 storage paths. Attempt to modify the URL path to Tenant B's folder and confirm AWS S3 rejects the request immediately.

## **Phase 7: Management Dashboard (Vue.js SPA)**

*Provide the developer portal interface to tie the system components together.*

### **7.1 Key Portal Views**

* [ ] Implement standard OIDC login with PKCE against the Custom IdP (Apogee-dev).  
* [ ] Build the Core Management Console view containing:  
  * **API Keys & Webhooks tab**: Manage endpoints, secrets, and trigger verification handshakes.  
  * **Routing Console**: Dynamically provision, activate, and deactivate 10-character email addresses.  
  * **Delivery Sandbox Log**: Browse inbound emails, inspect raw metadata payloads, view active webhook retry counters, and click "Re-deliver Webhook" to troubleshoot failed integrations.  
* **Verification Checkpoint**: Perform an end-to-end integration test: send an email to a newly provisioned address in the UI, watch the webhook dispatch successfully, and inspect the delivery logs in the developer dashboard.

## **Phase 8: Containerization & Infrastructure Deployment**

*Prepare the production-ready infrastructure stack, container routing profiles, and reverse proxy patterns.*

### **8.1 Docker Configurations & Build Stage**

* [ ] Write a multi-stage Dockerfile optimizing the Go application binary size and execution security (using a scratch or alpine base).  
* [ ] Implement the docker-compose.yml for production deployments, omitting Postgres container orchestration (since you utilize an external PostgreSQL instance), but including services for LocalStack/S3 and Traefik.

### **8.2 Traefik Routing & Production DNS**

* [ ] Create a Traefik routing configuration to handle automated Let's Encrypt SSL/TLS certificates and expose API ports securely.  
* [ ] Configure your system's public DNS MX records to point to the ingestion server's public IP address.  
* [ ] Implement clean TXT configurations (such as standard SPF strings, DKIM keys, and a basic _dmarc DMARC policy record) to prepare your domain for safe inbound validation.  
* **Verification Checkpoint**: Deploy the production stack via docker compose up -d and confirm that external mail clients can perform TLS handshakes and route traffic through Traefik to your Go SMTP Daemon.