package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Author holds the schema definition for the Author entity.
type Author struct {
	ent.Schema
}

// Fields of the Author.
func (Author) Fields() []ent.Field {
	return append([]ent.Field{
		field.String("name").Immutable().SchemaType(varchar255),
	}, defaultFields...)
}

// Edges of the Author.
func (Author) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("personal_books", PersonalBook.Type),
		edge.To("books", Book.Type),
	}
}
