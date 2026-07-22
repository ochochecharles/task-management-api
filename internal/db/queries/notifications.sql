-- name: CreateNotification :one
INSERT INTO notifications (user_id, type, message, entity_type, entity_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListNotifications :many
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUnreadNotifications :one
SELECT COUNT(*) FROM notifications
WHERE user_id = $1 AND read_at IS NULL;

-- name: MarkNotificationRead :one
UPDATE notifications
SET read_at = now()
WHERE id = $1 AND user_id = $2 AND read_at IS NULL
RETURNING *;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications
SET read_at = now()
WHERE user_id = $1 AND read_at IS NULL;

-- name: DeleteNotification :execrows
DELETE FROM notifications
WHERE id = $1 AND user_id = $2;

-- name: DeleteReadNotifications :execrows
DELETE FROM notifications
WHERE user_id = $1 AND read_at IS NOT NULL;