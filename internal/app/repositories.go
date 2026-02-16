package app

import (
	"github.com/nimyab/nim2book-back/ent"
	"github.com/nimyab/nim2book-back/internal/repository"
	"github.com/samber/do/v2"
)

func (a *App) registerRepositories() {
	// UserRepository
	do.Provide(a.injector, func(i do.Injector) (*repository.UserRepository, error) {
		ent := do.MustInvoke[*ent.Client](i)
		return repository.NewUserRepository(ent), nil
	})

	// AuthorRepository
	do.Provide(a.injector, func(i do.Injector) (*repository.AuthorRepository, error) {
		ent := do.MustInvoke[*ent.Client](i)
		return repository.NewAuthorRepository(ent), nil
	})

	// BookRepository
	do.Provide(a.injector, func(i do.Injector) (*repository.BookRepository, error) {
		ent := do.MustInvoke[*ent.Client](i)
		return repository.NewBookRepository(ent), nil
	})

	// DictionaryRepository
	do.Provide(a.injector, func(i do.Injector) (*repository.DictionaryRepository, error) {
		ent := do.MustInvoke[*ent.Client](i)
		return repository.NewDictionaryRepository(ent), nil
	})

	// GenreRepository
	do.Provide(a.injector, func(i do.Injector) (*repository.GenreRepository, error) {
		ent := do.MustInvoke[*ent.Client](i)
		return repository.NewGenreRepository(ent), nil
	})

	// PersonalBookRepository
	do.Provide(a.injector, func(i do.Injector) (*repository.PersonalBookRepository, error) {
		ent := do.MustInvoke[*ent.Client](i)
		return repository.NewPersonalBookRepository(ent), nil
	})

	// FcmTokenRepository
	do.Provide(a.injector, func(i do.Injector) (*repository.FcmTokenRepository, error) {
		ent := do.MustInvoke[*ent.Client](i)
		return repository.NewFcmTokenRepository(ent), nil
	})
}
