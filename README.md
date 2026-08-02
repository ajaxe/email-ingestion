# Email Ingestion Gateway

> **Note**: This project is currently in the **very early stages of development**. Features and architecture are actively being designed and implemented.

The Email Ingestion Gateway is a production-grade integrated microservices suite designed to handle inbound SMTP traffic, parse MIME emails, securely store attachments, and deliver webhooks to registered SaaS applications. It is built to support high throughput, horizontal scalability, and strict isolation between tenants.

## 🏗️ Architecture Overview & Deployable Units

The Email Ingestion Gateway compiles down to a **single Go binary** that encapsulates four distinct, independently scalable microservices. You can deploy any of these profiles by overriding the run command in your container orchestration (e.g., Docker, Kubernetes):

1. **SMTP Edge Proxy (`email-ingestion smtp`)**
   - **Role:** Faces the public internet, holds open TCP connections, rejects invalid recipients at the perimeter, and streams valid incoming payloads to the internal API.
   - **Scale:** Scales horizontally based on inbound SMTP traffic volume.

2. **API Service (`email-ingestion api`)**
   - **Role:** Handles the heavy lifting of streaming raw emails directly to S3 via HTTP, serves the secure REST endpoints for the developer dashboard, and issues S3 Presigned URLs for attachments.
   - **Scale:** Scales horizontally based on dashboard traffic and inbound email proxy requests.

3. **Stream Workers (`email-ingestion worker --streams email,webhook`)**
   - **Role:** Background engines listening to Redis consumer groups. They pull spooled emails from S3, run the CPU-intensive MIME parser (`enmime`), and execute outgoing HTTP Webhooks to downstream clients.
   - **Scale:** Scales horizontally based on backlog size. (You can split `email` and `webhook` into separate deployments for granular resource allocation).

4. **Cron Scheduler (`email-ingestion cron`)**
   - **Role:** The lightweight singleton loop that sweeps the database for scheduled webhook retries (using Exponential Backoff with Full Jitter) and pushes them into the Redis stream for the workers.
   - **Scale:** Deployed strictly as a single instance (`replicas: 1`) to prevent scheduling race conditions.

## 🛠️ Tech Stack

- **Backend**: Go (Golang)
- **Frontend**: Vue.js
- **Database**: PostgreSQL
- **Object Storage**: AWS S3 (LocalStack for local development)
- **SMTP**: `go-smtp`
- **MIME Parsing**: `enmime`

## ✨ Key Features (Planned / In Progress)

- **Multi-Tenancy**: Logical partitioning in PostgreSQL and strictly partitioned AWS S3 folders using Gateway-Brokered IAM Role Assumption.
- **Secure Webhooks**: Strict cryptographic and network barriers to prevent SSRF and webhook spoofing (using HMAC-SHA256 signatures).
- **High-Performance Validation**: Hybrid caching strategy backed by Postgres indexes for rapid SMTP recipient validation.
- **Reliable Delivery**: Transactional outbox pattern for webhooks with circuit breaking and full jitter exponential backoff.

## 🚀 Getting Started (Development)

The project includes a `docker-compose.yml` for local development, providing a PostgreSQL database and a LocalStack environment for S3 emulation.

```bash
docker-compose up -d
```

This will spin up:
- PostgreSQL database (`db`) on port `5432`
- LocalStack (`localstack`) for S3 emulation on port `4566`
- The Gateway API (`app`) on port `8080` (HTTP) and `2525` (SMTP)

## 📄 Documentation

For detailed technical specifications, please see:
- [Technical Specification & Architecture](docs/email-ingestion-initial-refined.md)
