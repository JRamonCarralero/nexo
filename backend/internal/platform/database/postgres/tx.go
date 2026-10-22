package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/scope"
)

type txKey struct{}

// TxManager runs functions inside a transaction carried by the context, so
// several repositories (even from different modules) share the same transaction.
//
// Every transaction publishes the request scope to PostgreSQL (app.user_id and
// app.workspace_id) so Row-Level Security policies can filter rows. Queries
// outside a transaction run without that scope and RLS returns no tenant rows:
// the database fails closed.
type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// WithinTx runs fn in a transaction. If ctx already carries one, fn joins it
// instead of opening a nested transaction. The transaction commits when fn
// returns nil and rolls back otherwise (including on panic).
//
// The scope of ctx is applied every time, also when joining: each operation
// runs with the user and workspace of the context it was called with.
func (m *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		if err := applyScope(ctx, tx); err != nil {
			return err
		}
		return fn(ctx)
	}

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return TranslateError(fmt.Errorf("beginning transaction: %w", err))
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				err = errors.Join(err, fmt.Errorf("rolling back: %w", rbErr))
			}
		}
	}()

	if err = applyScope(ctx, tx); err != nil {
		return err
	}
	if err = fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return TranslateError(fmt.Errorf("committing transaction: %w", err))
	}
	return nil
}

// Conn returns the transaction in ctx if there is one, or the pool otherwise.
// Repositories use it inside WithinTx; outside a transaction RLS sees no scope.
func (m *TxManager) Conn(ctx context.Context) DBTX {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return m.pool
}

// applyScope sets the transaction-local settings read by the RLS helper
// functions nexo_user_id() and nexo_workspace_id(). An empty value means "none".
func applyScope(ctx context.Context, tx pgx.Tx) error {
	var userID, workspaceID string
	if id, ok := scope.UserID(ctx); ok {
		userID = id.String()
	}
	if id, ok := scope.WorkspaceID(ctx); ok {
		workspaceID = id.String()
	}
	_, err := tx.Exec(ctx,
		`SELECT set_config('app.user_id', $1, true), set_config('app.workspace_id', $2, true)`,
		userID, workspaceID)
	if err != nil {
		return TranslateError(fmt.Errorf("applying request scope: %w", err))
	}
	return nil
}
