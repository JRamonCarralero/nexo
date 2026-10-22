package postgres

import (
	"encoding/base64"
	"encoding/json"
	"reflect"

	"github.com/google/uuid"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
)

// cursor marks the last row of a page: the value of the sort column and the id
// (the tie-breaker). It is opaque to clients: base64url-encoded JSON.
type cursor struct {
	Column string          `json:"c"`
	Value  json.RawMessage `json:"v"`
	ID     uuid.UUID       `json:"id"`
}

func encodeCursor(col column, value any, id uuid.UUID) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(cursor{Column: col.name, Value: raw, ID: id})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// decodeCursor returns the sort value (typed like the column) and the id.
func decodeCursor(s string, col column) (any, uuid.UUID, error) {
	invalid := apperror.Validation("INVALID_CURSOR", "invalid pagination cursor")

	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, uuid.Nil, invalid
	}
	var c cursor
	if err := json.Unmarshal(b, &c); err != nil || c.ID == uuid.Nil {
		return nil, uuid.Nil, invalid
	}
	if c.Column != col.name {
		return nil, uuid.Nil, apperror.Validation("INVALID_CURSOR", "cursor belongs to a different ordering")
	}
	value := reflect.New(col.typ)
	if err := json.Unmarshal(c.Value, value.Interface()); err != nil {
		return nil, uuid.Nil, invalid
	}
	return value.Elem().Interface(), c.ID, nil
}
