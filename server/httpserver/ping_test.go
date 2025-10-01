package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestPing(t *testing.T) {
	port := 55928
	ch := HttpServerChannels{Err: make(chan error)}

	go func() {
		err := <-ch.Err
		t.Fatalf("server had error: %s", err) // i've found no good way to do this error listening in the test goroutine
	}()

	srv := RunServer(port, &ch)
	defer func() {
		if err := srv.Shutdown(context.Background()); err != nil && err != http.ErrServerClosed {
			t.Fatalf("failed to shutdown: %s", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/ping", port))
	if err != nil {
		t.Fatalf("error making GET: %s", err)
	}

	buffer := make([]byte, 32)
	_, err = resp.Body.Read(buffer)
	if err != nil {
		t.Fatalf("error reading body: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		t.Fatalf("error closing body: %s", err)
	}

	if !strings.Contains(string(buffer), "fedupd") {
		t.Fatalf("wanted response body containing 'fedupd', got '%s'", string(buffer))
	}

}
