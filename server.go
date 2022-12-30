package resweave

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
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
	// AddHost adds a new Host to the Server instance with the Host being returned on success.
	// On error, Host will be nil and the relevant error will be returned.
	AddHost(name HostName) (Host, error)
	// GetHost finds an existing Host in the Server instance with the Host being returned on success.
	// On error, Host will be nil and the boolean will be false.
	GetHost(name HostName) (Host, bool)
	// AddResource adds the provided resource to the Server instance.
	// If the resource is nil or can otherwise not be added to the Server, an error will be returned.
	AddResource(r Resource) error
	// GetResource retrieves a Resource by name from the resources at the root level of this node.
	// If the provided name cannot be found, the returned resource will be nil and the boolean will be false.
	GetResource(name ResourceName) (res Resource, found bool)
	// Run runs the actual server and will return an error on failure.
	Run() error
	// Port returns the port number the server will run on.
	Port() int
	// Sets the logger to use for the server, and if recursive is true, to each of the hosts and resources.
	SetLogger(logger *zap.SugaredLogger, recursive bool)
	// AddInterceptor adds a new interceptor at the start of the handling chain.
	// For example, on an incoming request, _next_ will be called first, with any current interceptors being
	AddInterceptor(func(next http.Handler) http.Handler)
}

const (
	defaultHostName = HostName("")
)

// NewServer creates a new instance of a resweave Server.
// By default this serer will not log.
// If logging is desired, a zap.SugaredLogger may be provided to the Server.SetLogger(...) function.
// Logging can be disabled by calling Server.SetLogger(nil).
//
// Parameters:
//
// * port: The port number to run the server on
func NewServer(port int) Server {
	s := &server{port: port, hosts: make(HostMap)}
	s.hosts[defaultHostName] = newHost(defaultHostName)
	s.logHolder = newLogholder("<srv>", s.recurse)
	s.interceptor = http.HandlerFunc(s.serve)
	return s
}

type server struct {
	port        int
	hosts       HostMap
	interceptor http.Handler
	logHolder
}

func (s *server) recurse(l *zap.SugaredLogger) {
	for _, h := range s.hosts {
		h.SetLogger(l, true)
	}
}

func (s *server) getRunAddr() string {
	return fmt.Sprintf(":%d", s.port)
}

func (s *server) Port() int {
	return s.port
}

func (s *server) createHTTPServer() *http.Server {
	s.Infow("createHTTPServer", "Address", s.getRunAddr())
	return &http.Server{
		Addr:              s.getRunAddr(),
		ReadHeaderTimeout: 3 * time.Second,
	}
}

func (s *server) getDefaultHost() Host {
	return s.hosts[defaultHostName]
}

func (s *server) serve(w http.ResponseWriter, req *http.Request) {
	s.Infow("serve", "Request URI", req.URL, "Host", req.Host, "Header", req.Header)
	var host Host = s.getDefaultHost()
	hostname := HostName(req.Host)

	if h, f := s.hosts[hostname.StripPort()]; f {
		s.Infow("serve", "Hostname", hostname.StripPort(), "Found?", true, "Default?", false)
		host = h
	} else {
		s.Infow("serve", "Hostname", hostname.StripPort(), "Found?", false, "Default?", true)
	}
	host.Serve(w, req)
}

func (s *server) Run() error {
	http.Handle("/", s.interceptor)
	return s.createHTTPServer().ListenAndServe()
}

func (s *server) AddResource(r Resource) error {
	return s.getDefaultHost().AddResource(r)
}

func (s *server) GetResource(name ResourceName) (Resource, bool) {
	return s.getDefaultHost().GetResource(name)
}

func (s *server) AddHost(name HostName) (Host, error) {
	if _, found := s.hosts[name]; found {
		return nil, fmt.Errorf("host '%s' already exists", name)
	}
	h := newHost(name)
	s.hosts[name] = h
	return h, nil
}

func (s *server) GetHost(name HostName) (h Host, f bool) {
	h, f = s.hosts[name]
	return
}

func (s *server) AddInterceptor(f func(next http.Handler) http.Handler) {
	s.interceptor = f(s.interceptor)
}
