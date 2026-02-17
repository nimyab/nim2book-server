package repository

import (
	"context"

	entclient "github.com/nimyab/nim2book-back/ent"
	"github.com/nimyab/nim2book-back/internal/domain"
)

// QueryOptions для фильтрации и пагинации
type QueryOptions struct {
	Limit  int
	Offset int
	IDs    []domain.ID
}

// BaseRepository содержит общую логику для всех репозиториев
type BaseRepository struct {
	client *entclient.Client
}

// NewBaseRepository создает новый базовый репозиторий
func NewBaseRepository(client *entclient.Client) *BaseRepository {
	return &BaseRepository{client: client}
}

// Client возвращает ent.Client
func (r *BaseRepository) Client() *entclient.Client {
	return r.client
}

// HandleError преобразует ошибки ent в domain ошибки
func HandleError(err error) error {
	if err == nil {
		return nil
	}

	// Проверка на NotFound ошибку
	if entclient.IsNotFound(err) {
		return ErrNotFound
	}

	// Проверка на ошибку нарушения уникальности
	if entclient.IsConstraintError(err) {
		return ErrDuplicateKey
	}

	// Проверка на ошибку валидации
	if entclient.IsValidationError(err) {
		return ErrInvalidInput
	}

	// Любая другая ошибка
	return ErrInternal
}

// DoInTx выполняет функцию внутри транзакции
// Если функция возвращает ошибку, транзакция откатывается
// Если функция завершается успешно, транзакция коммитится
// ВАЖНО: Используйте эту функцию ТОЛЬКО для операций записи (Create, Update, Delete)
// или когда нужно несколько операций атомарно.
// Для read-only операций используйте обычный client напрямую.
func DoInTx[T any](ctx context.Context, client *entclient.Client, fn func(tx *entclient.Tx) (T, error)) (T, error) {
	var zero T

	tx, err := client.Tx(ctx)
	if err != nil {
		return zero, HandleError(err)
	}

	// Defer для обработки паники и отката транзакции
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// Выполняем функцию
	result, err := fn(tx)
	if err != nil {
		// Откатываем транзакцию при ошибке
		if rerr := tx.Rollback(); rerr != nil {
			return zero, HandleError(rerr)
		}
		return zero, err
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return zero, HandleError(err)
	}

	return result, nil
}

// TxClient возвращает Client из транзакции или обычный Client
// Используется для универсальных функций, которые могут работать
// как внутри транзакции, так и вне её
type TxClient interface {
	User() interface{}
	Book() interface{}
	PersonalBook() interface{}
	Dictionary() interface{}
	DictionaryExample() interface{}
	Author() interface{}
	Genre() interface{}
	GoogleAccount() interface{}
	BasicAccount() interface{}
	FCMToken() interface{}
}

// GetClientOrTx возвращает tx.Client(), если tx не nil, иначе client
func GetClientOrTx(client *entclient.Client, tx *entclient.Tx) *entclient.Client {
	if tx != nil {
		return tx.Client()
	}
	return client
}
