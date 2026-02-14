package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// BookChapter holds the schema definition for the BookChapter entity.
type BookChapter struct {
	ent.Schema
}

// Fields of the BookChapter.
func (BookChapter) Fields() []ent.Field {
	return append([]ent.Field{
		field.Int("order").NonNegative().Immutable(),
		field.String("title").SchemaType(varchar255),
		field.String("translated_title").SchemaType(varchar255),
		field.String("content_url").SchemaType(varchar255),
	}, defaultFields...)
}

// Edges of the BookChapter.
func (BookChapter) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("book", Book.Type).Ref("book_chapters").Unique(),
	}
}
