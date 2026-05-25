-- name: CreateCompany :one
INSERT INTO companies (user_id, name, description, website, location)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetCompanyByUserID :one
SELECT * FROM companies WHERE user_id = $1;

-- name: GetCompanyByID :one
SELECT * FROM companies WHERE id = $1;