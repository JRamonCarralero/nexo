package workspace

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/database/postgres"
)

var _ InvitationRepository = (*PostgresRepository)(nil)

func (r *PostgresRepository) CreateInvitation(ctx context.Context, inv Invitation, tokenHash []byte) error {
	return r.tm.WithinTx(ctx, func(ctx context.Context) error {
		_, err := r.tm.Conn(ctx).Exec(ctx, `SELECT nexo_create_invitation($1, $2, $3, $4, $5, $6)`,
			inv.ID, inv.Kind, inv.Email, inv.Role, tokenHash, inv.ExpiresAt)
		return invitationError(err)
	})
}

func (r *PostgresRepository) ListPendingInvitations(ctx context.Context) ([]Invitation, error) {
	ws, err := currentWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	var out []Invitation
	err = r.tm.WithinTx(ctx, func(ctx context.Context) error {
		// The workspace filter repeats what RLS enforces (belt and braces, like the Store).
		rows, err := r.tm.Conn(ctx).Query(ctx, `
			SELECT id, workspace_id, kind, email, role, expires_at, accepted_at, accepted_by,
			       revoked_at, created_by, created_at, updated_at
			  FROM invitations
			 WHERE workspace_id = $1 AND accepted_at IS NULL AND revoked_at IS NULL AND expires_at > now()
			 ORDER BY id DESC
			 LIMIT 500`, ws)
		if err != nil {
			return postgres.TranslateError(err)
		}
		out, err = pgx.CollectRows(rows, pgx.RowToStructByName[Invitation])
		return postgres.TranslateError(err)
	})
	return out, err
}

func (r *PostgresRepository) RevokeInvitation(ctx context.Context, id uuid.UUID) error {
	_, err := r.invitations.Update(ctx, id, database.Changes{"revoked_at": time.Now()})
	return err
}

func (r *PostgresRepository) PreviewInvitation(ctx context.Context, tokenHash []byte) (InvitationPreview, error) {
	var out InvitationPreview
	err := r.tm.WithinTx(ctx, func(ctx context.Context) error {
		rows, err := r.tm.Conn(ctx).Query(ctx, `SELECT * FROM nexo_preview_invitation($1)`, tokenHash)
		if err != nil {
			return postgres.TranslateError(err)
		}
		out, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[InvitationPreview])
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.NotFound("invitation", nil)
		}
		return postgres.TranslateError(err)
	})
	return out, err
}

func (r *PostgresRepository) AcceptInvitation(ctx context.Context, tokenHash []byte, membershipID uuid.UUID) (uuid.UUID, Role, error) {
	var (
		wsID uuid.UUID
		role Role
	)
	err := r.tm.WithinTx(ctx, func(ctx context.Context) error {
		err := r.tm.Conn(ctx).QueryRow(ctx, `SELECT workspace_id, role FROM nexo_accept_invitation($1, $2)`,
			tokenHash, membershipID).Scan(&wsID, &role)
		return invitationError(err)
	})
	return wsID, role, err
}

// invitationError maps the 'nexo:<code>' exceptions raised by the invitation
// functions (migration 00006) to application errors.
func invitationError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || !strings.HasPrefix(pgErr.Message, "nexo:") {
		return postgres.TranslateError(err)
	}
	switch strings.TrimPrefix(pgErr.Message, "nexo:") {
	case "not_admin":
		return apperror.Wrap(err, apperror.KindForbidden, CodeAdminRequired, "only workspace admins can invite")
	case "link_cannot_grant_admin":
		return apperror.Wrap(err, apperror.KindValidation, CodeLinkCannotGrantAdmin, "invitation links cannot grant the admin role")
	case "already_member":
		return apperror.Wrap(err, apperror.KindConflict, CodeAlreadyMember, "that person is already a member")
	case "invitation_not_found":
		return apperror.Wrap(err, apperror.KindNotFound, CodeInvitationNotFound, "invitation not found")
	case "invitation_expired":
		return apperror.Wrap(err, apperror.KindConflict, CodeInvitationExpired, "the invitation has expired")
	case "invitation_email_mismatch":
		return apperror.Wrap(err, apperror.KindForbidden, CodeInvitationEmailMismatch, "the invitation is for another email address")
	case "email_not_verified":
		return apperror.Wrap(err, apperror.KindForbidden, CodeEmailNotVerified, "verify your email address to accept the invitation")
	case "no_user":
		return apperror.Unauthenticated("no authenticated user")
	}
	return postgres.TranslateError(err)
}
