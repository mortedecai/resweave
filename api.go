package resweave

import (
	"context"
	"net/http"
)

// APIResource is a basic APIResource which has a single point of entry for serving the supported access methods.
type APIResource interface {
	Resource
	logHolder
	SetList(f ResweaveFunc)
	SetCreate(f ResweaveFunc)
}

// BaseAPIRes supplies the basic building blocks for an APIResource.
// It may be used through composition
type BaseAPIRes struct {
	logHolder
	name       ResourceName
	listFunc   ResweaveFunc
	createFunc ResweaveFunc
}

// NewAPI creates a new APIResource instance with the provided name.
func NewAPI(name ResourceName) APIResource {
	bar := &BaseAPIRes{name: name, logHolder: newLogholder(name.String(), nil)}
	bar.listFunc = bar.defaultFunction
	bar.createFunc = bar.defaultFunction
	return bar
}

func (bar *BaseAPIRes) Name() ResourceName {
	return bar.name
}

func (bar *BaseAPIRes) defaultFunction(_ context.Context, w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (bar *BaseAPIRes) SetList(f ResweaveFunc) {
	bar.listFunc = f
}

func (bar *BaseAPIRes) SetCreate(f ResweaveFunc) {
	bar.createFunc = f
}

func (bar *BaseAPIRes) HandleCall(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		bar.listFunc(ctx, w, req)
	case http.MethodPost:
		bar.createFunc(ctx, w, req)
	default:
		bar.defaultFunction(ctx, w, req)
	}
}
