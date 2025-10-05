package rss

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/etree"
)

func channelElementToFeed(e *etree.Element) (*Feed, error) {
	if !strings.Contains(e.Tag, "channel") {
		return nil, errors.New("element is not a <channel>")
	}

	feed := &Feed{}

	if title := e.SelectElement("title"); title == nil {
		return nil, errors.New("<channel> does not contain <title>")
	} else {
		feed.Title = title.Text()
	}

	if description := e.SelectElement("description"); description == nil {
		return nil, errors.New("<channel> does not contain <description>")
	} else {
		feed.Description = description.Text()
	}

	if link := e.SelectElement("link"); link == nil {
		return nil, errors.New("<channel> does not contain <link>")
	} else {
		parsedUrl, err := url.Parse(link.Text())
		if err != nil {
			return nil, fmt.Errorf("failed to parse url: %w", err)
		}
		feed.Link = parsedUrl
	}

	// spec violation: allow any value for <language>
	if language := e.SelectElement("language"); language != nil {
		feed.Language = language.Text()
	}

	if ttl := e.SelectElement("ttl"); ttl != nil {
		parsedTTL, err := strconv.Atoi(ttl.Text())
		if err != nil {
			return nil, fmt.Errorf("malformed feed: ttl is not a number (got %s)", ttl.Text())
		}
		feed.TTL = parsedTTL
	}

	return feed, nil
}

func parseRSSDate(dateStr string) (time.Time, error) {
	layouts := []string{
		time.RFC1123Z, // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC1123,  // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC822Z,  // "02 Jan 06 15:04 -0700"
		time.RFC822,   // "02 Jan 06 15:04 MST"
	}

	var lastErr error
	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		} else {
			lastErr = err
		}
	}
	return time.Time{}, fmt.Errorf("could not parse RSS date %q: %w", dateStr, lastErr)
}

func itemElementToItem(e *etree.Element, feed *Feed) (*Item, error) {
	if !strings.Contains(e.Tag, "item") {
		return nil, errors.New("element is not an <item>")
	}

	// TODO: add error handling
	item := &Item{
		Feed:        feed,
		GUID:        e.SelectElement("guid").NotNil().Text(),
		Title:       e.SelectElement("title").NotNil().Text(),
		Description: e.SelectElement("description").NotNil().Text(),
		Link:        func() *url.URL { url, _ := url.Parse(e.SelectElement("link").NotNil().Text()); return url }(),
		Author:      e.SelectElement("author").NotNil().Text(),
		PubDate:     func() *time.Time { date, _ := parseRSSDate(e.SelectElement("pubDate").NotNil().Text()); return &date }(),
	}

	var enclosure *Enclosure
	if enclosureElem := e.SelectElement("enclosure"); enclosureElem != nil {
		// TODO: fix spec violation: allowing nullable enclosure elements
		enclosure = &Enclosure{
			URL:      func() *url.URL { url, _ := url.Parse(enclosureElem.SelectAttrValue("url", "")); return url }(),
			MimeType: enclosureElem.SelectAttrValue("type", ""),
			Length:   func() int { i, _ := strconv.Atoi(enclosureElem.SelectAttrValue("length", "0")); return i }(),
		}
	}
	item.Enclosure = enclosure

	return item, nil
}

// ParseRSS takes a reader with RSS XML and converts it to a Feed object
// This parser is not fully up to spec: it allows enclosure subelements to be null; allows both title and description of an item to be null
func ParseRSS(r io.Reader) (*Feed, error) {
	doc := etree.NewDocument()
	if _, err := doc.ReadFrom(r); err != nil {
		return nil, err
	}

	root := doc.SelectElement("rss")
	if root == nil {
		return nil, errors.New("xml does not have <rss> tag")
	}

	channel := root.SelectElement("channel")
	if channel == nil {
		return nil, errors.New("xml does not have <channel> tag")
	}

	feed, err := channelElementToFeed(channel)
	if err != nil {
		return nil, fmt.Errorf("failed to parse <channel> element: %w", err)
	}

	for itemElem := range channel.SelectElementsSeq("item") {
		item, err := itemElementToItem(itemElem, feed)
		if err != nil {
			log.Printf("failed to parse <item> element in feed %q: %s", feed.Title, err)
			continue
		}
		feed.Items = append(feed.Items, item)
	}

	return feed, nil
}
