package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	authEntity "github.com/sigif/sigif-go/internal/modules/auth/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/auth/domain/repository"
	userEntity "github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/shared/clock"
	"github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/jwt"
	"github.com/sigif/sigif-go/internal/shared/security"
)

type passwordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}

type passwordHasherImpl struct{}

func (passwordHasherImpl) Hash(password string) (string, error) {
	return security.HashPassword(password)
}

func (passwordHasherImpl) Verify(password, hash string) error {
	return security.VerifyPassword(password, hash)
}

type AuthService struct {
	userRepo       repository.UserRepo
	tokenRepo      repository.RefreshTokenRepository
	jwtManager     *jwt.JWTManager
	passwordHasher passwordHasher
	clock          clock.Clock
}

func NewAuthService(
	userRepo repository.UserRepo,
	tokenRepo repository.RefreshTokenRepository,
	jwtManager *jwt.JWTManager,
	clock clock.Clock,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		jwtManager:     jwtManager,
		passwordHasher: passwordHasherImpl{},
		clock:          clock,
	}
}

func (s *AuthService) Login(ctx context.Context, params LoginParams) (*LoginResult, error) {
	user, err := s.userRepo.GetByEmail(ctx, params.TenantID, params.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New(errors.CodeUnauthorized, "invalid credentials", 401)
	}

	if err := s.passwordHasher.Verify(params.Password, user.PasswordHash); err != nil {
		return nil, errors.New(errors.CodeUnauthorized, "invalid credentials", 401)
	}

	if user.Status != userEntity.UserStatusActive {
		return nil, errors.New(errors.CodeForbidden, "account is not active", 403)
	}

	tokenPair, err := s.jwtManager.GeneratePair(
		user.ID.String(),
		user.TenantID.String(),
		user.Email,
		userRolesToStrings(user.Roles),
	)
	if err != nil {
		return nil, err
	}

	refreshTokenEntity := authEntity.NewRefreshToken(
		s.clock,
		user.ID,
		tokenPair.RefreshToken,
		params.UserAgent,
		params.IPAddress,
		s.jwtManager.GetRefreshTokenExpiry(),
	)

	if err := s.tokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, err
	}

	if err := s.userRepo.RecordLogin(ctx, user.ID); err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
		User:         user,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	claims, err := s.jwtManager.Validate(refreshToken)
	if err != nil {
		return nil, errors.New(errors.CodeUnauthorized, "invalid refresh token", 401)
	}

	storedToken, err := s.tokenRepo.GetByTokenHash(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	if storedToken == nil || storedToken.IsRevoked() || storedToken.IsExpired() {
		return nil, errors.New(errors.CodeUnauthorized, "refresh token revoked or expired", 401)
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil || user.Status != userEntity.UserStatusActive {
		return nil, errors.New(errors.CodeUnauthorized, "user not found or inactive", 401)
	}

	if err := s.tokenRepo.Revoke(ctx, storedToken.ID, time.Now()); err != nil {
		return nil, err
	}

	tokenPair, err := s.jwtManager.GeneratePair(
		user.ID.String(),
		user.TenantID.String(),
		user.Email,
		userRolesToStrings(user.Roles),
	)
	if err != nil {
		return nil, err
	}

	newTokenEntity := authEntity.NewRefreshToken(
		s.clock,
		user.ID,
		tokenPair.RefreshToken,
		"",
		"",
		s.jwtManager.GetRefreshTokenExpiry(),
	)

	if err := s.tokenRepo.Create(ctx, newTokenEntity); err != nil {
		return nil, err
	}

	return &RefreshResult{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	_, err := s.jwtManager.Validate(refreshToken)
	if err != nil {
		return nil
	}

	storedToken, err := s.tokenRepo.GetByTokenHash(ctx, refreshToken)
	if err != nil {
		return err
	}
	if storedToken != nil {
		return s.tokenRepo.Revoke(ctx, storedToken.ID, time.Now())
	}
	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.tokenRepo.RevokeAllByUserID(ctx, userID, time.Now())
}

func (s *AuthService) ValidateToken(accessToken string) (*jwt.Claims, error) {
	return s.jwtManager.Validate(accessToken)
}

func userRolesToStrings(roles []userEntity.UserRole) []string {
	result := make([]string, len(roles))
	for i, r := range roles {
		result[i] = string(r)
	}
	return result
}

type LoginParams struct {
	TenantID  uuid.UUID
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	TokenType    string
	User         *userEntity.User
}

func (r *LoginResult) TokenPair() *jwt.TokenPair {
	return &jwt.TokenPair{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		ExpiresIn:    r.ExpiresIn,
		TokenType:    r.TokenType,
	}
}

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	TokenType    string
}