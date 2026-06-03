# AGENTS.md: Email Ingestion Gateway Context

## 1. Project Mission
The **Email Ingestion Gateway** is a production-grade microservices suite designed to handle inbound SMTP traffic, parse MIME emails, securely store attachments, and deliver webhooks to registered SaaS applications. It ensures high throughput, reliable delivery, and strict multi-tenant isolation.

## 2. Tech Stack & Core Dependencies
* **Backend:** Go (Golang)
  * **SMTP Daemon:** `go-smtp` (non-blocking receiver)
  * **MIME Parsing:** `enmime` (memory-efficient streaming and nested multipart handling)
  * **Database Access:** `sqlc` (type-safe SQL)
* **Frontend:** Vue.js (Management Dashboard SPA)
* **Database:** PostgreSQL 15 (Multi-tenant logical partitioning)
* **Object Storage:** AWS S3 (LocalStack for local development)
* **Identity Provider (IdP):** Apogee-dev IdP (OIDC Bearer JWT Requests)

## 3. Architecture & Design Patterns
* **Integrated Microservices Suite:** Combines Go SMTP daemon, ingestion engine, API service, and worker pool.
* **Multi-Tenant Logical Partitioning:** Database records are strictly partitioned by `application_id` to prevent index bloat and ensure isolation.
* **Gateway-Brokered IAM Role Assumption:** S3 folder isolation is enforced using dynamic AWS STS `AssumeRole`, mapping authenticated identities to dedicated, restricted tenant IAM Roles without polluting the IdP with custom claims.
* **Transactional Outbox Pattern:** Guarantees atomic writes for email metadata and scheduled webhook deliveries, utilizing **Exponential Backoff with Full Jitter** and circuit-breaking.
* **Hybrid Caching Strategy:** High-performance, low false-positive recipient validation using in-memory cache (Redis/Ristretto) backed by PostgreSQL indexes during the `RCPT TO` phase.
* **Webhook Security & SSRF Guard:** Webhooks use a secure challenge/response handshake at registration, block RFC 1918 internal IP resolution, and enforce payload integrity using **HMAC-SHA256 signatures**.

## 4. Directory Mental Model
* **`backend/`**: Go backend source code.
  * **`backend/cmd/`**: Entrypoints for the various services.
  * **`backend/internal/`**: Core internal business logic. Contains packages like `smtp/` (inbound routing) and `startup/` (service initialization).
  * **`backend/pkg/`**: Reusable Go libraries and domain helpers (e.g., database queries and schema definitions).
* **`docs/`**: Architecture and technical specifications (e.g., `email-ingestion-initial-refined.md`).
* **`logs/`**: Local application log files.
* **`misc/`**: Miscellaneous scripts and assets.

## 5. Development Standards
* **Database Queries:** All Postgres access must be managed via **`sqlc`**. Write pure SQL in `pkg/database/public/query.public.sql` and generate Go models.
* **Naming Conventions:** Use standard Go conventions (camelCase for internal, PascalCase for exported) and `snake_case` for all PostgreSQL schemas and tables.
* **API Responses:** The REST API expects JSON request payloads and returns **unified JSON error responses**. Authentication is managed via `Authorization: Bearer <JWT_Token>`.
* **Tenant Isolation Implementation:** Validate OIDC JWT locally via JWKS, map the `client_id` to `application_id`, fetch the `aws_iam_role_arn`, and generate transient S3 Presigned URLs via STS.

## 6. Hard Constraints & Anti-Patterns
* **DO NOT** use the standard library `net/mail` for MIME parsing. You **MUST** use `enmime`.
* **DO NOT** use Bloom Filters for address validation, as false positives violate SMTP reliability. Use the designated **Hybrid Caching Strategy**.
* **DO NOT** store or manage custom application/tenant claims in the IdP. The IdP should remain decoupled from service logic.
* **DO NOT** expose raw AWS credentials or leak DB IDs to the IdP. Use the **Gateway-Brokered S3 Access Control** pattern.
* **DO NOT** allow webhook deliveries to private, loopback, or RFC 1918 IP addresses. Maintain the strict SSRF DNS guard.

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
