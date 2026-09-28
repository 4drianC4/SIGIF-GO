package service

import (
	"context"
	"github.com/google/uuid"
	authEntity "github.com/sigif/sigif-go/internal/modules/auth/domain/entity"
	authRepo "github.com/sigif/sigif-go/internal/modules/auth/domain/repository"
	userEntity "github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	userRepo "github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/errors"
	"github.com/sigif/sigif-go/internal/shared/jwt"
)

type AuthService struct {
	userRepo  userRepo.UserRepository
	tokenRepo authRepo.RefreshTokenRepository
	jwtManager *jwt.JWTManager
}

func NewAuthService(
	userRepo userRepo.UserRepository,
	tokenRepo authRepo.RefreshTokenRepository,
	jwtManager *jwt.JWTManager,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtManager: jwtManager,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password, userAgent, ipAddress string) (*jwt.TokenPair, *userEntity.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, errors.New(errors.CodeUnauthorized, "invalid credentials", 401)
	}

	if err := user.CheckPassword(password); err != nil {
		return nil, nil, errors.New(errors.CodeUnauthorized, "invalid credentials", 401)
	}

	if user.Status != userEntity.UserStatusActive {
		return nil, nil, errors.New(errors.CodeForbidden, "account is not active", 403)
	}

	tokenPair, err := s.jwtManager.GeneratePair(
		user.ID.String(),
		user.TenantID.String(),
		user.Email,
		userRolesToStrings(user.Roles),
	)
	if err != nil {
		return nil, nil, err
	}

	refreshToken := authEntity.NewRefreshToken(
		nil,
		user.ID,
		tokenPair.RefreshToken,
		userAgent,
		ipAddress,
		s.jwtManager.GetRefreshTokenExpiry(),
	)

	if err := s.tokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, nil, err
	}

	if err := s.userRepo.RecordLogin(ctx, user.ID); err != nil {
		return nil, nil, err
	}

	return tokenPair, user, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
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

	user, err := s.userRepo.GetByID(ctx, uuid.MustParse(claims.UserID))
	if err != nil {
		return nil, err
	}
	if user == nil || user.Status != userEntity.UserStatusActive {
		return nil, errors.New(errors.CodeUnauthorized, "user not found or inactive", 401)
	}

	if err := s.tokenRepo.Revoke(ctx, storedToken.ID); err != nil {
		return nil, err
	}

	newTokenPair, err := s.jwtManager.GeneratePair(
		user.ID.String(),
		user.TenantID.String(),
		user.Email,
		userRolesToStrings(user.Roles),
	)
	if err != nil {
		return nil, err
	}

	newRefreshToken := authEntity.NewRefreshToken(
		nil,
		user.ID,
		newTokenPair.RefreshToken,
		"",
		"",
		s.jwtManager.GetRefreshTokenExpiry(),
	)

	if err := s.tokenRepo.Create(ctx, newRefreshToken); err != nil {
		return nil, err
	}

	return newTokenPair, nil
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
		return s.tokenRepo.Revoke(ctx, storedToken.ID)
	}
	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.tokenRepo.RevokeAllByUserID(ctx, userID)
}

func (s *AuthService) ValidateToken(ctx context.Context, accessToken string) (*jwt.Claims, error) {
	return s.jwtManager.Validate(accessToken)
}

func userRolesToStrings(roles []userEntity.UserRole) []string {
	result := make([]string, len(roles))
	for i, r := range roles {
		result[i] = string(r)
	}
	return result
}