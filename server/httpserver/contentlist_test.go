package httpserver_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/its-mrarsikk/fedup/server/httpserver"
)

func TestAddRemoveContent(t *testing.T) {
	c := httpserver.Content{
		Path: "test",
		Handler: func(path, contentType string, w http.ResponseWriter, r *http.Request) {

		},
		ContentType: "text/plain; charset=utf-8",
	}
	ch.ServeContent <- c
	time.Sleep(100 * time.Microsecond) // i haven't found a better solution than the ol' reliable sleep. i will not be changing code outside of the server because this isn't how the api is meant to be used, but this is a unit test
	if _, ok := srv.Contents[c.Path]; !ok {
		t.Fatalf("tried serving content at %s, expected %v got nil", c.Path, c)
	}

	ch.RemoveContent <- c.Path
	time.Sleep(100 * time.Microsecond)
	if _, ok := srv.Contents[c.Path]; ok {
		t.Fatalf("tried removing content at %s, expected nil got %v", c.Path, c)
	}
}
