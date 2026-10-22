// Package auth defines how the API verifies who is calling. The identity
// provider (Firebase today) lives in a subpackage; the rest of the code only
// sees this interface.
package auth

import "context"

// Identity is what the identity provider vouches for after verifying a token.
type Identity struct {
	// Subject is the provider's stable user id (Firebase UID).
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	AvatarURL     string
	// Provider is how the user signed in (password, google.com…).
	Provider string
}

// Authenticator verifies a bearer token. Invalid, expired or malformed tokens
// return an apperror of kind Unauthenticated.
type Authenticator interface {
	VerifyToken(ctx context.Context, token string) (Identity, error)
}
