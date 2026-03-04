package repository

import (
	"errors"
	"fmt"

	"github.com/nimyab/nim2book-back/ent"
)

// Repository errors
var (
	ErrNotFound            = errors.New("entity not found")
	ErrDuplicateKey        = errors.New("duplicate key violation")
	ErrForeignKeyViolation = errors.New("foreign key violation")
	ErrInvalidInput        = errors.New("invalid input")
	ErrInternal            = errors.New("internal repository error")
)

// HandleError преобразует ошибки ent в domain ошибки
func HandleError(err error) error {
	if err == nil {
		return nil
	}

	// Проверка на NotFound ошибку
	if ent.IsNotFound(err) {
		return ErrNotFound
	}

	// Проверка на ошибку нарушения уникальности
	if ent.IsConstraintError(err) {
		return ErrDuplicateKey
	}

	// Проверка на ошибку валидации
	if ent.IsValidationError(err) {
		return ErrInvalidInput
	}

	// Любая другая ошибка
	return fmt.Errorf("%w: %v", ErrInternal, err)
}
