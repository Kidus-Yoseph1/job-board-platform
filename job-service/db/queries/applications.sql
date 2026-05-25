-- name: CreateApplication :one
INSERT INTO applications (job_id, user_id, cover_letter)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetApplicationByID :one
SELECT * FROM applications WHERE id = $1;

-- name: ListApplicationsByJob :many
SELECT * FROM applications
WHERE job_id = $1
ORDER BY created_at DESC;

-- name: ListApplicationsByUser :many
SELECT * FROM applications
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateApplicationStatus :one
UPDATE applications SET status = $1, updated_at = now()
WHERE id = $2
RETURNING *;