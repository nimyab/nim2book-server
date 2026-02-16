package repository

import (
	"errors"
)

// Repository errors
var (
	ErrNotFound            = errors.New("entity not found")
	ErrDuplicateKey        = errors.New("duplicate key violation")
	ErrForeignKeyViolation = errors.New("foreign key violation")
	ErrInvalidInput        = errors.New("invalid input")
	ErrInternal            = errors.New("internal repository error")
)
