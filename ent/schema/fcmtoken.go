package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// FcmToken holds the schema definition for the FcmToken entity.
type FcmToken struct {
	ent.Schema
}

// Fields of the FcmToken.
func (FcmToken) Fields() []ent.Field {
	return append([]ent.Field{
		field.String("token").Unique().Immutable().SchemaType(varchar255),
	}, defaultFields...)
}

// Edges of the FcmToken.
func (FcmToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("fcm_tokens").Unique(),
	}
}
