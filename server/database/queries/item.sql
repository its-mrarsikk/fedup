-- name: UpsertItem :one
INSERT INTO items(guid, title, description, link, author, pubDate, read, starred)
    VALUES(?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(id, guid) DO UPDATE
    SET guid=excluded.guid,
    title=excluded.title,
    description=excluded.description,
    link=excluded.link,
    author=excluded.author,
    pubDate=excluded.pubDate,
    read=excluded.read,
    starred=excluded.starred
    RETURNING *;

-- name: GetItemByID :one
SELECT * FROM items WHERE items.id = ?;

-- name: GetItemByGUID :one
SELECT * FROM items WHERE items.guid = ?;

-- name: GetUnreadItems :many
SELECT * FROM items WHERE items.read = 0;

-- name: GetStarredItems :many
SELECT * FROM items WHERE items.starred = 1;

-- name: DeleteItemByID :exec
DELETE FROM items WHERE items.id = ?;

-- name: DeleteItemsOlderThan :exec
DELETE FROM items WHERE unixepoch() - unixepoch(items.pubDate) > unixepoch(?);
