package domain

import "github.com/google/uuid"

type ID = uuid.UUID

type JsonB = map[string]any

type ProcessStatus string

const (
	ProcessStatusInProgress ProcessStatus = "in_progress"
	ProcessStatusCompleted  ProcessStatus = "completed"
	ProcessStatusFailed     ProcessStatus = "failed"
)
