package interceptors

import (
	"errors"
	"net/http"
	"strings"
)

type CORSOption func(optVal string) func(w *http.ResponseWriter) *http.ResponseWriter

func WithOrigin(corsHost string) func(w *http.ResponseWriter) *http.ResponseWriter {
	return func(w *http.ResponseWriter) *http.ResponseWriter {
		(*w).Header().Set("Access-Control-Allow-Origin", corsHost)
		return w
	}
}

func WithMethods(methods string) func(w *http.ResponseWriter) *http.ResponseWriter {
	return func(w *http.ResponseWriter) *http.ResponseWriter {
		(*w).Header().Set("Access-Control-Allow-Methods", methods)
		return w
	}
}

func WithHeaders(headers string) func(w *http.ResponseWriter) *http.ResponseWriter {
	return func(w *http.ResponseWriter) *http.ResponseWriter {
		(*w).Header().Set("Access-Control-Allow-Headers", headers)
		return w
	}
}

func AllowCredentials(allowCreds string) func(w *http.ResponseWriter) *http.ResponseWriter {
	return func(w *http.ResponseWriter) *http.ResponseWriter {
		(*w).Header().Set("Access-Control-Allow-Credentials", allowCreds)
		return w
	}
}

func WithMaxAge(maxAge string) func(w *http.ResponseWriter) *http.ResponseWriter {
	return func(w *http.ResponseWriter) *http.ResponseWriter {
		(*w).Header().Set("Access-Control-Max-Age", maxAge)
		return w
	}
}

func NewCORS(next http.Handler, corsHost string) (http.Handler, error) {
	if strings.TrimSpace(corsHost) == "" {
		return nil, errors.New("cors host is empty")
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wp := &w
		(*wp).Header().Set("Access-Control-Allow-Origin", corsHost)
		next.ServeHTTP(*wp, r)
	}), nil
}
