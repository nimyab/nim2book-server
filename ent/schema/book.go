package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Book holds the schema definition for the Book entity.
type Book struct {
	ent.Schema
}

// Fields of the Book.
func (Book) Fields() []ent.Field {
	return append([]ent.Field{
		field.String("title").Immutable().SchemaType(varchar255),
		field.String("cover_url").Nillable().SchemaType(varchar255),
		field.String("original_lang").Default("en").SchemaType(varchar10),
		field.String("translated_lang").Default("ru").SchemaType(varchar10),
	}, defaultFields...)
}

// Edges of the Book.
func (Book) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("author", Author.Type).Ref("books").Unique(),
		edge.From("genres", Genre.Type).Ref("books"),
		edge.To("book_chapters", BookChapter.Type),
	}
}
