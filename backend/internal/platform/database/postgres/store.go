package postgres

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/scope"
)

// Store is the PostgreSQL implementation of database.Store.
//
// It centralizes the cross-cutting rules so repositories cannot forget them:
//   - tenant entities are always filtered by the workspace in the request scope
//     (on top of Row-Level Security, which enforces the same in the database);
//   - soft-deleted rows are invisible and Delete only marks them;
//   - id (UUIDv7), timestamps and created_by are filled automatically;
//   - every operation runs inside a transaction so RLS sees the scope;
//   - database errors become application errors.
type Store[T database.Entity] struct {
	tm   *TxManager
	meta *tableMeta
}

var _ database.Store[database.Entity] = (*Store[database.Entity])(nil)

// NewStore builds a Store for T. It panics if T is not a valid entity (a
// programming error that the first test touching the store reveals).
func NewStore[T database.Entity](tm *TxManager) *Store[T] {
	meta, err := metaFor[T]()
	if err != nil {
		panic(err)
	}
	return &Store[T]{tm: tm, meta: meta}
}

func (s *Store[T]) Create(ctx context.Context, e *T) error {
	return s.tm.WithinTx(ctx, func(ctx context.Context) error {
		b := &builder{}
		v := reflect.ValueOf(e).Elem()
		cols := make([]string, 0, len(s.meta.columns))
		vals := make([]string, 0, len(s.meta.columns))

		for _, c := range s.meta.columns {
			var val string
			switch c.name {
			case colCreatedAt, colUpdatedAt:
				val = "now()"
			case colDeletedAt:
				continue
			case colWorkspaceID:
				ws, err := s.workspace(ctx)
				if err != nil {
					return err
				}
				val = b.arg(ws)
			case colCreatedBy:
				user, ok := scope.UserID(ctx)
				if !ok {
					continue // NULL: created by the system (bootstrap, seed…)
				}
				val = b.arg(user)
			case colID:
				id := v.FieldByIndex(c.index).Interface().(uuid.UUID)
				if id == uuid.Nil {
					var err error
					if id, err = uuid.NewV7(); err != nil {
						return apperror.Internal(err, "generating id")
					}
				}
				val = b.arg(id)
			default:
				val = b.arg(v.FieldByIndex(c.index).Interface())
			}
			cols = append(cols, quote(c.name))
			vals = append(vals, val)
		}

		sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING %s",
			s.meta.table, strings.Join(cols, ", "), strings.Join(vals, ", "), s.meta.selectList)
		created, err := s.one(ctx, nil, sql, b.args)
		if err != nil {
			return err
		}
		*e = created
		return nil
	})
}

func (s *Store[T]) GetByID(ctx context.Context, id uuid.UUID) (T, error) {
	var out T
	err := s.tm.WithinTx(ctx, func(ctx context.Context) error {
		b := &builder{}
		conds, err := s.scopeConds(ctx, b)
		if err != nil {
			return err
		}
		conds = append(conds, quote(colID)+" = "+b.arg(id))
		sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", s.meta.selectList, s.meta.table, strings.Join(conds, " AND "))
		out, err = s.one(ctx, id, sql, b.args)
		return err
	})
	return out, err
}

func (s *Store[T]) Update(ctx context.Context, id uuid.UUID, changes database.Changes) (T, error) {
	if len(changes) == 0 {
		return s.GetByID(ctx, id)
	}
	var out T
	err := s.tm.WithinTx(ctx, func(ctx context.Context) error {
		b := &builder{}
		names := make([]string, 0, len(changes))
		for name := range changes {
			if !s.meta.has(name) {
				return apperror.Internal(fmt.Errorf("store %s: unknown column %q", s.meta.resource, name), "invalid update")
			}
			if managedColumns[name] {
				return apperror.Internal(fmt.Errorf("store %s: column %q is managed by the store", s.meta.resource, name), "invalid update")
			}
			names = append(names, name)
		}
		sort.Strings(names) // deterministic SQL

		sets := make([]string, 0, len(names)+1)
		for _, name := range names {
			sets = append(sets, quote(name)+" = "+b.arg(changes[name]))
		}
		sets = append(sets, quote(colUpdatedAt)+" = now()")

		conds, err := s.scopeConds(ctx, b)
		if err != nil {
			return err
		}
		conds = append(conds, quote(colID)+" = "+b.arg(id))
		sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s RETURNING %s",
			s.meta.table, strings.Join(sets, ", "), strings.Join(conds, " AND "), s.meta.selectList)
		out, err = s.one(ctx, id, sql, b.args)
		return err
	})
	return out, err
}

