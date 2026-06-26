// Notifications service entrypoint.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	notifApp "github.com/tinta/notifications/internal/notification/application"
	notifHTTP "github.com/tinta/notifications/internal/notification/infrastructure/http"
	notifPG "github.com/tinta/notifications/internal/notification/infrastructure/postgres"

	"github.com/tinta/notifications/internal/platform/config"
	"github.com/tinta/notifications/internal/platform/database"
	"github.com/tinta/notifications/internal/platform/server"

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
	log := logger.New("notifications", cfg.LogLevel)
	log.Info().Int("port", cfg.HTTPPort).Msg("starting notifications service")

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

	repo := notifPG.NewNotificationRepository(pool)
	handler := notifHTTP.NewHandler(
		notifApp.NewCreateNotificationUseCase(repo),
		notifApp.NewListNotificationsUseCase(repo),
		notifApp.NewMarkAsReadUseCase(repo),
		notifApp.NewMarkAllAsReadUseCase(repo),
		notifApp.NewDeleteNotificationUseCase(repo),
	)

	app := server.New("notifications")
	v1 := app.Group("/api/v1")
	authMW := middleware.RequireAuth(verifier)
	handler.Register(v1, authMW)

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
	log.Info().Msg("notifications stopped")
	return nil
}
