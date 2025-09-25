package main

import (
	"log"
	"runtime"

	"github.com/its-mrarsikk/fedup/shared"
)

func main() {
	log.Printf("fedupd %s (source code under LGPL v2.1) with Go %s", shared.Version, runtime.Version())
	go RunServer(4545)
	for {
	} // keep it busy until termination
}
