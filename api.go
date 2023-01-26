package resweave

import (
	"context"
	"net/http"
)

// APIResource is a basic APIResource which has a single point of entry for serving the supported access methods.
type APIResource interface {
	logHolder
	ResourceLister
	SetList(f ListFunc)
}

// BaseAPIRes supplies the basic building blocks for an APIResource.
// It may be used through composition
type BaseAPIRes struct {
	logHolder
	name     ResourceName
	listFunc ListFunc
}

// NewAPI creates a new APIResource instance with the provided name.
func NewAPI(name ResourceName) APIResource {
	bar := &BaseAPIRes{name: name, logHolder: newLogholder(name.String(), nil)}
	bar.listFunc = bar.defaultFunction
	return bar
}

func (bar *BaseAPIRes) Name() ResourceName {
	return bar.name
}

func (bar *BaseAPIRes) defaultFunction(_ context.Context, w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (bar *BaseAPIRes) List(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	bar.listFunc(ctx, w, req)
}

func (bar *BaseAPIRes) SetList(f ListFunc) {
	bar.listFunc = f
}
