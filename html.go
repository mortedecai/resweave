package resweave

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// HTMLResource represents an HTML file server.
// The resource itself only supports the Fetch functionality, with the remainder of the path being an input
// into the listener
type HTMLResource interface {
	Resource
	BaseDir() string
	FullPath() ResourceName
}

type htmlResource struct {
	LogHolder
	name    ResourceName
	base    string
	handler http.Handler
}

// NewHTML creats a new HTMLResource for use with a resweave Server
func NewHTML(name ResourceName, baseDir string) HTMLResource {
	// HTML resources never have sub resources; no recurser function necessary.
	return &htmlResource{name: name, base: baseDir, LogHolder: NewLogholder(name.String(), nil)}
}

func (h *htmlResource) Name() ResourceName {
	return h.name
}

func (h *htmlResource) HandleCall(_ context.Context, w http.ResponseWriter, req *http.Request) {
	f, err := os.Stat(h.base)
	if err != nil {
		h.Infow("Fetch", "Stat Base", h.base, "Error?", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.Infow("Fetch", "Is directory?", f.IsDir())
	if h.handler == nil {
		h.handler = http.StripPrefix(h.FullPath().String(), http.FileServer(http.Dir(h.base)))
	}
	h.handler.ServeHTTP(w, req)
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
