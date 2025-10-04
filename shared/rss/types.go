package rss

import (
	"net/url"
	"time"
)

type Feed struct {
	Title       string
	Description string
	Link        *url.URL
	FetchFrom   *url.URL
	Language    string
	TTL         int
	Items       []*Item
}

type Item struct {
	Feed      *Feed
	GUID      string
	Title     string
	Link      *url.URL
	Author    string
	PubDate   *time.Time
	Read      bool
	Enclosure *Enclosure
}

type Enclosure struct {
	URL      *url.URL
	MimeType string
	Length   int
}
