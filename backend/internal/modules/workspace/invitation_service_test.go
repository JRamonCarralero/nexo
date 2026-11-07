package workspace_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/JRamonCarralero/nexo/backend/internal/modules/workspace"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/workspace/mocks"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/email"
	emailmocks "github.com/JRamonCarralero/nexo/backend/internal/platform/email/mocks"
)

const appURL = "https://nexo.example.com"

type invEnv struct {
	repo        *mocks.Repository
	invitations *mocks.InvitationRepository
	users       *mocks.UserDirectory
	mailer      *emailmocks.Mailer
	svc         *workspace.InvitationService
}

func newInvEnv(t *testing.T) invEnv {
	e := invEnv{
		repo: mocks.NewRepository(t), invitations: mocks.NewInvitationRepository(t),
		users: mocks.NewUserDirectory(t), mailer: emailmocks.NewMailer(t),
	}
	e.svc = workspace.NewInvitationService(e.repo, e.invitations, e.users, e.mailer, appURL+"/")
	return e
}

// expectEmailContext sets up what an email invitation needs to render the email.
func (e invEnv) expectEmailContext() {
	w := workspace.Workspace{Name: "Product team"}
	w.ID = wsID
	e.repo.EXPECT().GetByID(mock.Anything, wsID).Return(w, nil).Once()
	e.users.EXPECT().Profiles(mock.Anything, []uuid.UUID{me}).
		Return(map[uuid.UUID]workspace.UserProfile{me: {ID: me, Name: "Ana"}}, nil).Once()
}

func tokenFromURL(t *testing.T, url string) string {
	t.Helper()
	prefix := appURL + "/invitations/"
	require.True(t, strings.HasPrefix(url, prefix), "unexpected url %q", url)
	return strings.TrimPrefix(url, prefix)
}

func hashOf(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func TestCreateInvitationIsForAdminsOnly(t *testing.T) {
	_, err := newInvEnv(t).svc.Create(inWorkspace(workspace.RoleMember), workspace.InvitationLink, "", workspace.RoleMember)
	assert.Equal(t, workspace.CodeAdminRequired, code(t, err))
}

func TestCreateInvitationValidatesInput(t *testing.T) {
	tests := []struct {
		name  string
		kind  workspace.InvitationKind
		email string
		role  workspace.Role
		want  string
	}{
		{"unknown kind", "carrier-pigeon", "", workspace.RoleMember, workspace.CodeInvalidInvitationKind},
		{"unknown role", workspace.InvitationLink, "", "owner", workspace.CodeInvalidRole},
		{"link as admin", workspace.InvitationLink, "", workspace.RoleAdmin, workspace.CodeLinkCannotGrantAdmin},
		{"missing email", workspace.InvitationEmail, "  ", workspace.RoleMember, workspace.CodeInvalidEmail},
		{"invalid email", workspace.InvitationEmail, "not-an-email", workspace.RoleMember, workspace.CodeInvalidEmail},
		{"display name", workspace.InvitationEmail, "Ana <ana@example.com>", workspace.RoleMember, workspace.CodeInvalidEmail},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newInvEnv(t).svc.Create(inWorkspace(workspace.RoleAdmin), tt.kind, tt.email, tt.role)
			assert.Equal(t, tt.want, code(t, err))
		})
	}
}

func TestLinkInvitationReturnsAURLWhoseTokenMatchesTheStoredHash(t *testing.T) {
	e := newInvEnv(t)
	var stored []byte
	var inv workspace.Invitation
	e.invitations.EXPECT().CreateInvitation(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, i workspace.Invitation, hash []byte) error {
			inv, stored = i, hash
			return nil
		}).Once()

	created, err := e.svc.Create(inWorkspace(workspace.RoleAdmin), workspace.InvitationLink, "", workspace.RoleGuest)

	require.NoError(t, err)
	token := tokenFromURL(t, created.URL)
	assert.Len(t, token, 43, "256-bit token, base64url without padding")
	assert.Equal(t, hashOf(token), stored, "only the hash is stored")
	assert.NotContains(t, string(stored), token)
	assert.Equal(t, workspace.RoleGuest, inv.Role)
	assert.Nil(t, inv.Email)
	assert.WithinDuration(t, time.Now().Add(workspace.LinkInvitationTTL), inv.ExpiresAt, time.Minute)
	assert.Equal(t, uuid.Version(7), inv.ID.Version())
}

