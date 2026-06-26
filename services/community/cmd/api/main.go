// Community service entrypoint.
//
// This service handles reading clubs, members and discussions.
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

	// Turno 2 — member module
	memberApp "github.com/tinta/community/internal/member/application"
	memberHTTP "github.com/tinta/community/internal/member/infrastructure/http"
	memberPG "github.com/tinta/community/internal/member/infrastructure/postgres"

	// Turno 2 — discussion module
	discApp "github.com/tinta/community/internal/discussion/application"
	discHTTP "github.com/tinta/community/internal/discussion/infrastructure/http"
	discPG "github.com/tinta/community/internal/discussion/infrastructure/postgres"

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

	// ---------- Club module ----------
	clubRepo := clubPG.NewClubRepository(pool)
	createUC := clubApp.NewCreateClubUseCase(clubRepo)
	listUC := clubApp.NewListClubsUseCase(clubRepo)
	getUC := clubApp.NewGetClubUseCase(clubRepo)
	updateUC := clubApp.NewUpdateClubUseCase(clubRepo)
	deleteUC := clubApp.NewDeleteClubUseCase(clubRepo)
	clubHandler := clubHTTP.NewHandler(createUC, listUC, getUC, updateUC, deleteUC)

	// ---------- Turno 2 · Member module ----------
	memberRepo := memberPG.NewMemberRepository(pool)
	joinUC := memberApp.NewJoinClubUseCase(memberRepo)
	leaveUC := memberApp.NewLeaveClubUseCase(memberRepo)
	listClubMembersUC := memberApp.NewListClubMembersUseCase(memberRepo)
	listMyClubsUC := memberApp.NewListMyClubsUseCase(memberRepo)
	checkMembershipUC := memberApp.NewCheckMembershipUseCase(memberRepo)
	memberHandler := memberHTTP.NewHandler(joinUC, leaveUC, listClubMembersUC, listMyClubsUC, checkMembershipUC)

	// ---------- Turno 2 · Discussion module ----------
	discRepo := discPG.NewDiscussionRepository(pool)
	postDiscUC := discApp.NewPostDiscussionUseCase(discRepo, memberRepo)
	listDiscUC := discApp.NewListDiscussionsUseCase(discRepo, memberRepo)
	updateDiscUC := discApp.NewUpdateDiscussionUseCase(discRepo)
	deleteDiscUC := discApp.NewDeleteDiscussionUseCase(discRepo)
	discHandler := discHTTP.NewHandler(postDiscUC, listDiscUC, updateDiscUC, deleteDiscUC)

	// ---------- HTTP server ----------
	app := server.New("community")
	v1 := app.Group("/api/v1")
	authMW := middleware.RequireAuth(verifier)
	clubHandler.Register(v1, authMW)
	memberHandler.Register(v1, authMW)
	discHandler.Register(v1, authMW)

	// ---------- Graceful shutdown ----------
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
