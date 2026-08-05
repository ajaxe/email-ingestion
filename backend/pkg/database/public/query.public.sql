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

-- name: CheckDuplicateWebhookJob :one  
SELECT EXISTS (
  SELECT 1 FROM webhook_delivery_jobs  
  WHERE application_id = $1 AND ingested_email_id = $2
);

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

-- name: CreateInboundSpooledEmail :one  
INSERT INTO inbound_spool_queue (id, s3_object_key, status, attempt_count, last_error_message, created_at, updated_at)  
VALUES ($1, $2, $3, $4, $5, $6, $7)  
RETURNING *;

-- name: UpdateSpooledEmailStatus :exec
UPDATE inbound_spool_queue
SET status = $2, attempt_count = attempt_count + 1, last_error_message = $3, updated_at = NOW()
WHERE id = $1;

-- name: GetApplicationByID :one
SELECT * FROM applications WHERE id = $1 LIMIT 1;

-- name: UpdateApplicationWebhook :exec
UPDATE applications
SET webhook_url = $2, webhook_secret = $3, updated_at = NOW()
WHERE id = $1;

-- name: GetIngestedEmailByID :one
SELECT * FROM ingested_emails WHERE id = $1 and application_id = $2 LIMIT 1;

-- name: GetWebhookJobByIDs :one
SELECT * FROM webhook_delivery_jobs WHERE application_id = $1 AND ingested_email_id = $2 LIMIT 1;
-- name: ListIngestedEmailsByApplication :many
SELECT * FROM ingested_emails WHERE application_id = $1 ORDER BY received_at DESC LIMIT $2 OFFSET $3;

-- name: GetApiKeyByKeyHash :one
SELECT * FROM api_keys WHERE key_hash = $1;

-- name: CreateApiKey :exec
INSERT INTO api_keys (application_id, name, key_prefix, key_hash, created_at, expires_at)
VALUES ($1, $2, $3, $4, $5, $6);

