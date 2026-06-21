// Package pipeline implements ports.MLPipeline.
//
// V1: a logger stub that records the regeneration request. In the next
// iteration we will replace this with an Asynq task published to Redis,
// consumed by the Python pipeline that produces the actual recommendations.
package pipeline

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// LoggerPipeline is a placeholder implementation that logs the regeneration request.
// It satisfies ports.MLPipeline so the rest of the system can compile and run.
type LoggerPipeline struct {
	log zerolog.Logger
}

// NewLoggerPipeline returns a logger-based pipeline stub.
func NewLoggerPipeline(log zerolog.Logger) *LoggerPipeline {
	return &LoggerPipeline{log: log}
}

// RegenerateForUser logs the request. The real implementation will publish
// an Asynq task: tasks.Enqueue(ctx, "recommendations:regenerate", userID).
func (p *LoggerPipeline) RegenerateForUser(ctx context.Context, userID uuid.UUID) error {
	p.log.Info().
		Str("user_id", userID.String()).
		Msg("ML pipeline regeneration requested (stub: no actual job dispatched)")
	return nil
}
