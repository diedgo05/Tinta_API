// Community service entrypoint.
//
// This service handles reading clubs and their discussions.
// It does NOT sign JWTs; it only verifies them with the public key
// produced by the Identity service.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	clubApp "github.com/tinta/community/internal/club/application"
	clubHTTP "github.com/tinta/community/internal/club/infrastructure/http"
	clubPG "github.com/tinta/community/internal/club/infrastructure/postgres"

	"github.com/tinta/community/internal/platform/config"
	"github.com/tinta/community/internal/platform/database"
	"github.com/tinta/community/internal/platform/server"

	"github.com/tinta/shared/jwtauth"
	"github.com/tinta/shared/logger"
	"github.com/tinta/shared/middleware"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	log := logger.New("community", cfg.LogLevel)
	log.Info().Int("port", cfg.HTTPPort).Msg("starting community service")

	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()
	log.Info().Msg("postgres connected")

	verifier, err := jwtauth.NewVerifier(cfg.JWTPublicKeyPath)
	if err != nil {
		return fmt.Errorf("load jwt verifier: %w", err)
	}

	// Club module
	clubRepo := clubPG.NewClubRepository(pool)
	createUC := clubApp.NewCreateClubUseCase(clubRepo)
	listUC := clubApp.NewListClubsUseCase(clubRepo)
	getUC := clubApp.NewGetClubUseCase(clubRepo)
	updateUC := clubApp.NewUpdateClubUseCase(clubRepo)
	deleteUC := clubApp.NewDeleteClubUseCase(clubRepo)
	clubHandler := clubHTTP.NewHandler(createUC, listUC, getUC, updateUC, deleteUC)

	// HTTP server
	app := server.New("community")
	v1 := app.Group("/api/v1")
	authMW := middleware.RequireAuth(verifier)
	clubHandler.Register(v1, authMW)

	// Graceful shutdown
	go func() {
		if err := app.Listen(fmt.Sprintf(":%d", cfg.HTTPPort)); err != nil {
			log.Error().Err(err).Msg("server stopped")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}
	log.Info().Msg("community service stopped")
	return nil
}
