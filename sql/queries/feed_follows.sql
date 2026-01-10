-- name: CreateFeedFollow :many
with inserted AS (
  INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
  VALUES ($1, $2, $3, $4, $5)
  RETURNING *
)
SELECT 
  inserted.*,
  users.name AS user_name,
  feeds.name AS feed_name
FROM inserted
INNER JOIN users ON users.id = inserted.user_id
INNER JOIN feeds ON feeds.id = inserted.feed_id;

-- name: GetFeedFollowsForUser :many
SELECT
  users.name AS user_name,
  feeds.name AS feed_name,
  feeds.url AS feed_url
FROM feed_follows
INNER JOIN users ON feed_follows.user_id=users.id
INNER JOIN feeds ON feed_follows.feed_id=feeds.id
WHERE users.name = $1;

-- name: DeleteFeedFollow :one
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2
RETURNING *;
