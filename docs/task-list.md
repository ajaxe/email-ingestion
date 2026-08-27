# **Email Ingestion Gateway: Active Implementation Roadmap**

This document tracks active and upcoming development phases. Completed phases are archived in [task-list-archive.md](file:///C:/CodeWorkspace/projects/email-ingestion/docs/task-list-archive.md).

## **Project Status Overview**

**Current Phase:** Phase 8 (Deployment & Infra)  
**Overall Progress:** 8 / 9 Phases Completed (88.9%)

| Phase | Status | Key Scope | Archive Details |
| :--- | :--- | :--- | :--- |
| **Phase 1: Local Foundations** | Completed | Bootstrap, Config & SQLC Generation | [View Details](file:///C:/CodeWorkspace/projects/email-ingestion/docs/task-list-archive.md#phase-1-local-development-foundations--data-engine) |
| **Phase 2: Perimeter SMTP & Cache** | Completed | SMTP Daemon & Address Validation | [View Details](file:///C:/CodeWorkspace/projects/email-ingestion/docs/task-list-archive.md#phase-2-in-memory-address-caching--perimeter-smtp-handshake) |
| **Phase 3: S3 Spooling & Edge Proxy** | Completed | S3 Direct Uploads, Redis Outbox & Echo Proxy | [View Details](file:///C:/CodeWorkspace/projects/email-ingestion/docs/task-list-archive.md#phase-3-stateless-s3-spooling-architecture) |
| **Phase 4: MIME Engine & Workers** | Completed | Redis Consumer Group Pool & MIME Parsing | [View Details](file:///C:/CodeWorkspace/projects/email-ingestion/docs/task-list-archive.md#phase-4-spool-queue-worker--mime-parsing-engine) |
| **Phase 5: Secure Webhook Dispatch** | Completed | SSRF Guard Handshake & Jitter Retries | [View Details](file:///C:/CodeWorkspace/projects/email-ingestion/docs/task-list-archive.md#phase-5-secure-webhook--callback-dispatch-engine) |
| **Phase 6: REST API & Presigned S3 Downloads** | Completed | OIDC JWT Auth & Direct S3 Presigned Attachment Downloader | [View Details](file:///C:/CodeWorkspace/projects/email-ingestion/docs/task-list-archive.md#phase-6-application-api--direct-s3-presigned-downloads) |
| **Phase 7: Management Dashboard** | Completed | Vue.js SPA & Developer Management Console | [View Details](file:///C:/CodeWorkspace/projects/email-ingestion/docs/task-list-archive.md#phase-7-management-dashboard-vuejs-spa) |
| **Phase 8: Deployment & Infra** | Pending | Docker Multi-Stage, Traefik & MX DNS | *Active Below* |
| **Phase 9: Database Migrations** | Completed | Goose Migration Engine, SemVer v0 & SQLC Integration | *Active Below* |

---

## **Phase 8: Containerization & Infrastructure Deployment**

*Prepare the production-ready infrastructure stack, container routing profiles, and reverse proxy patterns.*

### **8.1 Docker Configurations & Build Stage**

* [x] Write a multi-stage Dockerfile optimizing the Go application binary size and execution security (using a scratch or alpine base).  
* [x] Implement the `docker-compose.yml` for production deployments, omitting Postgres container orchestration (since you utilize an external PostgreSQL instance), but including services for LocalStack/S3 and Traefik.

### **8.2 Traefik Routing & Production DNS**

* [x] Create a Traefik routing configuration to handle automated Let's Encrypt SSL/TLS certificates and expose API ports securely.  
* [x] Configure your system's public DNS MX records to point to the ingestion server's public IP address.  
* [x] Implement clean TXT configurations (such as standard SPF strings, DKIM keys, and a basic `_dmarc` DMARC policy record) to prepare your domain for safe inbound validation.  
* **Verification Checkpoint**: Deploy the production stack via `docker compose up -d` and confirm that external mail clients can perform TLS handshakes and route traffic through Traefik to your Go SMTP Daemon.

---

## **Phase 9: Database Migration Engine & Versioning (Pressly Goose)**

*Implement a versioned, modular database migration pipeline using `pressly/goose` with Go `embed` support and SemVer folder organization.*

### **9.1 SemVer Directory Structure & Schema Decomposition**

* [x] Create initial semantic version directory `backend/pkg/database/migrations/v0/`.  
* [x] Decompose monolithic `backend/pkg/database/migrations/public.sql` into sequential, numbered migration scripts:  
  * `00001_init_extensions_and_enums.sql`: Enable `uuid-ossp` extension and `spool_status` / `webhook_status` ENUM types.  
  * `00002_create_applications.sql`: Tenant `applications` table.  
  * `00003_create_assigned_emails.sql`: `assigned_emails` table and lookup indexes.  
  * `00004_create_ingested_emails.sql`: `ingested_emails` metadata table and search indexes.  
  * `00005_create_webhook_jobs_and_logs.sql`: `webhook_delivery_jobs` & `webhook_logs` outbox tables and scheduled lookup indexes.  
  * `00006_create_spool_queue.sql`: `inbound_spool_queue` buffer table.  
  * `00007_create_users_and_access.sql`: `users`, `user_application_access`, and `api_keys` security tables + indexes.  
  * `00008_create_organizations.sql`: `organizations` container table, FK linking to `applications`, and personal owner index.  
* [x] Add explicit `-- +goose Up` and `-- +goose Down` annotation blocks to every script for full rollback support.


### **9.2 SQLC Schema Path Configuration**

* [x] Update `backend/sqlc.yaml` `schema` key to point to the semver folder list:  
  ```yaml
  schema:
    - "pkg/database/migrations/v0"
  ```
* [x] Run `sqlc generate` in `backend/` and verify Go database models in `pkg/database/public/` generate cleanly with zero schema regressions.

### **9.3 Embedded Goose Runner & CLI Subcommand**

* [x] Add `github.com/pressly/goose/v3` package dependency to `backend/go.mod`.  
* [x] Create `backend/pkg/database/migrations.go` migration wrapper using Go `//go:embed migrations/*/*.sql` and `fs.Sub`.  
* [x] Create Cobra CLI migration command (`backend/cmd/migrate.go`) supporting subcommands: `up`, `down`, `status`, `version`, and `skip`.  
* [x] Integrate automatic migration checks into service startup routines (`backend/internal/startup/`).

### **9.4 End-to-End Verification Checkpoint**

* [x] Run `go run cmd/api.go migrate up` against a clean Postgres container and confirm table creation.  
* [x] Verify migration history tracked in PostgreSQL `goose_db_version` table.  
* [x] Test `goose status` and `goose skip` commands to confirm migration skipping and status reporting functionality.  
* [x] Run `sqlc generate` and `go test ./...` to verify complete system compatibility.
