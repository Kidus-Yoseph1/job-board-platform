-- name: CreateEmailJob :one
INSERT INTO email_jobs (to_email, subject, body)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateEmailJobStatus :exec
UPDATE email_jobs SET status = $1, attempts = attempts + 1, updated_at = now()
WHERE id = $2;

-- name: GetPendingEmailJobs :many
SELECT * FROM email_jobs
WHERE (status = 'pending' OR status = 'failed') AND attempts < 3
ORDER BY created_at ASC
LIMIT $1;