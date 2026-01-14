package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/models"
	"github.com/nimyab/nim2book-back/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookRepository_CreateBook(t *testing.T) {
	cleanupDatabase(t)

	book := &models.Book{
		Title:        "Test Book",
		Author:       "Test Author",
		ChapterPaths: models.StringArray{"chapter1.json", "chapter2.json"},
		Cover:        nil,
	}

	createdBook, err := bookRepo.CreateBook(context.Background(), book)
	require.NoError(t, err)
	assert.NotNil(t, createdBook)
	assert.NotEqual(t, uuid.Nil, createdBook.ID)
	assert.Equal(t, "Test Book", createdBook.Title)
	assert.Equal(t, "Test Author", createdBook.Author)
	assert.Len(t, createdBook.ChapterPaths, 2)
}

func TestBookRepository_GetBookById(t *testing.T) {
	cleanupDatabase(t)

	book := createTestBook(t, "Find Book", "Find Author", []string{"ch1.json"})

	foundBook, err := bookRepo.GetBookById(context.Background(), book.ID)
	require.NoError(t, err)
	assert.Equal(t, book.ID, foundBook.ID)
	assert.Equal(t, book.Title, foundBook.Title)
	assert.Equal(t, book.Author, foundBook.Author)
}

func TestBookRepository_GetBookByAuthorAndTitle(t *testing.T) {
	cleanupDatabase(t)

	title := "Search Book"
	author := "Search Author"
	createTestBook(t, title, author, []string{"ch1.json"})

	book, err := bookRepo.GetBookByAuthorAndTitle(context.Background(), author, title)
	require.NoError(t, err)
	assert.Equal(t, title, book.Title)
	assert.Equal(t, author, book.Author)
}

func TestBookRepository_GetBookNotFound(t *testing.T) {
	cleanupDatabase(t)

	_, err := bookRepo.GetBookById(context.Background(), uuid.New())
	assert.Error(t, err)
	assert.Equal(t, repositories.ErrBookNotFound, err)

	_, err = bookRepo.GetBookByAuthorAndTitle(context.Background(), "Unknown", "Book")
	assert.Error(t, err)
	assert.Equal(t, repositories.ErrBookNotFound, err)
}

func TestBookRepository_UpdateBook(t *testing.T) {
	cleanupDatabase(t)

	book := createTestBook(t, "Update Book", "Update Author", []string{"ch1.json"})

	newTitle := "Updated Title"
	newAuthor := "Updated Author"
	coverPath := "cover.jpg"

	err := bookRepo.UpdateBook(context.Background(), book.ID, newTitle, newAuthor, &coverPath)
	require.NoError(t, err)

	updatedBook, err := bookRepo.GetBookById(context.Background(), book.ID)
	require.NoError(t, err)
	assert.Equal(t, newTitle, updatedBook.Title)
	assert.Equal(t, newAuthor, updatedBook.Author)
	assert.NotNil(t, updatedBook.Cover)
	assert.Equal(t, coverPath, *updatedBook.Cover)
}

func TestBookRepository_DeleteBook(t *testing.T) {
	cleanupDatabase(t)

	book := createTestBook(t, "Delete Book", "Delete Author", []string{"ch1.json"})

	err := bookRepo.DeleteBook(context.Background(), book.ID)
	require.NoError(t, err)

	_, err = bookRepo.GetBookById(context.Background(), book.ID)
	assert.Error(t, err)
	assert.Equal(t, repositories.ErrBookNotFound, err)
}

func TestBookRepository_GetBooks_NoFilter(t *testing.T) {
	cleanupDatabase(t)

	createTestBook(t, "Book 1", "Author A", []string{"ch1.json"})
	createTestBook(t, "Book 2", "Author B", []string{"ch1.json"})
	createTestBook(t, "Book 3", "Author C", []string{"ch1.json"})

	books, err := bookRepo.GetBooks(context.Background(), "", "", 1)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(books), 3)
}

func TestBookRepository_GetBooks_FilterByAuthor(t *testing.T) {
	cleanupDatabase(t)

	createTestBook(t, "Harry Potter", "J.K. Rowling", []string{"ch1.json"})
	createTestBook(t, "Lord of the Rings", "J.R.R. Tolkien", []string{"ch1.json"})
	createTestBook(t, "The Hobbit", "J.R.R. Tolkien", []string{"ch1.json"})

	books, err := bookRepo.GetBooks(context.Background(), "Tolkien", "", 1)
	require.NoError(t, err)
	assert.Len(t, books, 2)
}

func TestBookRepository_GetBooks_FilterByTitle(t *testing.T) {
	cleanupDatabase(t)

	createTestBook(t, "Harry Potter", "J.K. Rowling", []string{"ch1.json"})
	createTestBook(t, "Lord of the Rings", "J.R.R. Tolkien", []string{"ch1.json"})

	books, err := bookRepo.GetBooks(context.Background(), "", "Potter", 1)
	require.NoError(t, err)
	assert.Len(t, books, 1)
	assert.Equal(t, "Harry Potter", books[0].Title)
}

