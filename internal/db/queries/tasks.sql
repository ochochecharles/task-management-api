-- name: CreateTask :one
INSERT INTO tasks (project_id, title, description, priority, due_date, created_by, assigned_to)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetTaskByID :one
SELECT * FROM tasks
WHERE id = $1;

-- name: ListTasksByProject :many
SELECT * FROM tasks
WHERE project_id = $1
ORDER BY created_at DESC;

-- name: ListTasksAssignedToUser :many
SELECT * FROM tasks
WHERE assigned_to = $1
ORDER BY created_at DESC;

-- name: UpdateTask :one
UPDATE tasks
SET title = $2, description = $3, status = $4, priority = $5, due_date = $6, assigned_to = $7, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateTaskStatus :one
UPDATE tasks
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;