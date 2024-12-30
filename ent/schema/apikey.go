package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// ApiKey holds the schema definition for the ApiKey entity.
type ApiKey struct {
	ent.Schema
}

// Fields of the ApiKey.
func (ApiKey) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			Unique(),
		field.String("description").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Bool("is_admin").
			Default(false),
	}
}
