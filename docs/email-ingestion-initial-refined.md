# **Email Ingestion Gateway: Production-Grade Technical Specification**

This document provides a refined, production-grade technical specification and system architecture for the **Email Ingestion Gateway**. It addresses open questions from the initial draft, including secure webhooks, multi-tenancy mechanics, storage partitioning, high-performance address validation, and standardizing the Golang/Vue.js tech stack.

## **1. System Architecture & Component Design**

The system runs as an integrated microservices suite structured to support high throughput, horizontal scalability, and strict isolation between SaaS clients (Registered Applications).

```mermaid
flowchart TD  
    %% Entrypoint  
    InboundSMTP[Inbound SMTP Traffic<br/>Port 25/587] --> |MIME SMTP stream| GoSMTP["Go SMTP Daemon (go-smtp)"]  
      
    subgraph Go Gateway Service ["Go Ingestion Gateway Backend"]
        GoSMTP -->|Validate Local Part| CacheCheck{In-Memory Cache<br/>& Postgres Index}  
        CacheCheck -->|Valid| Parser["MIME Parser & Ingestion Engine (enmime)"]  
        CacheCheck -->|Invalid| Reject[SMTP 550 User Unknown]  
          
        Parser -->|1. Write Meta & Outbox Job| DB[(PostgreSQL Database)]  
        Parser -->|2. Write contents.json & attachments| S3[(AWS S3 Object Storage)]  
          
        APIService[Application Control API<br/>Go HTTP Router]  
        WorkerPool[Outbox Worker Pool<br/>Exponential Backoff & Jitter]  
    end

    %% Database relations  
    DB <-->|Read/Write Application State| APIService  
    DB -->|Poll Scheduled Jobs| WorkerPool  
      
    %% S3 relations  
    S3 -.->|S3 Presigned Download URLs| APIService

    %% Outside Integrations  
    DeveloperPortal[Management Dashboard<br/>Vue.js SPA] -->|Auth Code + PKCE| IdP[Apogee-dev IdP]  
    DeveloperPortal <-->|OIDC Bearer JWT Requests| APIService  
      
    ExternalSaaS[External SaaS Backend] -->|M2M Token Request| IdP  
    ExternalSaaS <-->|Client Credentials Bearer JWT| APIService  
      
    WorkerPool -->|HTTP POST with HMAC-SHA256| ExternalSaaS  
      
    %% Styling  
    classDef primary fill:#1e293b,stroke:#38bdf8,stroke-width:2px,color:#f8fafc;  
    classDef storage fill:#0f172a,stroke:#64748b,stroke-dasharray: 5 5,stroke-width:2px,color:#cbd5e1;  
    classDef external fill:#1e1b4b,stroke:#818cf8,stroke-width:1px,color:#e0e7ff;  
      
    class GoSMTP,Parser,APIService,WorkerPool primary;  
    class DB,S3 storage;  
    class IdP,DeveloperPortal,ExternalSaaS,InboundSMTP,Reject external;
```

### **1.1 Components**

1. **Go SMTP Daemon**: A lightweight, non-blocking SMTP receiver. It handles inbound connections, negotiates TLS, performs fast routing lookups, and streams inbound MIME streams.  
2. **Ingestion Engine & Parser**: A Go module responsible for parsing MIME messages, storing binary attachments securely in AWS S3 object storage, and serializing email metadata into PostgreSQL.  
3. **Application Control API**: A secure REST API for registered applications to manage addresses, query inbound email logs, and generate signed S3 download URLs.  
4. **Outbox Worker Pool**: A reliable transactional-outbox runner that handles webhook deliveries to client applications using exponential backoff, jitter, and automated circuit breaking.  
5. **Management Dashboard (Vue.js)**: A single-page application (SPA) enabling developers to manage their application configurations, inspect API keys, monitor webhook delivery history, and configure routing addresses.

## **2. Multi-Tenancy & Data Partitioning**

