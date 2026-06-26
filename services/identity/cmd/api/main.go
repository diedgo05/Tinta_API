// Identity service entrypoint.
//
// This service is responsible for:
//   - User registration and profile management
//   - Authentication via email + password
//   - Issuance and rotation of JWT tokens
//   - Email verification (Turno 2)
//   - Password recovery (Turno 2)
//
// It is the ONLY service that signs JWTs; the rest only verify them with
// the public key.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	authApp "github.com/tinta/identity/internal/auth/application"
	authHTTP "github.com/tinta/identity/internal/auth/infrastructure/http"
	authJWT "github.com/tinta/identity/internal/auth/infrastructure/jwt"
	authPG "github.com/tinta/identity/internal/auth/infrastructure/postgres"

	"github.com/tinta/identity/internal/platform/config"
	"github.com/tinta/identity/internal/platform/database"
	"github.com/tinta/identity/internal/platform/server"

	userApp "github.com/tinta/identity/internal/user/application"
	userArgon2 "github.com/tinta/identity/internal/user/infrastructure/argon2"
	userHTTP "github.com/tinta/identity/internal/user/infrastructure/http"
	userPG "github.com/tinta/identity/internal/user/infrastructure/postgres"

	// Turno 2 — email verification
	verifApp "github.com/tinta/identity/internal/verification/application"
	verifHTTP "github.com/tinta/identity/internal/verification/infrastructure/http"
	verifPG "github.com/tinta/identity/internal/verification/infrastructure/postgres"

	// Turno 2 — password reset
	resetApp "github.com/tinta/identity/internal/passwordreset/application"
	resetHTTP "github.com/tinta/identity/internal/passwordreset/infrastructure/http"
	resetPG "github.com/tinta/identity/internal/passwordreset/infrastructure/postgres"

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
	// ---------- Configuration ----------
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	log := logger.New("identity", cfg.LogLevel)
	log.Info().Int("port", cfg.HTTPPort).Msg("starting identity service")

	// ---------- Database ----------
	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()
	log.Info().Msg("postgres connected")

	// ---------- JWT ----------
	signer, err := jwtauth.NewSigner(cfg.JWTPrivateKeyPath, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	if err != nil {
		return fmt.Errorf("load jwt signer: %w", err)
	}
	verifier, err := jwtauth.NewVerifier(cfg.JWTPublicKeyPath)
	if err != nil {
		return fmt.Errorf("load jwt verifier: %w", err)
	}

	// ---------- User module ----------
	userRepo := userPG.NewUserRepository(pool)
	hasher := userArgon2.New()

	createUserUC := userApp.NewCreateUserUseCase(userRepo, hasher)
	getUserUC := userApp.NewGetUserUseCase(userRepo)
	updateUserUC := userApp.NewUpdateUserUseCase(userRepo)
	deleteUserUC := userApp.NewDeleteUserUseCase(userRepo)
	userHandler := userHTTP.NewHandler(createUserUC, getUserUC, updateUserUC, deleteUserUC)

	// ---------- Auth module ----------
	refreshRepo := authPG.NewRefreshTokenRepository(pool)
	signerAdapter := authJWT.NewSignerAdapter(signer)

	loginUC := authApp.NewLoginUseCase(userRepo, hasher, refreshRepo, signerAdapter)
	refreshUC := authApp.NewRefreshUseCase(userRepo, refreshRepo, signerAdapter)
	logoutUC := authApp.NewLogoutUseCase(refreshRepo)
	authHandler := authHTTP.NewHandler(loginUC, refreshUC, logoutUC)

	// ---------- Turno 2 · Verification module ----------
	verifRepo := verifPG.NewVerificationRepository(pool)
	requestCodeUC := verifApp.NewRequestCodeUseCase(verifRepo, log)
	verifyCodeUC := verifApp.NewVerifyCodeUseCase(verifRepo)
	verifHandler := verifHTTP.NewHandler(requestCodeUC, verifyCodeUC)

	// ---------- Turno 2 · Password reset module ----------
	resetRepo := resetPG.NewPasswordResetRepository(pool)
	// Adapter: convert userArgon2.Hasher to the reset module's hasher interface.
	resetHasher := &argon2HasherAdapter{h: hasher}
	requestResetUC := resetApp.NewRequestResetUseCase(resetRepo, log)
	confirmResetUC := resetApp.NewConfirmResetUseCase(resetRepo, resetHasher)
	resetHandler := resetHTTP.NewHandler(requestResetUC, confirmResetUC)

	// ---------- HTTP server ----------
	app := server.New("identity")

	v1 := app.Group("/api/v1")
	authMW := middleware.RequireAuth(verifier)
	userHandler.Register(v1, authMW)
	authHandler.Register(v1)
	verifHandler.Register(v1, authMW)
	resetHandler.Register(v1) // public endpoints (no auth)

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
	log.Info().Msg("identity service stopped")
	return nil
}

// argon2HasherAdapter bridges the user-module's Argon2 hasher to the
// password-reset module's expected interface (which only needs Hash()).
type argon2HasherAdapter struct {
	h *userArgon2.Hasher
}

func (a *argon2HasherAdapter) Hash(password string) (string, error) {
	return a.h.Hash(password)
}
