package database

import (
	"fmt"

	"github.com/its-mrarsikk/fedup/shared/rss"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Wrapper around [gorm.DB].
type Database struct {
	*gorm.DB
}

// InitDB creates or migrates a GORM database. The name is either a filename or :memory: for a transient database.
func InitDB(name string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(name), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}

	if err = db.AutoMigrate(&rss.Feed{}, &rss.Item{}, &rss.Enclosure{}); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate DB: %w", err)
	}

	return &Database{DB: db}, nil
}
