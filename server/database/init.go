package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "embed"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

//go:embed sql/init.sql
var init_query string

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
