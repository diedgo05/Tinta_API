// Catalog service entrypoint.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	bookApp "github.com/tinta/catalog/internal/book/application"
	bookHTTP "github.com/tinta/catalog/internal/book/infrastructure/http"
	bookPG "github.com/tinta/catalog/internal/book/infrastructure/postgres"

	genreApp "github.com/tinta/catalog/internal/genre/application"
	genreHTTP "github.com/tinta/catalog/internal/genre/infrastructure/http"
	genrePG "github.com/tinta/catalog/internal/genre/infrastructure/postgres"

	"github.com/tinta/catalog/internal/platform/config"
	"github.com/tinta/catalog/internal/platform/database"
	"github.com/tinta/catalog/internal/platform/server"

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
	log := logger.New("catalog", cfg.LogLevel)
	log.Info().Int("port", cfg.HTTPPort).Msg("starting catalog service")

	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	verifier, err := jwtauth.NewVerifier(cfg.JWTPublicKeyPath)
	if err != nil {
		return fmt.Errorf("load jwt verifier: %w", err)
	}

	// Book module
	bookRepo := bookPG.NewBookRepository(pool)
	bookHandler := bookHTTP.NewHandler(
		bookApp.NewCreateBookUseCase(bookRepo),
		bookApp.NewGetBookUseCase(bookRepo),
		bookApp.NewListBooksUseCase(bookRepo),
		bookApp.NewUpdateBookUseCase(bookRepo),
		bookApp.NewDeleteBookUseCase(bookRepo),
	)

	// Genre module
	genreRepo := genrePG.NewGenreRepository(pool)
	genreHandler := genreHTTP.NewHandler(
		genreApp.NewCreateGenreUseCase(genreRepo),
		genreApp.NewGetGenreUseCase(genreRepo),
		genreApp.NewListGenresUseCase(genreRepo),
		genreApp.NewUpdateGenreUseCase(genreRepo),
		genreApp.NewDeleteGenreUseCase(genreRepo),
	)

	app := server.New("catalog")
	v1 := app.Group("/api/v1")
	authMW := middleware.RequireAuth(verifier)
	bookHandler.Register(v1, authMW)
	genreHandler.Register(v1, authMW)

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
	log.Info().Msg("catalog stopped")
	return nil
}
