package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

// Repository provides generic CRUD operations for GORM models
type Repository[T any] struct {
	db *gorm.DB
}

// NewRepository creates a new generic repository
func NewRepository[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{db: db}
}

// Create inserts a new record
func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return fmt.Errorf("failed to create: %w", result.Error)
	}
	return nil
}

// CreateInBatches inserts multiple records in batches
func (r *Repository[T]) CreateInBatches(ctx context.Context, entities []T, batchSize int) error {
	result := r.db.WithContext(ctx).CreateInBatches(entities, batchSize)
	if result.Error != nil {
		return fmt.Errorf("failed to create in batches: %w", result.Error)
	}
	return nil
}

// FindByID retrieves a record by ID
func (r *Repository[T]) FindByID(ctx context.Context, id uuid.UUID) (*T, error) {
	var entity T
	result := r.db.WithContext(ctx).First(&entity, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, fmt.Errorf("failed to find by id: %w", result.Error)
	}
	return &entity, nil
}

// FindOne retrieves a single record matching the conditions
func (r *Repository[T]) FindOne(ctx context.Context, conditions map[string]interface{}) (*T, error) {
	var entity T
	result := r.db.WithContext(ctx).Where(conditions).First(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, fmt.Errorf("failed to find one: %w", result.Error)
	}
	return &entity, nil
}

// FindAll retrieves all records matching the conditions
func (r *Repository[T]) FindAll(ctx context.Context, conditions map[string]interface{}) ([]T, error) {
	var entities []T
	result := r.db.WithContext(ctx).Where(conditions).Find(&entities)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find all: %w", result.Error)
	}
	return entities, nil
}

// FindWithPagination retrieves records with pagination
func (r *Repository[T]) FindWithPagination(ctx context.Context, page, pageSize int, conditions map[string]interface{}) ([]T, int64, error) {
	var entities []T
	var total int64

	db := r.db.WithContext(ctx).Model(new(T))
	if conditions != nil {
		db = db.Where(conditions)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count: %w", err)
	}

	offset := (page - 1) * pageSize
	result := db.Limit(pageSize).Offset(offset).Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to find with pagination: %w", result.Error)
	}

	return entities, total, nil
}

// Update updates a record
func (r *Repository[T]) Update(ctx context.Context, entity *T) error {
	result := r.db.WithContext(ctx).Save(entity)
	if result.Error != nil {
		return fmt.Errorf("failed to update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// UpdateFields updates specific fields
func (r *Repository[T]) UpdateFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(fields)
	if result.Error != nil {
		return fmt.Errorf("failed to update fields: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// Delete deletes a record by ID
func (r *Repository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(new(T), "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// HardDelete is deprecated - use Delete instead as models don't support soft delete
// Kept for backward compatibility
func (r *Repository[T]) HardDelete(ctx context.Context, id uuid.UUID) error {
	return r.Delete(ctx, id)
}

// Exists checks if a record exists with given conditions
func (r *Repository[T]) Exists(ctx context.Context, conditions map[string]interface{}) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(new(T)).Where(conditions).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to check existence: %w", result.Error)
	}
	return count > 0, nil
}

// Count returns the number of records matching the conditions
func (r *Repository[T]) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(new(T)).Where(conditions).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count: %w", result.Error)
	}
	return count, nil
}

// Transaction executes a function within a transaction
func (r *Repository[T]) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// WithPreload adds preload to the query
func (r *Repository[T]) WithPreload(preloads ...string) *Repository[T] {
	db := r.db
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	return &Repository[T]{db: db}
}

// WithLock adds locking to the query
func (r *Repository[T]) WithLock() *Repository[T] {
	return &Repository[T]{db: r.db.Clauses(clause.Locking{Strength: "UPDATE"})}
}

// DB returns the underlying GORM DB instance for custom queries
func (r *Repository[T]) DB() *gorm.DB {
	return r.db
}

// Query builder methods

type QueryBuilder[T any] struct {
	db *gorm.DB
}

func (r *Repository[T]) Query() *QueryBuilder[T] {
	return &QueryBuilder[T]{db: r.db}
}

func (q *QueryBuilder[T]) Where(query interface{}, args ...interface{}) *QueryBuilder[T] {
	q.db = q.db.Where(query, args...)
	return q
}

func (q *QueryBuilder[T]) Or(query interface{}, args ...interface{}) *QueryBuilder[T] {
	q.db = q.db.Or(query, args...)
	return q
}

func (q *QueryBuilder[T]) Order(order string) *QueryBuilder[T] {
	q.db = q.db.Order(order)
	return q
}

func (q *QueryBuilder[T]) Limit(limit int) *QueryBuilder[T] {
	q.db = q.db.Limit(limit)
	return q
}

func (q *QueryBuilder[T]) Offset(offset int) *QueryBuilder[T] {
	q.db = q.db.Offset(offset)
	return q
}

func (q *QueryBuilder[T]) Preload(preload string, args ...interface{}) *QueryBuilder[T] {
	q.db = q.db.Preload(preload, args...)
	return q
}

func (q *QueryBuilder[T]) Join(query string, args ...interface{}) *QueryBuilder[T] {
	q.db = q.db.Joins(query, args...)
	return q
}

func (q *QueryBuilder[T]) Group(group string) *QueryBuilder[T] {
	q.db = q.db.Group(group)
	return q
}

func (q *QueryBuilder[T]) Having(query interface{}, args ...interface{}) *QueryBuilder[T] {
	q.db = q.db.Having(query, args...)
	return q
}

func (q *QueryBuilder[T]) Find(ctx context.Context) ([]T, error) {
	var entities []T
	result := q.db.WithContext(ctx).Find(&entities)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find: %w", result.Error)
	}
	return entities, nil
}

func (q *QueryBuilder[T]) First(ctx context.Context) (*T, error) {
	var entity T
	result := q.db.WithContext(ctx).First(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, fmt.Errorf("failed to find first: %w", result.Error)
	}
	return &entity, nil
}

func (q *QueryBuilder[T]) Count(ctx context.Context) (int64, error) {
	var count int64
	result := q.db.WithContext(ctx).Model(new(T)).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count: %w", result.Error)
	}
	return count, nil
}

func (q *QueryBuilder[T]) Delete(ctx context.Context) error {
	result := q.db.WithContext(ctx).Delete(new(T))
	if result.Error != nil {
		return fmt.Errorf("failed to delete: %w", result.Error)
	}
	return nil
}

func (q *QueryBuilder[T]) Update(ctx context.Context, fields map[string]interface{}) error {
	result := q.db.WithContext(ctx).Updates(fields)
	if result.Error != nil {
		return fmt.Errorf("failed to update: %w", result.Error)
	}
	return nil
}
