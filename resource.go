package resweave

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

// ResourceName is the identifier for the resource
type ResourceName string

type ResweaveFunc func(ctx context.Context, w http.ResponseWriter, req *http.Request)

// ResourceNames is a slice of ResourceName instances
func ResourceNames(n []string) []ResourceName {
	names := make([]ResourceName, len(n))
	for i, v := range n {
		names[i] = ResourceName(v)
	}
	return names
}

// String converts the ResourceName to a string
func (rn ResourceName) String() string {
	return string(rn)
}

// Resource defines the base operations necessary for a Resource
type Resource interface {
	Name() ResourceName
	Logger() *zap.SugaredLogger
	HandleCall(context.Context, http.ResponseWriter, *http.Request)
	SetLogger(logger *zap.SugaredLogger, recursive bool)
}

// ResourceMap is a type alias for a map of ResourceNames to Resources (map[ResourceName]Resource)
type ResourceMap map[ResourceName]Resource
