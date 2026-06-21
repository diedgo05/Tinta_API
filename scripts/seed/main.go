// Tinta seed script
//
// Inserts the four admin users that bypass email verification, used by the
// team to develop and demo the platform without going through the regular
// registration + verification flow.
//
// Usage:
//
//	go run scripts/seed/main.go
//
// Reads DATABASE_URL from the environment if set; otherwise uses the local
// development default. Idempotent: running it twice does not create duplicates.
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/argon2"
)

// adminSeed describes one of the admin accounts to be inserted.
type adminSeed struct {
	Email string
	Name  string
	Role  string
}

var admins = []adminSeed{
	{Email: "adrian@tinta.app", Name: "Adrian Sanchez",   Role: "admin"},
	{Email: "diego@tinta.app",  Name: "Diego Jimenez",    Role: "admin"},
	{Email: "gael@tinta.app",   Name: "Gael Hueytlelt",   Role: "admin"},
	{Email: "system@tinta.app", Name: "System Account",   Role: "system"},
}

const defaultPassword = "admin123"

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://tinta:tinta_dev_pass@localhost:5432/tinta?sslmode=disable&search_path=identity"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping postgres: %v", err)
	}

	// Hash the shared default password once; reuse for every admin.
	hash, err := hashPassword(defaultPassword)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	const q = `
		INSERT INTO users (email, password_hash, name, role, email_verified, language)
		VALUES ($1, $2, $3, $4, TRUE, 'es')
		ON CONFLICT (email) DO NOTHING
		RETURNING id`

	inserted := 0
	skipped := 0
	for _, a := range admins {
		var id string
		err := pool.QueryRow(ctx, q, a.Email, hash, a.Name, a.Role).Scan(&id)
		switch {
		case err == nil:
			fmt.Printf("✅ Inserted %s (id=%s, role=%s)\n", a.Email, id, a.Role)
			inserted++
		default:
			// ON CONFLICT DO NOTHING returns no rows → ErrNoRows. Treat as skip.
			fmt.Printf("↷  Skipped %s (already exists)\n", a.Email)
			skipped++
		}
	}

	fmt.Printf("\nDone. Inserted: %d, Skipped: %d\n", inserted, skipped)
	fmt.Printf("Default password for all admins: %s\n", defaultPassword)
	fmt.Println("Remember to change these passwords in production.")
}

// hashPassword replicates the Argon2id format used by the Identity service.
// Keep these parameters in sync with services/identity/internal/user/infrastructure/argon2.
const (
	saltLen     = 16
	keyLen      = 32
	timeCost    = 3
	memoryCost  = 64 * 1024
	parallelism = 2
)

func hashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, parallelism, keyLen)
	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memoryCost, timeCost, parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return encoded, nil
}
