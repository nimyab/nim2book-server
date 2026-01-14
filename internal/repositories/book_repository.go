package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/models"
	"gorm.io/gorm"
)

var (
	ErrBookNotFound      = errors.New("book not found")
	ErrBookAlreadyExists = errors.New("book already exists")
)

const (
	DefaultPageSize = 100
)

type BookRepository struct {
	*Repository[models.Book]
	db *gorm.DB
}

func NewBookRepository(db *gorm.DB) *BookRepository {
	return &BookRepository{
		Repository: NewRepository[models.Book](db),
		db:         db,
	}
}

// GetBookByAuthorAndTitle retrieves a book by author and title
func (r *BookRepository) GetBookByAuthorAndTitle(ctx context.Context, author, title string) (*models.Book, error) {
	book, err := r.Query().
		Where("author = ? AND title = ?", author, title).
		First(ctx)

	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return nil, ErrBookNotFound
		}
		return nil, err
	}

	return book, nil
}

// CreateBook creates a new book
func (r *BookRepository) CreateBook(ctx context.Context, book *models.Book) (*models.Book, error) {
	var result *models.Book
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := NewRepository[models.Book](tx)

		// Check if book already exists
		existingBook, err := repo.Query().
			Where("author = ? AND title = ?", book.Author, book.Title).
			First(ctx)

		if err == nil {
			result = existingBook
			return ErrBookAlreadyExists
		}

		if !errors.Is(err, ErrRecordNotFound) {
			return err
		}

		// Create new book
		if err := repo.Create(ctx, book); err != nil {
			return err
		}

		result = book
		return nil
	})

	if err != nil && !errors.Is(err, ErrBookAlreadyExists) {
		return nil, err
	}

	return result, err
}

// GetBookById retrieves a book by ID
func (r *BookRepository) GetBookById(ctx context.Context, id uuid.UUID) (*models.Book, error) {
	book, err := r.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return nil, ErrBookNotFound
		}
		return nil, err
	}

	return book, nil
}

// GetBooks retrieves books with filtering and pagination
func (r *BookRepository) GetBooks(ctx context.Context, author, title string, page int) ([]*models.Book, error) {
	qb := r.Query()

	// Apply filters if provided
	if author != "" {
		qb = qb.Where("author ILIKE ?", "%"+author+"%")
	}
	if title != "" {
		qb = qb.Where("title ILIKE ?", "%"+title+"%")
	}

	// Apply pagination
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * DefaultPageSize
	qb = qb.Limit(DefaultPageSize).Offset(offset)

	books, err := qb.Find(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*models.Book, len(books))
	for i := range books {
		result[i] = &books[i]
	}

	return result, nil
}

// UpdateBook updates an existing book
func (r *BookRepository) UpdateBook(ctx context.Context, id uuid.UUID, title, author string, cover *string) error {
	updates := map[string]interface{}{
		"title":  title,
		"author": author,
	}

	if cover != nil {
		updates["cover"] = cover
	}

	err := r.UpdateFields(ctx, id, updates)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return ErrBookNotFound
		}
		return err
	}

	return nil
}

// DeleteBook deletes a book by ID
func (r *BookRepository) DeleteBook(ctx context.Context, id uuid.UUID) error {
	err := r.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return ErrBookNotFound
		}
		return err
	}

	return nil
}

// CountBooks returns the total number of books matching the filters
func (r *BookRepository) CountBooks(ctx context.Context, author, title string) (int64, error) {
	qb := r.Query()

	if author != "" {
		qb = qb.Where("author ILIKE ?", "%"+author+"%")
	}
	if title != "" {
		qb = qb.Where("title ILIKE ?", "%"+title+"%")
	}

	count, err := qb.Count(ctx)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// SearchBooks searches books by pattern
func (r *BookRepository) SearchBooks(ctx context.Context, pattern string, limit int) ([]*models.Book, error) {
	if limit < 1 {
		limit = 20
	}

	books, err := r.Query().
		Where("title ILIKE ? OR author ILIKE ?", "%"+pattern+"%", "%"+pattern+"%").
		Limit(limit).
		Find(ctx)

	if err != nil {
		return nil, err
	}

	result := make([]*models.Book, len(books))
	for i := range books {
		result[i] = &books[i]
	}

	return result, nil
}

// GetBooksByAuthor retrieves all books by a specific author
func (r *BookRepository) GetBooksByAuthor(ctx context.Context, author string) ([]*models.Book, error) {
	books, err := r.Query().
		Where("author ILIKE ?", "%"+author+"%").
		Find(ctx)

	if err != nil {
		return nil, err
	}

	result := make([]*models.Book, len(books))
	for i := range books {
		result[i] = &books[i]
	}

	return result, nil
}

// GetRecentBooks retrieves the most recently added books
func (r *BookRepository) GetRecentBooks(ctx context.Context, limit int) ([]*models.Book, error) {
	if limit < 1 {
		limit = 10
	}

	books, err := r.Query().
		Limit(limit).
		Find(ctx)

	if err != nil {
		return nil, err
	}

	result := make([]*models.Book, len(books))
	for i := range books {
		result[i] = &books[i]
	}

	return result, nil
}
