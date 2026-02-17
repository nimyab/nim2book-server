package repository

import (
	"github.com/nimyab/nim2book-back/ent"
	"github.com/nimyab/nim2book-back/internal/domain"
)

// MapUserToDomain преобразует ent.User в domain.User
func MapUserToDomain(entUser *ent.User) *domain.User {
	if entUser == nil {
		return nil
	}

	user := &domain.User{
		ID:        entUser.ID,
		CreatedAt: entUser.CreatedAt,
		IsAdmin:   entUser.IsAdmin,
		IsVIP:     entUser.IsVip,
		Metadata:  entUser.Metadata,
	}

	// Маппинг связанных сущностей
	if entUser.Edges.GoogleAccount != nil {
		user.GoogleAccount = MapGoogleAccountToDomain(entUser.Edges.GoogleAccount)
	}

	if entUser.Edges.BasicAccount != nil {
		user.BasicAccount = MapBasicAccountToDomain(entUser.Edges.BasicAccount)
	}

	if entUser.Edges.PersonalBooks != nil {
		user.PersonalBooks = make([]*domain.PersonalBook, len(entUser.Edges.PersonalBooks))
		for i, pb := range entUser.Edges.PersonalBooks {
			user.PersonalBooks[i] = MapPersonalBookToDomain(pb)
		}
	}

	if entUser.Edges.FcmTokens != nil {
		user.FcmTokens = make([]*domain.FcmToken, len(entUser.Edges.FcmTokens))
		for i, token := range entUser.Edges.FcmTokens {
			user.FcmTokens[i] = MapFcmTokenToDomain(token)
		}
	}

	return user
}

// MapGoogleAccountToDomain преобразует ent.GoogleAccount в domain.GoogleAccount
func MapGoogleAccountToDomain(entAccount *ent.GoogleAccount) *domain.GoogleAccount {
	if entAccount == nil {
		return nil
	}

	account := &domain.GoogleAccount{
		ID:            entAccount.ID,
		CreatedAt:     entAccount.CreatedAt,
		Sub:           entAccount.Sub,
		Email:         entAccount.Email,
		EmailVerified: entAccount.EmailVerified,
		Name:          entAccount.Name,
		Picture:       entAccount.Picture,
	}

	if entAccount.Edges.User != nil {
		account.User = MapUserToDomain(entAccount.Edges.User)
	}

	return account
}

// MapBasicAccountToDomain преобразует ent.BasicAccount в domain.BasicAccount
func MapBasicAccountToDomain(entAccount *ent.BasicAccount) *domain.BasicAccount {
	if entAccount == nil {
		return nil
	}

	account := &domain.BasicAccount{
		ID:           entAccount.ID,
		CreatedAt:    entAccount.CreatedAt,
		Email:        entAccount.Email,
		PasswordHash: entAccount.PasswordHash,
		IsVerified:   entAccount.IsVerified,
		VerifyLink:   entAccount.VerifyLink,
	}

	if entAccount.Edges.User != nil {
		account.User = MapUserToDomain(entAccount.Edges.User)
	}

	return account
}

// MapBookToDomain преобразует ent.Book в domain.Book
func MapBookToDomain(entBook *ent.Book) *domain.Book {
	if entBook == nil {
		return nil
	}

	book := &domain.Book{
		ID:             entBook.ID,
		CreatedAt:      entBook.CreatedAt,
		Title:          entBook.Title,
		CoverURL:       entBook.CoverURL,
		OriginalLang:   entBook.OriginalLang,
		TranslatedLang: entBook.TranslatedLang,
	}

	// Маппинг связанных сущностей
	if entBook.Edges.Author != nil {
		book.Author = MapAuthorToDomain(entBook.Edges.Author)
	}

	if entBook.Edges.Genres != nil {
		book.Genres = make([]*domain.Genre, len(entBook.Edges.Genres))
		for i, genre := range entBook.Edges.Genres {
			book.Genres[i] = MapGenreToDomain(genre)
		}
	}

	if entBook.Edges.BookChapters != nil {
		book.BookChapters = make([]*domain.BookChapter, len(entBook.Edges.BookChapters))
		for i, chapter := range entBook.Edges.BookChapters {
			book.BookChapters[i] = MapBookChapterToDomain(chapter)
		}
	}

	return book
}

// MapPersonalBookToDomain преобразует ent.PersonalBook в domain.PersonalBook
func MapPersonalBookToDomain(entBook *ent.PersonalBook) *domain.PersonalBook {
	if entBook == nil {
		return nil
	}

	book := &domain.PersonalBook{
		ID:             entBook.ID,
		CreatedAt:      entBook.CreatedAt,
		Title:          entBook.Title,
		CoverURL:       entBook.CoverURL,
		OriginalLang:   entBook.OriginalLang,
		TranslatedLang: entBook.TranslatedLang,
		ProcessStatus:  domain.ProcessStatus(entBook.ProcessStatus),
	}

	// Маппинг связанных сущностей
	if entBook.Edges.User != nil {
		book.User = MapUserToDomain(entBook.Edges.User)
	}

	if entBook.Edges.Author != nil {
		book.Author = MapAuthorToDomain(entBook.Edges.Author)
	}

	if entBook.Edges.Genres != nil {
		book.Genres = make([]*domain.Genre, len(entBook.Edges.Genres))
		for i, genre := range entBook.Edges.Genres {
			book.Genres[i] = MapGenreToDomain(genre)
		}
	}

	if entBook.Edges.PersonalBookChapters != nil {
		book.PersonalBookChapters = make([]*domain.PersonalBookChapter, len(entBook.Edges.PersonalBookChapters))
		for i, chapter := range entBook.Edges.PersonalBookChapters {
			book.PersonalBookChapters[i] = MapPersonalBookChapterToDomain(chapter)
		}
	}

	return book
}

