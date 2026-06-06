-- name: CreateNotification :one
INSERT INTO notifications (user_id, type, title, body)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListNotificationsByUser :many
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: MarkNotificationRead :exec
UPDATE notifications SET read = true
WHERE id = $1;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications SET read = true
WHERE user_id = $1 AND read = false;