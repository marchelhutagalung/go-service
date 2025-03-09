package middleware

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/marchelhutagalung/go-service/internal/logger"
	"net/http"
	"time"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			// Create fields map for structured logging
			fields := map[string]interface{}{
				"method":        r.Method,
				"path":          r.URL.Path,
				"remote_addr":   r.RemoteAddr,
				"user_agent":    r.UserAgent(),
				"status":        ww.Status(),
				"bytes_written": ww.BytesWritten(),
				"duration_ms":   time.Since(start).Milliseconds(),
			}

			// Add request ID if available
			if reqID := middleware.GetReqID(r.Context()); reqID != "" {
				fields["request_id"] = reqID
			}

			// Log with appropriate level based on status code
			entry := logger.WithFields(fields)
			statusCode := ww.Status()

			switch {
			case statusCode >= 500:
				entry.Error("Server error")
			case statusCode >= 400:
				entry.Warn("Client error")
			case statusCode >= 300:
				entry.Info("Redirection")
			default:
				entry.Info("Success")
			}
		}()

		next.ServeHTTP(ww, r)
	})
}