// MapAuthorToDomain преобразует ent.Author в domain.Author
func MapAuthorToDomain(entAuthor *ent.Author) *domain.Author {
	if entAuthor == nil {
		return nil
	}

	author := &domain.Author{
		ID:        entAuthor.ID,
		CreatedAt: entAuthor.CreatedAt,
		Name:      entAuthor.Name,
	}

	// Маппинг связанных сущностей (только если загружены)
	if entAuthor.Edges.Books != nil {
		author.Books = make([]*domain.Book, len(entAuthor.Edges.Books))
		for i, book := range entAuthor.Edges.Books {
			author.Books[i] = MapBookToDomain(book)
		}
	}

	if entAuthor.Edges.PersonalBooks != nil {
		author.PersonalBooks = make([]*domain.PersonalBook, len(entAuthor.Edges.PersonalBooks))
		for i, book := range entAuthor.Edges.PersonalBooks {
			author.PersonalBooks[i] = MapPersonalBookToDomain(book)
		}
	}

	return author
}

// MapGenreToDomain преобразует ent.Genre в domain.Genre
func MapGenreToDomain(entGenre *ent.Genre) *domain.Genre {
	if entGenre == nil {
		return nil
	}

	genre := &domain.Genre{
		ID:        entGenre.ID,
		CreatedAt: entGenre.CreatedAt,
		Name:      entGenre.Name,
	}

	// Маппинг связанных сущностей (только если загружены)
	if entGenre.Edges.Books != nil {
		genre.Books = make([]*domain.Book, len(entGenre.Edges.Books))
		for i, book := range entGenre.Edges.Books {
			genre.Books[i] = MapBookToDomain(book)
		}
	}

	if entGenre.Edges.PersonalBooks != nil {
		genre.PersonalBooks = make([]*domain.PersonalBook, len(entGenre.Edges.PersonalBooks))
		for i, book := range entGenre.Edges.PersonalBooks {
			genre.PersonalBooks[i] = MapPersonalBookToDomain(book)
		}
	}

	return genre
}

// MapBookChapterToDomain преобразует ent.BookChapter в domain.BookChapter
func MapBookChapterToDomain(entChapter *ent.BookChapter) *domain.BookChapter {
	if entChapter == nil {
		return nil
	}

	chapter := &domain.BookChapter{
		ID:              entChapter.ID,
		CreatedAt:       entChapter.CreatedAt,
		Order:           entChapter.Order,
		Title:           entChapter.Title,
		TranslatedTitle: entChapter.TranslatedTitle,
		ContentURL:      entChapter.ContentURL,
	}

	if entChapter.Edges.Book != nil {
		chapter.Book = MapBookToDomain(entChapter.Edges.Book)
	}

	return chapter
}

// MapPersonalBookChapterToDomain преобразует ent.PersonalBookChapter в domain.PersonalBookChapter
func MapPersonalBookChapterToDomain(entChapter *ent.PersonalBookChapter) *domain.PersonalBookChapter {
	if entChapter == nil {
		return nil
	}

	chapter := &domain.PersonalBookChapter{
		ID:              entChapter.ID,
		CreatedAt:       entChapter.CreatedAt,
		Order:           entChapter.Order,
		Title:           entChapter.Title,
		TranslatedTitle: entChapter.TranslatedTitle,
		ContentURL:      entChapter.ContentURL,
	}

	if entChapter.Edges.PersonalBook != nil {
		chapter.PersonalBook = MapPersonalBookToDomain(entChapter.Edges.PersonalBook)
	}

	return chapter
}

// MapDictionaryToDomain преобразует ent.Dictionary в domain.DictionaryWord
func MapDictionaryToDomain(entDict *ent.Dictionary) *domain.DictionaryWord {
	if entDict == nil {
		return nil
	}

	dict := &domain.DictionaryWord{
		ID:            entDict.ID,
		Text:          entDict.Text,
		FromLangCode:  entDict.FromLangCode,
		ToLangCode:    entDict.ToLangCode,
		PartOfSpeech:  entDict.PartOfSpeech,
		Translations:  entDict.Translations,
		Transcription: entDict.Transcription,
	}

	// Маппинг примеров
	if entDict.Edges.DictionaryExamples != nil {
		dict.Examples = make([]domain.DictionaryExample, len(entDict.Edges.DictionaryExamples))
		for i, example := range entDict.Edges.DictionaryExamples {
			dict.Examples[i] = MapDictionaryExampleToDomain(example, entDict.ID)
		}
	}

	return dict
}

