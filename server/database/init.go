package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

const (
	init_query = `-- Table: feeds
CREATE TABLE IF NOT EXISTS feeds (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
	link TEXT,
	fetchFrom TEXT,
    language TEXT,
    ttl INTEGER
);

-- Table: items
CREATE TABLE IF NOT EXISTS items (
    id INTEGER PRIMARY KEY,
    feed_id INTEGER NOT NULL,
    guid TEXT UNIQUE,
    title TEXT,
    description TEXT,
    link TEXT,
    author TEXT,
    pubDate DATETIME,
    read BOOLEAN NOT NULL DEFAULT 0,
    enclosure_url TEXT,
    enclosure_type TEXT,
    enclosure_length INTEGER,
    source_url TEXT,
    source_name TEXT,
    FOREIGN KEY(feed_id) REFERENCES feeds(id)
);`
)

func InitDB(name string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", name)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to establish database connection: %w", err)
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()
	if _, err := db.ExecContext(ctx2, init_query); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}
