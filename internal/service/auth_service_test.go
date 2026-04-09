package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/thwqsz/uptime-monitor/internal/mocks"
	"github.com/thwqsz/uptime-monitor/internal/models"
	"github.com/thwqsz/uptime-monitor/internal/service"
)

func TestAuthService_Register_UserAlreadyExistsReturnsErrUserAlreadyExists(t *testing.T) {
	ctx := context.Background()

	userRepo := mocks.NewUserRepository(t)
	svc := service.NewAuthService(userRepo, "secret")

	existingUser := &models.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: "hash",
	}

	userRepo.
		On("GetByEmail", mock.Anything, "test@example.com").
		Return(existingUser, nil)

	err := svc.Register(ctx, "test@example.com", "password123")

	require.Error(t, err)
	require.ErrorIs(t, err, service.ErrUserAlreadyExists)

	userRepo.AssertNotCalled(t, "CreateUser", mock.Anything, mock.Anything)
}

func TestAuthService_Register_WhenUserNotFoundCreatesUser(t *testing.T) {
	ctx := context.Background()

	userRepo := mocks.NewUserRepository(t)
	svc := service.NewAuthService(userRepo, "secret")

	userRepo.
		On("GetByEmail", mock.Anything, "test@example.com").
		Return((*models.User)(nil), sql.ErrNoRows)

	userRepo.
		On("CreateUser", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
			return user.Email == "test@example.com" && user.PasswordHash != ""
		})).
		Return(nil)

	err := svc.Register(ctx, "test@example.com", "password123")

	require.NoError(t, err)
}

func TestAuthService_Register_CreateUserErrorReturnsError(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("create user failed")

	userRepo := mocks.NewUserRepository(t)
	svc := service.NewAuthService(userRepo, "secret")

	userRepo.
		On("GetByEmail", mock.Anything, "test@example.com").
		Return((*models.User)(nil), sql.ErrNoRows)

	userRepo.
		On("CreateUser", mock.Anything, mock.Anything).
		Return(repoErr)

	err := svc.Register(ctx, "test@example.com", "password123")

	require.Error(t, err)
	require.ErrorIs(t, err, repoErr)
}

func TestAuthService_Register_GetByEmailUnexpectedErrorReturnsError(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db unavailable")

	userRepo := mocks.NewUserRepository(t)
	svc := service.NewAuthService(userRepo, "secret")

	userRepo.
		On("GetByEmail", mock.Anything, "test@example.com").
		Return((*models.User)(nil), repoErr)

	err := svc.Register(ctx, "test@example.com", "password123")

	require.Error(t, err)
	require.ErrorIs(t, err, repoErr)

	userRepo.AssertNotCalled(t, "CreateUser", mock.Anything, mock.Anything)
}

func TestAuthService_Login_UserNotFoundReturnsErrInvalidCredentials(t *testing.T) {
	ctx := context.Background()

	userRepo := mocks.NewUserRepository(t)
	svc := service.NewAuthService(userRepo, "secret")

	userRepo.
		On("GetByEmail", mock.Anything, "test@example.com").
		Return((*models.User)(nil), sql.ErrNoRows)

	token, err := svc.Login(ctx, "test@example.com", "password123")

	require.Error(t, err)
	require.Empty(t, token)
	require.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestAuthService_Login_WrongPasswordReturnsErrInvalidCredentials(t *testing.T) {
	ctx := context.Background()

	userRepo := mocks.NewUserRepository(t)
	svc := service.NewAuthService(userRepo, "secret")

	hash, err := service.HashPassword("correct-password")
	require.NoError(t, err)

	user := &models.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: hash,
	}

	userRepo.
		On("GetByEmail", mock.Anything, "test@example.com").
		Return(user, nil)

	token, err := svc.Login(ctx, "test@example.com", "wrong-password")

	require.Error(t, err)
	require.Empty(t, token)
	require.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestAuthService_Login_SuccessReturnsToken(t *testing.T) {
	ctx := context.Background()

	userRepo := mocks.NewUserRepository(t)
	svc := service.NewAuthService(userRepo, "secret")

	hash, err := service.HashPassword("password123")
	require.NoError(t, err)

	user := &models.User{
		ID:           42,
		Email:        "test@example.com",
		PasswordHash: hash,
	}

	userRepo.
		On("GetByEmail", mock.Anything, "test@example.com").
		Return(user, nil)

	token, err := svc.Login(ctx, "test@example.com", "password123")

	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestAuthService_Login_GetByEmailUnexpectedErrorReturnsError(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db unavailable")

	userRepo := mocks.NewUserRepository(t)
	svc := service.NewAuthService(userRepo, "secret")

	userRepo.
		On("GetByEmail", mock.Anything, "test@example.com").
		Return((*models.User)(nil), repoErr)

	token, err := svc.Login(ctx, "test@example.com", "password123")

	require.Error(t, err)
	require.Empty(t, token)
	require.ErrorIs(t, err, repoErr)
}
