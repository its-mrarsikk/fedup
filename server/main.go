package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/its-mrarsikk/fedup/server/httpserver"
	"github.com/its-mrarsikk/fedup/shared"
)

// graceful shutdown logic from https://stackoverflow.com/a/42533360
// i found it by accident while looking for docs on ListenAndServe, but good to have anyway lol
func stopHttp(srv *http.Server) {
	log.Printf("Shutting down HTTP server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		panic(err) // failure/timeout shutting down the server gracefully
	}
}

func main() {
	log.Printf("fedupd %s (source code under LGPL v2.1) with Go %s", shared.Version, runtime.Version())

	httpChannels := httpserver.HttpServerChannels{
		Err:          make(chan error),
		ServeContent: make(chan httpserver.Content),
	}
	mainShouldQuit := make(chan error)

	port := 4545

	go func() {
		err := <-httpChannels.Err
		switch {
		case strings.Contains(err.Error(), "address already in use"):
			log.Printf("Port %d is already in use (is fedupd already running?)", port)
			mainShouldQuit <- fmt.Errorf("Port in use")
		default:
			log.Printf("The HTTP server encountered an error: %s", err)
			mainShouldQuit <- fmt.Errorf("HTTP error: %w", err)
		}
	}()

	signalCh := make(chan os.Signal, 2)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		mainShouldQuit <- fmt.Errorf("Got signal %s", <-signalCh)
	}()

	srv := httpserver.RunServer(port, &httpChannels)

	quitReason := <-mainShouldQuit

	if quitReason != nil {
		log.Printf("Exit requested with reason: %s", quitReason)
	}

	stopHttp(srv.Server)

	log.Printf("All done, clocking out.")
}
