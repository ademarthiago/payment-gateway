package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", wrapped.status).
			Dur("latency", time.Since(start)).
			Str("remote_addr", r.RemoteAddr).
			Str("request_id", r.Header.Get("X-Request-ID")).
			Msg("request")
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
