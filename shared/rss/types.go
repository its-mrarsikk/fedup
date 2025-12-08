package rss

import (
	"net/url"
	"time"
)

// Feed represents an RSS channel
type Feed struct {
	Title       string
	Description string
	// The link to the HTML representation of the feed. Different from FetchFrom
	Link string
	// The URL that this feed can be retrieved from. Different from Link
	FetchFrom string
	Language  string
	// The time-to-live of the feed. Time in minutes that the reader should wait between each refresh
	TTL          int
	ETag         string
	LastModified time.Time
	Items        []Item
}

// Item represents an RSS item/post
type Item struct {
	GUID        string
	Title       string
	Description string
	Link        *url.URL
	Author      string
	PubDate     time.Time
	Read        bool
	Starred     bool
	Enclosure   *Enclosure
}

// Enclosure represents an RSS enclosure, usually media associated with an item
// See https://www.rssboard.org/rss-specification
type Enclosure struct {
	URL      string
	MimeType string
	FilePath string
}