func (s *Store[T]) Delete(ctx context.Context, id uuid.UUID) error {
	return s.tm.WithinTx(ctx, func(ctx context.Context) error {
		b := &builder{}
		conds, err := s.scopeConds(ctx, b)
		if err != nil {
			return err
		}
		conds = append(conds, quote(colID)+" = "+b.arg(id))

		var sql string
		if s.meta.softDelete {
			sql = fmt.Sprintf("UPDATE %s SET %s = now(), %s = now() WHERE %s",
				s.meta.table, quote(colDeletedAt), quote(colUpdatedAt), strings.Join(conds, " AND "))
		} else {
			sql = fmt.Sprintf("DELETE FROM %s WHERE %s", s.meta.table, strings.Join(conds, " AND "))
		}
		tag, err := s.tm.Conn(ctx).Exec(ctx, sql, b.args...)
		if err != nil {
			return TranslateError(err)
		}
		if tag.RowsAffected() == 0 {
			return apperror.NotFound(s.meta.resource, id)
		}
		return nil
	})
}

func (s *Store[T]) List(ctx context.Context, q database.Query) (database.Page[T], error) {
	var page database.Page[T]

	limit := q.Limit
	if limit <= 0 {
		limit = database.DefaultLimit
	}
	limit = min(limit, database.MaxLimit)
	if q.Offset < 0 {
		return page, apperror.Validation("INVALID_QUERY", "offset must not be negative")
	}
	if q.After != "" && q.Offset > 0 {
		return page, apperror.Validation("INVALID_QUERY", "cursor and offset cannot be combined")
	}
	orderName := q.OrderBy
	if orderName == "" {
		orderName = colID
	}
	order, ok := s.meta.byName[orderName]
	if !ok || order.nullable {
		return page, apperror.Internal(fmt.Errorf("store %s: cannot order by %q", s.meta.resource, orderName), "invalid query")
	}

	err := s.tm.WithinTx(ctx, func(ctx context.Context) error {
		b := &builder{}
		conds, err := s.scopeConds(ctx, b)
		if err != nil {
			return err
		}
		for _, f := range q.Filters {
			cond, err := s.filterCond(f, b)
			if err != nil {
				return err
			}
			conds = append(conds, cond)
		}

		dir, cmp := "ASC", ">"
		if q.Desc {
			dir, cmp = "DESC", "<"
		}
		if q.After != "" {
			value, id, err := decodeCursor(q.After, order)
			if err != nil {
				return err
			}
			if order.name == colID {
				conds = append(conds, fmt.Sprintf("%s %s %s", quote(colID), cmp, b.arg(id)))
			} else {
				conds = append(conds, fmt.Sprintf("(%s, %s) %s (%s, %s)",
					quote(order.name), quote(colID), cmp, b.arg(value), b.arg(id)))
			}
		}

		orderBy := quote(colID) + " " + dir
		if order.name != colID {
			orderBy = fmt.Sprintf("%s %s, %s", quote(order.name), dir, orderBy)
		}
		sql := fmt.Sprintf("SELECT %s FROM %s%s ORDER BY %s LIMIT %s",
			s.meta.selectList, s.meta.table, where(conds), orderBy, b.arg(limit+1)) // one extra row tells if there is more
		if q.Offset > 0 {
			sql += " OFFSET " + b.arg(q.Offset)
		}

		items, err := s.many(ctx, sql, b.args)
		if err != nil {
			return err
		}
		if len(items) > limit {
			items = items[:limit]
			page.HasMore = true
		}
		page.Items = items

		if page.HasMore && q.Offset == 0 {
			last := reflect.ValueOf(items[len(items)-1])
			lastID := last.FieldByIndex(s.meta.byName[colID].index).Interface().(uuid.UUID)
			next, err := encodeCursor(order, last.FieldByIndex(order.index).Interface(), lastID)
			if err != nil {
				return apperror.Internal(err, "encoding cursor")
			}
			page.NextCursor = next
		}
		return nil
	})
	return page, err
}

