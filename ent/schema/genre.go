package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Genre holds the schema definition for the Genre entity.
type Genre struct {
	ent.Schema
}

// Fields of the Genre.
func (Genre) Fields() []ent.Field {
	return append([]ent.Field{
		field.
			String("name").
			Unique().
			Immutable().
			SchemaType(varchar255),
	}, defaultFields...)
}

// Edges of the Genre.
func (Genre) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("books", Book.Type),
		edge.To("personal_books", PersonalBook.Type),
	}
}
