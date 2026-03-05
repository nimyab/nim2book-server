package app

import (
	"github.com/nimyab/nim2book-back/ent"
	"github.com/nimyab/nim2book-back/internal/repository"
)

type Repositories struct {
	User         *repository.UserRepository
	Author       *repository.AuthorRepository
	Book         *repository.BookRepository
	Dictionary   *repository.DictionaryRepository
	Genre        *repository.GenreRepository
	PersonalBook *repository.PersonalBookRepository
	FcmToken     *repository.FcmTokenRepository
}

func newRepositories(entClient *ent.Client) *Repositories {
	return &Repositories{
		User:         repository.NewUserRepository(entClient),
		Author:       repository.NewAuthorRepository(entClient),
		Book:         repository.NewBookRepository(entClient),
		Dictionary:   repository.NewDictionaryRepository(entClient),
		Genre:        repository.NewGenreRepository(entClient),
		PersonalBook: repository.NewPersonalBookRepository(entClient),
		FcmToken:     repository.NewFcmTokenRepository(entClient),
	}
}
