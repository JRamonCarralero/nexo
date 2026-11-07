package workspace

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/database"
)

// InvitationKind says how an invitation is delivered.
type InvitationKind string

const (
	// InvitationEmail is for one person: single use, sent by email, accepted
	// only by an account with that email, verified.
	InvitationEmail InvitationKind = "email"
	// InvitationLink is a shareable link: reusable until it expires or is
	// revoked. It never grants the admin role.
	InvitationLink InvitationKind = "link"
)

// Valid reports whether k is a known kind.
func (k InvitationKind) Valid() bool { return k == InvitationEmail || k == InvitationLink }

// How long invitations stay valid.
const (
	EmailInvitationTTL = 7 * 24 * time.Hour
	LinkInvitationTTL  = 30 * 24 * time.Hour
)

// Invitation lets someone join a workspace with a role. The token itself is
// never stored, only its SHA-256 hash (not mapped here: it is never read back).
type Invitation struct {
	database.TenantModel
	Kind       InvitationKind `db:"kind"`
	Email      *string        `db:"email"`
	Role       Role           `db:"role"`
	ExpiresAt  time.Time      `db:"expires_at"`
	AcceptedAt *time.Time     `db:"accepted_at"`
	AcceptedBy *uuid.UUID     `db:"accepted_by"`
	RevokedAt  *time.Time     `db:"revoked_at"`
	CreatedBy  *uuid.UUID     `db:"created_by"`
}

func (Invitation) TableName() string    { return "invitations" }
func (Invitation) ResourceName() string { return "invitation" }

// InvitationStatus is the state of an invitation as seen by whoever holds its token.
type InvitationStatus string

const (
	InvitationPending  InvitationStatus = "pending"
	InvitationExpired  InvitationStatus = "expired"
	InvitationRevoked  InvitationStatus = "revoked"
	InvitationAccepted InvitationStatus = "accepted"
)

// InvitationPreview is what the confirmation screen shows before accepting.
type InvitationPreview struct {
	WorkspaceID   uuid.UUID        `db:"workspace_id"`
	WorkspaceName string           `db:"workspace_name"`
	Kind          InvitationKind   `db:"kind"`
	Email         *string          `db:"email"`
	Role          Role             `db:"role"`
	InvitedByName string           `db:"invited_by_name"`
	ExpiresAt     time.Time        `db:"expires_at"`
	Status        InvitationStatus `db:"status"`
}

// CreatedInvitation is a new invitation and, for links, the URL to share.
// The URL contains the token and is only available right after creating it.
type CreatedInvitation struct {
	Invitation Invitation
	URL        string
}

// InvitationRepository is the persistence of invitations.
type InvitationRepository interface {
	// CreateInvitation stores an invitation in the current workspace (admins only).
	CreateInvitation(ctx context.Context, inv Invitation, tokenHash []byte) error
	// ListPendingInvitations returns the current workspace's usable invitations, newest first.
	ListPendingInvitations(ctx context.Context) ([]Invitation, error)
	// RevokeInvitation makes an invitation of the current workspace unusable.
	RevokeInvitation(ctx context.Context, id uuid.UUID) error
	// PreviewInvitation describes the invitation of a token, or INVITATION_NOT_FOUND.
	PreviewInvitation(ctx context.Context, tokenHash []byte) (InvitationPreview, error)
	// AcceptInvitation joins the acting user to the invitation's workspace.
	AcceptInvitation(ctx context.Context, tokenHash []byte, membershipID uuid.UUID) (uuid.UUID, Role, error)
}
