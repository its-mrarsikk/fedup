package database

import (
	"context"
	"fmt"

	"github.com/its-mrarsikk/fedup/shared/rss"
	"gorm.io/gorm"
)

// Constants for Upsert conditions.
const ()

// Upsert updates or inserts `m` in the database. `where` is a WHERE clause condition
// that is used to check for existence.
// Example:
//
//	upsert(db, feed, "id = ?, fetch_from = ?", feed.ID, "https://example.com")
func upsert[T any](db *gorm.DB, m T, where string, args ...any) error {
	if where == "" {
		return fmt.Errorf("expected non-empty where")
	}

	ctx := context.Background()

	w := gorm.G[T](db).Where(where, args...)
	rows, err := w.Updates(ctx, m)
	if err != nil {
		return fmt.Errorf("failed to update object: %w", err)
	}

	if rows == 0 {
		err := gorm.G[T](db).Create(ctx, &m)
		if err != nil {
			return fmt.Errorf("failed to create object: %w", err)
		}
		return nil
	}

	// gorm doesn't update zero values, so we need a separate case for read and starred going from true to false
	// and yea this isn't very generic but whatever
	// if i, ok := any(m).(rss.Item); ok {
	// 	_, err := w.Update(ctx, "read", i.Read)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to update object (read): %w", err)
	// 	}
	// 	_, err = w.Update(ctx, "starred", i.Starred)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to update object (starred): %w", err)
	// 	}
	// }

	return nil
}

// UpsertFeed inserts or updates a Feed. If `recurse` is true, the items and
// their enclosures are also upserted
func (db *Database) UpsertFeed(feed rss.Feed, recurse bool, where string, args ...any) error {
	if err := upsert(db.DB, feed, where, args...); err != nil {
		return fmt.Errorf("failed to upsert feed: %w", err)
	}

	if recurse {
		// TODO
	}

	return nil
}
