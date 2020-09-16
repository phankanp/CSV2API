package middleware

import (
	"context"
	"net/http"
)

func MiddlewareAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("key")

		ctx := context.WithValue(r.Context(), "key", key)

		next(w, r.WithContext(ctx))
	}
}
