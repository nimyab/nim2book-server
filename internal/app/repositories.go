package app

import (
	"github.com/nimyab/nim2book-back/ent"
	"github.com/nimyab/nim2book-back/internal/repository"
	"github.com/samber/do/v2"
)

func (a *App) registerRepositories() {
	// UserRepository
	do.Provide(a.injector, func(i do.Injector) (*repository.UserRepository, error) {
		entClient := do.MustInvoke[*ent.Client](i)
		return repository.NewUserRepository(entClient), nil
	})

	// AuthorRepository
	do.Provide(a.injector, func(i do.Injector) (*repository.AuthorRepository, error) {
		entClient := do.MustInvoke[*ent.Client](i)
		return repository.NewAuthorRepository(entClient), nil
	})

	// BookRepository
	do.Provide(a.injector, func(i do.Injector) (*repository.BookRepository, error) {
		entClient := do.MustInvoke[*ent.Client](i)
		return repository.NewBookRepository(entClient), nil
	})

	// DictionaryRepository
	do.Provide(a.injector, func(i do.Injector) (*repository.DictionaryRepository, error) {
		entClient := do.MustInvoke[*ent.Client](i)
		return repository.NewDictionaryRepository(entClient), nil
	})

	// GenreRepository
	do.Provide(a.injector, func(i do.Injector) (*repository.GenreRepository, error) {
		entClient := do.MustInvoke[*ent.Client](i)
		return repository.NewGenreRepository(entClient), nil
	})

	// PersonalBookRepository
	do.Provide(a.injector, func(i do.Injector) (*repository.PersonalBookRepository, error) {
		entClient := do.MustInvoke[*ent.Client](i)
		return repository.NewPersonalBookRepository(entClient), nil
	})

	// FcmTokenRepository
	do.Provide(a.injector, func(i do.Injector) (*repository.FcmTokenRepository, error) {
		entClient := do.MustInvoke[*ent.Client](i)
		return repository.NewFcmTokenRepository(entClient), nil
	})
}
