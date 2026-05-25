-- name: CreateJob :one
INSERT INTO jobs (company_id, title, description, category, location, type, is_negotiable)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetJobByID :one
SELECT * FROM jobs WHERE id = $1 AND deleted_at IS NULL;

-- name: ListJobs :many
SELECT * FROM jobs
WHERE deleted_at IS NULL AND status = 'open'
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListJobsByCompany :many
SELECT * FROM jobs
WHERE company_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateJobStatus :one
UPDATE jobs SET status = $1, updated_at = now()
WHERE id = $2
RETURNING *;

-- name: DeleteJob :exec
UPDATE jobs SET deleted_at = now(), updated_at = now()
WHERE id = $1;