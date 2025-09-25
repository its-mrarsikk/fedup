package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/its-mrarsikk/fedup/shared"
)

func handlePing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "fedupd %s (go %s)\n", shared.Version, runtime.Version())
}

func RunServer(port int) {
	http.HandleFunc("/ping", handlePing)

	log.Printf("Listening on 0.0.0.0:%d", port)

	log.Print(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
