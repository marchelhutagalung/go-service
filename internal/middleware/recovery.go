package middleware

import (
	"github.com/marchelhutagalung/go-service/internal/logger"
	"github.com/marchelhutagalung/go-service/internal/response"
	"net/http"
	"runtime/debug"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				stackTrace := string(debug.Stack())

				// Log the panic with details
				logger.WithFields(map[string]interface{}{
					"error":       err,
					"stack":       stackTrace,
					"path":        r.URL.Path,
					"method":      r.Method,
					"remote_addr": r.RemoteAddr,
				}).Error("PANIC RECOVERED")

				// Return a 500 error response
				response.ErrorResponse(w, http.StatusInternalServerError, "Internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
