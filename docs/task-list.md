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

*Provide the developer portal interface to tie the system components together.*

### **7.1 Key Portal Views**

* [ ] Implement standard OIDC login with PKCE against the Custom IdP (`Apogee-dev`).  
* [ ] Build the Core Management Console view containing:  
  * **API Keys & Webhooks tab**: Manage endpoints, secrets, and trigger verification handshakes.  
  * **Routing Console**: Dynamically provision, activate, and deactivate 10-character email addresses.  
  * **Delivery Sandbox Log**: Browse inbound emails, inspect raw metadata payloads, view active webhook retry counters, and click "Re-deliver Webhook" to troubleshoot failed integrations.  
* **Verification Checkpoint**: Perform an end-to-end integration test: send an email to a newly provisioned address in the UI, watch the webhook dispatch successfully, and inspect the delivery logs in the developer dashboard.

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