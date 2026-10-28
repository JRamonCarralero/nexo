package app

import (
	"context"

	"github.com/google/uuid"

	"github.com/JRamonCarralero/nexo/backend/internal/modules/user"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/workspace"
)

// Modules never import each other: when one needs another, it declares a
// port and these adapters connect them here, in the composition root.

// userDirectory adapts the user service to workspace.UserDirectory.
type userDirectory struct{ users *user.Service }

func (d userDirectory) Profiles(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]workspace.UserProfile, error) {
	list, err := d.users.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	profiles := make(map[uuid.UUID]workspace.UserProfile, len(list))
	for _, u := range list {
		profiles[u.ID] = workspace.UserProfile{ID: u.ID, Email: u.Email, Name: u.Name, AvatarURL: u.AvatarURL}
	}
	return profiles, nil
}
