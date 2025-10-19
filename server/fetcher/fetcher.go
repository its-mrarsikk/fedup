package fetcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/its-mrarsikk/fedup/shared"
	"github.com/its-mrarsikk/fedup/shared/rss"
)

var userAgent = fmt.Sprintf("fedupd/%s (+%s)", shared.Version, shared.ContactEmail)

type FetchFeed struct { // dont got a good name for this one
	url     *url.URL
	ttl     time.Duration
	ticker  *time.Ticker
	tickCh  chan time.Time
	done    chan struct{}
	fetcher *Fetcher

	last_etag     string
	last_modified time.Time
}

type Fetcher struct {
	feeds   []*FetchFeed
	mu      sync.RWMutex
	started bool
	Ch      *FetcherChannels
	client  *http.Client
}

type FetcherChannels struct {
	FetchedFeeds chan *rss.Feed
	Err          chan error
}

var ErrNewTTL = errors.New("ttl changed")

// NewFetcher constructs and returns a Fetcher with initialized FetchedFeeds and Err channels, and an http.Client.
func NewFetcher() *Fetcher {
	client := &http.Client{Timeout: 15 * time.Second}
	return &Fetcher{Ch: &FetcherChannels{FetchedFeeds: make(chan *rss.Feed, 6), Err: make(chan error, 2)}, client: client}
}

// NewFetcherWithClient constructs and returns a Fetcher with the provided Client.
func NewFetcherWithClient(client *http.Client) *Fetcher {
	return &Fetcher{Ch: &FetcherChannels{FetchedFeeds: make(chan *rss.Feed, 6), Err: make(chan error, 2)}, client: client}
}

// fetch requests the feed from the url. nil is returned for io.ReadCloser if the feed is cached.
func (ff *FetchFeed) fetch() (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, ff.url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w on feed %q", err, ff.url.String())
	}

	req.Header.Set("User-Agent", userAgent)
	if ff.last_etag != "" {
		req.Header.Set("If-None-Match", ff.last_etag)
	}
	if !ff.last_modified.IsZero() {
		req.Header.Set("If-Modified-Since", ff.last_modified.Format(http.TimeFormat))
	}

	if ff.fetcher == nil {
		return nil, errors.New("fetcher is nil")
	}
	if ff.fetcher.client == nil {
		return nil, errors.New("fetcher.client is nil")
	}

	resp, err := ff.fetcher.client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("timed out: %w", err)
		}

		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return nil, fmt.Errorf("timed out: %w", err)
		}

		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		// Get returns "" on no header
		ff.last_etag = resp.Header.Get("ETag")
		resp_lm := resp.Header.Get("Last-Modified")
		if resp_lm != "" {
			lm, err := time.Parse(http.TimeFormat, resp_lm)
			if err != nil {
				log.Printf("failed to parse Last-Modified value %q on feed %q (non-fatal)", resp_lm, ff.url.String())
			} else {
				ff.last_modified = lm
			}
		}
		return resp.Body, nil
	case http.StatusNotModified:
		resp.Body.Close()
		return nil, nil
	default:
		resp.Body.Close()
		return nil, fmt.Errorf("got unhappy status code on feed %q: %s", ff.url.String(), resp.Status)
	}
}

func (ff *FetchFeed) fetchAndParse() error {
	r, err := ff.fetch()
	if err != nil {
		return fmt.Errorf("failed to fetch feed %q: %w", ff.url.String(), err)
	}
	if r == nil {
		return nil
	}
	defer r.Close()

	parsed, err := rss.ParseRSS(r)

	if err != nil {
		return fmt.Errorf("failed to parse feed %q: %w", ff.url.String(), err)
	}

	parsed.FetchFrom = ff.url
	ff.fetcher.Ch.FetchedFeeds <- parsed

	if parsed.TTL != 0 && !(time.Duration(parsed.TTL)*time.Minute == ff.ttl) {
		ff.ttl = time.Duration(parsed.TTL) * time.Minute
		return ErrNewTTL
	}

	return nil
}

func (ff *FetchFeed) watch() {
	for {
		select {
		case <-ff.done:
			if ff.ticker != nil {
				ff.ticker.Stop()
				ff.ticker = nil
			}
			return
		case <-ff.tickCh:
			err := ff.fetchAndParse()
			if err != nil {
				if errors.Is(err, ErrNewTTL) {
					log.Printf("Fetcher %q: discovered new TTL %s, restarting", ff.url.String(), ff.ttl)
					if ff.ticker != nil {
						ff.ticker.Stop()
						ff.ticker = nil
					}
					startFeed(ff)
					return
				}
				go func(e error) { ff.fetcher.Ch.Err <- e }(err)
			}
		}
	}
}

func (f *Fetcher) AddFeed(rawurl string, optTtl *time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	parsedURL, err := url.Parse(rawurl)
	if err != nil {
		return err
	}

	var ttl time.Duration
	if optTtl != nil {
		ttl = *optTtl
		if ttl <= 0 {
			return errors.New("ttl is 0 or negative")
		}
	} else {
		ttl = 60 * time.Minute
	}

	ff := &FetchFeed{url: parsedURL, ttl: ttl, fetcher: f}
	if f.started {
		ff.ticker = time.NewTicker(ttl)
		go ff.watch()
	}

	f.feeds = append(f.feeds, ff)

	return nil
}

func startFeed(ff *FetchFeed) {
	ff.ticker = time.NewTicker(ff.ttl)
	ff.done = make(chan struct{})
	ff.tickCh = make(chan time.Time, 1)
	ff.tickCh <- time.Time{}
	go func() {
		for v := range ff.ticker.C {
			ff.tickCh <- v
		}
		close(ff.tickCh)
	}()

	go ff.watch()
}

func (f *Fetcher) Start() error {
	if f.started == true {
		return nil
	}

	if f.Ch == nil {
		return errors.New("Ch is nil")
	}
	if f.Ch.FetchedFeeds == nil {
		return errors.New("FetchedFeeds is nil")
	}
	if f.Ch.Err == nil {
		return errors.New("Err is nil")
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	for _, ff := range f.feeds {
		startFeed(ff)
	}

	f.started = true

	return nil
}

func (f *Fetcher) Stop() error {
	if f.started == false {
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	for _, ff := range f.feeds {
		close(ff.done)
		ff.done = nil
	}

	f.started = false

	return nil
}
