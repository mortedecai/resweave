package resweave

import (
	"net/http"

	"go.uber.org/zap"
)

// ResourceName is the identifier for the resource
type ResourceName string

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
	SetLogger(logger *zap.SugaredLogger, recursive bool)
}

// ResourceLister interface defines the operations necessary for any Resource which provides resource List functionality
type ResourceLister interface {
	Resource
	List(http.ResponseWriter, *http.Request)
}

// ResourceCreator interface defines the operations necessary for any Resource which provides resource Create functionality
type ResourceCreator interface {
	Resource
	Create(http.ResponseWriter, *http.Request)
}

// ResourceFetcher interface defines the operations necessary for any Resource which provides resource Fetch functionality
type ResourceFetcher interface {
	Resource
	Fetch(http.ResponseWriter, *http.Request)
}

// ResourceUpdater interface defines the operations necessary for any Resource which provides resource Update functionality
type ResourceUpdater interface {
	Resource
	Update(http.ResponseWriter, *http.Request)
}

// ResourceDeleter interface defines the operations necessary for any Resource which provides resource Delete functionality
type ResourceDeleter interface {
	Resource
	Delete(http.ResponseWriter, *http.Request)
}

// ResourceMap is a type alias for a map of ResourceNames to Resources (map[ResourceName]Resource)
type ResourceMap map[ResourceName]Resource
