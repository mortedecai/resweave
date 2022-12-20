package resweave

import (
	"net/http"
)

// ResourceName is the identifier for the resource
type ResourceName string

// Resource defines the base operations necessary for a Resource
type Resource interface {
	Name() ResourceName
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
