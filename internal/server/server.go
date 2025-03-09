package server

import (
	"context"
	"fmt"
	"github.com/marchelhutagalung/go-service/internal/config"
	"github.com/marchelhutagalung/go-service/internal/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	server *http.Server
	config *config.ServerConfig
	router chi.Router
}

func NewServer(config *config.ServerConfig, router chi.Router) *Server {
	return &Server{
		config: config,
		router: router,
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.config.Port)

	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info("Server listening", logger.Field("port", s.config.Port))
		serverErrors <- s.server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case <-shutdown:
		logger.Info("Server shutdown initiated")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		logger.Info("Gracefully shutting down server")
		if err := s.server.Shutdown(ctx); err != nil {
			logger.Error("Shutdown error", logger.Field("error", err))
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}

		logger.Info("Server shutdown complete")
	}

	return nil
}
