package resweave

import "net/http"

// HTMLResource represents an HTML file server.
// The resource itself only supports the Fetch functionality, with the remainder of the path being an input
// into the listener
type HTMLResource interface {
	ResourceFetcher
	BaseDir() string
}

type htmlResource struct {
	name ResourceName
	base string
}

// NewHTML creats a new HTMLResource for use with a resweave Server
func NewHTML(name ResourceName, baseDir string) HTMLResource {
	return &htmlResource{name: name, base: baseDir}
}

func (h *htmlResource) Name() ResourceName {
	return ""
}

func (h *htmlResource) Fetch(w http.ResponseWriter, req *http.Request) {
	hndlr := http.FileServer(http.Dir(h.base))
	hndlr.ServeHTTP(w, req)
}

func (h *htmlResource) BaseDir() string {
	return h.base
}
