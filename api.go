package resweave

import (
	"net/http"
)

// APIResource is a basic APIResource which has a single point of entry for serving the supported access methods.
type APIResource interface {
	logHolder
	Resource
	List(w http.ResponseWriter, req *http.Request)
}

// BaseAPIRes supplies the basic building blocks for an APIResource.
// It may be used through composition
type BaseAPIRes struct {
	logHolder
	name ResourceName
}

// NewAPI creates a new APIResource instance with the provided name.
func NewAPI(name ResourceName) APIResource {
	bar := &BaseAPIRes{name: name, logHolder: newLogholder(name.String(), nil)}
	return bar
}

func (bar *BaseAPIRes) Name() ResourceName {
	return bar.name
}

func (bar *BaseAPIRes) List(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	respBytes := []byte("Hello, World!")
	if bw, err := w.Write(respBytes); err != nil {
		bar.Infow("List", "WriteError", err, "BytesWritten", bw)
	} else {
		bar.Debugw("List", "BytesWritten", bw)
	}
}
