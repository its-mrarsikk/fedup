package rss_test

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/its-mrarsikk/fedup/shared/rss"
)

//go:embed testcase/rss2.xml
var rss2String string

func TestBaseRSS2(t *testing.T) {
	feed, err := rss.ParseRSS(strings.NewReader(rss2String))
	if err != nil {
		t.Fatalf("ParseRSS: %s", err)
	}

	const (
		expectedTitle           = "NASA Space Station News"
		expectedItemTitle       = "Louisiana Students to Hear from NASA Astronauts Aboard Space Station"
		expectedEnclosureLength = 1032272
	)

	if feed.Title != expectedTitle {
		t.Fatalf("expected feed title %s, got %s", expectedTitle, feed.Title)
	}

	if feed.Items[0].Title != expectedItemTitle {
		t.Fatalf("expected item title %s, got %s", feed.Items[0].Title, expectedItemTitle)
	}

	if feed.Items[2].Enclosure.Length != expectedEnclosureLength {
		t.Fatalf("expected enclosure length %d, got %d", expectedEnclosureLength, feed.Items[2].Enclosure.Length)
	}
}