// MapDictionaryExampleToDomain преобразует ent.DictionaryExample в domain.DictionaryExample
func MapDictionaryExampleToDomain(entExample *ent.DictionaryExample, dictionaryID domain.ID) domain.DictionaryExample {
	example := domain.DictionaryExample{
		ID:                entExample.ID,
		Text:              entExample.Text,
		TranslatedText:    entExample.Translation,
		WordPositionStart: entExample.TargetPositionStart,
		WordPositionEnd:   entExample.TargetPositionEnd,
		DictionaryID:      dictionaryID,
	}

	return example
}

// MapFcmTokenToDomain преобразует ent.FcmToken в domain.FcmToken
func MapFcmTokenToDomain(entToken *ent.FcmToken) *domain.FcmToken {
	if entToken == nil {
		return nil
	}

	token := &domain.FcmToken{
		ID:        entToken.ID,
		CreatedAt: entToken.CreatedAt,
		Token:     entToken.Token,
	}

	if entToken.Edges.User != nil {
		token.User = MapUserToDomain(entToken.Edges.User)
	}

	return token
}

// MapUsersToDomain преобразует срез ent.User в срез domain.User
func MapUsersToDomain(entUsers []*ent.User) []*domain.User {
	if entUsers == nil {
		return nil
	}

	users := make([]*domain.User, len(entUsers))
	for i, user := range entUsers {
		users[i] = MapUserToDomain(user)
	}
	return users
}

// MapBooksToDomain преобразует срез ent.Book в срез domain.Book
func MapBooksToDomain(entBooks []*ent.Book) []*domain.Book {
	if entBooks == nil {
		return nil
	}

	books := make([]*domain.Book, len(entBooks))
	for i, book := range entBooks {
		books[i] = MapBookToDomain(book)
	}
	return books
}

// MapPersonalBooksToDomain преобразует срез ent.PersonalBook в срез domain.PersonalBook
func MapPersonalBooksToDomain(entBooks []*ent.PersonalBook) []*domain.PersonalBook {
	if entBooks == nil {
		return nil
	}

	books := make([]*domain.PersonalBook, len(entBooks))
	for i, book := range entBooks {
		books[i] = MapPersonalBookToDomain(book)
	}
	return books
}

// MapAuthorsToDomain преобразует срез ent.Author в срез domain.Author
func MapAuthorsToDomain(entAuthors []*ent.Author) []*domain.Author {
	if entAuthors == nil {
		return nil
	}

	authors := make([]*domain.Author, len(entAuthors))
	for i, author := range entAuthors {
		authors[i] = MapAuthorToDomain(author)
	}
	return authors
}

// MapGenresToDomain преобразует срез ent.Genre в срез domain.Genre
func MapGenresToDomain(entGenres []*ent.Genre) []*domain.Genre {
	if entGenres == nil {
		return nil
	}

	genres := make([]*domain.Genre, len(entGenres))
	for i, genre := range entGenres {
		genres[i] = MapGenreToDomain(genre)
	}
	return genres
}

// MapBookChaptersToDomain преобразует срез ent.BookChapter в срез domain.BookChapter
func MapBookChaptersToDomain(entChapters []*ent.BookChapter) []*domain.BookChapter {
	if entChapters == nil {
		return nil
	}

	chapters := make([]*domain.BookChapter, len(entChapters))
	for i, chapter := range entChapters {
		chapters[i] = MapBookChapterToDomain(chapter)
	}
	return chapters
}

// MapPersonalBookChaptersToDomain преобразует срез ent.PersonalBookChapter в срез domain.PersonalBookChapter
func MapPersonalBookChaptersToDomain(entChapters []*ent.PersonalBookChapter) []*domain.PersonalBookChapter {
	if entChapters == nil {
		return nil
	}

	chapters := make([]*domain.PersonalBookChapter, len(entChapters))
	for i, chapter := range entChapters {
		chapters[i] = MapPersonalBookChapterToDomain(chapter)
	}
	return chapters
}

// MapDictionariesToDomain преобразует срез ent.Dictionary в срез domain.DictionaryWord
func MapDictionariesToDomain(entDicts []*ent.Dictionary) []*domain.DictionaryWord {
	if entDicts == nil {
		return nil
	}

	dicts := make([]*domain.DictionaryWord, len(entDicts))
	for i, dict := range entDicts {
		dicts[i] = MapDictionaryToDomain(dict)
	}
	return dicts
}

// MapFcmTokensToDomain преобразует срез ent.FcmToken в срез domain.FcmToken
func MapFcmTokensToDomain(entTokens []*ent.FcmToken) []*domain.FcmToken {
	if entTokens == nil {
		return nil
	}

	tokens := make([]*domain.FcmToken, len(entTokens))
	for i, token := range entTokens {
		tokens[i] = MapFcmTokenToDomain(token)
	}
	return tokens
}
