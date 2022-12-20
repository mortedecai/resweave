package resweave

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Server is the resweave implementation of an opinionated resource server.
// Outside of a path prefix, the full path of a resource is determined by the resource names alone.
// This means that to have a path `v1/api/users/<id>/profile`, you will require a `v1/api` path prefix and two resources:
// * users, with an ID regex setup; and
// * profile, which is added to the users resource.
//
// This also applies to HTML resources. For example, `v1/html/somedir/index.html`, would require a `v1` path prefix,
// and an HTML resource `html`, with a top level directory `somedir` containing an `index.html` file.
type Server interface {
	AddResource(r Resource) error
	GetResource(name ResourceName) (Resource, bool)
	Run() error
	Port() int
}

// NewServer creates a new instance of a resweave Server.
func NewServer(port int) Server {
	return &server{port: port, resources: make(ResourceMap)}
}

type server struct {
	port      int
	resources ResourceMap
}

func (s *server) getRunAddr() string {
	return fmt.Sprintf(":%d", s.port)
}

func (s *server) Port() int {
	return s.port
}

func (s *server) createHTTPServer() *http.Server {
	return &http.Server{
		Addr:              s.getRunAddr(),
		ReadHeaderTimeout: 3 * time.Second,
	}
}

func (s *server) serve(w http.ResponseWriter, req *http.Request) {
	// TODO: Assumes a single, unnamed HTML resource exists; lean development
	s.resources[""].(HTMLResource).Fetch(w, req)
}

func (s *server) Run() error {
	http.HandleFunc("/", s.serve)
	return s.createHTTPServer().ListenAndServe()
}

func (s *server) AddResource(r Resource) error {
	if r == nil {
		return errors.New("cannot add a nil resource")
	}

	if _, found := s.resources[r.Name()]; found {
		return fmt.Errorf("resource '%s' already exists", r.Name())
	}

	s.resources[r.Name()] = r
	return nil
}

func (s *server) GetResource(name ResourceName) (Resource, bool) {
	r, f := s.resources[name]
	return r, f
}
