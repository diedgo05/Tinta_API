// Package logger provides a shared zerolog logger initialization.
package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// New returns a zerolog logger configured with the given level and service name.
func New(serviceName, level string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	lvl, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	return zerolog.New(os.Stdout).
		Level(lvl).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()
}
