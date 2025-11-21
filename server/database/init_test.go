package database_test

import (
	"testing"

	"github.com/its-mrarsikk/fedup/server/database"
)

func TestInit(t *testing.T) {
	_, err := database.InitDB("shitass.db")
	if err != nil {
		t.Fatalf("failed to init db: %s", err)
	}
}
