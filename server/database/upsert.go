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
func (db *Database) UpsertFeed(feed *rss.Feed, recurse bool, ctx context.Context) error {
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
	last_modified=excluded.last_modified
	RETURNING id;
`

	id, err := gorm.G[uint](db.DB).Raw(stmt,
		time.Now(), time.Now(), // created_at, updated_at
		feed.Title, feed.Description, feed.Link,
		feed.FetchFrom, feed.Language,
		feed.TTL, feed.ETag, feed.LastModified).Take(ctx)

	if err != nil {
		return fmt.Errorf("failed to execute sql: %w", err)
	}

	feed.ID = id

	if recurse {
		for _, item := range feed.Items {
			item.Feed = *feed
			item.FeedID = id
			err := db.UpsertItem(&item, true, ctx)
			if err != nil {
				return fmt.Errorf("failed to upsert item: %w", err)
			}
		}
	}

	return nil
}

func (db *Database) UpsertItem(item *rss.Item, recurse bool, ctx context.Context) error {
	const stmt = `
	INSERT INTO items(created_at, updated_at, deleted_at, feed_id, guid, title, description, link, author, pub_date, read, starred)
	VALUES(?, ?, NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(guid) DO UPDATE
	SET created_at=excluded.created_at,
	updated_at=excluded.updated_at,
	feed_id=excluded.feed_id,
	guid=excluded.guid,
	title=excluded.title,
	description=excluded.description,
	link=excluded.link,
	author=excluded.author,
	pub_date=excluded.pub_date,
	read=excluded.read,
	starred=excluded.starred
	RETURNING id;
`

	id, err := gorm.G[uint](db.DB).Raw(stmt,
		time.Now(), time.Now(),
		item.FeedID, item.GUID,
		item.Title, item.Description,
		item.Link, item.Author,
		item.PubDate, item.Read, item.Starred).Take(ctx)

	if err != nil {
		return fmt.Errorf("failed to execute sql: %w", err)
	}

	item.ID = id

	return nil
}
