package workspace

import (
	"context"

	"github.com/google/uuid"
)

// UserProfile is the public profile of a person, as shown in member lists.
type UserProfile struct {
	ID        uuid.UUID
	Email     string
	Name      string
	AvatarURL string
}

// UserDirectory gives the profiles of people the acting user can see. The
// user module implements it; internal/app wires them together.
type UserDirectory interface {
	// Profiles returns the profiles of the given users, keyed by id. Users the
	// acting user cannot see are missing from the result.
	Profiles(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]UserProfile, error)
}
