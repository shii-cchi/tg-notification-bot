-- +goose Up

CREATE TABLE tasks (
    id BIGSERIAL NOT NULL PRIMARY KEY,
    task TEXT NOT NULL,
    task_time TEXT NOT NULL,
    chat_id bigint NOT NULL,
    status TEXT NOT NULL DEFAULT 'in_progress',
    created_at TIMESTAMP NOT NULL
);

-- +goose Down

DROP TABLE tasks;