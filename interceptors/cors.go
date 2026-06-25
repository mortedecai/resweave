package interceptors

import (
	"net/http"
	"strings"

	"github.com/mortedecai/resweave"
)

type CORSOption func(w *http.ResponseWriter) *http.ResponseWriter

func WithOrigin(corsHost string) CORSOption {
	return func(w *http.ResponseWriter) *http.ResponseWriter {
		(*w).Header().Set("Access-Control-Allow-Origin", corsHost)
		(*w).Header().Add("Vary", "Origin")
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

func NewCORS(opts ...CORSOption) (resweave.Interceptor, error) {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wp := &w
			for _, opt := range opts {
				wp = opt(wp)
			}
			if r.Method == http.MethodOptions {
				// Preflight responses are completed here and never reach next.
				// Any auth middleware downstream is bypassed; place it upstream of this interceptor.
				(*wp).WriteHeader(http.StatusNoContent)
				return
			}
			(*wp).Header().Del("Access-Control-Allow-Methods")
			(*wp).Header().Del("Access-Control-Allow-Headers")
			(*wp).Header().Del("Access-Control-Max-Age")
			next.ServeHTTP(*wp, r)
		})
	}, nil
}
