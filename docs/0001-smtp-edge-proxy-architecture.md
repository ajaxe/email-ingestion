# ADR 0001: Decoupling SMTP Daemon as a Stateless Edge Proxy

## Status

Accepted

## Context

The current architecture closely couples the SMTP daemon with our core infrastructure. When an email is received during the SMTP `DATA` phase, the SMTP daemon directly performs multipart uploads to AWS S3, writes outbox tasks to PostgreSQL, and communicates with Redis for processing pipelines. 

However, residential internet service providers (ISPs) typically block inbound port 25, which complicates local deployment or hosting the main infrastructure on a restricted network (like a homelab). Furthermore, exposing an SMTP daemon—which inherently handles untrusted public internet traffic—provides a potential attack vector. A compromise of this daemon would grant attackers direct access to our core database, Redis, and S3 credentials.

## Decision

We will decouple the SMTP daemon from the core storage and caching layers, converting it into a **stateless edge proxy**. 

1. **Ingestion API:** We will create a new internal API endpoint (`POST /api/internal/ingest`) on the core backend. This endpoint will accept a raw MIME data stream and handle the S3 upload, DB transactions, and Redis enqueueing.
2. **Synchronous Streaming:** During the SMTP `DATA` phase, the edge SMTP daemon will stream the incoming payload (`io.Reader`) directly over HTTP to the ingestion API. 
3. **Blocking Handshake:** The SMTP daemon will block and wait for a synchronous success response (`200 OK`) from the ingestion API before responding with `250 OK` to the sending MTA. If the API returns an error or times out, the SMTP daemon will return a `4xx/5xx` SMTP error, ensuring the sender retries and preventing message loss without requiring a local queue.
4. **Stateless Edge:** The SMTP daemon will no longer contain S3 credentials, Redis configurations, or direct DB access.

## Consequences

### Positive

* **Reduced Attack Surface:** Edge nodes exposed to the public internet will no longer hold sensitive cloud storage credentials or internal database access.
* **Flexible Deployment:** The SMTP edge proxy can be deployed on a cloud VPS (where port 25 is accessible) while the main API and database can safely reside in a highly restricted network (e.g., residential homelab).
* **Stateless Operations:** The edge nodes hold no state or local queues. They can be scaled horizontally behind a TCP load balancer or recreated effortlessly.

### Negative

* **Increased Internal Latency:** We introduce a secondary HTTP hop (SMTP Edge -> Core API) during the `DATA` phase.
* **Synchronous Coupling:** If the core API is temporarily down, the edge node cannot accept mail. (However, standard SMTP retry semantics by sending MTAs naturally mitigate this issue).
