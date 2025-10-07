package database_test

import (
	"database/sql"
	"fmt"
	"net/url"
	"testing"

	"github.com/its-mrarsikk/fedup/server/database"
	"github.com/its-mrarsikk/fedup/shared/rss"
)

const (
	expectedFeedTitle = "Testing Feed"
	expectedFeedLink  = "https://example.com"
	expectedFeedTTL   = 60
)

type mockRow struct {
	values []any
}

func (m *mockRow) Scan(dest ...any) error {
	if len(dest) != len(m.values) {
		return fmt.Errorf("expected %d values, got %d", len(m.values), len(dest))
	}

	for i := range dest {
		switch d := dest[i].(type) {
		case *string:
			*d = m.values[i].(string)
		case *int:
			*d = m.values[i].(int)
		case *sql.NullString:
			if v, ok := m.values[i].(string); ok {
				*d = sql.NullString{String: v, Valid: true}
			} else {
				*d = sql.NullString{Valid: false}
			}
		case *sql.NullInt64:
			if v, ok := m.values[i].(int64); ok {
				*d = sql.NullInt64{Int64: v, Valid: true}
			} else {
				*d = sql.NullInt64{Valid: false}
			}
		default:
			return fmt.Errorf("unsupported scan type %T", d)
		}
	}

	return nil
}

func TestFeedSerialize(t *testing.T) {
	uLink, err := url.Parse(expectedFeedLink)
	if err != nil {
		panic(err)
	}

	uFetchFrom, err := url.Parse("https://example.com/rss")
	if err != nil {
		panic(err)
	}

	f := &rss.Feed{
		DatabaseID:  4,
		Title:       expectedFeedTitle,
		Description: "Test",
		Link:        uLink,
		FetchFrom:   uFetchFrom,
		TTL:         expectedFeedTTL,
	}

	placeholders, _ := database.FeedSerialize(f)
	gotTitle, gotLink, gotTTL := placeholders[1].(string), placeholders[3].(string), placeholders[6].(int)

	if gotTitle != expectedFeedTitle {
		t.Fatalf("expected title %q, got %q", expectedFeedTitle, gotTitle)
	}

	if gotLink != expectedFeedLink {
		t.Fatalf("expected link %q, got %q", expectedFeedLink, gotLink)
	}

	if gotTTL != expectedFeedTTL {
		t.Fatalf("expected ttl %d, got %d", expectedFeedTTL, gotTTL)
	}
}

func TestFeedDeserialize(t *testing.T) {
	m := mockRow{values: []any{4, expectedFeedTitle, "Test", expectedFeedLink, nil, "", int64(expectedFeedTTL)}}

	f, err := database.FeedDeserialize(&m)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if f.Title != expectedFeedTitle {
		t.Fatalf("expected title %q, got %q", expectedFeedTitle, f.Title)
	}

	if f.Link.String() != expectedFeedLink {
		t.Fatalf("expected link %q, got %q", expectedFeedLink, f.Link.String())
	}

	if f.TTL != expectedFeedTTL {
		t.Fatalf("expected TTL %d, got %d", expectedFeedTTL, f.TTL)
	}
}
