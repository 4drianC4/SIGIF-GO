package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/auth/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/auth/domain/repository"
	userEntity "github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/jwt"
	"github.com/sigif/sigif-go/internal/shared/security"
)

type AuthService struct {
	userRepo         repository.UserRepo
	sessionRepo      repository.SessionRepository
	loginAttemptRepo repository.LoginAttemptRepository
	jwtManager       *jwt.JWTManager
	clock            clock.Clock
}

func NewAuthService(
	userRepo repository.UserRepo,
	sessionRepo repository.SessionRepository,
	loginAttemptRepo repository.LoginAttemptRepository,
	jwtManager *jwt.JWTManager,
	clock clock.Clock,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		sessionRepo:      sessionRepo,
		loginAttemptRepo: loginAttemptRepo,
		jwtManager:       jwtManager,
		clock:            clock,
	}
}

type LoginParams struct {
	Email     string
	Password  string
	IPAddress string
	Device    string
}

type LoginResult struct {
	User        *userEntity.AppUser
	Permissions []string
	AccessToken string
	ExpiresIn   int
	TokenType   string
}

// Login authenticates the user, creates a session and records the attempt.
// It always returns a generic error to avoid account enumeration.
func (s *AuthService) Login(ctx context.Context, params LoginParams) (LoginResult, error) {
	user, err := s.userRepo.GetByEmail(ctx, params.Email)
	if err != nil {
		return LoginResult{}, err
	}

	fail := func(reason string) (LoginResult, error) {
		_ = s.loginAttemptRepo.Create(ctx, entity.NewLoginAttempt(
			s.clock, params.Email, companyIDOf(user), params.IPAddress, false, reason,
		))
		return LoginResult{}, sharedErrors.New(sharedErrors.CodeUnauthorized, "invalid credentials", 401)
	}

	if user == nil {
		return fail("user_not_found")
	}
	if err := security.VerifyPassword(params.Password, user.PasswordHash); err != nil {
		return fail("invalid_password")
	}
	if !user.IsActive() {
		return fail("account_not_active")
	}

	sessionID := uuid.New()
	token, expiresIn, err := s.jwtManager.GenerateAccessToken(
		user.ID.String(),
		sessionID.String(),
		companyIDString(user.CompanyID),
		user.Email,
		user.RoleName,
	)
	if err != nil {
		return LoginResult{}, err
	}

	session := entity.NewSession(
		s.clock,
		sessionID,
		user.ID,
		jwt.HashToken(token),
		params.IPAddress,
		params.Device,
		s.jwtManager.GetAccessTokenExpiry(),
	)
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return LoginResult{}, err
	}

	_ = s.loginAttemptRepo.Create(ctx, entity.NewLoginAttempt(
		s.clock, params.Email, user.CompanyID, params.IPAddress, true, "",
	))
	_ = s.userRepo.RecordAccess(ctx, user.ID)

	permissions, err := s.userRepo.PermissionsByRole(ctx, user.RoleID)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		User:        user,
		Permissions: permissions,
		AccessToken: token,
		ExpiresIn:   expiresIn,
		TokenType:   "Bearer",
	}, nil
}

// Logout ends the session identified by its token hash.
func (s *AuthService) Logout(ctx context.Context, tokenHash string) error {
	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return err
	}
	if session == nil {
		return nil
	}
	return s.sessionRepo.End(ctx, session.ID, s.clock.NowUTC(), entity.CloseReasonManual)
}

// Me returns the authenticated user together with the codes of the
// permissions granted to its role.
func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*userEntity.AppUser, []string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, nil
	}
	if !user.IsActive() {
		return user, []string{}, nil
	}

	permissions, err := s.userRepo.PermissionsByRole(ctx, user.RoleID)
	if err != nil {
		return nil, nil, err
	}
	return user, permissions, nil
}

func companyIDOf(user *userEntity.AppUser) *uuid.UUID {
	if user == nil {
		return nil
	}
	return user.CompanyID
}

func companyIDString(companyID *uuid.UUID) string {
	if companyID == nil {
		return ""
	}
	return companyID.String()
}
