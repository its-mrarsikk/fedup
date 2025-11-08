package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/its-mrarsikk/fedup/shared/rss"
)

// txDrill is a combined type for a transaction and its context. This is an internal type that is used to
// propagate a transaction when recursing in [StoreFeed] and [StoreItem].
//
// This is jank. I couldn't think of a better solution, and since Go doesn't have optional parameters,
// external callers just have to add a nil at the end of their calls to [StoreItem] and [StoreEnclosure].
type txDrill struct {
	tx     *sql.Tx
	ctx    context.Context
	cancel func()
}

// StoreFeed adds or updates the Feed in the database. Existence is determined by a) checking for the DatabaseID or b) checking for the fetchFrom.
// If the entry does not exist, it is INSERTed, otherwise it is UPDATEd.
//
// If recurse is true, the Items are also stored via [StoreItem] with recurse = true.
//
// The feed's DatabaseID is modified to be consistent with the database. The FetchFrom value is used for
// finding the entry after insertion, so it is assumed that it is present.
func (d *Database) StoreFeed(feed *rss.Feed, recurse bool) error {
	if feed == nil {
		return fmt.Errorf("feed is nil")
	}

	txCtx, cancelTx := context.WithCancel(context.Background())
	defer cancelTx()
	tx, err := d.BeginTx(txCtx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	var dbid int
	row := tx.QueryRow("SELECT id FROM feeds WHERE id = ? OR fetchFrom = ?;", feed.DatabaseID, feed.FetchFrom.String())
	err = row.Scan(&dbid)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to scan SELECT query: %w", err)
	}

	var (
		placeholders []any
		query        string
	)
	op := "UPDATE"

	if errors.Is(err, sql.ErrNoRows) {
		placeholders, query = feedSerializeInsert(feed)
		op = "INSERT"
	} else {
		feed.DatabaseID = dbid
		placeholders, query = feedSerializeUpdate(feed)
	}

	_, err = tx.Exec(query, placeholders...)
	if err != nil {
		return fmt.Errorf("failed to %s feed: %w", op, err)
	}

	// update database id
	// i have to do this because INSERT and UPDATE don't return affected rows
	row = tx.QueryRow("SELECT id FROM feeds WHERE fetchFrom = ?", feed.FetchFrom.String())
	if err := row.Scan(&dbid); err != nil {
		return fmt.Errorf("failed to scan post-%s SELECT query: %w", op, err)
	}
	feed.DatabaseID = dbid

	if recurse && len(feed.Items) > 0 {
		td := &txDrill{tx: tx, ctx: txCtx, cancel: cancelTx}
		for _, item := range feed.Items {
			err := d.StoreItem(item, true, td)
			if err != nil {
				return fmt.Errorf("failed to store item: %w", err)
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// StoreItem adds or updates the Item in the database. Existence is determined by a) checking for the DatabaseID or b) checking for the guid.
// If the entry does not exist, it is INSERTed, otherwise it is UPDATEd.
//
// If recurse is true, the Enclosure is also stored via [StoreEnclosure].
//
// The item's DatabaseID is modified to be consistent with the database. The GUID value is used for
// finding the entry after insertion, so it is assumed that it is present.
func (d *Database) StoreItem(item *rss.Item, recurse bool, td *txDrill) error {
	if item == nil {
		return fmt.Errorf("item is nil")
	}

	// This is also jank, but this is the only way I can think of to determine who commits the transaction.
	txRoot := false

	if td == nil {
		txRoot = true
		txCtx, cancelTx := context.WithCancel(context.Background())
		tx, err := d.BeginTx(txCtx, nil)
		if err != nil {
			cancelTx()
			return fmt.Errorf("failed to start transaction: %w", err)
		}
		td = &txDrill{tx: tx, ctx: txCtx, cancel: cancelTx}
	}

	var dbid int
	row := td.tx.QueryRow("SELECT id FROM items WHERE id = ? OR guid = ?;", item.DatabaseID, item.GUID)
	err := row.Scan(&dbid)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to scan SELECT query: %w", err)
	}

	var (
		placeholders []any
		query        string
	)
	op := "UPDATE"

	if errors.Is(err, sql.ErrNoRows) {
		placeholders, query = itemSerializeInsert(item)
		op = "INSERT"
	} else {
		item.DatabaseID = dbid
		placeholders, query = itemSerializeUpdate(item)
	}

	_, err = td.tx.Exec(query, placeholders...)
	if err != nil {
		return fmt.Errorf("failed to %s item: %w", op, err)
	}

	// update database id
	// i have to do this because INSERT and UPDATE don't return affected rows
	row = td.tx.QueryRow("SELECT id FROM items WHERE guid = ?", item.GUID)
	if err := row.Scan(&dbid); err != nil {
		return fmt.Errorf("failed to scan post-%s SELECT query: %w", op, err)
	}
	item.DatabaseID = dbid

	if recurse && item.Enclosure != nil {
		err = d.StoreEnclosure(item.Enclosure, td)
		if err != nil {
			return fmt.Errorf("failed to store enclosure: %w", err)
		}
	}
	if txRoot {
		if err = td.tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}
	return nil
}

// StoreEnclosure adds or updates the Item in the database. Existence is determined by a) checking for the DatabaseID or b) checking for the url. If the entry does not exist, it is INSERTed, otherwise it is UPDATEd.
//
// The enclosure's DatabaseID is modified to be consistent with the database. The URL value is used for
// finding the entry after insertion, so it is assumed that it is present.
func (d *Database) StoreEnclosure(enclosure *rss.Enclosure, td *txDrill) error {
	if enclosure == nil {
		return fmt.Errorf("enclosure is nil")
	}
	txRoot := false

	if td == nil {
		txRoot = true
		txCtx, cancelTx := context.WithCancel(context.Background())
		tx, err := d.BeginTx(txCtx, nil)
		if err != nil {
			cancelTx()
			return fmt.Errorf("failed to start transaction: %w", err)
		}
		td = &txDrill{tx: tx, ctx: txCtx, cancel: cancelTx}
	}

	var dbid int
	row := td.tx.QueryRow("SELECT id FROM enclosures WHERE id = ? OR url = ?;", enclosure.DatabaseID, enclosure.URL.String())
	err := row.Scan(&dbid)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to scan SELECT query: %w", err)
	}

	var (
		placeholders []any
		query        string
	)
	op := "UPDATE"

	if errors.Is(err, sql.ErrNoRows) {
		placeholders, query = enclosureSerializeInsert(enclosure)
		op = "INSERT"
	} else {
		enclosure.DatabaseID = dbid
		placeholders, query = enclosureSerializeUpdate(enclosure)
	}

	_, err = td.tx.Exec(query, placeholders...)
	if err != nil {
		return fmt.Errorf("failed to %s enclosure: %w", op, err)
	}

	// update database id
	// i have to do this because INSERT and UPDATE don't return affected rows
	row = td.tx.QueryRow("SELECT id FROM enclosures WHERE url = ?", enclosure.URL.String())
	if err := row.Scan(&dbid); err != nil {
		return fmt.Errorf("failed to scan post-%s SELECT query: %w", op, err)
	}
	enclosure.DatabaseID = dbid

	if txRoot {
		if err = td.tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}
	return nil
}
