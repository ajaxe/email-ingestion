# **Email Ingestion Gateway: Progressive Implementation Roadmap**

This document outlines a high-level, production-grade implementation roadmap. The tasks are strictly ordered so that each completed item provides the baseline infrastructure, code libraries, or data layers required for the subsequent step.

## **Phase 1: Local Development Foundations & Data Engine**

*Before building active server daemons, establish the local environment boundaries, schemas, and typed data wrappers, leveraging your existing running PostgreSQL instance.*

### **1.1 Local Project Bootstrap**

* [ ] Initialize the project repository with a clean Go module structure (go mod init).  
* [ ] Create a local environment configuration file (.env or .env.local) to define development parameters:  
  * DB_DSN (pointing to your existing running PostgreSQL instance).  
  * S3_BUCKET (configured for local folder-spool testing or a dedicated development AWS bucket).  
* [ ] Implement a system configuration parser in Go (e.g., using cleanenv or standard library os.LookupEnv) to validate these database connections on application startup.  
* **Verification Checkpoint**: Run a simple main.go ping script that parses your .env file and successfully establishes a database connection pool (sql.DB) to your existing PostgreSQL instance.

### **1.2 DB Engine & SQLC Generation**

* [ ] Write the PostgreSQL schema in schema.sql outlining all relational tables, indexes, and custom enum types (spool_status, webhook_status).  
* [ ] Draft SQL query patterns in query.sql for all state modifications (such as transactional outbox locks, polling queries, address allocations).  
* [ ] Setup and run sqlc generate to generate safe, strongly-typed Go query files.  
* **Verification Checkpoint**: Apply the schemas to your running PostgreSQL instance and verify that the SQLC compiled Go files match database columns and types flawlessly.

## **Phase 2: In-Memory Address Caching & Perimeter SMTP Handshake**

*Accept inbound TCP packets and block unauthorized messages at the perimeter before storing any payload files on disk.*

### **2.1 SMTP Daemon Integration**

* [ ] Import github.com/emersion/go-smtp and instantiate a basic SMTP server listening on local port 2525.  
* [ ] Set up the TLS handshake configurations and map standard debug logging formats.  
* [ ] Implement Go SMTP Backend and Session interfaces to intercept standard hooks (Mail, Rcpt, Data).  
* **Verification Checkpoint**: Use telnet localhost 2525 or netcat to verify the server negotiates a connection and responds to standard EHLO greetings.

### **2.2 Inbound Address Validation Layer (RAM Cache + DB Index)**

* [ ] Implement a lightweight in-memory cache layer (using go-cache, ristretto, or simple thread-safe Redis clients).  
* [ ] Inside the Rcpt() connection hook:  
  1. Strip sub-address parameters (e.g., extracting a8f3g9j2k1 from a8f3g9j2k1+token123@domain.com).  
  2. Query the fast in-memory cache first.  
  3. Fall back to a targeted PostgreSQL query using the SQLC generated index search.  
  4. Cache the result aside and return 550 User Unknown to the client if the email local-part is unassigned.  
* [ ] Implement envelope-level SPF check inside the Mail() connection hook utilizing github.com/emersion/go-msgauth/spf based on the sender's connecting TCP IP.  
* **Verification Checkpoint**: Send test handshakes with both registered and fake email addresses. Validate that fake addresses are immediately dropped with a 550 SMTP error code during the socket connection.

## **Phase 3: Zero-Memory Disk Spooling Queue**

*Once an email is accepted at the socket level, get it out of memory as quickly as possible to protect against server crashes and memory leaks.*

### **3.1 Raw Stream Archiving (Zero-Memory Copy)**

* [ ] Implement the Data() SMTP hook. Inside, generate a secure UUID for the transaction.  
* [ ] Set up a target file descriptor pointing to the local spool directory (e.g., /tmp/spool/{uuid}.eml).  
* [ ] Use io.Copy combined with a constrained chunk buffer (e.g., 32KB) to stream raw MIME data from the socket reader directly to the local disk.  
* [ ] Simultaneously pipe the stream into a single-pass DKIM signature checker using a Go io.MultiWriter wrapper.  
* **Verification Checkpoint**: Send a large mock email containing attachments. Verify the Go process memory (RAM) consumption remains flat while the .eml file grows directly on your host file system.

