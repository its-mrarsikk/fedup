package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/its-mrarsikk/fedup/shared/rss"
)

// RowScanner is an abstraction for an SQL row that can be used as Scan(&id, &val), etc.
//
// This interface only exists to allow unit tests to provide mock data.
type RowScanner interface {
	Scan(...any) error
}

// safeURLParse uses [url.Parse] to create a [*url.URL] from s, returning nil if any errors occur.
//
// This is a convenience function to save typing on optional URL parameters.
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

// feedSerializeInsert constructs a prepared SQL INSERT statement from an [rss.Feed].
// Note that it assigns id to NULL to let SQLite autogenerate an ID.
func feedSerializeInsert(f *rss.Feed) ([]any, string) {
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

// feedSerializeUpdate constructs a prepared SQL UPDATE statement from an [rss.Feed].
// Note that it adds all fields to SET as using some kind of diff system would be complex and overkill.
func feedSerializeUpdate(f *rss.Feed) ([]any, string) {
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

// feedDeserialize constructs an [rss.Feed] from a [RowScanner].
// It scans the columns in the order specified in sql/init.sql.
//
// Items is not constructed as it is out of the scope of this function.
func feedDeserialize(r RowScanner) (*rss.Feed, error) {
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

// itemSerializeInsert constructs a prepared SQL INSERT statement from an [rss.Item].
// Note that it assigns id to NULL to let SQLite autogenerate an ID.
func itemSerializeInsert(i *rss.Item) ([]any, string) {
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

// itemSerializeUpdate constructs a prepared SQL UPDATE statement from an [rss.Item].
// Note that it adds all fields to SET as using some kind of diff system would be complex and overkill.
func itemSerializeUpdate(i *rss.Item) ([]any, string) {
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

// itemDeserialize constructs an [rss.Item] from a [RowScanner]. It scans the columns in the order specified in sql/init.sql.
//
// This does not set Feed; it must be set manually by the caller if needed.
func itemDeserialize(r RowScanner) (*rss.Item, error) {
	var (
		dbid, discardFeedId                                   int
		guid, title, description, strLink, author, strPubDate sql.NullString
		rawRead, rawStarred                                   int
		pubDate                                               time.Time
		link                                                  *url.URL
		read, starred                                         bool
	)

	err := r.Scan(&dbid, &discardFeedId, &guid, &title, &description, &strLink, &author, &strPubDate, &rawRead, &rawStarred)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	link = safeURLParse(strLink)

	if strPubDate.Valid {
		pubDate, err = time.Parse(time.RFC3339, strPubDate.String)
	}

	read = rawRead == 1
	starred = rawStarred == 1

	return &rss.Item{
		DatabaseID:  dbid,
		GUID:        guid.String,
		Title:       title.String,
		Description: description.String,
		Link:        link,
		Author:      author.String,
		PubDate:     &pubDate,
		Read:        read,
		Starred:     starred,
	}, nil

}

// ENCLOSURES //

// enclosureSerializeInsert constructs a prepared SQL INSERT statement from an [rss.Enclosure].
// Note that it assigns id to NULL to let SQLite autogenerate an ID.
func enclosureSerializeInsert(e *rss.Enclosure) ([]any, string) {
	return []any{
		nil,
		e.Item.DatabaseID,
		e.MimeType,
		e.URL.String(),
		e.FilePath,
	}, "INSERT INTO enclosures VALUES (?, ?, ?, ?, ?);"
}

// enclosureSerializeUpdate constructs a prepared SQL UPDATE statement from an [rss.Enclosure].
// Note that it adds all fields to SET as using some kind of diff system would be complex and overkill.
func enclosureSerializeUpdate(e *rss.Enclosure) ([]any, string) {
	return []any{
		e.MimeType,
		e.URL.String(),
		e.FilePath,
		e.DatabaseID,
	}, "UPDATE enclosures SET type = ?, url = ?, filePath = ? WHERE id = ?;"
}

// enclosureDeserialize constructs an [rss.Enclosure] from a [RowScanner]. It scans the columns in the order specified in sql/init.sql.
//
// This does not set Item; it must be set manually by the caller if needed.
func enclosureDeserialize(r RowScanner) (*rss.Enclosure, error) {
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
