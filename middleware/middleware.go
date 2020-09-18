package middleware

import (
	"context"
	"net/http"
)

// Gets api key from header and adds to context
func MiddlewareAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("key")

		ctx := context.WithValue(r.Context(), "key", key)

		next(w, r.WithContext(ctx))
	}
}