### **3.2 Atomic Outbox Enqueueing**

* [ ] Write a transactional DB hook to insert the raw spool pointer path (/tmp/spool/{uuid}.eml) into the inbound_spool_queue database table with a PENDING status.  
* [ ] Only after a successful PostgreSQL commit, return 250 OK to the sender.  
* **Verification Checkpoint**: Send an email. Verify that an inbound_spool_queue row appears in the database and the connection gracefully terminates.

## **Phase 4: Spool Queue Worker & MIME Parsing Engine**

*Process spooled email files concurrently, parse nested attachments, and upload results securely.*

### **4.1 SKIP LOCKED Spool Poller**

* [ ] Implement a concurrent, multi-threaded worker pool utilizing Go channels and Goroutines.  
* [ ] Write a poller loop that queries GetNextSpoolJob utilizing PostgreSQL's native FOR UPDATE SKIP LOCKED.  
* [ ] Ensure that even if multiple Go worker nodes scale up, each spool job is acquired by exactly one thread without locking others.  
* **Verification Checkpoint**: Force-populate the DB queue with multiple test logs and verify that your workers process them concurrently with zero collisions or race conditions.

### **4.2 MIME Engine & S3 Storage Ingestion**

* [ ] Inside the worker thread:  
  1. Open the physical .eml file from disk.  
  2. Pass the file descriptor to the enmime parser (enmime.ReadEnvelope).  
  3. Extract standard text bodies, metadata headers, and attachments.  
  4. Store the parsed email body structure as a contents.json file in S3 inside the application folder path (apps/{application_id}/...).  
  5. Upload raw attachment binaries as separate files (attachments/{id}.bin).  
  6. Insert meta rows into the ingested_emails database table.  
  7. Delete the raw spool file from disk and execute DeleteSpoolJob in PostgreSQL.  
* **Verification Checkpoint**: Send a test email containing multiple files (e.g., a PDF and an image). Verify the spool database row is cleared, the file is deleted from your local directory, and S3 contains the mapped directory structure with all binary files matching their original byte sizes.

## **Phase 5: Secure Webhook & Callback Dispatch Engine**

*Notify client applications using a secure transaction outbox backed by strict SSRF defenses and replay protection.*

### **5.1 Callback Setup & SSRF Defense Handshake**

* [ ] Implement the webhook configuration and subscription logic inside the Application API.  
* [ ] Create an outbound DNS-resolving network dialer that overrides standard lookups and drops connection targets resolving to private, loopback, or link-local address blocks (RFC 1918 limits).  
* [ ] Implement the Challenge Handshake:  
  1. Generate a cryptographic hex challenge token.  
  2. Send an outbound POST containing the challenge to the candidate's webhook endpoint.  
  3. Verify the endpoint returns an HTTP 200 echoing the exact challenge.  
* **Verification Checkpoint**: Try registering http://127.0.0.1:5432/callback as a webhook destination. Verify the DNS hook blocks the setup. Register a public mock endpoint and verify the challenge handshakes cleanly.

### **5.2 Outbox Runner & Jitter Retries**

* [ ] Write the background loop to poll webhook_delivery_jobs.  
* [ ] Formulate the payload containing parsed JSON content metadata.  
* [ ] Implement HMAC-SHA256 signature generator. Append the signature in the custom header X-Gateway-Signature alongside the transmission timestamp.  
* [ ] Implement Exponential Backoff with Full Jitter retry calculations to handle failed callback targets.  
* [ ] Write audit attempts to the webhook_logs table.  
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
* [ ] Mount the persistent host directory /tmp/spool inside the app container volume map (spool_data) to prevent container layers from bloating with transient spooled email files.

### **8.2 Traefik Routing & Production DNS**

* [ ] Create a Traefik routing configuration to handle automated Let's Encrypt SSL/TLS certificates and expose API ports securely.  
* [ ] Configure your system's public DNS MX records to point to the ingestion server's public IP address.  
* [ ] Implement clean TXT configurations (such as standard SPF strings, DKIM keys, and a basic _dmarc DMARC policy record) to prepare your domain for safe inbound validation.  
* **Verification Checkpoint**: Deploy the production stack via docker compose up -d and confirm that external mail clients can perform TLS handshakes and route traffic through Traefik to your Go SMTP Daemon.