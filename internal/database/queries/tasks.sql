-- name: CreateTask :one
INSERT INTO tasks (task, task_time, chat_id, created_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetTaskId :many
SELECT id, task_time, created_at
FROM tasks
WHERE task = $1 AND chat_id = $2;

-- name: UpdateTaskStatus :exec
UPDATE tasks
SET status = 'completed'
WHERE id = $1;

-- name: GetAllTasks :many
SELECT id, task, task_time, created_at
FROM tasks
WHERE chat_id = $1 AND status = 'in_progress';

-- name: DeleteTask :exec
UPDATE tasks
SET status = 'deleted', deleted_at = $2
WHERE id = $1;