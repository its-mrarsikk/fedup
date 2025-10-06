package database_test

import (
	"database/sql"
	"testing"

	"github.com/its-mrarsikk/fedup/server/database"
)

func TestInit(t *testing.T) {
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("%s", err)
	}
	defer db.Close()

	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", "feeds").Scan(&name)
	if err == sql.ErrNoRows {
		t.Fatalf("table feeds does not exist in database")
	} else if err != nil {
		t.Fatalf("%s", err)
	}
}
