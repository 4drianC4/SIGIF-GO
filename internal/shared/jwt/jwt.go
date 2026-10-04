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
	TokenTypeAccess = "access"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Claims struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	CompanyID string `json:"company_id,omitempty"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type JWTManager struct {
	secret            []byte
	accessTokenExpiry time.Duration
	issuer            string
}

func NewManager(cfg *config.Config) *JWTManager {
	return &JWTManager{
		secret:            []byte(cfg.JWT.Secret),
		accessTokenExpiry: time.Duration(cfg.JWT.AccessTokenExpiry) * time.Minute,
		issuer:            cfg.JWT.Issuer,
	}
}

// GenerateAccessToken issues an access token bound to a specific session.
func (m *JWTManager) GenerateAccessToken(userID, sessionID, companyID, email, role string) (string, int, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID,
		SessionID: sessionID,
		CompanyID: companyID,
		Email:     email,
		Role:      role,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{"sigif"},
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        generateID(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", 0, err
	}

	return signed, int(m.accessTokenExpiry.Seconds()), nil
}

func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
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
	if claims.TokenType != TokenTypeAccess {
		return nil, ErrInvalidToken
	}
	if claims.UserID == "" || claims.SessionID == "" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func generateID() string {
	return uuid.NewString()
}

// HashToken returns a deterministic SHA-256 hex digest of a token so it can be
// stored and looked up without keeping the raw token in the database.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (m *JWTManager) GetAccessTokenExpiry() time.Duration {
	return m.accessTokenExpiry
}
