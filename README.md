# Email Ingestion Gateway

> **Note**: This project is currently in the **very early stages of development**. Features and architecture are actively being designed and implemented.

The Email Ingestion Gateway is a production-grade integrated microservices suite designed to handle inbound SMTP traffic, parse MIME emails, securely store attachments, and deliver webhooks to registered SaaS applications. It is built to support high throughput, horizontal scalability, and strict isolation between tenants.

## 🏗️ Architecture Overview

The system operates using the following core components:
- **Go SMTP Daemon**: A lightweight, non-blocking SMTP receiver utilizing `go-smtp`.
- **Ingestion Engine & Parser**: Parses MIME messages using `enmime`, storing binary attachments securely in AWS S3, and serializing metadata into PostgreSQL.
- **Application Control API**: A secure REST API for registered applications to manage routing and query logs.
- **Outbox Worker Pool**: Handles reliable webhook delivery using the transactional outbox pattern with exponential backoff.
- **Management Dashboard**: A Vue.js SPA for developers to manage application configurations, inspect API keys, and monitor webhooks.

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
