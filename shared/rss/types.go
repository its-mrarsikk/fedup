package rss

import (
	"net/url"
	"time"

	"gorm.io/gorm"
)

// Feed represents an RSS channel
type Feed struct {
	gorm.Model
	Title       string `gorm:"not null;"`
	Description string `gorm:"not null;"`
	// The link to the HTML representation of the feed. Different from FetchFrom
	Link string `gorm:"not null;"`
	// The URL that this feed can be retrieved from. Different from Link
	FetchFrom string `gorm:"not null;unique"`
	Language  string
	// The time-to-live of the feed. Time in minutes that the reader should wait between each refresh
	TTL          int `gorm:"default:60"`
	ETag         string
	LastModified time.Time
	Items        []Item `gorm:"constraint:OnDelete:CASCADE;"`
}

// Item represents an RSS item/post
type Item struct {
	gorm.Model
	// The feed this item came from.
	Feed Feed
	// ID of the feed. Automatically creates a belongs-to relationship with Feed -> Item.
	FeedID      uint   `gorm:"index;not null;"`
	GUID        string `gorm:"not null;unique"`
	Title       string
	Description string
	Link        *url.URL
	Author      string
	PubDate     time.Time  `gorm:"not null;"`
	Read        bool       `gorm:"not null;"`
	Starred     bool       `gorm:"not null;"`
	Enclosure   *Enclosure `gorm:"constraint:OnDelete:CASCADE;"`
}

// Enclosure represents an RSS enclosure, usually media associated with an item
// See https://www.rssboard.org/rss-specification
type Enclosure struct {
	gorm.Model
	ItemID   uint `gorm:"index"`
	Item     Item
	URL      string `gorm:"not null;"`
	MimeType string `gorm:"not null;"`
	FilePath string `gorm:"not null;"`
}
