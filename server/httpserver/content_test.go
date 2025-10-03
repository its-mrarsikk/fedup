package httpserver_test

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/its-mrarsikk/fedup/server/httpserver"
)

const ct_path = "test"

func TestContent(t *testing.T) {
	c := httpserver.Content{
		Path: ct_path,
		Handler: func(path string, contentType string, w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Invoked at %s with %s", path, contentType)
		},
		ContentType: "text/plain; charset=utf-8",
	}
	ch.ServeContent <- c

	var resp *http.Response
	var err error
	retries := 0

	for {
		resp, err = http.Get(fmt.Sprintf("http://localhost:%d/content/test", port))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode == 200 {
			break
		}
		retries++
		if retries > 5 {
			t.Fatalf("never got 200, last status %d", resp.StatusCode)
		}
		time.Sleep(10 * time.Millisecond)
	}

	b, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("failed to read body: %s", err)
	}
	text := string(b)
	resp.Body.Close()

	expect := fmt.Sprintf("Invoked at %s with %s", c.Path, c.ContentType)
	if text != expect {
		t.Fatalf("expected response body '%s', got '%s'", expect, text)
	}
}

func TestContentNotFound(t *testing.T) {
	url := fmt.Sprintf("http://127.0.0.1:%d/content/go_is_a_great_language", port)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to make GET: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Fatalf("requested %s, expected status code %d, got %d", url, 404, resp.StatusCode)
	}
}
