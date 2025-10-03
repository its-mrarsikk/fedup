package httpserver_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/its-mrarsikk/fedup/server/httpserver"
)

var (
	srv *httpserver.Server
	ch  *httpserver.HttpServerChannels
)

const port = 55928

func TestMain(m *testing.M) {
	ch = &httpserver.HttpServerChannels{Err: make(chan error), ServeContent: make(chan httpserver.Content), RemoveContent: make(chan string)}

	srv = httpserver.RunServer(port, ch)

	code := m.Run()

	fmt.Println("shutting down server")

	srv.Shutdown(context.Background())
	os.Exit(code)
}
