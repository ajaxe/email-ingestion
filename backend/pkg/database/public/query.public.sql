-- name: GetApplicationByAPIKey :one  
SELECT * FROM applications a
JOIN api_keys ak ON ak.application_id = a.id
WHERE key_hash = $1 LIMIT 1;

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
INSERT INTO webhook_logs (webhook_delivery_job_id, attempt_number, http_status_code, request_payload, response_body, is_retry, duration_ms)  
VALUES ($1, $2, $3, $4, $5, $6, $7);

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

-- name: GetAdminApplications :many
SELECT * FROM applications;

-- name: GetApplications :many
SELECT a.* FROM applications a
JOIN user_application_access ua ON ua.application_id = a.id
WHERE ua.user_id = $1;

-- name: UpdateApplicationWebhook :exec
UPDATE applications
SET webhook_url = $2, webhook_secret = $3, max_retries = $4, updated_at = NOW()
WHERE id = $1;

-- name: GetIngestedEmailByID :one
SELECT i.*, a.local_part
FROM ingested_emails i
JOIN assigned_emails a ON a.id = i.assigned_email_id
WHERE i.id = $1 and i.application_id = $2 LIMIT 1;

-- name: GetWebhookJobByIDs :one
SELECT * FROM webhook_delivery_jobs WHERE application_id = $1 AND ingested_email_id = $2 ORDER BY created_at DESC LIMIT 1;

-- name: CancelWebhookJob :exec
UPDATE webhook_delivery_jobs
SET status = 'DEAD'
WHERE id = $1 AND application_id = $2 AND status IN ('PENDING', 'FAILED');

-- name: GetWebhookJobByIDAndAppID :one
SELECT * FROM webhook_delivery_jobs
WHERE id = $1 AND application_id = $2
LIMIT 1;

-- name: ResetWebhookJobForRedelivery :one
UPDATE webhook_delivery_jobs
SET status = 'PENDING',
    retry_count = 0,
    next_delivery_at = CURRENT_TIMESTAMP
WHERE id = $1 AND application_id = $2
RETURNING *;
-- name: ListIngestedEmailsByApplication :many
SELECT i.*, a.local_part FROM ingested_emails i
JOIN assigned_emails a ON a.id = i.assigned_email_id
WHERE i.application_id = $1
  AND (sqlc.narg('local_part')::text IS NULL OR sqlc.narg('local_part')::text = '' OR a.local_part = sqlc.narg('local_part')::text)
  AND (sqlc.narg('search')::text IS NULL OR sqlc.narg('search')::text = '' OR (
    i.from_address ILIKE '%' || sqlc.narg('search')::text || '%' OR
    i.subject ILIKE '%' || sqlc.narg('search')::text || '%'
  ))
ORDER BY i.received_at DESC LIMIT $2 OFFSET $3;

-- name: CountIngestedEmailsByApplication :one
SELECT count(*) FROM ingested_emails i
JOIN assigned_emails a ON a.id = i.assigned_email_id
WHERE i.application_id = $1
  AND (sqlc.narg('local_part')::text IS NULL OR sqlc.narg('local_part')::text = '' OR a.local_part = sqlc.narg('local_part')::text)
  AND (sqlc.narg('search')::text IS NULL OR sqlc.narg('search')::text = '' OR (
    i.from_address ILIKE '%' || sqlc.narg('search')::text || '%' OR
    i.subject ILIKE '%' || sqlc.narg('search')::text || '%'
  ));


-- name: GetApiKeyByKeyHash :one
SELECT * FROM api_keys WHERE key_hash = $1;

-- name: ListApiKeysByApplication :many
SELECT id, application_id, name, key_prefix, created_at, expires_at, last_used_at
FROM api_keys
WHERE application_id = $1
ORDER BY created_at DESC;

-- name: CreateApiKey :exec
INSERT INTO api_keys (application_id, name, key_prefix, key_hash, created_at, expires_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: DeleteApiKey :exec
DELETE FROM api_keys
WHERE id = $1 AND application_id = $2;


-- name: GetUserBySubject :one
SELECT * FROM users WHERE idp_user_sub = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :exec
UPDATE users
SET email = $1, idp_user_sub = $2, status = $3, created_at = $4, activated_at = $5, last_login_at = $6 WHERE id = $7;

-- name: ListWebhookJobsByApplication :many
SELECT wj.id, wj.application_id, wj.ingested_email_id, wj.status, wj.retry_count, wj.next_delivery_at, wj.created_at,
       COALESCE(wl.http_status_code, 0), COALESCE(wl.duration_ms, 0), COALESCE(wl.attempt_number, 0),
       COALESCE(wl.request_payload, ''), COALESCE(wl.response_body, '')
FROM webhook_delivery_jobs wj
LEFT JOIN LATERAL (
  SELECT http_status_code, duration_ms, attempt_number, request_payload, response_body
  FROM webhook_logs
  WHERE webhook_delivery_job_id = wj.id
  ORDER BY attempt_number DESC
  LIMIT 1
) wl ON TRUE
WHERE wj.application_id = $1
  AND (sqlc.narg('status')::text IS NULL OR sqlc.narg('status')::text = '' OR wj.status = sqlc.narg('status')::webhook_status)
ORDER BY wj.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetWebhookDeliveryStatsByApplication :one
    SELECT
        COUNT(*)::bigint AS total,
        COUNT(*) FILTER (WHERE status = 'SUCCESS')::bigint AS success,
        COUNT(*) FILTER (WHERE status IN ('FAILED', 'DEAD'))::bigint AS failures,
        COUNT(*) FILTER (WHERE status IN ('PENDING', 'PROCESSING'))::bigint AS pending_or_processing
    FROM public.webhook_delivery_jobs
    WHERE application_id = $1;

-- name: GetWebhookLogsByJobID :many
SELECT * FROM webhook_logs
WHERE webhook_delivery_job_id = $1
ORDER BY attempt_number DESC;

-- name: ListAssignedEmailsByApplication :many
SELECT * FROM assigned_emails
WHERE application_id = $1
ORDER BY created_at DESC;

-- name: UpdateAssignedEmailStatus :exec
UPDATE assigned_emails
SET is_active = $2
WHERE id = $1 AND application_id = $3;

-- name: CreatePersonalOrganization :one
INSERT INTO organizations (name, owner_user_id, is_personal)
VALUES ($1, $2, true)
ON CONFLICT (owner_user_id) WHERE is_personal = true
DO UPDATE SET name = organizations.name
RETURNING *;

-- name: GetPersonalOrganizationByUserID :one
SELECT * FROM organizations
WHERE owner_user_id = $1 AND is_personal = true
LIMIT 1;

-- name: GetOrganizationByID :one
SELECT * FROM organizations
WHERE id = $1 LIMIT 1;

-- name: InsertApplication :one
INSERT INTO applications (name, organization_id, webhook_url, webhook_secret, aws_iam_role_arn)
VALUES ($1, $2, '', '', '')
RETURNING *;

-- name: InsertUserApplication :exec
INSERT INTO user_application_access (user_id, application_id)
VALUES ($1, $2);

