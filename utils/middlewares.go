package utils

import (
	"crypto/subtle"
	"net/http"

	"github.com/rs/zerolog"
)

// ZerologMiddleware returns a middleware that logs HTTP requests using the provided zerolog.Logger
func ZerologMiddleware(logger zerolog.Logger) func(http.Handler) http.Handler {
	childLog := ChildLogger("http").With().Logger()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			childLog.Info().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Msg("HTTP request")
			next.ServeHTTP(w, r)
		})
	}
}

// APIKeyMiddleware rejects requests that don't carry a valid X-API-Key header.
// If apiKey is empty, all requests are rejected (fail-closed).
func APIKeyMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			provided := r.Header.Get("X-API-Key")
			if apiKey == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(apiKey)) != 1 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
