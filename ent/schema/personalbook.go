package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// PersonalBook holds the schema definition for the PersonalBook entity.
type PersonalBook struct {
	ent.Schema
}

// Fields of the PersonalBook.
func (PersonalBook) Fields() []ent.Field {
	return append([]ent.Field{
		field.String("title").Immutable().SchemaType(varchar255),
		field.String("cover_url").SchemaType(varchar255),
		field.String("original_lang").Default("en").SchemaType(varchar10),
		field.String("translated_lang").Default("ru").SchemaType(varchar10),
	}, defaultFields...)
}

// Edges of the PersonalBook.
func (PersonalBook) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("personal_books").Unique(),
		edge.From("author", Author.Type).Ref("personal_books").Unique(),
		edge.From("genres", Genre.Type).Ref("personal_books"),
		edge.To("personal_book_chapters", PersonalBookChapter.Type),
	}
}
