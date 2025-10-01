package httpserver

import (
	"net/http"
	"strings"
)

// A handler to be called whenever a /content/ endpoint is requested.
type ContentHandler func(path string, contentType string, w http.ResponseWriter, r *http.Request)

// A map of Content items to serve. The key is the path with no prefix, like "images/tux.jpg"
type ContentList map[string]Content

/*
Content represents an item to be served via HTTP. The path is appended to /content/ to form a /content/<path> endpoint.
When a request to the endpoint is made, the handler func must be called.
The contentType is identical to a Content-Type header. It must be added to the response, and can be used as a guide for what to return in the handler.
*/
type Content struct {
	Path        string
	Handler     ContentHandler
	ContentType string
}

func (self *Server) handleContent(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Path
	path := strings.TrimPrefix(uri, "/content/")
	c, ok := self.Contents[path]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	c.Handler(path, c.ContentType, w, r)
}

func (self *Server) AddContent(c Content) {
	self.contentMutex.Lock()
	defer self.contentMutex.Unlock()

	self.Contents[c.Path] = c
}

func (self *Server) RemoveContent(p string) {
	self.contentMutex.Lock()
	defer self.contentMutex.Unlock()

	delete(self.Contents, p)
}
