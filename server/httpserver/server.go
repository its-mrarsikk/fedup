package httpserver

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"

	"github.com/its-mrarsikk/fedup/shared"
)

// An extension of http.Server
type Server struct {
	*http.Server // embed
	Contents     ContentList
	contentMutex sync.RWMutex
}

type HttpServerChannels struct {
	Err           chan error
	ServeContent  chan Content
	RemoveContent chan string
}

/*
Returns a [net/http/HandlerFunc] closure that calls http.Error to respond with the provided code and msg.
*/
func statusCode(code int, msg string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, msg, code)
	}
}

func (self *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "fedupd %s (go %s)\n", shared.Version, runtime.Version())
}

// graceful shutdown logic from https://stackoverflow.com/a/42533360
// i found it by accident while looking for docs on ListenAndServe, but good to have anyway lol
func RunServer(port int, ch *HttpServerChannels) *Server {
	srv := &Server{
		Server:   &http.Server{Addr: fmt.Sprintf(":%d", port)},
		Contents: make(ContentList),
	}

	http.HandleFunc("/", statusCode(400, "use subroutes"))
	http.HandleFunc("/ping", srv.handlePing)
	http.HandleFunc("/content", statusCode(400, "use subroutes"))
	http.HandleFunc("/content/", srv.handleContent)

	go func() {
		// always returns error. ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("ListenAndServe(): %T: %v", err, err)
			ch.Err <- err
		}
	}()

	go func() {
		for {
			c := <-ch.ServeContent
			srv.AddContent(c)
		}
	}()

	go func() {
		for {
			c := <-ch.RemoveContent
			srv.RemoveContent(c)
		}
	}()

	log.Printf("Listening on 0.0.0.0:%d", port)

	return srv
}
