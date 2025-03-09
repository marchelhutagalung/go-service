package middleware

import (
	"context"
	"errors"
	"github.com/marchelhutagalung/go-service/internal/auth"
	"github.com/marchelhutagalung/go-service/internal/response"
	"net/http"
	"strings"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
)

type Middleware struct {
	jwtService *auth.JWTService
}

func NewMiddleware(jwtService *auth.JWTService) *Middleware {
	return &Middleware{
		jwtService: jwtService,
	}
}

// RequireAuth is a middleware that requires JWT authentication
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.ErrorResponse(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			response.ErrorResponse(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		tokenString := headerParts[1]

		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			var statusCode int
			var message string

			if errors.Is(err, auth.ErrExpiredToken) {
				statusCode = http.StatusUnauthorized
				message = "Token has expired"
			} else {
				statusCode = http.StatusForbidden
				message = "Invalid token"
			}

			response.ErrorResponse(w, statusCode, message)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID extracts the user ID from the request context
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}
