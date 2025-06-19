-- name: GetAllUnsent :many
SELECT id, recipient, content
FROM message
WHERE sent_at IS NULL
ORDER BY created_at;

-- name: GetNextUnsent :one
SELECT id, recipient, content
FROM message
WHERE sent_at IS NULL
ORDER BY created_at
LIMIT 1;

-- name: GetAllSent :many
SELECT message_id, sent_at
FROM message
WHERE sent_at NOTNULL
ORDER BY created_at;

-- name: SetMessageSent :exec
UPDATE message
SET message_id = $2,
    sent_at    = $3
WHERE id = $1;

-- name: InsertMessage :exec
INSERT INTO message (recipient, content)
VALUES ($1, $2);