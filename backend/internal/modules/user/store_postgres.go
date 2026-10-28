package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database/postgres"
)

// PostgresRepository implements Repository. Users are created only through
// the SECURITY DEFINER functions of migration 00003; reads go through the
// generic Store, so Row-Level Security limits them to what the acting user may see.
type PostgresRepository struct {
	tm    *postgres.TxManager
	users database.Store[User]
}

var _ Repository = (*PostgresRepository)(nil)

func NewPostgresRepository(tm *postgres.TxManager) *PostgresRepository {
	return &PostgresRepository{tm: tm, users: postgres.NewStore[User](tm)}
}

func (r *PostgresRepository) Resolve(ctx context.Context, newID uuid.UUID, id auth.Identity) (uuid.UUID, error) {
	var userID uuid.UUID
	err := r.tm.WithinTx(ctx, func(ctx context.Context) error {
		err := r.tm.Conn(ctx).QueryRow(ctx,
			`SELECT nexo_resolve_user($1, $2, $3, $4, $5, $6)`,
			newID, id.Subject, id.Email, id.EmailVerified, id.Name, id.AvatarURL,
		).Scan(&userID)
		return postgres.TranslateError(err)
	})
	if appErr := apperror.From(err); appErr != nil && appErr.Kind == apperror.KindConflict {
		return uuid.Nil, apperror.Wrap(err, apperror.KindConflict, "EMAIL_IN_USE",
			"the email address belongs to another account")
	}
	return userID, err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	return r.users.GetByID(ctx, id)
}

func (r *PostgresRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var all []User
	q := database.Query{Filters: []database.Filter{database.In("id", ids)}, Limit: database.MaxLimit}
	for {
		page, err := r.users.List(ctx, q)
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

func (r *PostgresRepository) EnsureSuperadmin(ctx context.Context, newID uuid.UUID, email string) (BootstrapResult, error) {
	var result string
	err := r.tm.WithinTx(ctx, func(ctx context.Context) error {
		err := r.tm.Conn(ctx).QueryRow(ctx,
			`SELECT nexo_bootstrap_superadmin($1, $2)`, newID, email,
		).Scan(&result)
		return postgres.TranslateError(err)
	})
	return BootstrapResult(result), err
}
