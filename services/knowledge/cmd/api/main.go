// Knowledge service entrypoint.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	fragApp "github.com/tinta/knowledge/internal/fragment/application"
	fragHTTP "github.com/tinta/knowledge/internal/fragment/infrastructure/http"
	fragPG "github.com/tinta/knowledge/internal/fragment/infrastructure/postgres"

	topicApp "github.com/tinta/knowledge/internal/topic/application"
	topicHTTP "github.com/tinta/knowledge/internal/topic/infrastructure/http"
	topicPG "github.com/tinta/knowledge/internal/topic/infrastructure/postgres"

	"github.com/tinta/knowledge/internal/platform/config"
	"github.com/tinta/knowledge/internal/platform/database"
	"github.com/tinta/knowledge/internal/platform/server"

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
	log := logger.New("knowledge", cfg.LogLevel)
	log.Info().Int("port", cfg.HTTPPort).Msg("starting knowledge service")

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

	// Topic module
	topicRepo := topicPG.NewTopicRepository(pool)
	userTopicRepo := topicPG.NewUserTopicRepository(pool)
	topicHandler := topicHTTP.NewHandler(
		topicApp.NewListTopicsUseCase(topicRepo),
		topicApp.NewGetTopicUseCase(topicRepo),
		topicApp.NewSelectTopicsUseCase(topicRepo, userTopicRepo),
		topicApp.NewListMyTopicsUseCase(userTopicRepo),
		topicApp.NewMarkDownloadedUseCase(userTopicRepo),
		topicApp.NewRemoveTopicUseCase(userTopicRepo),
	)

	// Fragment module
	fragRepo := fragPG.NewFragmentRepository(pool)
	docRepo := fragPG.NewDocumentRepository(pool)
	fragHandler := fragHTTP.NewHandler(
		fragApp.NewListFragmentsUseCase(fragRepo),
		fragApp.NewCreateFragmentUseCase(fragRepo),
		fragApp.NewListDocumentsUseCase(docRepo),
		fragApp.NewCreateDocumentUseCase(docRepo),
	)

	app := server.New("knowledge")
	v1 := app.Group("/api/v1")
	authMW := middleware.RequireAuth(verifier)
	topicHandler.Register(v1, authMW)
	fragHandler.Register(v1, authMW)

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
	log.Info().Msg("knowledge stopped")
	return nil
}