Each **Registered Application** represents a distinct tenant (an external SaaS platform). Within a Registered Application, there can be millions of **Assigned Email Addresses** (representing the SaaS platform's individual users).

### **2.1 Database Partitioning (PostgreSQL)**

To ensure performance and multi-tenant security, we adopt a **Logical Partitioning** strategy keyed on application_id.

* The applications table stores top-level SaaS registrations, API credentials (hashed), global webhook configurations, and the dedicated AWS IAM Role ARN mapped during registration.  
* The assigned_emails table maps the 10-character unique token to an application_id and tracks tenant-specific metadata.  
* Large transaction tables, specifically ingested_emails and webhook_delivery_jobs, use PostgreSQL declarative partitioning based on application_id range or hash partitions to prevent index bloat and ensure fast query/delete times.

### **2.2 Object Storage Partitioning (AWS S3 & Brokered IAM Role Assumption)**

To prevent cross-tenant data leaks and allow clients to leverage secure, isolated access to S3 without exposing raw AWS credentials or leaking database/service IDs to your OIDC Identity Provider (IdP), objects are partitioned strictly by application_id directory paths:

```
s3://{bucket-name}/apps/{application_id}/emails/{assigned_email_id}/{email_uuid}/  
├── contents.json  
└── attachments/  
    ├── {attachment_uuid_1}.bin  
    └── {attachment_uuid_2}.bin
```

#### **Gateway-Brokered S3 Access Control (Dynamic STS Role Assumption)**

To keep the IdP clean and decoupled from service logic, the Custom IdP does **not** store or manage custom application/tenant claims (like custom:app_id). Instead, AWS S3 folder access is secured entirely by the **Go Ingestion Gateway** which maps authenticated identities to a dedicated, restricted AWS IAM Role assigned to each tenant during registration.

##### **Path A: Interactive Developer Console (User-facing Sequence)**

```mermaid
sequenceDiagram  
    autonumber
    actor Dev as Developer (Vue.js SPA)  
    participant IdP as Apogee-dev IdP  
    participant GW as Go Gateway API  
    participant DB as PostgreSQL DB  
    participant STS as AWS STS  
    participant S3 as AWS S3

    Dev->>IdP: Auth Code Flow + PKCE Challenge  
    IdP-->>Dev: Auth Code  
    Dev->>IdP: Exchange Code + Verifier  
    IdP-->>Dev: ID Token (Standard JWT)  
    Dev->>GW: Request Download (Authorization: Bearer <id_token>)  
    Note over GW: Validate JWT locally via JWKS  
    GW->>DB: Fetch associated Tenant Application & IAM Role ARN  
    DB-->>GW: Return application_id & aws_iam_role_arn  
    GW->>STS: AssumeRole (App-Specific Role ARN)  
    STS-->>GW: Temporary Scoped AWS Credentials  
    GW->>GW: Cryptographically Sign S3 URL using Scoped Credentials  
    GW-->>Dev: HTTP 200 (Short-lived Presigned S3 URL)  
    Dev->>S3: Get Attachment (Direct S3 Stream)  
    S3-->>Dev: Stream Attachment Data
```

##### **Path B: Headless Machine-to-Machine (M2M) SaaS Integration Sequence**

```mermaid
sequenceDiagram  
    autonumber
    actor SaaS as External SaaS Backend  
    participant IdP as Custom IdP  
    participant GW as Go Gateway API  
    participant DB as PostgreSQL DB  
    participant STS as AWS STS  
    participant S3 as AWS S3

    SaaS->>IdP: Client Credentials Request (client_id, client_secret)  
    IdP-->>SaaS: Access Token (Standard JWT)  
    SaaS->>GW: Request Download (Authorization: Bearer <access_token>)  
    Note over GW: Validate JWT locally via JWKS  
    GW->>DB: Map Client ID to application_id & Fetch IAM Role ARN  
    DB-->>GW: Return application_id & aws_iam_role_arn  
    GW->>STS: AssumeRole (App-Specific Role ARN)  
    STS-->>GW: Temporary Scoped AWS Credentials  
    GW->>GW: Cryptographically Sign S3 URL using Scoped Credentials  
    GW-->>SaaS: HTTP 200 (Short-lived Presigned S3 URL)  
    SaaS->>S3: Get Attachment (Direct S3 Stream)  
    S3-->>SaaS: Stream Attachment Data
```

##### **Tenant IAM Role Isolation (Configured at Registration Time)**

When an Application is registered on the Gateway:

1. A dedicated AWS IAM Role is mapped/provisioned for the tenant: arn:aws:iam::{account-id}:role/EmailIngestion-App-{application_id}.  
2. The role possesses an inline IAM policy restricting its data planes strictly to its prefix:  
   {  
     "Version": "2012-10-17",  
     "Statement": [  
       {  
         "Effect": "Allow",  
         "Action": ["s3:GetObject", "s3:PutObject", "s3:ListBucket"],  
         "Resource": [  
           "arn:aws:s3:::{bucket-name}",  
           "arn:aws:s3:::{bucket-name}/apps/{application_id}/*"  
         ]  
       }  
     ]  
   }

3. The trust policy of this application-specific IAM role strictly permits the primary Go Ingestion Gateway Host Role (EC2 Instance Profile or ECS Task Role) to assume it via sts:AssumeRole.  
4. The generated role ARN is stored in the Gateway's database (applications.aws_iam_role_arn).

##### **Path A: Interactive Developer Portal (OIDC Auth Code Flow with PKCE)**

Used when developers interact with the Management Dashboard (Vue.js SPA).

1. **Authentication Handshake**:  
   * The Vue.js SPA initiates standard OIDC Authorization Code Flow with PKCE against the Apogee-dev IdP to acquire an ID Token (id_token).  
2. **Accessing Ingested Attachments**:  
   * The Vue.js SPA calls the Gateway REST API, passing the OIDC ID Token in the request's Authorization: Bearer <id_token> header.  
   * The Gateway verifies the JWT cryptographic signature locally using the public keys from Apogee-dev's JWKS endpoint.  
   * The Gateway identifies the application context associated with the user, and fetches the matching aws_iam_role_arn from PostgreSQL.  
   * Using its host credentials, the Gateway executes sts:AssumeRole for the retrieved tenant-specific role ARN.  
   * With the transient, scoped AWS credentials returned by STS, the Gateway generates a temporary, cryptographically signed AWS S3 Presigned URL.  
   * S3 natively enforces folder isolation because the signing credentials belong to a role restricted purely to s3://{bucket-name}/apps/{application_id}/*.  
   * The short-lived URL is returned to the Vue.js SPA, enabling secure, isolated browser downloads directly from AWS S3.

##### **Path B: Headless Machine-to-Machine (M2M) SaaS Integration**

Used by external backend systems to ingest emails programmatically.

1. **Acquiring the Bearer Token**:  
   * The SaaS backend requests an Access Token directly from your Custom IdP using its client credentials:  
     POST /oauth/token  
     Content-Type: application/x-www-form-urlencoded

     grant_type=client_credentials&  
     client_id={m2m_client_id}&  
     client_secret={m2m_client_secret}

   * The Custom IdP validates the credentials and returns a standard JWT Access Token (no service-specific application_id claims required).  
2. **Accessing Ingested Attachments**:  
   * The SaaS backend invokes the Gateway API to download an email attachment, passing the JWT Access Token in the Authorization: Bearer <token> header.  
   * The Gateway validates the incoming JWT locally against your Custom IdP's JWKS.  
   * The Gateway maps the token's authenticated client ID to its registered application_id using Postgres, and retrieves the corresponding aws_iam_role_arn.  
   * The Gateway assumes this role via AWS STS, obtains dynamic scoped credentials, and signs an S3 Presigned URL for the requested object.  
   * The signed URL is returned to the SaaS backend, which pulls the payload directly from S3.

This brokered setup prevents service data leakage into your identity system, removes OIDC claim synchronization overhead, and keeps AWS S3 security strictly governed by AWS IAM access-control boundaries.

## **3. Database Schema & SQLC Configuration**

### **3.1 PostgreSQL Schema (schema.sql)**

```sql
-- Enable UUID extension  
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Applications (Tenants)  
CREATE TABLE applications (  
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),  
    name VARCHAR(255) NOT NULL,  
    api_key_hash VARCHAR(64) NOT NULL UNIQUE,      -- SHA256 hash of API key  
    webhook_url VARCHAR(2048) NOT NULL,  
    webhook_secret VARCHAR(128) NOT NULL,         -- Key used to sign HMAC payloads  
    aws_iam_role_arn VARCHAR(2048) NOT NULL,      -- Dedicated IAM Role mapped at registration  
    max_retries INT NOT NULL DEFAULT 5,  
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,  
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL  
);

-- Assigned Email Addresses (Unique 10-char local parts)  
CREATE TABLE assigned_emails (  
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),  
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,  
    local_part VARCHAR(10) NOT NULL UNIQUE, -- EXACTLY 10 characters  
    description VARCHAR(500),  
    is_active BOOLEAN NOT NULL DEFAULT TRUE,  
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,  
    CONSTRAINT chk_local_part_len CHECK (char_length(local_part) = 10)  
);

CREATE INDEX idx_assigned_emails_lookup ON assigned_emails(local_part) WHERE is_active = TRUE;

-- Ingested Emails Metadata  
CREATE TABLE ingested_emails (  
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),  
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,  
    assigned_email_id UUID NOT NULL REFERENCES assigned_emails(id) ON DELETE RESTRICT,  
    reference_token VARCHAR(53),             -- Extracted from local-part + addressing  
    from_address VARCHAR(512) NOT NULL,  
    subject VARCHAR(998) NOT NULL,          -- RFC 2822 max subject length  
    message_id VARCHAR(255) NOT NULL,       -- External Message-ID header  
    s3_key_prefix VARCHAR(1024) NOT NULL,    -- S3 base path of contents & attachments  
    received_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL  
);

CREATE INDEX idx_ingested_emails_app_search ON ingested_emails(application_id, received_at DESC);

-- Webhook Delivery Transaction Outbox  
CREATE TYPE webhook_status AS ENUM ('PENDING', 'PROCESSING', 'SUCCESS', 'FAILED', 'DEAD');

CREATE TABLE webhook_delivery_jobs (  
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),  
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,  
    ingested_email_id UUID NOT NULL REFERENCES ingested_emails(id) ON DELETE CASCADE,  
    status webhook_status NOT NULL DEFAULT 'PENDING',  
    retry_count INT NOT NULL DEFAULT 0,  
    next_delivery_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,  
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL  
);

CREATE INDEX idx_webhook_jobs_scheduled ON webhook_delivery_jobs(status, next_delivery_at)   
WHERE status IN ('PENDING', 'PROCESSING');

-- Webhook Invocation History (Logs)  
CREATE TABLE webhook_logs (  
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),  
    webhook_delivery_job_id UUID NOT NULL REFERENCES webhook_delivery_jobs(id) ON DELETE CASCADE,  
    attempt_number INT NOT NULL,  
    http_status_code INT,  
    response_body TEXT,  
    is_retry BOOLEAN NOT NULL,  
    duration_ms INT NOT NULL,  
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL  
);
```

### **3.2 SQLC Query Configuration (query.sql)**

```sql
-- name: GetApplicationByAPIKey :one  
SELECT * FROM applications WHERE api_key_hash = $1 LIMIT 1;

-- name: GetApplicationWithEmails :many  
SELECT a.*, e.id AS email_id, e.local_part, e.description, e.is_active, e.created_at AS email_created_at  
FROM applications a  
LEFT JOIN assigned_emails e ON a.id = e.application_id  
WHERE a.id = $1;

-- name: GetAssignedEmailByLocalPart :one  
SELECT * FROM assigned_emails WHERE local_part = $1 AND is_active = TRUE LIMIT 1;

-- name: CreateAssignedEmail :one  
INSERT INTO assigned_emails (application_id, local_part, description)  
VALUES ($1, $2, $3)  
RETURNING *;

-- name: CreateIngestedEmail :one  
INSERT INTO ingested_emails (application_id, assigned_email_id, reference_token, from_address, subject, message_id, s3_key_prefix)  
VALUES ($1, $2, $3, $4, $5, $6, $7)  
RETURNING *;

-- name: EnqueueWebhookJob :one  
INSERT INTO webhook_delivery_jobs (application_id, ingested_email_id, next_delivery_at)  
VALUES ($1, $2, CURRENT_TIMESTAMP)  
RETURNING *;

-- name: GetPendingWebhookJobs :many  
SELECT * FROM webhook_delivery_jobs  
WHERE status = 'PENDING' AND next_delivery_at <= CURRENT_TIMESTAMP  
LIMIT $1;

-- name: UpdateWebhookJobStatus :exec  
UPDATE webhook_delivery_jobs  
SET status = $2, retry_count = $3, next_delivery_at = $4  
WHERE id = $1;

-- name: LogWebhookAttempt :exec  
INSERT INTO webhook_logs (webhook_delivery_job_id, attempt_number, http_status_code, response_body, is_retry, duration_ms)  
VALUES ($1, $2, $3, $4, $5, $6);
```

## **4. Email Ingestion & Parsing Deep Dive (UC:2 & UC:3)**

### **4.1 Optimized Library Selection**

* **Go SMTP Server Daemon**: Use github.com/emersion/go-smtp. It is the industry standard for Golang, actively maintained, supports pipelining, LMTP extensions, and features a clean, stream-oriented backend interface.  
* **MIME Email Parser**: Use github.com/jhillyerd/enmime. Compared to the standard library net/mail, enmime is heavily optimized, gracefully handles nested multipart layouts, parses malformed headers, safely decodes complex character sets (e.g., ISO-8859-1, EUC-JP), and extracts attachments directly into memory-efficient standard streams.

### **4.2 Inbound Address Validation Optimization**

Using a Bloom filter was initially proposed to check address validity. However, in transactional mail delivery, a Bloom Filter's **false positive rate** means the server could incorrectly signal HTTP 200 or accept a mail context for an unassigned email address, only to silently drop it later. This violates SMTP reliability expectations.

#### **The High-Performance Hybrid Caching Strategy:**

1. **In-Memory Cache (Redis / Ristretto)**: Keep a fast, thread-safe in-memory cache of valid 10-character local_part values.  
2. **Execution Flow**:  
   * During the RCPT TO:<local_part+ref_token@domain.com> phase of the SMTP handshake:  
   * **Step A**: Strip sub-address parameters. Extract the base 10-character string before the + sign.  
   * **Step B**: Check the in-memory cache.  
   * **Step C**: If missing, run a fast read on the Postgres assigned_emails table index.  
   * **Step D**: If found, cache it and permit the SMTP stream. If not, reject the RCPT block immediately with a 550 User Unknown response.  
3. This ensures invalid emails are blocked at the TCP perimeter, preventing expensive MIME parsing and storage of spam.

```mermaid
flowchart TD  
    Start([SMTP RCPT TO Phase]) --> Extract[Extract 10-char local_part]  
    Extract --> CacheCheck{Exists in RAM Cache<br/>Redis / Ristretto?}  
      
    CacheCheck -->|Yes| AcceptSMTP[Accept Recipient<br/>Permit SMTP Stream]  
    CacheCheck -->|No| QueryDB[(Query Postgres Index<br/>assigned_emails table)]  
      
    QueryDB --> DBCheck{Local Part Found?}  
    DBCheck -->|Yes| WriteCache[Write to RAM Cache]  
    WriteCache --> AcceptSMTP  
    DBCheck -->|No| RejectSMTP[Reject RCPT TO<br/>Respond SMTP 550 User Unknown]  
      
    classDef default fill:#1e293b,stroke:#64748b,stroke-width:1px,color:#f8fafc;  
    classDef decision fill:#0f172a,stroke:#38bdf8,stroke-width:2px,color:#f8fafc;  
    classDef terminal fill:#1e1b4b,stroke:#ef4444,stroke-width:2px,color:#f8fafc;  
    classDef accept fill:#064e3b,stroke:#10b981,stroke-width:2px,color:#f8fafc;  
      
    class CacheCheck,DBCheck decision;  
    class RejectSMTP terminal;  
    class AcceptSMTP accept;
```

### **4.3 Detailed Parsing and Storage Pipeline**

Upon accepting the message body, the SMTP daemon executes the pipeline:

1. **MIME Stream Parsing**: Stream the body into enmime.ReadEnvelope().  
2. **Metadata Extraction**:  
   * Extract Subject, From, To, Message-ID.  
   * Parse the envelope recipient to extract the sub-address reference token:  
     * Format: [10-char local_part]+[up-to-53-char ref_token]@domain.com  
     * Validated using regular expressions to prevent injection attacks.  
3. **Structured Payload Assembly (contents.json)**:  
   Create a single JSON payload detailing the email contents:  
    ```json
   {  
     "messageId": "<id-header-from-email>",  
     "from": "sender@external.com",  
     "to": "a1b2c3d4e5+ref994823@domain.com",  
     "subject": "Invoice Processing",  
     "bodyHtml": "<h1>Please see attached invoice</h1>",  
     "bodyText": "Please see attached invoice",  
     "attachments": [  
       {  
         "attachmentId": "9a2f1b40-9fe5-4c07-ba77-626a5704d9c1",  
         "filename": "invoice.pdf",  
         "contentType": "application/pdf",  
         "size": 142850  
       }  
     ]  
   }
   ```

4. **S3 Storage Engine Write**:  
   * Store contents.json in the targeted application folder.  
   * Store attachment bytes as separate object payloads at the companion key structure attachments/{attachmentId}.bin.  
5. **Database Sync & Outbox Trigger**:  
   * Insert metadata into ingested_emails.  
   * **Transactionally** write to webhook_delivery_jobs (Transactional Outbox Pattern) in the same DB transaction to guarantee atomicity.

## **5. Webhook System & Security Engine (UC:1 & UC:4)**

To prevent security vulnerabilities like Server-Side Request Forgery (SSRF) and webhook spoofing, the callback gateway implements strict cryptographic and network barriers.

### **5.1 Callback Endpoint Security Verification (Handshake)**

When registering or updating a callback endpoint, a validation handshake prevents arbitrary endpoint subscription.

```mermaid
sequenceDiagram  
    autonumber  
    actor Dev as Developer / Admin  
    participant GW as Go Ingestion Gateway  
    participant DNS as DNS Resolver (with SSRF Guard)  
    participant SaaS as Candidate SaaS Endpoint

    Dev->>GW: Request Webhook Registration (Callback URL)  
    rect rgb(30, 41, 59)  
        Note over GW: SSRF Prevention Check  
        GW->>DNS: Resolve Endpoint Domain Name  
        DNS-->>GW: Return IP Address  
        alt IP matches Private, Loopback, Link-Local or RFC 1918 space  
            GW-->>Dev: HTTP 400 Bad Request (Prohibited Target Destination)  
        else IP is Public & Routeable  
            Note over GW: Generate Secure Random challenge Hex Token  
        end  
    end

    GW->>SaaS: POST /callback-verify {"challenge": "8f92bd3a09e0..."}  
    Note over SaaS: Receive JSON, extract challenge string  
    SaaS-->>GW: HTTP 200 OK {"challenge": "8f92bd3a09e0..."}

    rect rgb(15, 23, 42)  
        Note over GW: Match Echoed Challenge to Generated Value  
        alt Challenge Matches & Response within 5s  
            GW-->>Dev: HTTP 201 Created (Webhook Registered Successfully)  
        else Challenge Mismatch or Timeout  
            GW-->>Dev: HTTP 400 Bad Request (Endpoint Handshake Verification Failed)  
        end  
    end
```

1. **Challenge Issuance**: The Gateway issues a transient, cryptographically secure hex-encoded string (challenge) and writes a JSON POST to the candidate callback URL containing:  
   ```json
   {  
     "challenge": "8f92bd3a09e03949a882e987c9bc348c"  
   }
   ```

2. The endpoint must respond within 5 seconds with an HTTP status code 200 OK and a matching response payload echoing the exact challenge.  
3. **SSRF Guard**: The HTTP client resolving external callback URLs restricts connections strictly to public IP spaces. Local loops, RFC 1918 addresses (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16), and AWS metadata endpoints (169.254.169.254) are blocked via custom DNS resolver hooks in Go.

### **5.2 Delivery Signature (HMAC-SHA256)**

Every webhook delivery includes a signature header to verify integrity and origin.

1. During Application registration, the system generates a secure, random string webhook_secret (prefix: whsec_).  
2. When dispatching webhooks, the Outbox worker computes an HMAC-SHA256 signature over the payload concatenated with a millisecond timestamp to prevent replay attacks.  
3. **Webhook Signature Header Formulation**:  
   * Header: X-Gateway-Signature  
   * Format: t=1782806400000,v1=f6c8d7e63b...  
   * Signature String calculated over: timestamp_string + "." + json_body

mac := hmac.New(sha256.New, []byte(webhookSecret))  
mac.Write([]byte(timestampStr + "." + string(jsonBody)))  
signature := hex.EncodeToString(mac.Sum(nil))

### **5.3 Reliability & Retry Engine**

The Outbox runner processes scheduled jobs concurrently. If a client server returns a non-2xx response or times out (10s threshold), the system reschedules the webhook using an **Exponential Backoff with Full Jitter** model.

#### **The Backoff Formula:**

The time wait (![][image1]) before retry is governed by:

![][image2]![][image3]Where:

* ![][image4]  
* ![][image5]  
* ![][image6] is the current zero-indexed try count.  
* Jitter prevents a synchronized thundering herd against client applications.

#### **Circuit Breaking:**

If an application's endpoint fails consecutively over 100 unique webhooks within a 15-minute window, the system flags the Webhook status as **SUSPENDED**. Admin notification triggers, and deliveries queue in a PENDING state until manual activation.

## **6. Email Gateway REST API Specification (UC:5)**

All endpoints expect JSON request payloads and return unified JSON error responses.

### **6.1 Authentication**

Clients pass their issued API Key or JWT Bearer Token in the HTTP header:

Authorization: Bearer <JWT_Token>

*(Note: API keys are validated as fallback tokens by lookup in PostgreSQL, but standard communication utilizes verified OIDC bearer JWTs).*

### **6.2 Endpoint Inventory**

#### **1. Register Assigned Email Address**

* **Endpoint**: POST /api/v1/addresses  
* **Description**: Generates a new unique 10-character email local-part mapped to the tenant application.  
* **Payload**:  
  {  
    "description": "Customer Sandbox Account #4"  
  }

* **Response (201 Created)**:  
  {  
    "id": "e3a891ba-77d1-412e-8395-64bc410d0fba",  
    "localPart": "a8f3g9j2k1",  
    "emailAddress": "a8f3g9j2k1@ingest.yourdomain.com",  
    "description": "Customer Sandbox Account #4",  
    "isActive": true,  
    "createdAt": "2026-05-31T16:01:00Z"  
  }

#### **2. Get Application Info & Active Addresses**

* **Endpoint**: GET /api/v1/application  
* **Response (200 OK)**:  
  {  
    "id": "8aa121fe-cc2d-45db-9c3f-c5643ef09d12",  
    "name": "Acme SaaS CRM",  
    "webhookUrl": "[https://api.acme-crm.com/webhooks/emails](https://api.acme-crm.com/webhooks/emails)",  
    "maxRetries": 5,  
    "addresses": [  
      {  
        "localPart": "a8f3g9j2k1",  
        "emailAddress": "a8f3g9j2k1@ingest.yourdomain.com",  
        "isActive": true,  
        "createdAt": "2026-05-31T16:01:00Z"  
      }  
    ]  
  }

#### **3. List Ingested Email Logs**

* **Endpoint**: GET /api/v1/emails  
* **Parameters**: page (default 1), limit (default 20), localPart (optional filter)  
* **Response (200 OK)**:  
  {  
    "emails": [  
      {  
        "id": "4b68e412-f12a-45c1-840a-9d663ef01041",  
        "localPart": "a8f3g9j2k1",  
        "referenceToken": "invoice-9923a",  
        "from": "accounting@client-corp.com",  
        "subject": "May Invoice #9923a",  
        "messageId": "<msg-923184@client-corp.com>",  
        "receivedAt": "2026-05-31T16:05:12Z"  
      }  
    ],  
    "pagination": {  
      "currentPage": 1,  
      "totalCount": 142  
    }  
  }

#### **4. Fetch Secure S3 Download Link for Attachment**

* **Endpoint**: GET /api/v1/emails/{emailId}/attachments/{attachmentId}  
* **Description**: Returns a short-lived, cryptographically signed AWS S3 download link for the binary attachment.  
* **Response (200 OK)**:  
  {  
    "attachmentId": "9a2f1b40-9fe5-4c07-ba77-626a5704d9c1",  
    "filename": "invoice.pdf",  
    "contentType": "application/pdf",  
    "expiresAt": "2026-05-31T16:21:00Z",  
    "downloadUrl": "[https://ingest-bucket.s3.amazonaws.com/apps/8aa121fe/emails/a8f3g9j2k1/.../invoice.pdf?AWSAccessKeyId=AKIAIOSFODNN7EXAMPLE&Expires=1782807600&Signature=vjbyPxybdZaNmGa%2ByT272YEAiv4%3D](https://ingest-bucket.s3.amazonaws.com/apps/8aa121fe/emails/a8f3g9j2k1/.../invoice.pdf?AWSAccessKeyId=AKIAIOSFODNN7EXAMPLE&Expires=1782807600&Signature=vjbyPxybdZaNmGa%2ByT272YEAiv4%3D)"  
  }

## **7. Frontend Interface Specifications (Vue.js)**

The frontend is implemented as a functional developer portal in Vue.js.

+---------------------------------------------------------------------------------+  
|  EMAIL INGESTION GATEWAY                                           [User Profile] |  
+---------------------------------------------------------------------------------+  
|  [Dashboard] [Active Addresses] [Ingestion Logs] [Webhook Config] [API Keys]     |  
+---------------------------------------------------------------------------------+  
|  API Keys Management                                                            |  
|                                                                                 |  
|  Active API Key:   eg_live_a1b2******************** [Regenerate Key]           |  
|  Webhook Secret:   whsec_7d2f********************** [Reveal Secret]           |  
|                                                                                 |  
|  -----------------------------------------------------------------------------  |  
|  Webhook Settings                                                               |  
|  Endpoint URL: [ [https://api.acme-crm.com/webhooks/emails](https://api.acme-crm.com/webhooks/emails)                     ] |  
|  Max Retries:  [ 5 ]                                                            |  
|                                                                                 |  
|  [ Test & Verify Endpoint ]                                        [Save Changes]|  
+---------------------------------------------------------------------------------+

### **Key Modules:**

1. **Interactive Sandbox Console**: Allows developers to view live webhook payloads, inspect JSON formatting, and review processing histories alongside outbound webhook delivery attempt statuses (HTTP codes, payloads, retries).  
2. **Address Management Panel**: A simple layout where users can quickly provision new 10-character email addresses, monitor current addresses, and toggles to dynamically pause/resume individual routing paths.  
3. **Delivery Log Analyzer**: An operations-oriented view displaying real-time delivery performance metrics. The screen graphs latency, summarizes webhook error patterns (e.g., categorizing 5xx servers), and features a "Re-deliver Webhook" trigger for manual debugging.

## **8. Development & Infrastructure Setup**

### **8.1 Docker Compose Configuration (docker-compose.yml)**

version: '3.8'

services:  
  db:  
    image: postgres:15-alpine  
    container_name: gateway_db  
    environment:  
      POSTGRES_DB: email_gateway  
      POSTGRES_USER: gateway_user  
      POSTGRES_PASSWORD: production_secure_db_password  
    ports:  
      - "5432:5432"  
    volumes:  
      - postgres_data:/var/lib/postgresql/data  
    healthcheck:  
      test: ["CMD-SHELL", "pg_isready -U gateway_user -d email_gateway"]  
      interval: 10s  
      timeout: 5s  
      retries: 5

  localstack:  
    image: localstack/localstack  
    container_name: gateway_s3_emulator  
    ports:  
      - "4566:4566"  
    environment:  
      - SERVICES=s3  
      - DEFAULT_REGION=us-east-1  
    volumes:  
      - "./infra/localstack:/etc/localstack/init/ready.d"

  app:  
    build:  
      context: .  
      dockerfile: Dockerfile  
    container_name: gateway_service  
    depends_on:  
      db:  
        condition: service_healthy  
    ports:  
      - "8080:8080" # Web API Port  
      - "2525:25"   # SMTP Port mapped locally  
    environment:  
      - DB_DSN=postgres://gateway_user:production_secure_db_password@db:5432/email_gateway?sslmode=disable  
      - AWS_S3_BUCKET=ingestion-store  
      - AWS_ENDPOINT=http://localstack:4566  
    restart: unless-stopped

volumes:  
  postgres_data:  


[image1]: <data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACgAAAAaCAYAAADFTB7LAAADj0lEQVR4AeyWW0jTcRTHt79uzm1J3kHH3AbdLCsa9FAUlPUQCEp7jIIIKil7EOyhh54qgqioEEqonrpAUfTgSwQ9hEXlIjMQq903a82WVkuxXfr8RsrmNpxOVoHjfHeuv985v/M//98myf7xz0KBuT6g/6ODdXV1W4FNr9e7s8S2XDuT7XrRQTnBO+Vyea8kSZvdbncd0KM/BDps+4UOjMi7sUVjsdgoa/JCktForCJpVUFBQZuTD1ljdLEUvga4w+FwP1xQJBQKWRGeRKPRj/C8kBSJRBrI9Mxms32Gx4li9QjLQK/X6/XD41RcXKxA+KrVavPXQRKW8ui64VNEh1ahLOJRvoSHQZw4jITt7eDg4I+4IQ9fErN1x263v5uWawu6KKwHPkU+n+8L8V0YYiAvJF6SpEQJ8+ejk44k519QUgqkBhOoB0nzh54z1dbW6rjOzgqe7WYpBfJGL2exmllLmj9sOZNSqSxnk/WFhYWL4ILkFHyosrJSK5R0mF6gnMLEJZwyf+kWz9bmcDj6XC7XJjAg1oorjnziFhFqWiQVyPwtpoPiDU6ZP4PBoOK0HaCLuNPg6KSMby248AcwQwu+cwhXQKvZbFbAJ9ff0+l0Daw3MeO3yGfh+rpE/L50FSYVSEDG+eMOH+fk54kpY9N+3uYzyBrkF/heIzvBzbGxsQBdOYDsxX4QuX54eHgP8rhKpbqMXcm1pmS9HV8H6GFNG3tfxZdCkslkWkr1r0CQZL1EqIEFfQQMcFIz+iSFibkLmhn0Mow1oKm6ulpDInHYPr/fHyLZDoq4wdom/GX4KuCyYDCIGBPjI9SsIIk7kA3XgTIgT4AWeQUnFT9vU5uR4TkwMegt8GughuFvJGBAdEk8Tg7WyeM7zsX+FP8bfDMSj30JQYUgicSpkwwzKRTtJKn4fRZ/HB7Rzfd0ay/r4oMfCARWI28EJ8TFjr8EaCh6F7aMxIEbeZtV0wNmXSAbUF/sPryfYj+hPKCAb8gebDISfcDmQm7nxWiFK4GFmHK1Wn0EvgEcw2dQKBQe5BKNRnOYmBEOl/ITOpcCZR6Pp7uioqKdTWOMwOOJiQlRSHy2GJlRbM0kPon/OoV3gJU8/ovwU6AKWNCd4g8K8vaioqJO9NvEp9CcCmSXiNVq/QUXFBkaGvophESQcASMJ9oyyBH+fHzP4JPNtcBM+827faHAXFs6Tx3MtYzM638DAAD//2sa0HIAAAAGSURBVAMAbCK6RJPAcuAAAAAASUVORK5CYII=>

[image2]: <data:image/png;base64,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>

[image3]: <data:image/png;base64,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>

[image4]: <data:image/png;base64,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>

[image5]: <data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAQMAAAAaCAYAAACzfzksAAAQAElEQVR4AeyaCXhVxRXH70uAbA9ZAglbSAIoRZDFAAIKVESFL6W1oohApYqKpYqKIlWMpguosVKlAoKCCBZlLagtVFxA3EBwwQXZAoRNtohAIECW/v6Xd9/37nsvycsGmO/mm5OZOXPmzMyZM2fOnPvCDOfPkYAjAUcCSMAxBgjBSY4EHAkYhmMMHC2oEhJo1qxZrZSUlOpVYjH2RYQ3bdq0DigXUKnJMQaVKl6HeTAJJCUl1U4CjAr6S0xMTM3Pz3+sItjVr1/f3bhx41h4VfrhY4xQUoHL5boNed0HcchzgrbUyTQGCLMXsBULlBUi9C71SOd3h3DWnwpMBiYmJCRcU9Qt06JFiwuQ0b3QTQMeUd1vaS427jLaJgLil0p7OGBL6ufDJw3ahjaCKliRXFlnfmFh4Y8s7zmg3Al+HWDyYLVq1Z5at27dacq2xF4oDbQhg1SaNGlyCbz2R0dHH4XXQoxCTBCyc4EqrFev3kRk1oH5XV+ZE5AxkLW5HuuzNiwsrEdWVlYi0JT620ATcHeqDiRT/h24Aib2U2VO6mzyZtPdCPl1xkxGCdILCgpmsc4pBw8efEOuJ3hvgq7X6dOn14GQMj9OTvX0TBQpirKSiwM+Gh4ZVP6Rl5eXRj6IflN8jUtycnIiHVfRdhx5jiRfi0xXgu9MucqmnTt3vo1M4ljgF0C5k+SO3NKR4dStW7futxgi71bACOA92reC7wsUm3bt2vU1h64xRIuB8yrJyHl06sFGjRolVNbkwlDAOIQZFx4efs92/hioEIWuQ94OyGLzviZXys/JydFBWMXE9gpRFSAqKqoL6+iP0rRGobJR2LXUZRz6sE7vjYJMWoOfjawexTDOwmAUUL8LSEF2tckNaC6l/QHKY3fs2LFt9+7dh+Arg3DtoUOHrgKvVA2+wm2Gz8uIPBfapfSbj6v7Fym4iKoqnDp16iRrOwKUO2G8eyK3JOB9f2bg9oB7FAhZV3Xo6Fchc2PcCk3o5fcw3F6jRo1h5JWSwlDAS+D8CQfBa1lR7qbgWgJrsZj7yM3EwVGA5ke3211lPAMOdR6LK0QJGuBPVqOslKN/HGTLVaxG+/3gDqPM75Ib27Ztk7z0Tv0zh/mAcNDIjcsl120klIHM1JaFAbgJhDwHyTYV3p9R19hkZlpHv+7M52Kz5vwrSQIuj0xXshdeHVUn9mMDRlY3/Dbq+cDPOnkmL115C71J5dLRZe1BV1ymZ0IdFPA/viwRchvqNRnYprAYjjBw32zcuPEY7VUioTQr8X7kHd1MOVc3M2vsweKOk68kN3A3E8j7ABv0pmzevHkcQaa6KN0cYAb4PAxJJHl7wJays7MLQeRz0NuwifIgmlGPB2yJsfTejYbuF7YGn0p8fHwMc+kLPIBH1yMhIaETsYf6PiQu8O1oT0tKShpKHhCH0HMFvGJE48iHsd4WPv1VVPQ6hbZRzHcI61QwTXhBuNbOuG1o70vfKMaLh64feQ/xFpEfuHhuXQTNDfTrGBERUcOv3azK/dWcgeHwvQT6K2iwjDNFe4I+FlnpWbXC3lIxtVq1aoVZ82YuKXANiPtovVo3sngEmn7aH+jM5GmLp70za/olSK3FlB/ru5A+qeR1wbskY8omjnXVg1cK/QLkyTn9Bvo4dMV/z0CXP4Xhqs7PzMzc5MfqSuqyRB+Re5PcXuingZCCk1WJVKh1yRCwGpyi8EEoWXfKY1jr5+QGwm9O3giIxHBMBIbgoj7Phq5iI5uAN3hCabMtT0Iof5DBkcFQfEFxGv92s85YJj+z4vOPsVpFRkb+j7n9CHoOhlnu8UI8FXkahhQRJZqBIU+nfQG59uhL+slbAWXIqLUiFvIVlYH0n8lYMSx4I4f01+AMlLYBPP5L+UYUbyGwi3W+B+4WcC7aa9LvefDi8SJtirHcTT2G8Z6E91IOUC1ozaQy48+j7VXmHQFdO+A1GhX0IzuT4H9H9erVM6BZASxjTn8g/yvxHMnrDJHff8ZOZP66sHb4NVVENRK5ZrDWX8FMBnoe67DFfahLlh+ztrbog56N0ezPZ+B70cfgWdiY/lr7auaZrrVIfsSKJrK+b6GZhyySwUWwljHgFEeZzzNgEm23w3fZgQMHnqbs1RVw8kYN+ulyoqlikzwDG0c2pg4IxQt2M7jcLKrnR0JwNzO/UL94ZLExa7G4IVlReN8Fb613PKt9Mjc392VyHSgywzrAVyKTaXgDEwg26YAcQIlnsNFuPIaaECYDJaWyPgMUBNuB0VrN+HsxVLM4MDOtwSIiIu6mfjX1UbRvoH025fnA4xisWB1M2qeimFko6b08/zZT13PIhXLVR05RtL1I+SfWk4brrbFWsF7xfQ7Z9Gbsw8AAeCrm0Rj6nYyVBk4xlr+D78XB6ERu6Gak/CxjKFjaF7p/AdMZexDt3meUjBg0ekLN15gCDsxj8NYbGdIiUyP6mc/WIinK3tAZOSxmvhOQ4yzYTAf6Y+xMXZJ3RF2yXQbNJC6TXchgLnOW7ryGrFpT305feQRzoDUTuMPgBkIn+Vk4xYweAnczCOnZduQ2nrq8csnJ0kHj+PHjR6HZRtuF5BWeAowBI8iNlcLa4gXgz3lCmK8hzKahAhvVEaXfEsrE4f2C+KKsSdC3w8qvx5i0ouybVqAksuqGgk1syqdAb+ICXcAroChvypc+oIwC6zkQgC8JQT8d3IHMaSIuZGcOrBuD9RT5ermWtMs4fUVuBcykRO/BtwW3ThIHrCdzVUzi38jkBHgD+cg7iCWfwfwvB6fPoO9qbZTNBF7yOwLf4SDk/chT0js8m7oZPyEXTuvSLSaFNrjV2tJnAGMu57AcEo2AG9cWQER24iXcJNY2AqN0IYb1KPiH4KE1q1tRcIoxzLUURVBG/AbmqUCy2Z01aI6R5NFCoCO/Jb+Yup6RkjNVM23gv5s5yeBRNBSLsumEkMUBxvfDPXv27EQXewL/DEYL/wuC4cuLCzAGDKQ3azQLlWUq1ULKO5nzob/noMiayyhmoJxRyMSUA/lBjEau3zx1AFKOHTsmxd3p1xZQRa6bA5AhIDiUun0XQXo3CrPac2AUwDRwMxPBy4hJQfX7Bv0GQs+5Psz5JfoeIjcNG30VZYfcTPkonJ4dUtq2YBSU820H5U0dMEK+8QPFjQ57WwMLelZJj4qNzkuezG0scpExmYTh2oTh2o+hU8zA96AFjlB5mAPMwX+fvaMxV3k/MkKSnRfvU7hMHo9PvaKLATGnihjA3xi4WKh+UCTlt8ULKmKw8vLAlY/k9mgYKqC88XJXixsXl+4q+E3Alfa+1ZGB3qHH6dcW5azNYdIBkYsGKnjiFpPyHAzeamKz8DZ0gKRAkq+J9P/HwQhqLDIzM3/iBh9AezLwe/qtYJ6jeZsOo665avx10IwA7rSAw3YvsB36YhM8ZMwM8qJunWxuxFPFMiljI/P7Eg+nOYZKh2wsbGRkZrE3+pxL9bxLkrcCoaYXFGR2h/ft2ydvJ0hTsSjxla4VS8S+B9WRYjuF0GgzBgi/NsqgLwlB4wUcmmHQLAZu4WCmU18O9Ae6UZ9LPt06VIpyU/8bkAG8xJv1Is2HsqLh5s1Fn/bUhwGqj6ZeZMBIfZlbPIK4vBTQhU2pqb7BAAWUS6f32/0c+sE+NOatRv0YN8RpxtVh+o5xFSA0XWXarKTbS7+/yKNd3pSbhtqAmYgpaE2K+H/BV5ijKLziErvhqR+4mDT6h8HRmHK95WoKZQNkNJpAXz8OjtIrxCyugccCiLpySLd7xk5kb2zrxSC2015A+wG0cqv1uwqKZ5JuMNHQfz0YHfZEcm8CL4WX4q/Gi9Ah9bYVV2A8PS+yyW3r9O+jPWDOL7NPcXxLX4sRG8969JsMeQryzvy7WHUZ1erITT9isnBnJWdNHzOQPjf7j23JSl84ijP4AV8m4Fdi4umkGEkM40s2JdKXlsBmDOgs4V9MHjRewEbNYCI5QJ/Y2NhxbMSD0E4EGqChCgJtxWV9iLpBECQFRRoAzTMcgGfZ4MkYilh4LAU3jjYFKSOgjYDf++AnwEO3G6jgScElFHJBqAC/Jbj9OmBBGXpu8y3MZTPGYKmHiOm49MlK1Rm83w7qVgap93Vb3ufm5zp5HOC6AO+cOHHiUxGzzoXkJ8BpbRQNA75yz2UMFEU3mI8sv4J7nTGQZuRdvCDWAXgbOQS1+swxDF5DPLSKWcj47GfM7+Apl/UZeLRHtr8hdwGG5E37cOZXgPHQG/gV8LdidPUUpGgYeCvXwbsL435Grv29Tv3MxjP/upHpk/IUchk+eQ9S5nD2WDnowAS/7+GndaYynp4wJhEKLdl0os1rWJFXEvPWFyyTBgMqL0zemWRl4vz/0ceMQ6BXkq1/s63OWNVAmDIhLy7JMxZteN26dYuk58vRW4y/HBho7QdMqbquJd/IGZhHbibG1m9yEtxut9ZraM8h1GdqjeV//vTlwt/AmHz0j7WKR114ynALVaFgfkvl1vkckBWXwihIotv+MLgNWG19Y7UG1dtSSviVgkwo2n4WthWlWCMCJqnv6aaCcxCXYfVb05YATVfaayEk3TD6wc4OcBn01y/6CqHVe1hBGsjOamJ6eeko4g/MJY31Dma9igiP0fw4QDJ05oQ4UFLsj1DUJSj3UCLLijI3g+4OjIrcfx103YYjkcMT8LlTdPCeCoN0DodpMCgXcqgVec9mTHlEfeGlKH8jJjOa9qA3CuNAXpAI7SrmOQ3++lFNHWT6An0MDORyaHoz9ljaPwCm0/Y6uGkK4Gm/eIffD+0CaNYwN3lyikG0xLCbv5VggFHAAvqZ/eEhpf8j8+0O/28VRaef9nocfBpDlwmNPLq5jKP5yFAshmZufHx8NLwehm4R470L3XRgEXRDwW0CpGPb+QrSmvYjyGkka1oCzTT6LaF9NmOan3YpByQi64rPbKKfPvva2uHTGz67QAr06XUQ9aPMaz0xIP3IjiZ7Eh4a7Z+Cf1cgq73wGQLuE+b8JNQ6qJ9Qn6D9Zu03MO997McacNoPeQOtoO3t+wM+DvCL9GVJBTNFR/0F+r0BLoq5q+9k5uWVH/i3oPtG86FsS+yDjGouzIJ6jzbiMlTCuPU2oaiXAvoRjYvcAjflVmyIXOAiWbOwAgST70+AS9sRoX2IcOogAN265pvUokPxV9CmA6FfOlros55zUPRZSLeSPmfpnfcxc2uIgRqjA2RNCDc2B1ncxoYMZDNy2MgpGIuO8lYsGuXIbBH9O1LeC1025e7gJlM3b1VyA5nr/d+P9gzkFwNMgqan5qL2YMBzZRLjd2NMfRFIwzgNps9g8bLomfNKcBcx5nXsycOU+4D70mrXGsCNRO764cp9HKhbqD9mrVMeBmNkMIY8jEegG0J7V8bQ4TWk5PDrDM7SkRqUnwZ3E3k4IHy46hpL/MCNho/e/o+SD6I+nPn1plyHchNo1pw8eXIQfTprTcg1jfEDZGatvdYjvwAAAqdJREFUwcrRrWPwUNxEn1N1m1tNMozviDdgzUnzcjNGW8azfl7vpVdBeOibA6IVuJHFq9S7AqoLtN5Rot+yZcsR8PcwB110jzP3VOo3MMYParfAw1d7MoR1a0/0Oxb9VqMB6xS/EfSxyQ8+bdTP4mHlyEbxvPW02X5xabWXN/d3U8rLz9ufQ3MPlUUIdDlleQSCbrJ4cq3A3c7ihiNM/fLK+8MY+pyLlM8G6Nu8niBvcihNFzTIRAp1MFjTAozAB9Yh8qdTf2jeFKjs3+6pK5K/DhqNKYMbYFA9dGbGc0XBpXyNydj7dBjMhsB/5o+odHBpCsoT5dO37b1F8bDGgC7kGAFjFZngY46nHCJzfpQt3oUyHOD1OfKY1qbxVS8JOFyvQtOCi8f77KF+VhPrMNdWlCw9kzHX7NEFecK5pVmneOjpxqXRgwtEF0tQ71F05YFSGQPcmeFM6BoO8I24MrJw6QzeibqCiarfSHsP2oaRy9W7nj5DadcPKuQZDOLGulSuFbiB3HaKvkspFDl+Flpv4A2+TnIkUKwEuCHl1s/mUrkNwiLf+LSVmM53Ai7PvpypTXFxcdZzs8KnXCpjgBWcyg3aEOgIzKE+nDzGk+t3+sIng5sOLCJmcAXW+03K44Grgf7QvkLeQcBm7iaXuxvDDXkfbTIMhvPnSCBUCeBqK65TnQvoXHuXoU651HR40y24PAdjEJ4I1Wsq9SB0KJUxgL5USRPnwCua730vl4qBQ+xIoAQJSMcI9v0JsvYE/PQ1jGLVSS1btqyJEbiVS/UOnhb6wlJpi6tUY1Bps3YYOxLwkYBiDniYaXiXmT7oKlHUb1NY21jiDfoyUqlrcoxBpYrXYX4uJOCMWTYJOMagbHJzejkSqHIScIxBldtSZ0GOBMomgf8DAAD//yfHGuYAAAAGSURBVAMAj2f3vNfqmHsAAAAASUVORK5CYII=>

[image6]: <data:image/png;base64,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>