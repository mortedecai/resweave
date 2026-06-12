package interceptors

import (
	"net/http"
	"strings"
)

type CORSOption func(w *http.ResponseWriter) *http.ResponseWriter

func WithOrigin(corsHost string) CORSOption {
	return func(w *http.ResponseWriter) *http.ResponseWriter {
		(*w).Header().Set("Access-Control-Allow-Origin", corsHost)
		return w
	}
}

func WithMethods(methods ...string) CORSOption {
	return func(w *http.ResponseWriter) *http.ResponseWriter {
		(*w).Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
		return w
	}
}

func WithHeaders(headers ...string) CORSOption {
	return func(w *http.ResponseWriter) *http.ResponseWriter {
		(*w).Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
		return w
	}
}

func AllowCredentials(allowCreds string) CORSOption {
	return func(w *http.ResponseWriter) *http.ResponseWriter {
		(*w).Header().Set("Access-Control-Allow-Credentials", allowCreds)
		return w
	}
}

func WithMaxAge(maxAge string) CORSOption {
	return func(w *http.ResponseWriter) *http.ResponseWriter {
		(*w).Header().Set("Access-Control-Max-Age", maxAge)
		return w
	}
}

func NewCORS(next http.Handler, opts ...CORSOption) (http.Handler, error) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wp := &w
		for _, opt := range opts {
			opt(wp)
		}
		next.ServeHTTP(*wp, r)
	}), nil
}
