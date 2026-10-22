package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/JRamonCarralero/nexo/backend/internal/modules/user"
	"github.com/JRamonCarralero/nexo/backend/internal/modules/user/mocks"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/apperror"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/auth"
	"github.com/JRamonCarralero/nexo/backend/internal/platform/scope"
)

var identity = auth.Identity{Subject: "uid-1", Email: "ana@example.com", EmailVerified: true}

func TestResolveReturnsUserID(t *testing.T) {
	repo := mocks.NewRepository(t)
	want := uuid.New()
	repo.EXPECT().Resolve(mock.Anything, mock.Anything, identity).Return(want, nil).Once()

	got, err := user.NewService(repo).Resolve(context.Background(), identity)

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestResolveRetriesOnceAfterARace(t *testing.T) {
	repo := mocks.NewRepository(t)
	want := uuid.New()
	race := apperror.Conflict("ALREADY_EXISTS", "created by a concurrent request")
	repo.EXPECT().Resolve(mock.Anything, mock.Anything, identity).Return(uuid.Nil, race).Once()
	repo.EXPECT().Resolve(mock.Anything, mock.Anything, identity).Return(want, nil).Once()

	got, err := user.NewService(repo).Resolve(context.Background(), identity)

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestResolveGivesUpAfterTwoConflicts(t *testing.T) {
	repo := mocks.NewRepository(t)
	inUse := apperror.Conflict("EMAIL_IN_USE", "taken")
	repo.EXPECT().Resolve(mock.Anything, mock.Anything, identity).Return(uuid.Nil, inUse).Twice()

	_, err := user.NewService(repo).Resolve(context.Background(), identity)

	assert.ErrorIs(t, err, inUse)
}

func TestResolveDoesNotRetryOtherErrors(t *testing.T) {
	repo := mocks.NewRepository(t)
	boom := errors.New("db down")
	repo.EXPECT().Resolve(mock.Anything, mock.Anything, identity).Return(uuid.Nil, boom).Once()

	_, err := user.NewService(repo).Resolve(context.Background(), identity)

	assert.ErrorIs(t, err, boom)
}

func TestResolveSendsUUIDv7(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().Resolve(mock.Anything, mock.MatchedBy(func(id uuid.UUID) bool { return id.Version() == 7 }), identity).
		Return(uuid.New(), nil).Once()

	_, err := user.NewService(repo).Resolve(context.Background(), identity)

	require.NoError(t, err)
}

func TestMeRequiresAnAuthenticatedUser(t *testing.T) {
	_, err := user.NewService(mocks.NewRepository(t)).Me(context.Background())

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.KindUnauthenticated, appErr.Kind)
}

func TestMeReturnsTheScopedUser(t *testing.T) {
	repo := mocks.NewRepository(t)
	id := uuid.New()
	u := user.User{Email: "ana@example.com"}
	u.ID = id
	repo.EXPECT().GetByID(mock.Anything, id).Return(u, nil).Once()

	got, err := user.NewService(repo).Me(scope.WithUser(context.Background(), id))

	require.NoError(t, err)
	assert.Equal(t, u, got)
}

func TestEnsureSuperadminValidatesEmail(t *testing.T) {
	for _, email := range []string{"", "not-an-email", "Ana <ana@example.com>"} {
		t.Run(email, func(t *testing.T) {
			_, err := user.NewService(mocks.NewRepository(t)).EnsureSuperadmin(context.Background(), email)

			var appErr *apperror.Error
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, "INVALID_EMAIL", appErr.Code)
		})
	}
}

func TestEnsureSuperadminTrimsAndDelegates(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().EnsureSuperadmin(mock.Anything, mock.Anything, "admin@example.com").
		Return(user.BootstrapCreated, nil).Once()

	got, err := user.NewService(repo).EnsureSuperadmin(context.Background(), "  admin@example.com ")

	require.NoError(t, err)
	assert.Equal(t, user.BootstrapCreated, got)
}
