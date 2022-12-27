package resweave

import (
	"fmt"
	"net/http"
	"strings"
)

// HTMLResource represents an HTML file server.
// The resource itself only supports the Fetch functionality, with the remainder of the path being an input
// into the listener
type HTMLResource interface {
	ResourceFetcher
	BaseDir() string
	FullPath() ResourceName
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
	return h.name
}

func (h *htmlResource) Fetch(w http.ResponseWriter, req *http.Request) {
	hndlr := http.StripPrefix(h.FullPath().String(), http.FileServer(http.Dir(h.base)))
	hndlr.ServeHTTP(w, req)
}

func (h *htmlResource) BaseDir() string {
	return h.base
}

func (h *htmlResource) FullPath() ResourceName {
	fp := "/%s"
	if !strings.HasPrefix(h.name.String(), "/") {
		fp = fmt.Sprintf(fp, h.name)
	} else {
		fp = fmt.Sprintf(fp, h.name[1:])
	}
	if !strings.HasSuffix(fp, "/") {
		fp = fp + "/"
	}
	return ResourceName(fp)
}
