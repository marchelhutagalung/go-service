package router

import (
	"github.com/marchelhutagalung/go-service/internal/handlers"
	customMiddleware "github.com/marchelhutagalung/go-service/internal/middleware"
	"github.com/marchelhutagalung/go-service/internal/response"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func SetupRouter(
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	movieHandler *handlers.MovieHandler,
	authMiddleware *customMiddleware.Middleware,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(customMiddleware.RequestLogger)
	r.Use(customMiddleware.Recovery)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)

			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequireAuth)
				r.Post("/logout", authHandler.Logout)
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			r.Get("/me", userHandler.GetCurrentUser)
		})

		// Movie routes
		r.Route("/movies", func(r chi.Router) {
			r.Get("/{id}", movieHandler.GetMovie)
			r.Get("/", movieHandler.ListMovies)
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequireAuth)
				r.Post("/", movieHandler.CreateMovie)
				r.Put("/{id}", movieHandler.UpdateMovie)
				r.Delete("/{id}", movieHandler.DeleteMovie)
			})
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.SuccessResponse(w, http.StatusOK, "Service is healthy", nil)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.ErrorResponse(w, http.StatusNotFound, "Resource not found")
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		response.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	})

	return r
}
