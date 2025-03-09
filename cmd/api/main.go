package main

import (
	"github.com/marchelhutagalung/go-service/internal/auth"
	"github.com/marchelhutagalung/go-service/internal/config"
	"github.com/marchelhutagalung/go-service/internal/database"
	"github.com/marchelhutagalung/go-service/internal/handlers"
	"github.com/marchelhutagalung/go-service/internal/logger"
	"github.com/marchelhutagalung/go-service/internal/middleware"
	"github.com/marchelhutagalung/go-service/internal/repository"
	"github.com/marchelhutagalung/go-service/internal/router"
	"github.com/marchelhutagalung/go-service/internal/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	logger.Init(cfg.Server.Env)
	logger.Info("Starting application", logger.Field("env", cfg.Server.Env))

	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", logger.Field("error", err))
	}
	defer db.Close()
	logger.Info("Connected to PostgreSQL")

	redisClient, err := database.NewRedisClient(&cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", logger.Field("error", err))
	}
	defer redisClient.Close()
	logger.Info("Connected to Redis")

	userRepo := repository.NewUserRepository(db)
	movieRepo := repository.NewMovieRepository(db)

	jwtService := auth.NewJWTService(&cfg.JWT, redisClient)
	authMiddleware := middleware.NewMiddleware(jwtService)
	authHandler := handlers.NewAuthHandler(userRepo, jwtService)
	userHandler := handlers.NewUserHandler(userRepo)
	movieHandler := handlers.NewMovieHandler(movieRepo)

	r := router.SetupRouter(authHandler, userHandler, movieHandler, authMiddleware)
	logger.Info("Router configured")

	srv := server.NewServer(&cfg.Server, r)

	// Capture shutdown signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-stopChan
		logger.Info("Received shutdown signal", logger.Field("signal", sig.String()))
		// The server will handle graceful shutdown
	}()

	logger.Info("Server starting", logger.Field("port", cfg.Server.Port))
	if err := srv.Start(); err != nil {
		logger.Fatal("Server failed", logger.Field("error", err))
	}
}
