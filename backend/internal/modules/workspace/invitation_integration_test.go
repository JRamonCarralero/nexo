//go:build integration

package workspace_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JRamonCarralero/nexo/backend/internal/modules/workspace"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/email"
)

// outbox is a Mailer that keeps the messages instead of sending them.
type outbox struct{ sent []email.Message }

func (o *outbox) Send(_ context.Context, m email.Message) error {
	o.sent = append(o.sent, m)
	return nil
}

var linkToken = regexp.MustCompile(`/invitations/([A-Za-z0-9_-]{43})`)

// token extracts the invitation token from the last email sent.
func (o *outbox) token(t *testing.T) string {
	t.Helper()
	require.NotEmpty(t, o.sent, "no email sent")
	m := linkToken.FindStringSubmatch(o.sent[len(o.sent)-1].Text)
	require.NotNil(t, m, "no invitation link in the email")
	return m[1]
}

type invWorld struct {
	world
	mail *outbox
	inv  *workspace.InvitationService
}

func newInvWorld(t *testing.T) invWorld {
	t.Helper()
	w := newWorld(t)
	_, err := testDB.OwnerPool.Exec(context.Background(), `TRUNCATE invitations`)
	require.NoError(t, err)
	mail := &outbox{}
	return invWorld{world: w, mail: mail,
		inv: workspace.NewInvitationService(w.repo, w.repo, directory{w.users}, mail, "http://localhost:5173")}
}

// person signs someone up with the given email and verification state.
func (w invWorld) person(t *testing.T, name, address string, verified bool) uuid.UUID {
	t.Helper()
	id, err := w.users.Resolve(context.Background(),
		auth.Identity{Subject: "uid-" + name, Email: address, EmailVerified: verified, Name: name})
	require.NoError(t, err)
	return id
}

func (w invWorld) asAdmin(t *testing.T) context.Context { return w.in(t, w.ana, w.ws) }

func (w invWorld) inviteByEmail(t *testing.T, address string, role workspace.Role) string {
	t.Helper()
	_, err := w.inv.Create(w.asAdmin(t), workspace.InvitationEmail, address, role)
	require.NoError(t, err)
	return w.mail.token(t)
}

func (w invWorld) inviteByLink(t *testing.T, role workspace.Role) string {
	t.Helper()
	created, err := w.inv.Create(w.asAdmin(t), workspace.InvitationLink, "", role)
	require.NoError(t, err)
	m := linkToken.FindStringSubmatch(created.URL)
	require.NotNil(t, m)
	return m[1]
}

func TestEmailInvitationFlow(t *testing.T) {
	w := newInvWorld(t)
	token := w.inviteByEmail(t, "bob@example.com", workspace.RoleGuest)

	preview, err := w.inv.Preview(asUserID(w.eva), token)
	require.NoError(t, err)
	assert.Equal(t, "Product team", preview.WorkspaceName)
	assert.Equal(t, "ana", preview.InvitedByName)
	assert.Equal(t, workspace.InvitationPending, preview.Status)

	pending, err := w.inv.List(w.asAdmin(t))
	require.NoError(t, err)
	require.Len(t, pending, 1)

	bob := w.person(t, "bob", "bob@example.com", true)
	joined, err := w.inv.Accept(asUserID(bob), token)
	require.NoError(t, err)
	assert.Equal(t, w.ws, joined.ID)
	assert.Equal(t, workspace.RoleGuest, joined.Role)

	mine, err := w.svc.ListMine(asUserID(bob))
	require.NoError(t, err)
	require.Len(t, mine, 1)
	assert.Equal(t, workspace.RoleGuest, mine[0].Role)

	pending, err = w.inv.List(w.asAdmin(t))
	require.NoError(t, err)
	assert.Empty(t, pending, "an accepted email invitation is no longer pending")

	again, err := w.inv.Accept(asUserID(bob), token)
	require.NoError(t, err, "accepting twice is idempotent for the same person")
	assert.Equal(t, w.ws, again.ID)

	preview, err = w.inv.Preview(asUserID(bob), token)
	require.NoError(t, err)
	assert.Equal(t, workspace.InvitationAccepted, preview.Status)
}

func TestEmailInvitationNeedsTheSameVerifiedEmail(t *testing.T) {
	w := newInvWorld(t)
	token := w.inviteByEmail(t, "bob@example.com", workspace.RoleMember)

	_, err := w.inv.Accept(asUserID(w.eva), token)
	assert.Equal(t, workspace.CodeInvitationEmailMismatch, code(t, err), "someone else with the link cannot use it")

	impostor := w.person(t, "bob", "BOB@example.com", false)
	_, err = w.inv.Accept(asUserID(impostor), token)
	assert.Equal(t, workspace.CodeEmailNotVerified, code(t, err), "an unverified account with that email cannot use it")

	verified := w.person(t, "bob", "bob@example.com", true) // same identity, now verified
	_, err = w.inv.Accept(asUserID(verified), token)
	require.NoError(t, err)
}

