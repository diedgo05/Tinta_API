// Package domain contains the ReadingProgress aggregate.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusReading   Status = "reading"
	StatusPaused    Status = "paused"
	StatusFinished  Status = "finished"
	StatusAbandoned Status = "abandoned"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusReading, StatusPaused, StatusFinished, StatusAbandoned:
		return true
	}
	return false
}

type Progress struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	BookID       uuid.UUID
	CurrentPage  int
	TotalTimeMin int
	Status       Status
	StartedAt    time.Time
	LastReadAt   time.Time
	FinishedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

var (
	ErrProgressNotFound = errors.New("reading progress not found")
	ErrInvalidStatus    = errors.New("invalid status")
	ErrInvalidPage      = errors.New("invalid page")
)
