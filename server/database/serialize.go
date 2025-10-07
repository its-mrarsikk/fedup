package database

import (
	"database/sql"
	"fmt"
	"net/url"

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
