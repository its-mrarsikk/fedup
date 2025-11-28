package database

import (
	"context"
	"fmt"
	"time"

	"github.com/its-mrarsikk/fedup/shared/rss"
	"gorm.io/gorm"
)

// UpsertFeed inserts or updates a Feed. If `recurse` is true, the items and
// their enclosures are also upserted
func (db *Database) UpsertFeed(feed rss.Feed, recurse bool, ctx context.Context) error {
	const stmt = `
	INSERT INTO feeds(created_at, updated_at, deleted_at, title, description, link, fetch_from, language, ttl, e_tag, last_modified)
	VALUES (?, ?, NULL, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(fetch_from) DO UPDATE
	SET created_at=excluded.created_at,
	updated_at=excluded.updated_at,
	title=excluded.title,
	description=excluded.description,
	link=excluded.link,
	fetch_from=excluded.fetch_from,
	language=excluded.language,
	ttl=excluded.ttl,
	e_tag=excluded.e_tag,
	last_modified=excluded.last_modified;
`

	err := gorm.G[rss.Feed](db.DB).Exec(ctx, stmt,
		time.Now(), time.Now(), // created_at, updated_at
		feed.Title, feed.Description, feed.Link,
		feed.FetchFrom, feed.Language,
		feed.TTL, feed.ETag, feed.LastModified)

	if err != nil {
		return fmt.Errorf("failed to execute sql: %w", err)
	}

	if recurse {

	}

	return nil
}

func (db *Database) UpsertItem(item rss.Item, recurse bool) {
	_ = `
	INSERT INTO items(id, created_at, updated_at, deleted_at, guid, title, description, link, author, pub_date, read, starred)
	VALUES(?, ?, ?, NULL, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id, guid) DO UPDATE
	SET id=excluded.id,
	created_at=excluded.created_at,
	updated_at=excluded.updated_at,
	guid=excluded.guid,
	title=excluded.title,
	description=excluded.description,
	link=excluded.link,
	author=excluded.author,
	pub_date=excluded.pub_date,
	read=excluded.read,
	starred=excluded.starred;
`
}
