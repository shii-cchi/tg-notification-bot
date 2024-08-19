-- name: CreateTask :one
INSERT INTO tasks (task, task_time, chat_id, created_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateTaskStatus :exec
UPDATE tasks
SET status = 'completed'
WHERE id = $1;

-- name: IsDeletedTask :one
SELECT
    CASE
        WHEN status = 'deleted' THEN TRUE
        ELSE FALSE
    END AS is_deleted
FROM tasks
WHERE id = $1;

-- name: GetAllTasks :many
SELECT id, task, task_time, created_at
FROM tasks
WHERE chat_id = $1 AND status = 'in_progress';

-- name: DeleteTask :exec
UPDATE tasks
SET status = 'deleted', deleted_at = $2
WHERE id = $1;