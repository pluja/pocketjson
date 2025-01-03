// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"pocketjson/ent/apikey"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
)

// ApiKey is the model entity for the ApiKey schema.
type ApiKey struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Key holds the value of the "key" field.
	Key string `json:"key,omitempty"`
	// Description holds the value of the "description" field.
	Description string `json:"description,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// IsAdmin holds the value of the "is_admin" field.
	IsAdmin      bool `json:"is_admin,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*ApiKey) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case apikey.FieldIsAdmin:
			values[i] = new(sql.NullBool)
		case apikey.FieldID:
			values[i] = new(sql.NullInt64)
		case apikey.FieldKey, apikey.FieldDescription:
			values[i] = new(sql.NullString)
		case apikey.FieldCreatedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the ApiKey fields.
func (ak *ApiKey) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case apikey.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			ak.ID = int(value.Int64)
		case apikey.FieldKey:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field key", values[i])
			} else if value.Valid {
				ak.Key = value.String
			}
		case apikey.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				ak.Description = value.String
			}
		case apikey.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				ak.CreatedAt = value.Time
			}
		case apikey.FieldIsAdmin:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field is_admin", values[i])
			} else if value.Valid {
				ak.IsAdmin = value.Bool
			}
		default:
			ak.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the ApiKey.
// This includes values selected through modifiers, order, etc.
func (ak *ApiKey) Value(name string) (ent.Value, error) {
	return ak.selectValues.Get(name)
}

// Update returns a builder for updating this ApiKey.
// Note that you need to call ApiKey.Unwrap() before calling this method if this ApiKey
// was returned from a transaction, and the transaction was committed or rolled back.
func (ak *ApiKey) Update() *ApiKeyUpdateOne {
	return NewApiKeyClient(ak.config).UpdateOne(ak)
}

// Unwrap unwraps the ApiKey entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (ak *ApiKey) Unwrap() *ApiKey {
	_tx, ok := ak.config.driver.(*txDriver)
	if !ok {
		panic("ent: ApiKey is not a transactional entity")
	}
	ak.config.driver = _tx.drv
	return ak
}

// String implements the fmt.Stringer.
func (ak *ApiKey) String() string {
	var builder strings.Builder
	builder.WriteString("ApiKey(")
	builder.WriteString(fmt.Sprintf("id=%v, ", ak.ID))
	builder.WriteString("key=")
	builder.WriteString(ak.Key)
	builder.WriteString(", ")
	builder.WriteString("description=")
	builder.WriteString(ak.Description)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(ak.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("is_admin=")
	builder.WriteString(fmt.Sprintf("%v", ak.IsAdmin))
	builder.WriteByte(')')
	return builder.String()
}

// ApiKeys is a parsable slice of ApiKey.
type ApiKeys []*ApiKey
