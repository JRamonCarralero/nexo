//go:build integration

package firebase_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth/firebase"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth/firebase/fbtest"
)

var emulator *fbtest.Emulator

func TestMain(m *testing.M) {
	ctx := context.Background()
	var err error
	emulator, err = fbtest.Start(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Setenv(firebase.EmulatorHostEnv, emulator.Host)
	code := m.Run()
	emulator.Stop(ctx)
	os.Exit(code)
}

func newAuthenticator(t *testing.T) *firebase.Authenticator {
	t.Helper()
	a, err := firebase.New(context.Background(), fbtest.ProjectID)
	require.NoError(t, err)
	return a
}

func TestVerifyTokenReturnsIdentity(t *testing.T) {
	ctx := context.Background()
	u, err := emulator.SignUp(ctx, "ana@example.com", "secret123")
	require.NoError(t, err)

	id, err := newAuthenticator(t).VerifyToken(ctx, u.IDToken)

	require.NoError(t, err)
	assert.Equal(t, u.UID, id.Subject)
	assert.Equal(t, "ana@example.com", id.Email)
	assert.False(t, id.EmailVerified)
	assert.Equal(t, "password", id.Provider)
}

func TestVerifiedEmailIsReported(t *testing.T) {
	ctx := context.Background()
	u, err := emulator.SignUp(ctx, "luis@example.com", "secret123")
	require.NoError(t, err)
	require.NoError(t, emulator.VerifyEmail(ctx, u.UID))
	fresh, err := emulator.SignIn(ctx, "luis@example.com", "secret123")
	require.NoError(t, err)

	id, err := newAuthenticator(t).VerifyToken(ctx, fresh.IDToken)

	require.NoError(t, err)
	assert.True(t, id.EmailVerified)
}

func TestInvalidTokenIsUnauthenticated(t *testing.T) {
	_, err := newAuthenticator(t).VerifyToken(context.Background(), "not-a-token")

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.KindUnauthenticated, appErr.Kind)
	assert.Equal(t, "INVALID_TOKEN", appErr.Code)
}
