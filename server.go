package resweave

import "net/http"

type Server interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}
