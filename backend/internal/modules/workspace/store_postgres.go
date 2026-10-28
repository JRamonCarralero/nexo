package workspace

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database/postgres"
)

// PostgresRepository implements Repository. Workspaces are created through
// nexo_create_workspace (migration 00005); Row-Level Security limits every
// read and write to what the acting user may see and do.
type PostgresRepository struct {
	tm          *postgres.TxManager
	workspaces  database.Store[Workspace]
	memberships database.Store[Membership]
}

var _ Repository = (*PostgresRepository)(nil)

func NewPostgresRepository(tm *postgres.TxManager) *PostgresRepository {
	return &PostgresRepository{
		tm:          tm,
		workspaces:  postgres.NewStore[Workspace](tm),
		memberships: postgres.NewStore[Membership](tm),
	}
}

func (r *PostgresRepository) Create(ctx context.Context, w Workspace, membershipID uuid.UUID) error {
	return r.tm.WithinTx(ctx, func(ctx context.Context) error {
		_, err := r.tm.Conn(ctx).Exec(ctx, `SELECT nexo_create_workspace($1, $2, $3)`, w.ID, membershipID, w.Name)
		return postgres.TranslateError(err)
	})
}

func (r *PostgresRepository) ListForUser(ctx context.Context, userID uuid.UUID) ([]WithRole, error) {
	var out []WithRole
	err := r.tm.WithinTx(ctx, func(ctx context.Context) error {
		rows, err := r.tm.Conn(ctx).Query(ctx, `
			SELECT w.id, w.name, w.created_by, w.created_at, w.updated_at, w.deleted_at, m.role
			  FROM memberships m
			  JOIN workspaces w ON w.id = m.workspace_id
			 WHERE m.user_id = $1 AND w.deleted_at IS NULL
			 ORDER BY lower(w.name), w.id`, userID)
		if err != nil {
			return postgres.TranslateError(err)
		}
		out, err = pgx.CollectRows(rows, pgx.RowToStructByName[WithRole])
		return postgres.TranslateError(err)
	})
	return out, err
}

func (r *PostgresRepository) RoleOf(ctx context.Context, workspaceID, userID uuid.UUID) (Role, error) {
	var role Role
	err := r.tm.WithinTx(ctx, func(ctx context.Context) error {
		err := r.tm.Conn(ctx).QueryRow(ctx, `
			SELECT m.role
			  FROM memberships m
			  JOIN workspaces w ON w.id = m.workspace_id AND w.deleted_at IS NULL
			 WHERE m.workspace_id = $1 AND m.user_id = $2`, workspaceID, userID).Scan(&role)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.NotFound("workspace", workspaceID)
		}
		return postgres.TranslateError(err)
	})
	return role, err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (Workspace, error) {
	return r.workspaces.GetByID(ctx, id)
}

func (r *PostgresRepository) Rename(ctx context.Context, id uuid.UUID, name string) (Workspace, error) {
	return r.workspaces.Update(ctx, id, database.Changes{"name": name})
}

func (r *PostgresRepository) ListMemberships(ctx context.Context) ([]Membership, error) {
	var all []Membership
	q := database.Query{Limit: database.MaxLimit}
	for {
		page, err := r.memberships.List(ctx, q)
		if err != nil {
			return nil, err
		}
		all = append(all, page.Items...)
		if !page.HasMore {
			return all, nil
		}
		q.After = page.NextCursor
	}
}

func (r *PostgresRepository) GetMembership(ctx context.Context, userID uuid.UUID) (Membership, error) {
	page, err := r.memberships.List(ctx, database.Query{
		Filters: []database.Filter{database.Eq("user_id", userID)},
		Limit:   1,
	})
	if err != nil {
		return Membership{}, err
	}
	if len(page.Items) == 0 {
		return Membership{}, apperror.NotFound("member", userID)
	}
	return page.Items[0], nil
}

func (r *PostgresRepository) UpdateRole(ctx context.Context, membershipID uuid.UUID, role Role) (Membership, error) {
	m, err := r.memberships.Update(ctx, membershipID, database.Changes{"role": role})
	return m, lastAdmin(err)
}

func (r *PostgresRepository) RemoveMembership(ctx context.Context, membershipID uuid.UUID) error {
	return lastAdmin(r.memberships.Delete(ctx, membershipID))
}

// lastAdmin turns the nexo_keep_one_admin trigger error into LAST_ADMIN.
func lastAdmin(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Message == "nexo:last_admin" {
		return apperror.Wrap(err, apperror.KindConflict, CodeLastAdmin, "a workspace needs at least one admin")
	}
	return err
}
