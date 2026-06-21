// Package argon2 implements the PasswordHasher port using Argon2id.
// Argon2id is the password hashing algorithm recommended by OWASP.
package argon2

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Parameters tuned for interactive logins (~50-100ms per hash on modern CPU).
// Adjust if your target hardware differs significantly.
const (
	saltLen     = 16
	keyLen      = 32
	timeCost    = 3
	memoryCost  = 64 * 1024 // 64 MB
	parallelism = 2
	version     = argon2.Version
)

// Hasher implements ports.PasswordHasher.
type Hasher struct{}

// New returns a new Argon2id hasher.
func New() *Hasher {
	return &Hasher{}
}

// Hash computes the Argon2id hash of a password.
// Returned format: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
func (h *Hasher) Hash(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, parallelism, keyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		version, memoryCost, timeCost, parallelism, b64Salt, b64Hash,
	)
	return encoded, nil
}

// Verify checks whether the given password matches the stored hash.
func (h *Hasher) Verify(password, encoded string) error {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return errors.New("invalid hash format")
	}

	var v, m, t, p uint32
	if _, err := fmt.Sscanf(parts[2], "v=%d", &v); err != nil {
		return fmt.Errorf("parse version: %w", err)
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return fmt.Errorf("parse params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("decode salt: %w", err)
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return fmt.Errorf("decode hash: %w", err)
	}

	actual := argon2.IDKey([]byte(password), salt, t, m, uint8(p), uint32(len(expected)))
	if subtle.ConstantTimeCompare(expected, actual) != 1 {
		return errors.New("password mismatch")
	}
	return nil
}
