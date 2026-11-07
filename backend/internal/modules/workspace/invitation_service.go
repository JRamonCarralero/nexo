package workspace

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/email"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/scope"
)

// Error codes specific to invitations.
const (
	CodeInvalidInvitationKind   = "INVALID_INVITATION_KIND"
	CodeInvalidEmail            = "INVALID_EMAIL"
	CodeLinkCannotGrantAdmin    = "LINK_CANNOT_GRANT_ADMIN"
	CodeAlreadyMember           = "ALREADY_MEMBER"
	CodeInvitationNotFound      = "INVITATION_NOT_FOUND"
	CodeInvitationExpired       = "INVITATION_EXPIRED"
	CodeInvitationEmailMismatch = "INVITATION_EMAIL_MISMATCH"
	CodeEmailNotVerified        = "EMAIL_NOT_VERIFIED"
	CodeEmailDeliveryFailed     = "EMAIL_DELIVERY_FAILED"
)

// tokenBytes is the entropy of an invitation token: 256 bits.
const tokenBytes = 32

// InvitationService creates, lists, revokes and accepts invitations.
type InvitationService struct {
	repo        Repository
	invitations InvitationRepository
	users       UserDirectory
	mailer      email.Mailer
	appURL      string
	newID       func() (uuid.UUID, error)
	now         func() time.Time
}

// NewInvitationService builds the service. appURL is the web app's base URL,
// used to build invitation links (APP_BASE_URL).
func NewInvitationService(repo Repository, invitations InvitationRepository, users UserDirectory,
	mailer email.Mailer, appURL string,
) *InvitationService {
	return &InvitationService{
		repo: repo, invitations: invitations, users: users, mailer: mailer,
		appURL: strings.TrimRight(appURL, "/"), newID: uuid.NewV7, now: time.Now,
	}
}

// Create invites someone to the current workspace. Admins only.
//   - email: sends the link by email; inviting the same email again revokes
//     the previous invitation.
//   - link: returns a shareable URL, shown only now; it cannot grant admin.
func (s *InvitationService) Create(ctx context.Context, kind InvitationKind, address string, role Role) (CreatedInvitation, error) {
	ws, err := currentWorkspace(ctx)
	if err != nil {
		return CreatedInvitation{}, err
	}
	if err := requireAdmin(ctx); err != nil {
		return CreatedInvitation{}, err
	}
	if !kind.Valid() {
		return CreatedInvitation{}, apperror.Validation(CodeInvalidInvitationKind, "unknown invitation kind: "+string(kind),
			apperror.FieldError{Field: "kind", Code: CodeInvalidInvitationKind})
	}
	if !role.Valid() {
		return CreatedInvitation{}, apperror.Validation(CodeInvalidRole, "unknown role: "+string(role),
			apperror.FieldError{Field: "role", Code: CodeInvalidRole})
	}
	if kind == InvitationLink && role == RoleAdmin {
		return CreatedInvitation{}, apperror.Validation(CodeLinkCannotGrantAdmin, "invitation links cannot grant the admin role",
			apperror.FieldError{Field: "role", Code: CodeLinkCannotGrantAdmin})
	}

	inv := Invitation{Kind: kind, Role: role}
	inv.WorkspaceID = ws
	if kind == InvitationEmail {
		normalized, err := validEmail(address)
		if err != nil {
			return CreatedInvitation{}, err
		}
		inv.Email = &normalized
		inv.ExpiresAt = s.now().Add(EmailInvitationTTL)
	} else {
		inv.ExpiresAt = s.now().Add(LinkInvitationTTL)
	}
	if inv.ID, err = s.newID(); err != nil {
		return CreatedInvitation{}, apperror.Internal(err, "generating invitation id")
	}
	inv.CreatedAt, inv.UpdatedAt = s.now(), s.now()
	if me, ok := scope.UserID(ctx); ok {
		inv.CreatedBy = &me
	}

	token, hash, err := newToken()
	if err != nil {
		return CreatedInvitation{}, apperror.Internal(err, "generating invitation token")
	}
	if err := s.invitations.CreateInvitation(ctx, inv, hash); err != nil {
		return CreatedInvitation{}, err
	}

	url := s.appURL + "/invitations/" + token
	if kind == InvitationLink {
		return CreatedInvitation{Invitation: inv, URL: url}, nil
	}
	if err := s.sendInvitationEmail(ctx, inv, url); err != nil {
		// Nobody received it: do not leave a usable invitation behind. The
		// transactional outbox (ADR-0015) will turn this into retries.
		_ = s.invitations.RevokeInvitation(ctx, inv.ID)
		return CreatedInvitation{}, apperror.Wrap(err, apperror.KindUnavailable, CodeEmailDeliveryFailed,
			"the invitation email could not be sent")
	}
	return CreatedInvitation{Invitation: inv}, nil
}

