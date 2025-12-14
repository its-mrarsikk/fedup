-- name: UpsertFeed :one
INSERT INTO feeds(title, description, link, fetch_from, language, ttl, etag, lastModified)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(fetch_from) DO UPDATE
    SET title=excluded.title,
    description=excluded.description,
    link=excluded.link,
    fetch_from=excluded.fetch_from,
    language=excluded.language,
    ttl=excluded.ttl,
    etag=excluded.etag,
    lastModified=excluded.lastModified
    RETURNING *;

-- name: GetFeedByID :one
SELECT * FROM feeds WHERE id = ?;

-- name: DeleteFeedByID :exec
DELETE FROM feeds WHERE id = ?;

-- name: DeleteFeedByFetchFrom :exec
DELETE FROM feeds WHERE fetch_from = ?;
