package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"

	"pocketjson/storage"
)

func RateLimit(store *storage.Store) func(http.Handler) http.Handler {
	cfg := store.Config()

	return func(next http.Handler) http.Handler {
		limiter := httprate.Limit(
			cfg.RequestLimit,
			1*time.Minute,
			httprate.WithKeyFuncs(httprate.KeyByIP),
		)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			isAuth, _, err := store.ValidateApiKey(r.Context(), apiKey)
			if err != nil {
				http.Error(w, "Server error", http.StatusInternalServerError)
				return
			}

			if !isAuth {
				limiter(next).ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