func (s *Store[T]) Count(ctx context.Context, filters ...database.Filter) (int64, error) {
	var n int64
	err := s.tm.WithinTx(ctx, func(ctx context.Context) error {
		b := &builder{}
		conds, err := s.scopeConds(ctx, b)
		if err != nil {
			return err
		}
		for _, f := range filters {
			cond, err := s.filterCond(f, b)
			if err != nil {
				return err
			}
			conds = append(conds, cond)
		}
		sql := fmt.Sprintf("SELECT count(*) FROM %s%s", s.meta.table, where(conds))
		return TranslateError(s.tm.Conn(ctx).QueryRow(ctx, sql, b.args...).Scan(&n))
	})
	return n, err
}

// one runs a query that must return exactly one row. id is used for the
// not-found error (nil when not applicable).
func (s *Store[T]) one(ctx context.Context, id any, sql string, args []any) (T, error) {
	rows, err := s.tm.Conn(ctx).Query(ctx, sql, args...)
	if err != nil {
		var zero T
		return zero, TranslateError(err)
	}
	out, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[T])
	if errors.Is(err, pgx.ErrNoRows) {
		return out, apperror.NotFound(s.meta.resource, id)
	}
	return out, TranslateError(err)
}

func (s *Store[T]) many(ctx context.Context, sql string, args []any) ([]T, error) {
	rows, err := s.tm.Conn(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, TranslateError(err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	return items, TranslateError(err)
}

// scopeConds returns the conditions every read and write of this entity must
// carry: the workspace of the request and, if soft-deletable, not deleted.
func (s *Store[T]) scopeConds(ctx context.Context, b *builder) ([]string, error) {
	var conds []string
	if s.meta.tenant {
		ws, err := s.workspace(ctx)
		if err != nil {
			return nil, err
		}
		conds = append(conds, quote(colWorkspaceID)+" = "+b.arg(ws))
	}
	if s.meta.softDelete {
		conds = append(conds, quote(colDeletedAt)+" IS NULL")
	}
	return conds, nil
}

func (s *Store[T]) workspace(ctx context.Context) (uuid.UUID, error) {
	ws, ok := scope.WorkspaceID(ctx)
	if !ok {
		return uuid.Nil, apperror.Internal(
			fmt.Errorf("store %s: tenant entity used without a workspace in the request scope", s.meta.resource),
			"missing workspace scope")
	}
	return ws, nil
}

func (s *Store[T]) filterCond(f database.Filter, b *builder) (string, error) {
	if !s.meta.has(f.Column) {
		return "", apperror.Internal(fmt.Errorf("store %s: unknown filter column %q", s.meta.resource, f.Column), "invalid query")
	}
	col := quote(f.Column)
	switch f.Op {
	case database.OpEq:
		return col + " = " + b.arg(f.Value), nil
	case database.OpIn:
		if k := reflect.TypeOf(f.Value); k == nil || (k.Kind() != reflect.Slice && k.Kind() != reflect.Array) {
			return "", apperror.Internal(fmt.Errorf("store %s: In(%q) needs a slice", s.meta.resource, f.Column), "invalid query")
		}
		return col + " = ANY(" + b.arg(f.Value) + ")", nil
	case database.OpIsNull:
		return col + " IS NULL", nil
	case database.OpNotNull:
		return col + " IS NOT NULL", nil
	default:
		return "", apperror.Internal(fmt.Errorf("store %s: unknown filter operator %d", s.meta.resource, f.Op), "invalid query")
	}
}

func where(conds []string) string {
	if len(conds) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(conds, " AND ")
}

// builder accumulates positional arguments ($1, $2…).
type builder struct{ args []any }

func (b *builder) arg(v any) string {
	b.args = append(b.args, v)
	return "$" + strconv.Itoa(len(b.args))
}
