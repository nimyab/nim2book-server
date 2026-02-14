package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// GoogleAccount holds the schema definition for the GoogleAccount entity.
type GoogleAccount struct {
	ent.Schema
}

// Fields of the GoogleAccount.
func (GoogleAccount) Fields() []ent.Field {
	return append([]ent.Field{
		field.String("sub").Unique().Immutable().SchemaType(varchar50),
		field.String("email").Unique().SchemaType(varchar255),
		field.Bool("email_verified"),
		field.String("name").SchemaType(varchar255),
		field.String("picture"),
	}, defaultFields...)
}

// Edges of the GoogleAccount.
func (GoogleAccount) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("google_account").Unique(),
	}
}
