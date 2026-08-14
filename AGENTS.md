# AGENTS.md: Email Ingestion Gateway Context

## 1. Project Mission
The **Email Ingestion Gateway** is a production-grade microservices suite designed to handle inbound SMTP traffic, parse MIME emails, securely store attachments, and deliver webhooks to registered SaaS applications. It ensures high throughput, reliable delivery, and strict multi-tenant isolation.

## 2. Tech Stack & Core Dependencies
* **Backend:** **Go** (Golang)
* **Frontend:** **Vue.js** (Management Dashboard SPA)
* **Database:** **PostgreSQL 15** (Multi-tenant logical partitioning)
* **Object Storage:** **AWS S3** (**LocalStack** for local development)
* **SMTP Daemon:** **`go-smtp`** (non-blocking receiver)
* **MIME Parsing:** **`enmime`** (memory-efficient streaming and nested multipart handling)
* **Database Access:** **`sqlc`** (type-safe SQL)
* **Identity Provider (IdP):** **Apogee-dev IdP** (OIDC Bearer JWT Requests)

## 3. Architecture & Design Patterns
* **Integrated Microservices Suite:** Combines Go SMTP daemon, ingestion engine, API service, and worker pool.
* **Multi-Tenant Logical Partitioning:** Database records are strictly partitioned by `application_id` to prevent index bloat and ensure isolation.
* **Gateway-Brokered IAM Role Assumption:** S3 folder isolation is enforced using dynamic **AWS STS `AssumeRole`**, mapping authenticated identities to dedicated, restricted tenant IAM Roles without polluting the IdP with custom claims.
* **Transactional Outbox Pattern:** Guarantees atomic writes for email metadata and scheduled webhook deliveries, utilizing **Exponential Backoff with Full Jitter** and circuit-breaking.
* **Hybrid Caching Strategy:** High-performance, low false-positive recipient validation using in-memory cache (**Redis/Ristretto**) backed by PostgreSQL indexes during the `RCPT TO` phase.
* **Webhook Security & SSRF Guard:** Webhooks use a secure challenge/response handshake at registration, block RFC 1918 internal IP resolution, and enforce payload integrity using **HMAC-SHA256 signatures**.

## 4. Directory Mental Model
* **`backend/cmd/`**: Entrypoints for the various services (`api.go`, `smtp.go`, `root.go`).
* **`backend/internal/`**: Core internal business logic (e.g., `smtp/` for inbound routing, `startup/` for initialization, `storage/`).
* **`backend/pkg/`**: Reusable Go libraries and domain helpers (e.g., database queries via `sqlc` and config).
* **`docs/`**: Architecture and technical specifications.
* **`logs/`**: Local application log files.
* **`misc/`**: Miscellaneous scripts and assets.

## 5. Development Standards
* **Database Queries:** All Postgres access must be managed via **`sqlc`**. Write pure SQL in `pkg/database/public/query.public.sql` (or `query.sql`) and generate Go models.
* **Database Migrations:** All database schema changes MUST be created as a new SQL migration file with a 5-digit zero-padded sequential prefix (e.g., `00001_*.sql`, `00009_*.sql`) in `backend/pkg/database/migrations/v0/` for compatibility with `pressly/goose`. Never modify existing migration files.
* **Naming Conventions:** Use standard Go conventions (`camelCase` for internal, `PascalCase` for exported) and `snake_case` for all PostgreSQL schemas and tables.
* **API Responses & Error Handling:** The REST API expects JSON request payloads and returns **unified JSON error responses**. Authentication is managed via `Authorization: Bearer <JWT_Token>`.
* **Frontend Component Modularity:** Decompose pages into single-responsibility Vue components. Logical UI/UX elements (e.g., config forms, data tables, modal dialogs) must be extracted into dedicated components in `src/components/` rather than monolithic page files.

## 6. Hard Constraints & Anti-Patterns
* **DO NOT** modify existing SQL migration files to alter database schema. Create a new migration file with a 5-digit prefix (e.g. `00009_*.sql`).
* **DO NOT** write monolithic Vue pages containing multiple distinct UI sections or inline dialogs. Extract them into smaller, single-responsibility components.
* **DO NOT** use the standard library `net/mail` for MIME parsing. You **MUST** use **`enmime`**.
* **DO NOT** use Bloom Filters for address validation, as false positives violate SMTP reliability. Use the designated **Hybrid Caching Strategy**.
* **DO NOT** store or manage custom application/tenant claims in the IdP. The IdP should remain decoupled from service logic.
* **DO NOT** expose raw AWS credentials or leak DB IDs to the IdP. Use the **Gateway-Brokered S3 Access Control** pattern.
* **DO NOT** allow webhook deliveries to private, loopback, or RFC 1918 IP addresses. Maintain the strict SSRF DNS guard.
* **DO NOT** output or expose secret passwords, connection string credentials, API keys, or `.env` file contents in chat responses, transcripts, or artifact logs.
* **DO NOT** read or inspect local configuration files containing secret credentials (e.g., `backend/config.yaml`, `.env*`) unless explicitly requested by the user.

## 7. Operational Commands
* **Start Local Environment (PostgreSQL, LocalStack, App):**
  ```bash
  docker-compose up -d
  ```
  *(Note: Spins up Postgres on `5432`, LocalStack on `4566`, API on `8080`, and SMTP on `2525`)*
* **Database Code Generation:**
  ```bash
  cd backend && sqlc generate
  ```

## 8. Architecture Decision Records (ADRs)
When making structural changes, API modifications, or data model updates, you **MUST** consult the Architecture Decision Records (ADRs) located in the `docs/adr/` directory to ensure alignment with past decisions.
* **Start here:** Always review **`docs/adr/INDEX.md`** to find relevant historical decisions before implementing major features.
* **Existing ADRs** cover critical topics such as: SMTP proxy architecture, the dual-principal API model, local identity registry, and JIT user provisioning.
* **New Decisions:** If proposing a new architectural change that deviates from or expands upon existing patterns, document it by creating a new ADR in this directory.
