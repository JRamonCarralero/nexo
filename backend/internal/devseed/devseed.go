// Package devseed fills a local database with example data so the app can be
// tried (and end-to-end tests can run) without creating everything by hand.
//
// It goes through the real services, so the data obeys the same rules and
// Row-Level Security as anything created from the app: people become users the
// way they do on their first login, workspaces are created by their admin and
// members join by accepting invitations.
//
// It is for local development only: cmd/seed refuses to run in production or
// without the Firebase Auth emulator.
package devseed

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/JRamonCarralero/nexo/backend/internal/modules/user"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/workspace"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/scope"
)

// Password is the password of every example account. They only exist in the
// local emulator.
const Password = "nexo1234"

// Accounts creates (or finds) accounts in the identity provider.
type Accounts interface {
	Ensure(ctx context.Context, email, password, name string) (auth.Identity, error)
}

// Services are the module services the seed uses.
type Services struct {
	Users       *user.Service
	Workspaces  *workspace.Service
	Invitations *workspace.InvitationService
}

type person struct{ email, name string }

var (
	ana   = person{"ana@nexo.local", "Ana García"}
	luis  = person{"luis@nexo.local", "Luis Pérez"}
	marta = person{"marta@nexo.local", "Marta Soler"}
)

// A membership of the example data: who, with which role.
type membership struct {
	who  person
	role workspace.Role
}

// A workspace of the example data. The admin creates it; the others join.
type space struct {
	name    string
	admin   person
	members []membership
	// pendingEmail is left as an open email invitation.
	pendingEmail string
}

var spaces = []space{
	{
		name:  "Lanzamiento",
		admin: ana,
		members: []membership{
			{luis, workspace.RoleMember},
			{marta, workspace.RoleGuest},
		},
		pendingEmail: "nuevo@nexo.local",
	},
	{
		name:    "Marketing",
		admin:   luis,
		members: []membership{{ana, workspace.RoleAdmin}},
	},
}

// People lists the emails of the example accounts, for messages.
func People() []string {
	return []string{ana.email, luis.email, marta.email}
}

type seeder struct {
	log      *slog.Logger
	svc      Services
	accounts Accounts
	users    map[string]uuid.UUID // by email
}

// Run creates the example data. It can be run again: workspaces that already
// exist are left as they are.
func Run(ctx context.Context, log *slog.Logger, svc Services, accounts Accounts) error {
	s := &seeder{log: log, svc: svc, accounts: accounts, users: map[string]uuid.UUID{}}

	for _, p := range []person{ana, luis, marta} {
		if err := s.person(ctx, p); err != nil {
			return err
		}
	}
	for _, sp := range spaces {
		if err := s.space(ctx, sp); err != nil {
			return fmt.Errorf("workspace %q: %w", sp.name, err)
		}
	}
	return nil
}

// person creates the account and the Nexo user, as a first login would.
func (s *seeder) person(ctx context.Context, p person) error {
	identity, err := s.accounts.Ensure(ctx, p.email, Password, p.name)
	if err != nil {
		return err
	}
	id, err := s.svc.Users.Resolve(ctx, identity)
	if err != nil {
		return fmt.Errorf("user %s: %w", p.email, err)
	}
	s.users[p.email] = id
	return nil
}

func (s *seeder) as(ctx context.Context, p person) context.Context {
	return scope.WithUser(ctx, s.users[p.email])
}

func (s *seeder) space(ctx context.Context, sp space) error {
	adminCtx := s.as(ctx, sp.admin)

	existing, err := s.svc.Workspaces.ListMine(adminCtx)
	if err != nil {
		return err
	}
	for _, w := range existing {
		if strings.EqualFold(w.Name, sp.name) {
			s.log.Info("workspace already exists, skipping", slog.String("name", sp.name))
			return nil
		}
	}

	ws, err := s.svc.Workspaces.Create(adminCtx, sp.name)
	if err != nil {
		return err
	}
	adminInWS := scope.WithWorkspaceRole(scope.WithWorkspace(adminCtx, ws.ID), string(workspace.RoleAdmin))

	for _, m := range sp.members {
		if err := s.join(ctx, adminInWS, m); err != nil {
			return err
		}
	}
	if sp.pendingEmail != "" {
		if _, err := s.svc.Invitations.Create(adminInWS, workspace.InvitationEmail, sp.pendingEmail, workspace.RoleMember); err != nil {
			return fmt.Errorf("pending invitation: %w", err)
		}
	}
	s.log.Info("workspace created", slog.String("name", sp.name), slog.Int("members", len(sp.members)+1))
	return nil
}

// join adds someone through a link invitation, like a person would, and then
// gives them their role (links never grant admin).
func (s *seeder) join(ctx, adminInWS context.Context, m membership) error {
	joinRole := m.role
	if joinRole == workspace.RoleAdmin {
		joinRole = workspace.RoleMember
	}
	created, err := s.svc.Invitations.Create(adminInWS, workspace.InvitationLink, "", joinRole)
	if err != nil {
		return fmt.Errorf("inviting %s: %w", m.who.email, err)
	}
	token := created.URL[strings.LastIndex(created.URL, "/")+1:]
	if _, err := s.svc.Invitations.Accept(s.as(ctx, m.who), token); err != nil {
		return fmt.Errorf("%s joining: %w", m.who.email, err)
	}
	// A link is reusable: revoke it so the example data leaves no open links.
	if err := s.svc.Invitations.Revoke(adminInWS, created.Invitation.ID); err != nil {
		return err
	}
	if m.role != joinRole {
		if _, err := s.svc.Workspaces.ChangeRole(adminInWS, s.users[m.who.email], m.role); err != nil {
			return fmt.Errorf("role of %s: %w", m.who.email, err)
		}
	}
	return nil
}
