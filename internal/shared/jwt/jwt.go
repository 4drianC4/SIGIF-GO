package jwt

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/shared/config"
)

const (
	AccessTokenType  = "access"
	RefreshTokenType = "refresh"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Claims struct {
	UserID    string   `json:"user_id"`
	TenantID  string   `json:"tenant_id"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	TokenType string   `json:"typ"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type JWTManager struct {
	secret             []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
	issuer             string
}

func NewManager(cfg *config.Config) *JWTManager {
	return &JWTManager{
		secret:             []byte(cfg.JWT.Secret),
		accessTokenExpiry:  time.Duration(cfg.JWT.AccessTokenExpiry) * time.Minute,
		refreshTokenExpiry: time.Duration(cfg.JWT.RefreshTokenExpiry) * time.Minute,
		issuer:             cfg.JWT.Issuer,
	}
}

func (m *JWTManager) GeneratePair(userID, tenantID, email string, roles []string) (*TokenPair, error) {
	now := time.Now()
	accessClaims := Claims{
		UserID:    userID,
		TenantID:  tenantID,
		Email:     email,
		Roles:     roles,
		TokenType: AccessTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			Audience:  []string{tenantID},
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        generateID(),
		},
	}

	refreshClaims := Claims{
		UserID:    userID,
		TenantID:  tenantID,
		Email:     email,
		Roles:     roles,
		TokenType: RefreshTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			Audience:  []string{tenantID},
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTokenExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        generateID(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(m.secret)
	if err != nil {
		return nil, err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(m.secret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int(m.accessTokenExpiry.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (m *JWTManager) Validate(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, "")
}

func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, AccessTokenType)
}

func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, RefreshTokenType)
}

func (m *JWTManager) validateToken(tokenString, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if expectedType != "" && claims.TokenType != expectedType {
		return nil, ErrInvalidToken
	}
	if claims.TokenType == "" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (m *JWTManager) Refresh(refreshTokenString string) (*TokenPair, error) {
	claims, err := m.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	return m.GeneratePair(claims.UserID, claims.TenantID, claims.Email, claims.Roles)
}

func generateID() string {
	return uuid.NewString()
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (m *JWTManager) GetRefreshTokenExpiry() time.Duration {
	return m.refreshTokenExpiry
}
