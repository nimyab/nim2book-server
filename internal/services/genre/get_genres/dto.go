package get_genres

import "github.com/nimyab/nim2book-back/internal/domain"

type Output struct {
	Genres []domain.Genre `json:"genres"`
}
