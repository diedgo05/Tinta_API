// Reading service entrypoint.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	annApp "github.com/tinta/reading/internal/annotation/application"
	annHTTP "github.com/tinta/reading/internal/annotation/infrastructure/http"
	annPG "github.com/tinta/reading/internal/annotation/infrastructure/postgres"

	progApp "github.com/tinta/reading/internal/progress/application"
	progHTTP "github.com/tinta/reading/internal/progress/infrastructure/http"
	progPG "github.com/tinta/reading/internal/progress/infrastructure/postgres"

	"github.com/tinta/reading/internal/platform/config"
	"github.com/tinta/reading/internal/platform/database"
	"github.com/tinta/reading/internal/platform/server"

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
		return err
	}
	log := logger.New("reading", cfg.LogLevel)
	log.Info().Int("port", cfg.HTTPPort).Msg("starting reading service")

	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	verifier, err := jwtauth.NewVerifier(cfg.JWTPublicKeyPath)
	if err != nil {
		return err
	}

	// Progress module
	progRepo := progPG.NewProgressRepository(pool)
	progHandler := progHTTP.NewHandler(
		progApp.NewStartReadingUseCase(progRepo),
		progApp.NewGetProgressUseCase(progRepo),
		progApp.NewListProgressUseCase(progRepo),
		progApp.NewUpdateProgressUseCase(progRepo),
		progApp.NewDeleteProgressUseCase(progRepo),
	)

	// Annotation module
	annRepo := annPG.NewAnnotationRepository(pool)
	annHandler := annHTTP.NewHandler(
		annApp.NewCreateAnnotationUseCase(annRepo),
		annApp.NewListAnnotationsUseCase(annRepo),
		annApp.NewUpdateAnnotationUseCase(annRepo),
		annApp.NewDeleteAnnotationUseCase(annRepo),
	)

	app := server.New("reading")
	v1 := app.Group("/api/v1")
	authMW := middleware.RequireAuth(verifier)
	progHandler.Register(v1, authMW)
	annHandler.Register(v1, authMW)

	go func() {
		if err := app.Listen(fmt.Sprintf(":%d", cfg.HTTPPort)); err != nil {
			log.Error().Err(err).Msg("server stopped")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = app.ShutdownWithContext(shutdownCtx)
	log.Info().Msg("reading stopped")
	return nil
}
