package postgres

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/database"
)

// Managed column names. The Store sets them; callers cannot write them.
const (
	colID          = "id"
	colWorkspaceID = "workspace_id"
	colCreatedAt   = "created_at"
	colUpdatedAt   = "updated_at"
	colCreatedBy   = "created_by"
	colDeletedAt   = "deleted_at"
)

var managedColumns = map[string]bool{
	colID: true, colWorkspaceID: true, colCreatedAt: true,
	colUpdatedAt: true, colCreatedBy: true, colDeletedAt: true,
}

type column struct {
	name     string
	index    []int
	typ      reflect.Type
	nullable bool // pointer types: cannot be used for cursor ordering
}

// tableMeta describes how an entity maps to its table. Built once per type.
type tableMeta struct {
	table      string // quoted identifier
	resource   string
	columns    []column
	byName     map[string]column
	selectList string // quoted, comma-separated column list

	tenant     bool // has workspace_id: scoped by the request workspace
	softDelete bool // has deleted_at
	createdBy  bool // has created_by
}

func (m *tableMeta) has(name string) bool {
	_, ok := m.byName[name]
	return ok
}

var metaCache sync.Map // reflect.Type → *tableMeta

func metaFor[T database.Entity]() (*tableMeta, error) {
	var zero T
	typ := reflect.TypeOf(zero)
	if cached, ok := metaCache.Load(typ); ok {
		return cached.(*tableMeta), nil
	}
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("store: entity %s must be a struct", typ)
	}

	m := &tableMeta{
		table:    pgx.Identifier{zero.TableName()}.Sanitize(),
		resource: zero.ResourceName(),
		byName:   map[string]column{},
	}
	quoted := make([]string, 0, typ.NumField())
	for _, f := range reflect.VisibleFields(typ) {
		name := f.Tag.Get("db")
		if f.Anonymous || !f.IsExported() || name == "" || name == "-" {
			continue
		}
		if _, dup := m.byName[name]; dup {
			return nil, fmt.Errorf("store: entity %s maps column %q twice", typ, name)
		}
		c := column{name: name, index: f.Index, typ: f.Type, nullable: f.Type.Kind() == reflect.Pointer}
		m.columns = append(m.columns, c)
		m.byName[name] = c
		quoted = append(quoted, pgx.Identifier{name}.Sanitize())
	}
	if !m.has(colID) || !m.has(colCreatedAt) || !m.has(colUpdatedAt) {
		return nil, fmt.Errorf("store: entity %s must embed database.Model (id, created_at, updated_at)", typ)
	}
	m.selectList = strings.Join(quoted, ", ")
	m.tenant = m.has(colWorkspaceID)
	m.softDelete = m.has(colDeletedAt)
	m.createdBy = m.has(colCreatedBy)

	actual, _ := metaCache.LoadOrStore(typ, m)
	return actual.(*tableMeta), nil
}

func quote(name string) string { return pgx.Identifier{name}.Sanitize() }
