package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/its-mrarsikk/fedup/shared"
)

type HttpServerChannels struct {
	Err chan error
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "fedupd %s (go %s)\n", shared.Version, runtime.Version())
}

// graceful shutdown logic from https://stackoverflow.com/a/42533360
// i found it by accident while looking for docs on ListenAndServe, but good to have anyway lol
func RunServer(port int, ch *HttpServerChannels) *http.Server {
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}

	http.HandleFunc("/ping", handlePing)

	go func() {
		// always returns error. ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("ListenAndServe(): %T: %v", err, err)
			ch.Err <- err
		}
	}()

	log.Printf("Listening on 0.0.0.0:%d", port)

	return srv
}
