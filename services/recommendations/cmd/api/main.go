// Recommendations service entrypoint.
//
// This service exposes the recommendations produced by the offline ML
// pipeline. It does NOT run ML inference here; it only stores and serves
// results, and emits regeneration requests to the pipeline.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tinta/recommendations/internal/platform/config"
	"github.com/tinta/recommendations/internal/platform/database"
	"github.com/tinta/recommendations/internal/platform/server"

	recoApp "github.com/tinta/recommendations/internal/recommendation/application"
	recoHTTP "github.com/tinta/recommendations/internal/recommendation/infrastructure/http"
	recoPipeline "github.com/tinta/recommendations/internal/recommendation/infrastructure/pipeline"
	recoPG "github.com/tinta/recommendations/internal/recommendation/infrastructure/postgres"

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
	log := logger.New("recommendations", cfg.LogLevel)
	log.Info().Int("port", cfg.HTTPPort).Msg("starting recommendations service")

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

	// Recommendation module
	recoRepo := recoPG.NewRecommendationRepository(pool)
	pipeline := recoPipeline.NewLoggerPipeline(log)

	listUC := recoApp.NewListRecommendationsUseCase(recoRepo)
	feedbackUC := recoApp.NewSubmitFeedbackUseCase(recoRepo)
	dismissUC := recoApp.NewDismissRecommendationUseCase(recoRepo)
	regenerateUC := recoApp.NewRegenerateRecommendationsUseCase(pipeline)

	recoHandler := recoHTTP.NewHandler(listUC, feedbackUC, dismissUC, regenerateUC)

	// HTTP server
	app := server.New("recommendations")
	v1 := app.Group("/api/v1")
	authMW := middleware.RequireAuth(verifier)
	recoHandler.Register(v1, authMW)

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
	log.Info().Msg("recommendations service stopped")
	return nil
}
