package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/auth/domain/entity"
	userEntity "github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/config"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/jwt"
)

type loginUsers struct {
	user     *userEntity.AppUser
	accesses int
}

func (r *loginUsers) GetByEmail(_ context.Context, email string) (*userEntity.AppUser, error) {
	if r.user != nil && r.user.Email == email {
		return r.user, nil
	}
	return nil, nil
}

func (r *loginUsers) GetByID(_ context.Context, id uuid.UUID) (*userEntity.AppUser, error) {
	if r.user != nil && r.user.ID == id {
		return r.user, nil
	}
	return nil, nil
}

func (r *loginUsers) RecordAccess(context.Context, uuid.UUID) error {
	r.accesses++
	return nil
}

type loginSessions struct {
	created []*entity.UserSession
}

func (r *loginSessions) Create(_ context.Context, session *entity.UserSession) error {
	r.created = append(r.created, session)
	return nil
}

func (r *loginSessions) GetByTokenHash(context.Context, string) (*entity.UserSession, error) {
	return nil, nil
}

func (r *loginSessions) IsActive(context.Context, string) (bool, error) {
	return false, nil
}

func (r *loginSessions) End(context.Context, uuid.UUID, time.Time, entity.SessionCloseReason) error {
	return nil
}

func (r *loginSessions) EndAllByUserID(context.Context, uuid.UUID, time.Time, entity.SessionCloseReason) error {
	return nil
}

type loginAttempts struct {
	attempts []*entity.LoginAttempt
}

func (r *loginAttempts) Create(_ context.Context, attempt *entity.LoginAttempt) error {
	r.attempts = append(r.attempts, attempt)
	return nil
}

func setupLogin(t *testing.T) (*AuthService, *loginUsers, *loginSessions, *loginAttempts, clock.Clock) {
	t.Helper()
	mockClock := clock.NewMockClock(time.Now())
	user, err := userEntity.NewUser(mockClock, userEntity.RegisterUserParams{
		RoleID:    uuid.New(),
		FirstName: "Juan",
		LastName:  "Pérez",
		Email:     "juan@sigif.com",
		Password:  "password123",
	})
	if err != nil {
		t.Fatalf("new user: %v", err)
	}
	users := &loginUsers{user: user}
	sessions := &loginSessions{}
	attempts := &loginAttempts{}
	manager := jwt.NewManager(&config.Config{JWT: config.JWTConfig{Secret: "test-secret", AccessTokenExpiry: 15, Issuer: "sigif-test"}})
	return NewAuthService(users, sessions, attempts, manager, mockClock), users, sessions, attempts, mockClock
}

func TestLoginSucceedsForActiveUser(t *testing.T) {
	svc, users, sessions, attempts, _ := setupLogin(t)

	result, err := svc.Login(context.Background(), LoginParams{Email: "juan@sigif.com", Password: "password123"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if result.AccessToken == "" || len(sessions.created) != 1 || users.accesses != 1 {
		t.Fatalf("active user must get a session: token=%q sessions=%d", result.AccessToken, len(sessions.created))
	}
	if len(attempts.attempts) != 1 || !attempts.attempts[0].Successful {
		t.Fatalf("unexpected attempts: %+v", attempts.attempts)
	}
}

func TestLoginRejectsDeactivatedUser(t *testing.T) {
	svc, users, sessions, attempts, mockClock := setupLogin(t)
	users.user.Deactivate(mockClock)

	result, err := svc.Login(context.Background(), LoginParams{Email: "juan@sigif.com", Password: "password123"})
	if !sharedErrors.Is(err, sharedErrors.CodeUnauthorized) {
		t.Fatalf("want UNAUTHORIZED, got %v", err)
	}
	if err.Error() != "invalid credentials" {
		t.Fatalf("login must not reveal the account state: %q", err.Error())
	}
	if result.AccessToken != "" || len(sessions.created) != 0 || users.accesses != 0 {
		t.Fatal("deactivated user must not get a session")
	}
	if len(attempts.attempts) != 1 || attempts.attempts[0].Successful || attempts.attempts[0].FailureReason != "account_not_active" {
		t.Fatalf("unexpected attempts: %+v", attempts.attempts)
	}
}

func TestLoginWorksAgainAfterReactivation(t *testing.T) {
	svc, users, _, _, mockClock := setupLogin(t)
	users.user.Deactivate(mockClock)
	users.user.Activate(mockClock)

	if _, err := svc.Login(context.Background(), LoginParams{Email: "juan@sigif.com", Password: "password123"}); err != nil {
		t.Fatalf("reactivated user must be able to log in: %v", err)
	}
}
