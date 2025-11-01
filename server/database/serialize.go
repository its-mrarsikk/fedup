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

func FeedSerializeInsert(f *rss.Feed) ([]any, string) {
	var link, fetchFrom, lastModified string
	if f.Link != nil {
		link = f.Link.String()
	}
	if f.FetchFrom != nil {
		fetchFrom = f.FetchFrom.String()
	}
	if !f.LastModified.IsZero() {
		lastModified = f.LastModified.Format(time.RFC3339)
	}

	return []any{nil, // sqlite automatically assigns a primary key on NULL
		f.Title,
		f.Description,
		link,
		fetchFrom,
		f.Language,
		f.TTL,
		f.ETag,
		lastModified,
	}, "INSERT INTO feeds VALUES (?,?,?,?,?,?,?,?,?);"
}

func FeedSerializeUpdate(f *rss.Feed) ([]any, string) {
	var link, fetchFrom, lastModified string
	if f.Link != nil {
		link = f.Link.String()
	}
	if f.FetchFrom != nil {
		fetchFrom = f.FetchFrom.String()
	}
	if !f.LastModified.IsZero() {
		lastModified = f.LastModified.Format(time.RFC3339)
	}

	return []any{
		f.Title,
		f.Description,
		link,
		fetchFrom,
		f.Language,
		f.TTL,
		f.ETag,
		lastModified,
		f.DatabaseID,
	}, "UPDATE feeds SET title = ?, description = ?, link = ?, fetchFrom = ?, language = ?, ttl = ?, etag = ?, lastModified = ? WHERE id = ?;"
}

func FeedDeserialize(r RowScanner) (*rss.Feed, error) {
	var dbid int
	var title, description string
	var link, fetchFrom, language, etag, strLastModified sql.NullString
	var ttl int

	var urlLink, urlFetchFrom *url.URL
	var strLanguage, strEtag string
	var lastModified time.Time

	err := r.Scan(&dbid, &title, &description, &link, &fetchFrom, &language, &ttl, &etag, &strLastModified)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	urlLink = safeURLParse(link)
	if !fetchFrom.Valid {
		return nil, fmt.Errorf("feed must have fetchFrom")
	}
	urlFetchFrom, err = url.Parse(fetchFrom.String)
	if err != nil {
		return nil, fmt.Errorf("invalid fetchFrom in feed: %w", err)
	}

	if language.Valid {
		strLanguage = language.String
	} else {
		strLanguage = ""
	}

	if etag.Valid {
		strEtag = etag.String
	}

	if strLastModified.Valid {
		lastModified, err = time.Parse(time.RFC3339, strLastModified.String)
		if err != nil {
			return nil, fmt.Errorf("invalid lastModified in feed: %w", err)
		}
	}

	return &rss.Feed{
		DatabaseID:   dbid,
		Title:        title,
		Description:  description,
		Link:         urlLink,
		FetchFrom:    urlFetchFrom,
		Language:     strLanguage,
		TTL:          ttl,
		ETag:         strEtag,
		LastModified: lastModified,
	}, nil
}

// ITEMS //

func ItemSerializeInsert(i *rss.Item) ([]any, string) {
	var link, pubDate string

	if i.Link != nil {
		link = i.Link.String()
	}
	if i.PubDate != nil {
		pubDate = i.PubDate.Format(time.RFC3339)
	}

	return []any{
		nil,
		i.Feed.DatabaseID,
		i.GUID,
		i.Title,
		i.Description,
		link,
		i.Author,
		pubDate,
		i.Read,
		i.Starred,
	}, "INSERT INTO items VALUES (?,?,?,?,?,?,?,?,?,?);"
}

func ItemSerializeUpdate(i *rss.Item) ([]any, string) {
	var link, pubDate string

	if i.Link != nil {
		link = i.Link.String()
	}
	if i.PubDate != nil {
		pubDate = i.PubDate.Format(time.RFC3339)
	}

	return []any{
		i.Feed.DatabaseID,
		i.GUID,
		i.Title,
		i.Description,
		link,
		i.Author,
		pubDate,
		i.Read,
		i.Starred,
		i.DatabaseID,
	}, "UPDATE items SET feed_id = ?, guid = ?, title = ?, description = ?, link = ?, author = ?, pubDate = ?, read = ?, starred = ? WHERE id = ?;"
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

// ENCLOSURES //

func EnclosureSerializeInsert(e *rss.Enclosure, itemId int) ([]any, string) {
	return []any{
		nil,
		itemId,
		e.MimeType,
		e.URL.String(),
		e.FilePath,
	}, "INSERT INTO enclosures VALUES (?, ?, ?, ?, ?);"
}

func EnclosureSerializeUpdate(e *rss.Enclosure) ([]any, string) {
	return []any{
		e.MimeType,
		e.URL.String(),
		e.FilePath,
		e.DatabaseID,
	}, "UPDATE enclosures SET type = ?, url = ?, filePath = ? WHERE id = ?;"
}

func EnclosureDeserialize(r RowScanner) (*rss.Enclosure, error) {
	var dbid int
	var discardItemId int // for some reason i cant use _ in Scan?
	var mimeType, rawUrl, filePath string

	err := r.Scan(&dbid, &discardItemId, &mimeType, &rawUrl, &filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	parsedURL, err := url.Parse(rawUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid URL in enclosure: %w", err)
	}

	return &rss.Enclosure{
		DatabaseID: dbid,
		MimeType:   mimeType,
		URL:        parsedURL,
		FilePath:   filePath,
	}, nil
}
