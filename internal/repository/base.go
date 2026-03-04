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

// DoInTx выполняет функцию внутри транзакции
// Если функция возвращает ошибку, транзакция откатывается
// Если функция завершается успешно, транзакция коммитится
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

// GetClientOrTx возвращает tx.Client(), если tx не nil, иначе client
func GetClientOrTx(client *entclient.Client, tx *entclient.Tx) *entclient.Client {
	if tx != nil {
		return tx.Client()
	}
	return client
}
