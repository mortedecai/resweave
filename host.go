package resweave

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// Host represents a unique Host to serve resources for.
type Host interface {
	// Name returns the HostName for this Host instance.
	Name() HostName
	// TopLevelResourceCount returns the number of top level resources this host has.
	TopLevelResourceCount() int
	// AddResource adds the provided resource to the Server instance.
	// If the resource is nil or can otherwise not be added to the Server (e.g. one with the same name has already been added),
	// an error will be returned.
	AddResource(r Resource) error
	// GetResource retrieves the top level resource identified by the provided name and sets found to true.
	// If the resource is not able to be found at the top level, nil with an explanatory error will be returned.
	GetResource(name ResourceName) (res Resource, found bool)
	// Serve handles serving the resources under the Host.
	Serve(w http.ResponseWriter, req *http.Request)
	LogHolder
}

// HostName is a type for managing host names in the resweave system.
type HostName string

// StripPort strips any port information from the provided HostName. e.g. localhost:8080 -> localhost
func (h HostName) StripPort() HostName {
	return HostName(strings.Split(string(h), ":")[0])
}

// HostMap is a convenience alias to a map of HostNames to Hosts
type HostMap map[HostName]Host

type host struct {
	name      HostName
	resources ResourceMap
	LogHolder
}

func newHost(name HostName) Host {
	h := &host{name: name, resources: make(ResourceMap)}
	h.LogHolder = NewLogholder(string(name.StripPort()), h.recurse)
	return h
}

func (h *host) Name() HostName {
	return h.name
}

func (h *host) TopLevelResourceCount() int {
	return len(h.resources)
}

func (h *host) AddResource(r Resource) error {
	if r == nil {
		h.Infow("AddResource", "Error", "resource was nil")
		return errors.New("cannot add a nil resource")
	}

	if _, found := h.resources[r.Name()]; found {
		h.Infow("AddResource", "Name", r.Name(), "Exists?", found)
		return fmt.Errorf(FmtResourceAlreadyExists, r.Name(), h.Name())
	}
	h.resources[r.Name()] = r
	h.Infow("AddResource", "Name", fmt.Sprintf("'%s'", r.Name()), "Added", true)
	return nil
}

func (h *host) GetResource(name ResourceName) (res Resource, found bool) {
	res, found = h.resources[name]
	return
}

func (h *host) Serve(w http.ResponseWriter, req *http.Request) {
	h.Infow("serve", "Host Name", h.Name(), "Request URI", req.RequestURI)
	var reqPaths []ResourceName
	if strings.HasSuffix(req.URL.Path, "/") {
		reqPaths = ResourceNames(strings.Split(req.URL.Path[:len(req.URL.Path)-1], "/"))
	} else {
		reqPaths = ResourceNames(strings.Split(req.URL.Path, "/"))
	}
	ctx := req.Context()
	pathIdx := 0
	leadSlash := false
	if strings.HasPrefix(req.URL.Path, "/") {
		pathIdx = 1
		leadSlash = true
	}

	if pathIdx >= len(reqPaths) {
		pathIdx = 0
	}
	res, found := h.GetResource(reqPaths[pathIdx])
	h.Infow("serve", "Request Path:", reqPaths[pathIdx], "Found?", found)
	if !found {
		pathIdx = 0
		if !leadSlash {
			reqPaths = append(ResourceNames([]string{""}), reqPaths...)
		}
		res, found = h.GetResource(ResourceName(""))
	}
	if found {
		ctx = context.WithValue(ctx, KeyURISegments, reqPaths[pathIdx:])
		res.HandleCall(ctx, w, req)
		return
	}
	h.Infow("serve", "Hard Return Code", http.StatusNotFound)
	w.WriteHeader(http.StatusNotFound)
}

func (h *host) recurse(logger *zap.SugaredLogger) {
	for _, v := range h.resources {
		v.SetLogger(logger.Named(string(h.name)), true)
	}
}
