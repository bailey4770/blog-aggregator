-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetFeedList :many
SELECT feeds.name, feeds.url, users.name FROM feeds
INNER JOIN users ON users.ID=feeds.user_id;

-- name: GetFeedRecord :one
SELECT * FROM feeds
WHERE name = $1;
