package fetcher_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/its-mrarsikk/fedup/server/fetcher"
)

const (
	ttl        = 1 * time.Second
	sampleFeed = `<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"><channel><title>a</title><link>b</link><description>c</description></channel></rss>`
)

func TestFetchAndParse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, sampleFeed)
	}))
	defer srv.Close()

	f := fetcher.NewFetcher()
	f.AddFeed(srv.URL, ttl)
	err := f.Start()
	if err != nil {
		t.Fatalf("%s", err)
	}

	select {
	case feed := <-f.Ch.FetchedFeeds:
		if feed.Title != "a" {
			t.Fatalf("expected feed title %q, got %q", "a", feed.Title)
		}
		return
	case <-time.After(3 * time.Second):
		t.Fatal("timed out (no response after 3s)")
	case err := <-f.Ch.Err:
		t.Fatalf("%s", err)
	}
}

func TestLastModified(t *testing.T) {
	t.Parallel()

	lastModified, err := time.Parse(http.TimeFormat, "Thu, 22 Oct 2009 00:00:00 GMT")
	if err != nil {
		panic(err)
	}

	first := true
	done := make(chan any, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqModifiedSince := r.Header.Get("If-Modified-Since")
		if first {
			if reqModifiedSince != "" {
				t.Fatalf("expected no If-Modified-Since on first request, got %q", reqModifiedSince)
			}
			first = false
		} else {
			parsed, err := time.Parse(http.TimeFormat, reqModifiedSince)
			if err != nil {
				t.Fatalf("invalid If-Modified-Since: %q", reqModifiedSince)
			}

			if parsed != lastModified {
				t.Fatalf("expected If-Modified-Since %q, got %q", lastModified, reqModifiedSince)
			}
			done <- nil
		}
		w.Header().Set("Last-Modified", lastModified.Format(http.TimeFormat))
		fmt.Fprint(w, sampleFeed)
	}))
	defer srv.Close()

	f := fetcher.NewFetcher()
	f.AddFeed(srv.URL, ttl)

	err = f.Start()
	if err != nil {
		t.Fatalf("failed to start fetcher: %s", err)
	}

	select {
	case <-done:
		// success
	case <-time.After(3 * time.Second):
		t.Fatal("timed out (no response after 3s)")
	case err := <-f.Ch.Err:
		t.Fatalf("%s", err)
	}
}

func TestETag(t *testing.T) {
	t.Parallel()

	const etag = "abc123"

	first := true
	done := make(chan any, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqETag := r.Header.Get("If-None-Match")
		if first {
			if reqETag != "" {
				t.Fatalf("expected no If-None-Match on first request, got %q", reqETag)
			}
			first = false
		} else {
			if reqETag != etag {
				t.Fatalf("expected If-None-Match %q, got %q", etag, reqETag)
			}
			done <- nil
		}

		w.Header().Set("ETag", etag)
		fmt.Fprint(w, sampleFeed)
	}))
	defer srv.Close()

	f := fetcher.NewFetcher()
	f.AddFeed(srv.URL, ttl)

	err := f.Start()
	if err != nil {
		t.Fatalf("failed to start fetcher: %s", err)
	}

	select {
	case <-done:
		// success
	case <-time.After(3 * time.Second):
		t.Fatal("timed out (no response after 3s)")
	case err := <-f.Ch.Err:
		t.Fatalf("%s", err)
	}
}

func TestRespectsTTL(t *testing.T) {
	t.Parallel()

	var last *time.Time
	done := make(chan any)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if last == nil {
			t := time.Now()
			last = &t
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, sampleFeed)
			return
		}
		since := time.Since(*last)
		if since < ttl-(100*time.Millisecond) {
			t.Fatalf("fetcher requested faster than %s (%s)", ttl, since)
		}
		done <- nil
	}))
	defer srv.Close()

	f := fetcher.NewFetcher()
	f.AddFeed(srv.URL, ttl)
	err := f.Start()
	if err != nil {
		t.Fatalf("failed to start fetcher: %s", err)
	}

	select {
	case <-done:
		// success
	case <-time.After(3 * time.Second):
		t.Fatal("timed out (no response after 3s)")
	case err := <-f.Ch.Err:
		t.Fatalf("%s", err)
	}
}

func TestCached304(t *testing.T) {
	t.Parallel()

	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.Header().Set("ETag", `"x"`)
			fmt.Fprint(w, sampleFeed)
			return
		}
		w.WriteHeader(http.StatusNotModified)
	}))
	defer srv.Close()

	f := fetcher.NewFetcher()
	if err := f.AddFeed(srv.URL, ttl); err != nil {
		t.Fatalf("AddFeed error: %v", err)
	}
	if err := f.Start(); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer func() { _ = f.Stop() }()

	select {
	case <-f.Ch.FetchedFeeds:
	case err := <-f.Ch.Err:
		t.Fatalf("unexpected error: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first feed")
	}

	// wait a few TTLs and assert no second feed
	select {
	case feed := <-f.Ch.FetchedFeeds:
		t.Fatalf("unexpected second feed: %+v", feed)
	case err := <-f.Ch.Err:
		t.Fatalf("unexpected error: %v", err)
	case <-time.After(3 * ttl):
		// success
	}
}

func TestTimeout(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		fmt.Fprint(w, sampleFeed)
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 50 * time.Millisecond}
	f := fetcher.NewFetcherWithClient(client)

	if err := f.AddFeed(srv.URL, ttl); err != nil {
		t.Fatalf("AddFeed error: %v", err)
	}
	if err := f.Start(); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer func() { _ = f.Stop() }()

	select {
	case e := <-f.Ch.Err:
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatalf("expected deadline exceeded (timeout), got: %v", e)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for timeout error") // should i test the timeout timeout as well?
	}
}

func TestStop(t *testing.T) {
	t.Parallel()

	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		fmt.Fprint(w, sampleFeed)
	}))
	defer srv.Close()

	f := fetcher.NewFetcher()
	if err := f.AddFeed(srv.URL, ttl); err != nil {
		t.Fatalf("AddFeed error: %v", err)
	}
	if err := f.Start(); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	select {
	case <-f.Ch.FetchedFeeds:
	case err := <-f.Ch.Err:
		t.Fatalf("unexpected error: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for initial fetch")
	}

	if err := f.Stop(); err != nil {
		t.Fatalf("Stop error: %v", err)
	}
	countAfterStop := atomic.LoadInt32(&calls)

	// wait more than a few TTLs and verify no more requests
	time.Sleep(3 * ttl)
	if atomic.LoadInt32(&calls) > countAfterStop {
		t.Fatalf("expected no more requests after Stop, calls grew from %d to %d", countAfterStop, calls)
	}
}
