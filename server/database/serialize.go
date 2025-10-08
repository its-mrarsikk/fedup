package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/its-mrarsikk/fedup/shared/rss"
)

type RowScanner interface {
	Scan(...any) error
}

func safeURLParse(s sql.NullString) *url.URL {
	if !s.Valid {
		return nil
	}
	u, err := url.Parse(s.String)
	if err != nil {
		return nil
	}
	return u
}

func FeedSerialize(f *rss.Feed) ([]any, string) {
	var link, fetchFrom string
	if f.Link != nil {
		link = f.Link.String()
	} else {
		link = ""
	}
	if f.FetchFrom != nil {
		fetchFrom = f.FetchFrom.String()
	} else {
		fetchFrom = ""
	}

	return []any{f.DatabaseID,
		f.Title,
		f.Description,
		link,
		fetchFrom,
		f.Language,
		f.TTL}, "(?,?,?,?,?,?,?)"
}

func FeedDeserialize(r RowScanner) (*rss.Feed, error) {
	var dbid int
	var title, description string
	var link, fetchFrom, language sql.NullString
	var ttl sql.NullInt64

	var urlLink, urlFetchFrom *url.URL
	var strLanguage string
	var intTTL int

	err := r.Scan(&dbid, &title, &description, &link, &fetchFrom, &language, &ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	urlLink = safeURLParse(link)
	urlFetchFrom = safeURLParse(fetchFrom)

	if language.Valid {
		strLanguage = language.String
	} else {
		strLanguage = ""
	}

	if ttl.Valid {
		intTTL = int(ttl.Int64)
	} else {
		intTTL = 0
	}

	return &rss.Feed{
		DatabaseID:  dbid,
		Title:       title,
		Description: description,
		Link:        urlLink,
		FetchFrom:   urlFetchFrom,
		Language:    strLanguage,
		TTL:         intTTL,
	}, nil
}

// ITEMS //

func ItemSerialize(i *rss.Item) ([]any, string) {
	var link, pubDate, encURL, encType string
	var encLength int

	if i.Link != nil {
		link = i.Link.String()
	} else {
		link = ""
	}

	if i.PubDate != nil {
		pubDate = i.PubDate.Format(time.RFC3339)
	} else {
		pubDate = ""
	}

	if i.Enclosure != nil {
		if i.Enclosure.URL != nil {
			encURL = i.Enclosure.URL.String()
		}
		encType = i.Enclosure.MimeType
		encLength = i.Enclosure.Length
	}

	return []any{
		i.DatabaseID,
		i.Feed.DatabaseID,
		i.GUID,
		i.Title,
		i.Description,
		link,
		i.Author,
		pubDate,
		i.Read,
		encURL,
		encType,
		encLength,
	}, "(?,?,?,?,?,?,?,?,?,?,?,?)"
}

func ItemDeserialize(r RowScanner, feed *rss.Feed) (*rss.Item, error) {
	var dbid, feedID int
	var guid, title, description, link, author sql.NullString
	var pubDate sql.NullString
	var read int
	var enclosureURL, enclosureType sql.NullString
	var enclosureLength sql.NullInt64

	err := r.Scan(&dbid, &feedID, &guid, &title, &description, &link, &author, &pubDate, &read, &enclosureURL, &enclosureType, &enclosureLength)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	var (
		urlLink, urlEnclosure *url.URL
		timePubDate           *time.Time
		enclosure             *rss.Enclosure
	)

	urlLink = safeURLParse(link)

	if pubDate.Valid {
		t, err := time.Parse(time.RFC3339, pubDate.String)
		if err == nil {
			timePubDate = &t
		}
	}

	urlEnclosure = safeURLParse(enclosureURL)
	if urlEnclosure != nil && enclosureType.Valid && enclosureLength.Valid {
		enclosure = &rss.Enclosure{
			URL:      urlEnclosure,
			MimeType: enclosureType.String,
			Length:   int(enclosureLength.Int64),
		}
	}

	return &rss.Item{
		DatabaseID: dbid,
		Feed:       feed,
		GUID:       guid.String,
		Title:      title.String,
		Description: func() string {
			if description.Valid {
				return description.String
			}
			return ""
		}(),
		Link:      urlLink,
		Author:    author.String,
		PubDate:   timePubDate,
		Read:      read != 0, // int to bool conversion
		Enclosure: enclosure,
	}, nil
}
