package database_test

import (
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
