package httpserver_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestPing(t *testing.T) {
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
