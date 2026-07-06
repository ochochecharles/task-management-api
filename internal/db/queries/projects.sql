-- name: CreateProject :one
INSERT INTO projects (name, description, owner_id, due_date)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetProjectByID :one
SELECT * FROM projects
WHERE id = $1;

-- name: ListProjectsByMember :many
SELECT p.* FROM projects p
INNER JOIN project_members pm ON pm.project_id = p.id
WHERE pm.user_id = $1;

-- name: UpdateProject :one
UPDATE projects
SET name = $2, description = $3, status = $4, due_date = $5, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1;

-- name: AddProjectMember :exec
INSERT INTO project_members (project_id, user_id, role)
VALUES ($1, $2, $3);

-- name: GetProjectMember :one
SELECT * FROM project_members
WHERE project_id = $1 AND user_id = $2;

-- name: ListProjectMembers :many
SELECT * FROM project_members
WHERE project_id = $1;

-- name: RemoveProjectMember :exec
DELETE FROM project_members
WHERE project_id = $1 AND user_id = $2;