func TestEmailInvitationSendsTheLinkByEmailOnly(t *testing.T) {
	e := newInvEnv(t)
	var stored []byte
	e.invitations.EXPECT().CreateInvitation(mock.Anything, mock.MatchedBy(func(i workspace.Invitation) bool {
		return i.Email != nil && *i.Email == "luis@example.com" && i.Role == workspace.RoleMember
	}), mock.Anything).RunAndReturn(func(_ context.Context, _ workspace.Invitation, hash []byte) error {
		stored = hash
		return nil
	}).Once()
	e.expectEmailContext()
	var sent email.Message
	e.mailer.EXPECT().Send(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, m email.Message) error {
		sent = m
		return nil
	}).Once()

	created, err := e.svc.Create(inWorkspace(workspace.RoleAdmin), workspace.InvitationEmail, " Luis@Example.com ", workspace.RoleMember)

	require.NoError(t, err)
	assert.Empty(t, created.URL, "the email token is only given to the recipient")
	assert.Equal(t, "luis@example.com", sent.To, "emails are trimmed and lower-cased")
	assert.Contains(t, sent.Subject, "Ana")
	assert.Contains(t, sent.Subject, "Product team")
	assert.Contains(t, sent.HTML, "Product team")

	start := strings.Index(sent.Text, appURL+"/invitations/")
	require.GreaterOrEqual(t, start, 0, "the email contains the invitation link")
	token := tokenFromURL(t, strings.Fields(sent.Text[start:])[0])
	assert.Equal(t, hashOf(token), stored)
}

func TestEmailDeliveryFailureRevokesTheInvitation(t *testing.T) {
	e := newInvEnv(t)
	var id uuid.UUID
	e.invitations.EXPECT().CreateInvitation(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, i workspace.Invitation, _ []byte) error {
			id = i.ID
			return nil
		}).Once()
	e.expectEmailContext()
	e.mailer.EXPECT().Send(mock.Anything, mock.Anything).Return(apperror.Unavailable(errors.New("smtp down"), "down")).Once()
	e.invitations.EXPECT().RevokeInvitation(mock.Anything, mock.MatchedBy(func(got uuid.UUID) bool { return got == id })).
		Return(nil).Once()

	_, err := e.svc.Create(inWorkspace(workspace.RoleAdmin), workspace.InvitationEmail, "luis@example.com", workspace.RoleMember)

	assert.Equal(t, workspace.CodeEmailDeliveryFailed, code(t, err))
}

func TestListAndRevokeAreForAdminsOnly(t *testing.T) {
	e := newInvEnv(t)
	member := inWorkspace(workspace.RoleMember)

	_, err := e.svc.List(member)
	assert.Equal(t, workspace.CodeAdminRequired, code(t, err))
	assert.Equal(t, workspace.CodeAdminRequired, code(t, e.svc.Revoke(member, uuid.New())))
}

func TestMalformedTokensAreRejectedWithoutQueryingTheDatabase(t *testing.T) {
	e := newInvEnv(t) // no repository expectation: any call would fail the test

	for _, token := range []string{"", "short", strings.Repeat("a", 44), "not base64!!"} {
		_, err := e.svc.Preview(asUser(), token)
		assert.Equal(t, "INVITATION_NOT_FOUND", code(t, err))
		_, err = e.svc.Accept(asUser(), token)
		assert.Equal(t, "INVITATION_NOT_FOUND", code(t, err))
	}
}

func TestAcceptRequiresAnAuthenticatedUser(t *testing.T) {
	_, err := newInvEnv(t).svc.Accept(context.Background(), strings.Repeat("A", 43))

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.KindUnauthenticated, appErr.Kind)
}