// List returns the pending invitations of the current workspace. Admins only.
func (s *InvitationService) List(ctx context.Context) ([]Invitation, error) {
	if _, err := currentWorkspace(ctx); err != nil {
		return nil, err
	}
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	return s.invitations.ListPendingInvitations(ctx)
}

// Revoke makes an invitation of the current workspace unusable. Admins only.
func (s *InvitationService) Revoke(ctx context.Context, id uuid.UUID) error {
	if _, err := currentWorkspace(ctx); err != nil {
		return err
	}
	if err := requireAdmin(ctx); err != nil {
		return err
	}
	return s.invitations.RevokeInvitation(ctx, id)
}

// Preview describes the invitation of a token for the confirmation screen.
func (s *InvitationService) Preview(ctx context.Context, token string) (InvitationPreview, error) {
	hash, ok := hashToken(token)
	if !ok {
		return InvitationPreview{}, apperror.NotFound("invitation", nil)
	}
	return s.invitations.PreviewInvitation(ctx, hash)
}

// Accept joins the acting user to the invitation's workspace and returns it.
func (s *InvitationService) Accept(ctx context.Context, token string) (WithRole, error) {
	if _, ok := scope.UserID(ctx); !ok {
		return WithRole{}, apperror.Unauthenticated("no authenticated user")
	}
	hash, ok := hashToken(token)
	if !ok {
		return WithRole{}, apperror.NotFound("invitation", nil)
	}
	membershipID, err := s.newID()
	if err != nil {
		return WithRole{}, apperror.Internal(err, "generating membership id")
	}
	wsID, role, err := s.invitations.AcceptInvitation(ctx, hash, membershipID)
	if err != nil {
		return WithRole{}, err
	}
	w, err := s.repo.GetByID(ctx, wsID)
	if err != nil {
		return WithRole{}, err
	}
	return WithRole{Workspace: w, Role: role}, nil
}

func (s *InvitationService) sendInvitationEmail(ctx context.Context, inv Invitation, url string) error {
	w, err := s.repo.GetByID(ctx, inv.WorkspaceID)
	if err != nil {
		return err
	}
	inviter := ""
	if inv.CreatedBy != nil {
		profiles, err := s.users.Profiles(ctx, []uuid.UUID{*inv.CreatedBy})
		if err != nil {
			return err
		}
		inviter = profiles[*inv.CreatedBy].Name
		if inviter == "" {
			inviter = profiles[*inv.CreatedBy].Email
		}
	}
	msg, err := invitationEmail(invitationEmailData{
		To: *inv.Email, Inviter: inviter, Workspace: w.Name, Role: inv.Role, URL: url, ExpiresAt: inv.ExpiresAt,
	})
	if err != nil {
		return apperror.Internal(err, "rendering invitation email")
	}
	return s.mailer.Send(ctx, msg)
}

// newToken returns a random URL-safe token and its SHA-256 hash.
func newToken() (string, []byte, error) {
	raw := make([]byte, tokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	return token, sum[:], nil
}

// hashToken returns the stored hash of a token, or false if it cannot be one
// of ours (so malformed tokens are rejected without a database round trip).
func hashToken(token string) ([]byte, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != tokenBytes {
		return nil, false
	}
	sum := sha256.Sum256([]byte(token))
	return sum[:], true
}

func validEmail(address string) (string, error) {
	address = strings.TrimSpace(address)
	parsed, err := mail.ParseAddress(address)
	if address == "" || err != nil || parsed.Address != address {
		return "", apperror.Validation(CodeInvalidEmail, "invalid email address",
			apperror.FieldError{Field: "email", Code: CodeInvalidEmail})
	}
	return strings.ToLower(address), nil
}