func TestReinvitingAnEmailRevokesThePreviousInvitation(t *testing.T) {
	w := newInvWorld(t)
	first := w.inviteByEmail(t, "bob@example.com", workspace.RoleMember)
	second := w.inviteByEmail(t, "bob@example.com", workspace.RoleMember)
	bob := w.person(t, "bob", "bob@example.com", true)

	_, err := w.inv.Accept(asUserID(bob), first)
	assert.Equal(t, workspace.CodeInvitationNotFound, code(t, err))
	_, err = w.inv.Accept(asUserID(bob), second)
	require.NoError(t, err)
}

func TestMembersCannotBeInvitedAgain(t *testing.T) {
	w := newInvWorld(t)

	_, err := w.inv.Create(w.asAdmin(t), workspace.InvitationEmail, "LUIS@example.com", workspace.RoleMember)

	assert.Equal(t, workspace.CodeAlreadyMember, code(t, err))
	assert.Empty(t, w.mail.sent)
}

func TestLinkInvitationsAreReusableAndKeepExistingRoles(t *testing.T) {
	w := newInvWorld(t)
	token := w.inviteByLink(t, workspace.RoleMember)

	joined, err := w.inv.Accept(asUserID(w.eva), token)
	require.NoError(t, err)
	assert.Equal(t, workspace.RoleMember, joined.Role)

	zoe := w.person(t, "zoe", "zoe@example.com", false)
	_, err = w.inv.Accept(asUserID(zoe), token)
	require.NoError(t, err, "links are reusable and need no verified email")

	stays, err := w.inv.Accept(asUserID(w.ana), token)
	require.NoError(t, err)
	assert.Equal(t, workspace.RoleAdmin, stays.Role, "accepting never changes an existing member's role")
}

func TestExpiredAndRevokedInvitationsCannotBeAccepted(t *testing.T) {
	w := newInvWorld(t)
	expired := w.inviteByLink(t, workspace.RoleMember)
	_, err := testDB.OwnerPool.Exec(context.Background(), `UPDATE invitations SET expires_at = now() - interval '1 minute'`)
	require.NoError(t, err)

	_, err = w.inv.Accept(asUserID(w.eva), expired)
	assert.Equal(t, workspace.CodeInvitationExpired, code(t, err))
	p, err := w.inv.Preview(asUserID(w.eva), expired)
	require.NoError(t, err)
	assert.Equal(t, workspace.InvitationExpired, p.Status)

	revoked := w.inviteByLink(t, workspace.RoleMember)
	pending, err := w.inv.List(w.asAdmin(t))
	require.NoError(t, err)
	require.Len(t, pending, 1, "expired invitations are not listed")
	require.NoError(t, w.inv.Revoke(w.asAdmin(t), pending[0].ID))

	_, err = w.inv.Accept(asUserID(w.eva), revoked)
	assert.Equal(t, workspace.CodeInvitationNotFound, code(t, err))
	p, err = w.inv.Preview(asUserID(w.eva), revoked)
	require.NoError(t, err)
	assert.Equal(t, workspace.InvitationRevoked, p.Status)
}

func TestUnknownTokensAreNotFound(t *testing.T) {
	w := newInvWorld(t)
	unknown := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	_, err := w.inv.Preview(asUserID(w.eva), unknown)
	assert.Equal(t, workspace.CodeInvitationNotFound, code(t, err))
	_, err = w.inv.Accept(asUserID(w.eva), unknown)
	assert.Equal(t, workspace.CodeInvitationNotFound, code(t, err))
}

// The database enforces the rules even if the service checks were bypassed.
func TestInvitationRulesAtTheDatabaseLevel(t *testing.T) {
	w := newInvWorld(t)
	asLuis := w.in(t, w.luis, w.ws)
	link := workspace.Invitation{Kind: workspace.InvitationLink, Role: workspace.RoleMember}
	link.ID = uuid.Must(uuid.NewV7())
	link.ExpiresAt = time.Now().Add(time.Hour)

	err := w.repo.CreateInvitation(asLuis, link, []byte("hash-1"))
	assert.Equal(t, workspace.CodeAdminRequired, code(t, err), "members cannot invite")

	admin := link
	admin.Role = workspace.RoleAdmin
	err = w.repo.CreateInvitation(w.asAdmin(t), admin, []byte("hash-2"))
	assert.Equal(t, workspace.CodeLinkCannotGrantAdmin, code(t, err))

	w.inviteByLink(t, workspace.RoleMember)
	list, err := w.repo.ListPendingInvitations(asLuis)
	require.NoError(t, err)
	assert.Empty(t, list, "only admins see invitations")
	outsider, err := w.repo.ListPendingInvitations(w.in(t, w.eva, w.ws))
	require.NoError(t, err)
	assert.Empty(t, outsider)
}

func TestOnlyTheTokenHashIsStored(t *testing.T) {
	w := newInvWorld(t)
	token := w.inviteByLink(t, workspace.RoleMember)

	var stored []byte
	require.NoError(t, testDB.OwnerPool.QueryRow(context.Background(), `SELECT token_hash FROM invitations`).Scan(&stored))
	assert.Len(t, stored, 32, "SHA-256")
	assert.NotContains(t, string(stored), token)
}
