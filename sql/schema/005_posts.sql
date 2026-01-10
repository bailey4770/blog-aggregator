-- +goose Up
CREATE TABLE posts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  title TEXT,
  url TEXT NOT NULL,
  description TEXT,
  published_at TIMESTAMP NOT NULL, 
  feed_id UUID NOT NULL,
  FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
  UNIQUE(url)
);

-- +goose Down
DROP TABLE posts;
