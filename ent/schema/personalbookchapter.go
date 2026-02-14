package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// PersonalBookChapter holds the schema definition for the PersonalBookChapter entity.
type PersonalBookChapter struct {
	ent.Schema
}

// Fields of the PersonalBookChapter.
func (PersonalBookChapter) Fields() []ent.Field {
	return append([]ent.Field{
		field.Int("order").NonNegative().Immutable(),
		field.String("title").SchemaType(varchar255),
		field.String("translated_title").SchemaType(varchar255),
		field.String("content_url").SchemaType(varchar255),
	}, defaultFields...)
}

// Edges of the PersonalBookChapter.
func (PersonalBookChapter) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("personal_book", PersonalBook.Type).Ref("personal_book_chapters").Unique(),
	}
}