func TestBookRepository_GetBooksByAuthor(t *testing.T) {
	cleanupDatabase(t)

	createTestBook(t, "Book 1", "George Orwell", []string{"ch1.json"})
	createTestBook(t, "Book 2", "George Orwell", []string{"ch1.json"})
	createTestBook(t, "Book 3", "Another Author", []string{"ch1.json"})

	books, err := bookRepo.GetBooksByAuthor(context.Background(), "George Orwell")
	require.NoError(t, err)
	assert.Len(t, books, 2)
}

func TestBookRepository_SearchBooks(t *testing.T) {
	cleanupDatabase(t)

	createTestBook(t, "The Great Gatsby", "F. Scott Fitzgerald", []string{"ch1.json"})
	createTestBook(t, "Great Expectations", "Charles Dickens", []string{"ch1.json"})
	createTestBook(t, "Other Book", "Other Author", []string{"ch1.json"})

	books, err := bookRepo.SearchBooks(context.Background(), "Great", 10)
	require.NoError(t, err)
	assert.Len(t, books, 2)
}

func TestBookRepository_CountBooks(t *testing.T) {
	cleanupDatabase(t)

	createTestBook(t, "Book 1", "Author A", []string{"ch1.json"})
	createTestBook(t, "Book 2", "Author B", []string{"ch1.json"})

	count, err := bookRepo.CountBooks(context.Background(), "", "")
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	count, err = bookRepo.CountBooks(context.Background(), "Author A", "")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestBookRepository_DuplicateBook(t *testing.T) {
	cleanupDatabase(t)

	title := "Duplicate Book"
	author := "Duplicate Author"

	book1, err := bookRepo.CreateBook(context.Background(), &models.Book{
		Title:        title,
		Author:       author,
		ChapterPaths: models.StringArray{"ch1.json"},
	})
	require.NoError(t, err)

	book2, err := bookRepo.CreateBook(context.Background(), &models.Book{
		Title:        title,
		Author:       author,
		ChapterPaths: models.StringArray{"ch2.json"},
	})
	assert.Error(t, err)
	assert.Equal(t, repositories.ErrBookAlreadyExists, err)
	assert.Equal(t, book1.ID, book2.ID)
}

func TestBookRepository_ChapterPaths_Empty(t *testing.T) {
	cleanupDatabase(t)

	book := createTestBook(t, "Empty Chapters", "Author", []string{})

	foundBook, err := bookRepo.GetBookById(context.Background(), book.ID)
	require.NoError(t, err)
	assert.NotNil(t, foundBook.ChapterPaths)
	assert.Len(t, foundBook.ChapterPaths, 0)
}

func TestBookRepository_ChapterPaths_Multiple(t *testing.T) {
	cleanupDatabase(t)

	chapters := []string{
		"chapters/book1/chapter1.json",
		"chapters/book1/chapter2.json",
		"chapters/book1/chapter3.json",
	}
	book := createTestBook(t, "Multi Chapter Book", "Author", chapters)

	foundBook, err := bookRepo.GetBookById(context.Background(), book.ID)
	require.NoError(t, err)
	assert.Len(t, foundBook.ChapterPaths, 3)
	assert.Equal(t, chapters, []string(foundBook.ChapterPaths))
}

func TestBookRepository_ChapterPaths_SpecialCharacters(t *testing.T) {
	cleanupDatabase(t)

	chapters := []string{
		"path/with spaces/chapter.json",
		"path/with-dashes/chapter.json",
		"path/with_underscores/chapter.json",
		"path/with.dots/chapter.json",
	}
	book := createTestBook(t, "Special Chars", "Author", chapters)

	foundBook, err := bookRepo.GetBookById(context.Background(), book.ID)
	require.NoError(t, err)
	assert.Equal(t, chapters, []string(foundBook.ChapterPaths))
}

func TestBookRepository_CoverOperations(t *testing.T) {
	cleanupDatabase(t)

	// Create book without cover
	book := createTestBook(t, "Book Without Cover", "Author", []string{"ch1.json"})
	assert.Nil(t, book.Cover)
	assert.False(t, book.HasCover())

	// Add cover
	coverPath := "covers/book-cover.jpg"
	err := bookRepo.UpdateBook(context.Background(), book.ID, book.Title, book.Author, &coverPath)
	require.NoError(t, err)

	updatedBook, err := bookRepo.GetBookById(context.Background(), book.ID)
	require.NoError(t, err)
	assert.NotNil(t, updatedBook.Cover)
	assert.Equal(t, coverPath, *updatedBook.Cover)
	assert.True(t, updatedBook.HasCover())
}

func TestBookRepository_GetChapterCount(t *testing.T) {
	cleanupDatabase(t)

	book1 := createTestBook(t, "Book 1", "Author", []string{})
	assert.Equal(t, 0, book1.GetChapterCount())

	book2 := createTestBook(t, "Book 2", "Author", []string{"ch1.json", "ch2.json", "ch3.json"})
	assert.Equal(t, 3, book2.GetChapterCount())
}
