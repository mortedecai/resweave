package resweave

import (
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
	logHolder
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
	logHolder
}

func newHost(name HostName) Host {
	h := &host{name: name, resources: make(ResourceMap)}
	h.logHolder = newLogholder(string(name.StripPort()), h.recurse)
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
	h.Infow("Serve", "Host Name", h.Name(), "Request URI", req.RequestURI)
	pathSegs := strings.Split(req.URL.Path, "/")[1:]
	reqPaths := ResourceNames(pathSegs)
	ctx := req.Context()

	res, found := h.GetResource(reqPaths[0])
	h.Infow("Serve", "Request Path:", reqPaths[0], "Found?", found)
	if !found {
		res, found = h.GetResource(ResourceName(""))
	}
	if found {
		res.HandleCall(ctx, w, req)
		return
	}
	h.Infow("Serve", "Hard Return Code", http.StatusNotFound)
	w.WriteHeader(http.StatusNotFound)
}

func (h *host) recurse(logger *zap.SugaredLogger) {
	for _, v := range h.resources {
		v.SetLogger(logger.Named(string(h.name)), true)
	}
}
