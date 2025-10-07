package database

import "github.com/its-mrarsikk/fedup/shared/rss"

func FeedSerialize(f *rss.Feed) ([]any, string) {
	return []any{f.DatabaseID,
		f.Title,
		f.Description,
		f.Link.String(),
		f.FetchFrom.String(),
		f.Language,
		f.TTL}, "(?,?,?,?,?,?,?)"
}
